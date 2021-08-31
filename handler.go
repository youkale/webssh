package edge

import (
	"context"
	"fmt"
	"github.com/gliderlabs/ssh"
	qr "github.com/mdp/qrterminal/v3"
	gossh "golang.org/x/crypto/ssh"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
)


type remoteForwardRequest struct {
	BindAddr string
	BindPort uint32
}

type remoteForwardSuccess struct {
	BindPort uint32
}

type ForwardHandler interface {
	TCPIPForward(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte)
	CancelTCPIPForward(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte)
	SessionHandler(sess ssh.Session)
}

func clientDial(sessionId,bindAddr,remoteAddr string) (*http.Client, error) {
	conn, err := createConnect(sessionId, bindAddr, remoteAddr)
	if nil != err {
		return nil, err
	}
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return conn, nil
			},
		},
	}, nil
}

type handle struct {
	config *Config
	listenAddr string
	port int
}


func newHandle(conf *Config) *handle {
	_, port, _ := net.SplitHostPort(conf.Addr)
	p, _ := strconv.Atoi(port)
	return &handle{
		port: p,
		config: conf,
		listenAddr: fmt.Sprintf(`localhost:%s`, port),
	}
}

func (f *handle) accessAddr(alias string) string {
	sb := strings.Builder{}
	sb.WriteString("https://")
	sb.WriteString(alias)
	sb.WriteString(".")
	sb.WriteString(f.config.Domain)
	return sb.String()
}

func (f *handle) SessionHandler(sess ssh.Session) {
	id := sess.Context().Value(ssh.ContextKeySessionID).(string)
	log.Printf(`ssh session %s is open`, id)
	alias := getSessionAlias(id)
	qrconf := qr.Config{
		Level: qr.L,
		Writer: sess,
		BlackChar: qr.WHITE,
		WhiteChar: qr.BLACK,
		QuietZone: 1,
	}
	for _, a := range alias {
		p := f.accessAddr(a)
		qr.GenerateWithConfig(p,qrconf)
		io.WriteString(sess, p)
		io.WriteString(sess, "\n")
	}
	<-sess.Context().Done()
	log.Printf(`ssh session %s is closed`, id)
}


func (f *handle) TCPIPForward(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
	conn := ctx.Value(ssh.ContextKeyConn).(*gossh.ServerConn)
	var reqPayload remoteForwardRequest
	if err := gossh.Unmarshal(req.Payload, &reqPayload); err != nil {
		// TODO: log parse failure
		return false, []byte{}
	}
	var destPort int
	if reqPayload.BindPort == 443 {
		destPort = 443
	} else {
		destPort = f.port
	}
	saveSession(ctx.SessionID(),conn)
	return true, gossh.Marshal(&remoteForwardSuccess{uint32(destPort)})
}

func (f *handle) CancelTCPIPForward(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
	deleteSession(ctx.SessionID())
	return true,nil
}

func (f *handle) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	h := request.Host
	n := strings.Split(h, ".")
	if len(n) == 0 {
		http.Error(writer,fmt.Sprintf("request host [%s] incorrect",h),http.StatusBadRequest)
		return
	}
	client, err := clientDial(n[0],f.listenAddr,request.RemoteAddr)
	if nil != err {
		http.Error(writer,"not found session",http.StatusNotFound)
		return
	}

	u := request.URL.String()
	req, e := http.NewRequest(request.Method, u, request.Body)
	if nil != e {
		http.Error(writer,"build handle request error",http.StatusBadRequest)
		return
	}
	req.Header = request.Header.Clone()
	if req.URL.Scheme == "" {
		req.URL.Scheme = "http"
	}
	if req.URL.Host == ""{
		req.URL.Host = request.Host
	}
	query := request.URL.Query()
	for key, vals := range query {
		for _, val := range vals {
			req.URL.Query().Add(key,val)
		}
	}
	resp, err := client.Do(req)
	if nil != err {
		http.Error(writer,fmt.Sprintf("handle origin request error, %s",err.Error()),http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	header := resp.Header
	for k, v := range header {
		for _, vv := range v {
			writer.Header().Add(k, vv)
		}
	}
	writer.WriteHeader(resp.StatusCode)
	io.Copy(writer, resp.Body)
}
