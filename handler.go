package webssh

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/gliderlabs/ssh"
	"github.com/karlseguin/ccache/v3"
	"github.com/teris-io/shortid"
	gossh "golang.org/x/crypto/ssh"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	sshRequestTypeForward       = "tcpip-forward"
	sshRequestTypeCancelForward = "cancel-tcpip-forward"
	sshForwardChannelName       = "forwarded-tcpip"

	sshContextKeyAlias = "sshSessionHubKey"

	sshContextKeyRequestHandler = "req-handler-name"

	NotFound = `HTTP/1.0 404 Not Found
Server: %s
Content-Length: %d

Tunnel %s not found
`

	BadRequest = `HTTP/1.0 400 Bad Request
Server: %s
Content-Length: 12

Bad Request
`
)

type requestForward func(*request)

type forward func(net.Conn)

type remoteForwardSuccess struct {
	BindPort uint32
}

type remoteForwardCancelRequest struct {
	BindAddr string
	BindPort uint32
}

type remoteForwardChannelData struct {
	DestAddr   string
	DestPort   uint32
	OriginAddr string
	OriginPort uint32
}

type request struct {
	requestId string
	conn      net.Conn
	method    string
	path      string
}

type repeatRead struct {
	r io.Reader
	net.Conn
}

func (re *repeatRead) Read(p []byte) (n int, err error) {
	return re.r.Read(p)
}

var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 32*1024)
	},
}

type forwarder struct {
	ctx      context.Context
	sessHub  sync.Map
	aliasHub sync.Map
	bindAddr string
	bindPort uint32
	domain   string
}

func newForwarder(ctx context.Context, bind, domain string) (*forwarder, error) {
	host, p, err := net.SplitHostPort(bind)
	if err != nil {
		return nil, fmt.Errorf("invalid bind address: %w", err)
	}
	bindPort, err := strconv.Atoi(p)
	if err != nil {
		return nil, fmt.Errorf("invalid port number: %w", err)
	}
	return &forwarder{
		ctx:      ctx,
		bindAddr: host,
		bindPort: uint32(bindPort),
		domain:   domain,
	}, nil
}

func (f *forwarder) forward(c net.Conn) {

	buf := bufferPool.Get().([]byte)
	defer bufferPool.Put(buf)

	n, err := c.Read(buf)
	if err != nil {
		log.Printf("read error: %v", err)
		c.Close()
		return
	}

	newReader := bytes.NewReader(buf[:n])
	reader := bufio.NewReader(bytes.NewReader(buf[:n]))
	req, err := http.ReadRequest(reader)
	if err != nil {
		log.Printf("parse request error: %v", err)
		c.Close()
		return
	}

	if _, err = newReader.Seek(0, io.SeekStart); err != nil {
		log.Printf("seek error: %v", err)
		c.Close()
		return
	}

	domainSep := strings.Split(req.Host, ".")
	if len(domainSep) <= 1 {
		c.Close()
		return
	}

	reqId := domainSep[0]
	fwd, err := f.findSshRequestHandle(reqId)
	if err != nil {
		c.Write([]byte(fmt.Sprintf(NotFound, f.domain, len(reqId)+18, reqId)))
		c.Close()
		return
	}

	fwd(&request{
		method:    req.Method,
		path:      req.URL.Path,
		requestId: reqId,
		conn: &repeatRead{
			r:    io.MultiReader(newReader, c),
			Conn: c,
		},
	})
}

func (f *forwarder) findSshRequestHandle(reqId string) (requestForward, error) {
	if sid, ok := f.aliasHub.Load(reqId); ok {
		if sess, ok := f.sessHub.Load(sid); ok {
			return sess.(ssh.Session).Context().Value(sshContextKeyRequestHandler).(func(req *request)), nil
		}
	}
	return nil, fmt.Errorf("reqId [%s] is not found", reqId)
}

var cache = ccache.New(ccache.Configure[string]().MaxSize(1_000_000))

