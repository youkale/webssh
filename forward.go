package echogy

import (
	"bufio"
	"context"
	"fmt"
	"github.com/gliderlabs/ssh"
	"github.com/youkale/echogy/logger"
	"github.com/youkale/echogy/tui"
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
	accessId   string
	pty        *tui.Tui
	reqChan    chan *facadeRequest
	bindAddr   string
	bindPort   uint32
}

type facadeRequest struct {
	net.Conn
	request *http.Request
}

func newForwarder(accessId, domain string, session ssh.Session) (*forwarder, error) {
	pty, err := tui.NewPty(session, map[string]string{
		"access": fmt.Sprintf("%s.%s", accessId, domain),
	})
	if err != nil {
		return nil, err
	}
	ctx, cancelFunc := context.WithCancel(session.Context())
	return &forwarder{
		context:    ctx,
		cancelFunc: cancelFunc,
		accessId:   accessId,
		pty:        pty,
		sess:       session,
		reqChan:    make(chan *facadeRequest, 4),
	}, nil
}

type remoteForwardChannelData struct {
	DestAddr   string
	DestPort   uint32
	OriginAddr string
	OriginPort uint32
}

func (fwd *forwarder) forward(request *facadeRequest) {
	fwd.reqChan <- request
}

func (fwd *forwarder) serve() {
	fwdReq := fwd.sess.Context().Value(sshRequestForward).(*remoteForwardRequest)
	remoteAddr := fwd.sess.RemoteAddr().String()
	svrConn := fwd.sess.Context().Value(ssh.ContextKeyConn).(*gossh.ServerConn)

	logger.Info("created forward session", map[string]interface{}{
		"module":     "session",
		"accessId":   fwd.accessId,
		"remoteAddr": remoteAddr,
	})

	go func() {
		err := fwd.pty.Start()
		if err != nil {
			logger.Error("start pty session", err, map[string]interface{}{
				"module":     "session",
				"accessId":   fwd.accessId,
				"remoteAddr": remoteAddr,
			})
		}
	}()

	for {
		select {
		case <-fwd.context.Done():
			return
		case facadeReq := <-fwd.reqChan:
			facadeRequestAddr, facadeRequestPortStr, _ := net.SplitHostPort(facadeReq.RemoteAddr().String())
			facadePort, _ := strconv.Atoi(facadeRequestPortStr)
			payload := gossh.Marshal(&remoteForwardChannelData{
				DestAddr:   fwdReq.BindAddr,
				DestPort:   fwdReq.BindPort,
				OriginAddr: facadeRequestAddr,
				OriginPort: uint32(facadePort),
			})
			gosshChan, _, err := svrConn.OpenChannel("forwarded-tcpip", payload)
			sshChan := wrapChannelConn(fwd.sess, gosshChan)

			if err != nil {
				logger.Error("open forward channel", err, map[string]interface{}{
					"module":     "session",
					"accessId":   fwd.accessId,
					"remoteAddr": remoteAddr,
				})
				return
			}
			go fwd.doCopy(sshChan, facadeReq)
			io.Copy(sshChan, facadeReq)
		}
	}
}

func (fwd *forwarder) doCopy(sessConn *wrappedConn, facadeReq *facadeRequest) {
	defer func() {
		facadeReq.Close()
		sessConn.Close()
	}()
	if nil != facadeReq.request {
		reader := sessConn.bufferedReader()
		response, err := http.ReadResponse(bufio.NewReader(reader), facadeReq.request)
		if nil != err {
			return
		} else {
			fwd.pty.Notify(response, facadeReq.request)
			logger.Debug("ssh <-> facade", map[string]interface{}{
				"module":   "session",
				"accessId": fwd.accessId,
				"method":   facadeReq.request.Method,
				"path":     facadeReq.request.URL.Path,
				"status":   response.StatusCode,
			})
			io.Copy(facadeReq, reader.toBufferedConn(sessConn))
		}
	} else {
		io.Copy(facadeReq, sessConn)
	}
}
