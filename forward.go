package webssh

import (
	"context"
	"github.com/gliderlabs/ssh"
	"github.com/youkale/webssh/logger"
	"github.com/youkale/webssh/tui"
	gossh "golang.org/x/crypto/ssh"
	"io"
	"net"
	"strconv"
)

type forwarder struct {
	context    context.Context
	cancelFunc context.CancelFunc
	sess       ssh.Session
	reqChan    chan net.Conn
	bindAddr   string
	bindPort   uint32
}

func newForwarder(session ssh.Session, host string, port uint32) (*forwarder, error) {
	ctx, cancelFunc := context.WithCancel(session.Context())
	return &forwarder{
		context:    ctx,
		cancelFunc: cancelFunc,
		sess:       session,
		bindAddr:   host,
		bindPort:   port,
		reqChan:    make(chan net.Conn, 4),
	}, nil
}

type remoteForwardChannelData struct {
	DestAddr   string
	DestPort   uint32
	OriginAddr string
	OriginPort uint32
}

func (ch *forwarder) forward(conn net.Conn) {
	ch.reqChan <- conn
}

func (ch *forwarder) serve() {

	accessId := ch.sess.Context().Value(sshAccessIdKey).(string)
	remoteAddr := ch.sess.RemoteAddr().String()
	svrConn := ch.sess.Context().Value(ssh.ContextKeyConn).(*gossh.ServerConn)

	logger.Info("start forward session", map[string]interface{}{
		"module":     "session",
		"accessId":   accessId,
		"remoteAddr": remoteAddr,
	})

	go tui.NewPty(ch.sess)

	for {
		select {
		case <-ch.context.Done():
			return
		case conn := <-ch.reqChan:
			originAddr, originPortStr, _ := net.SplitHostPort(conn.RemoteAddr().String())
			originPort, _ := strconv.Atoi(originPortStr)
			payload := gossh.Marshal(&remoteForwardChannelData{
				DestAddr:   ch.bindAddr,
				DestPort:   ch.bindPort,
				OriginAddr: originAddr,
				OriginPort: uint32(originPort),
			})
			sshChan, _, err := svrConn.OpenChannel("forwarded-tcpip", payload)
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
					conn.Close()
					sshChan.Close()
				}()
				io.Copy(sshChan, conn)
			}()
			io.Copy(conn, sshChan)
		}
	}
}