func (f *forwarder) generateRequestId(raddr net.Addr, regenerate bool) (string, error) {
	switch raddr.(type) {
	case *net.TCPAddr:
		addr := raddr.(*net.TCPAddr)
		cacheKey := addr.IP.String()
		if regenerate {
			cache.Delete(cacheKey)
		}
		fetch, err := cache.Fetch(cacheKey, time.Hour*12, func() (string, error) {
			return shortid.Generate()
		})
		if nil != fetch && nil == err {
			return fetch.Value(), nil
		} else {
			return shortid.Generate()
		}
	default:
		return shortid.Generate()
	}
}

type channelOpenMsg struct {
	ChanType         string `sshtype:"90"`
	PeersID          uint32
	PeersWindow      uint32
	MaxPacketSize    uint32
	TypeSpecificData []byte `ssh:"rest"`
}

func (f *forwarder) sessionRequestServe(sess ssh.Session, alias string) {
	sessChan := make(chan *request, 4) // Increased buffer size

	ctx := sess.Context()
	ctx.SetValue(sshContextKeyAlias, alias)
	ctx.SetValue(sshContextKeyRequestHandler, func(r *request) {
		select {
		case sessChan <- r:
		case <-time.After(5 * time.Second):
			log.Printf("request channel full, dropping request for %s", alias)
		}
	})

	_, _, isPty := sess.Pty()

	if isPty {
		fwd := fmt.Sprintf("https://%s", strings.Join([]string{alias, f.domain}, "."))
		sess.Write([]byte(fwd))
	}

	for {
		select {
		case req := <-sessChan:
			c := req.conn
			originAddr, originPortStr, _ := net.SplitHostPort(c.RemoteAddr().String())
			originPort, _ := strconv.Atoi(originPortStr)

			payload := gossh.Marshal(&remoteForwardChannelData{
				DestAddr:   f.bindAddr,
				DestPort:   f.bindPort,
				OriginAddr: originAddr,
				OriginPort: uint32(originPort),
			})

			svrConn := sess.Context().Value(ssh.ContextKeyConn).(*gossh.ServerConn)

			sshChan, _, err := svrConn.OpenChannel(sshForwardChannelName, payload)

			if err != nil {
				c.Write([]byte(fmt.Sprintf(BadRequest, f.domain)))
				c.Close()
			} else {
				go func() {
					//gossh.DiscardRequests(sshReqs)
					msg := channelOpenMsg{TypeSpecificData: []byte("hello")}
					sess.Write(gossh.Marshal(msg))
				}()

				go func() {
					defer func() {
						sshChan.Close()
						c.Close()
					}()
					io.Copy(sshChan, c)
				}()
				io.Copy(c, sshChan)
			}
		case <-sess.Context().Done():
			return
		}
	}
}

func (f *forwarder) sessionHandle(sess ssh.Session) {

	sessId := sess.Context().Value(ssh.ContextKeySessionID).(string)

	raddr := sess.RemoteAddr()

	alias, err := f.generateRequestId(raddr, false)

	if nil != err {
		return
	}

	log.Printf("create reqId [%s] -> sessId [%s...], addr [%s]",
		alias, sessId[0:8], raddr.String())

	f.aliasHub.Store(alias, sessId)
	f.sessHub.Store(sessId, sess)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		f.sessionRequestServe(sess, alias)
		wg.Done()
		sess.Close()
		f.sessHub.Delete(sessId)
		f.aliasHub.Delete(alias)
	}()
	wg.Wait()

	log.Printf("close  reqId [%s] -> sessId [%s...], addr [%s]",
		alias, sessId[0:8], raddr.String())
}

func (f *forwarder) handleRequest(ctx ssh.Context, _ *ssh.Server, req *gossh.Request) (bool, []byte) {
	switch req.Type {
	case sshRequestTypeForward:
		return true, gossh.Marshal(&remoteForwardSuccess{f.bindPort})

	case sshRequestTypeCancelForward:
		var reqPayload remoteForwardCancelRequest
		if err := gossh.Unmarshal(req.Payload, &reqPayload); err != nil {
			// TODO: log parse failure
			return false, []byte{}
		}
		id := ctx.Value(ssh.ContextKeySessionID).(string)
		if sshConn, ok := f.sessHub.Load(id); ok {
			sshConn.(ssh.Session).Close()
			alias := ctx.Value(sshContextKeyAlias).(string)
			f.aliasHub.Delete(alias)
			f.sessHub.Delete(id)
		}
		return true, nil
	default:
		return false, nil
	}
}
