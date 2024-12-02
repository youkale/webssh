package webssh

import (
	"bufio"
	"context"
	"github.com/gliderlabs/ssh"
	"github.com/youkale/webssh/logger"
	"github.com/youkale/webssh/tui"
	gossh "golang.org/x/crypto/ssh"
	"io"
	"net"
	"net/http"
	"strconv"
)

type forwarder struct {
	context    context.Context
	cancelFunc context.CancelFunc
	sess       ssh.Session
	reqChan    chan *facadeRequest
	bindAddr   string
	bindPort   uint32
}

type facadeRequest struct {
	net.Conn
	request *http.Request
}

func newForwarder(session ssh.Session, host string, port uint32) (*forwarder, error) {
	ctx, cancelFunc := context.WithCancel(session.Context())
	return &forwarder{
		context:    ctx,
		cancelFunc: cancelFunc,
		sess:       session,
		bindAddr:   host,
		bindPort:   port,
		reqChan:    make(chan *facadeRequest, 4),
	}, nil
}

type remoteForwardChannelData struct {
	DestAddr   string
	DestPort   uint32
	OriginAddr string
	OriginPort uint32
}

func (ch *forwarder) forward(request *facadeRequest) {
	ch.reqChan <- request
}

func (ch *forwarder) serve() {
	accessId := ch.sess.Context().Value(sshAccessIdKey).(string)
	remoteAddr := ch.sess.RemoteAddr().String()
	svrConn := ch.sess.Context().Value(ssh.ContextKeyConn).(*gossh.ServerConn)

	logger.Info("created forward session", map[string]interface{}{
		"module":     "session",
		"accessId":   accessId,
		"remoteAddr": remoteAddr,
	})

	pty, _ := tui.NewPty(ch.sess)

	go func() {
		err := pty.Start()
		if err != nil {
			logger.Error("start pty session", err, map[string]interface{}{
				"module":     "session",
				"accessId":   accessId,
				"remoteAddr": remoteAddr,
			})
		}
	}()

	for {
		select {
		case <-ch.context.Done():
			return
		case req := <-ch.reqChan:
			originAddr, originPortStr, _ := net.SplitHostPort(req.RemoteAddr().String())
			originPort, _ := strconv.Atoi(originPortStr)
			payload := gossh.Marshal(&remoteForwardChannelData{
				DestAddr:   ch.bindAddr,
				DestPort:   ch.bindPort,
				OriginAddr: originAddr,
				OriginPort: uint32(originPort),
			})
			gosshChan, _, err := svrConn.OpenChannel("forwarded-tcpip", payload)
			sshChan := wrapChannelConn(ch.sess, gosshChan)

			if err != nil {
				logger.Error("open forward channel", err, map[string]interface{}{
					"module":     "session",
					"sessionId":  accessId,
					"remoteAddr": remoteAddr,
				})
				return
			}
			go func() {
				defer func() {
					req.Close()
					sshChan.Close()
				}()
				io.Copy(sshChan, req)
			}()

			reader := sshChan.bufferedReader()
			response, err := http.ReadResponse(bufio.NewReader(reader), req.request)

			logger.Debug("ssh <-> facade", map[string]interface{}{
				"module":   "session",
				"accessId": accessId,
				"method":   req.request.Method,
				"path":     req.request.URL.Path,
				"status":   response.StatusCode,
			})
			io.Copy(req, reader.toBufferedConn(sshChan))
		}
	}
}
