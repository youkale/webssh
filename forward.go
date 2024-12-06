package echogy

import (
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
	"time"
)

type forwarder struct {
	context    context.Context
	cancelFunc context.CancelFunc
	sess       ssh.Session
	accessId   string
	pty        *tui.Tui
	reqChan    chan net.Conn
	bindAddr   string
	bindPort   uint32
}

type facadeRequest struct {
	net.Conn
	request *http.Request
}

func newForwarder(accessId, domain string, session ssh.Session) (*forwarder, error) {
	pty, err := tui.NewPty(session, fmt.Sprintf("%s.%s", accessId, domain))
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
		reqChan:    make(chan net.Conn, 4),
	}, nil
}

type remoteForwardChannelData struct {
	DestAddr   string
	DestPort   uint32
	OriginAddr string
	OriginPort uint32
}

func (fwd *forwarder) forward(hijackConn *hijackConn) {
	hijackConn.SetDispatch(fwd.pty.Notify)
	fwd.reqChan <- hijackConn
}

func (fwd *forwarder) serve() {
	ctxReq := fwd.sess.Context().Value(sshRequestForward)
	if nil == ctxReq {
		return
	}
	fwdReq := ctxReq.(*remoteForwardRequest)
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
		case <-time.After(time.Second * 30):
			_, err := fwd.sess.SendRequest("keepalive@openssh.com", true, nil)
			if err != nil {
				logger.Warn("Failed to send keepalive request", map[string]interface{}{})
			}
		case facadeConn := <-fwd.reqChan:
			logger.Debug("open forward channel", map[string]interface{}{
				"module":     "session",
				"accessId":   fwd.accessId,
				"remoteAddr": remoteAddr,
			})
			facadeRequestAddr, facadeRequestPortStr, _ := net.SplitHostPort(facadeConn.RemoteAddr().String())
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
				if nil != facadeConn {
					facadeConn.Close()
				}
				return
			}
			go func() {
				defer func() {
					facadeConn.Close()
					sshChan.Close()
				}()
				io.Copy(facadeConn, sshChan)
			}()
			io.Copy(sshChan, facadeConn)
		}
	}
}
