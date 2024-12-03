package webssh

import (
	"context"
	"fmt"
	"github.com/gliderlabs/ssh"
	"github.com/youkale/webssh/logger"
	gossh "golang.org/x/crypto/ssh"
	"sync"
)

var sessionHub sync.Map

func init() {
	sessionHub = sync.Map{}
}

const (
	sshRequestTypeForward       = "tcpip-forward"
	sshRequestTypeCancelForward = "cancel-tcpip-forward"
	sshAccessIdKey              = "sshAccessId"
	sshRequestForward           = "sshRequestForward"
)

type remoteForwardSuccess struct {
	BindPort uint32
}

type remoteForwardCancelRequest struct {
	BindAddr string
	BindPort uint32
}

type remoteForwardRequest struct {
	BindAddr string
	BindPort uint32
}

func requestHandler(bindPort uint32) func(ctx ssh.Context, _ *ssh.Server, req *gossh.Request) (bool, []byte) {
	return func(ctx ssh.Context, _ *ssh.Server, req *gossh.Request) (bool, []byte) {
		switch req.Type {
		case sshRequestTypeForward:
			var reqPayload remoteForwardRequest
			if err := gossh.Unmarshal(req.Payload, &reqPayload); err != nil {
				logger.Error("Unmarshal failed", err, map[string]interface{}{
					"module":  "serve",
					"payload": reqPayload,
				})
				return false, []byte{}
			}
			logger.Debug("Unmarshal forward request", map[string]interface{}{
				"module":  "serve",
				"payload": fmt.Sprintf("%v", reqPayload),
			})
			ctx.SetValue(sshRequestForward, &reqPayload)
			return true, gossh.Marshal(&remoteForwardSuccess{bindPort})

		case sshRequestTypeCancelForward:
			var reqPayload remoteForwardCancelRequest
			if err := gossh.Unmarshal(req.Payload, &reqPayload); err != nil {
				logger.Error("Unmarshal failed", err, map[string]interface{}{
					"module": "serve",
				})
				return false, []byte{}
			}
			id := ctx.Value(sshAccessIdKey).(string)
			sessionHub.Delete(id)
			return true, nil
		default:
			return false, nil
		}
	}
}

func newSshServer(sshAddr string, sshKey []byte, bindPort uint32) *ssh.Server {
	key, _ := gossh.ParseRawPrivateKey(sshKey)
	signer, _ := gossh.NewSignerFromKey(key)

	reqFunc := requestHandler(bindPort)

	return &ssh.Server{
		//IdleTimeout: 300 * time.Second,
		HostSigners: []ssh.Signer{signer},
		Addr:        sshAddr,
		PtyCallback: func(ctx ssh.Context, pty ssh.Pty) bool {
			return true
		},
		ReversePortForwardingCallback: func(ctx ssh.Context, bindHost string, bindPort uint32) bool {
			return true
		},
		RequestHandlers: map[string]ssh.RequestHandler{
			sshRequestTypeForward:       reqFunc,
			sshRequestTypeCancelForward: reqFunc,
		},
	}
}

func Serve(_ctx context.Context, sshAddr, facadeAddr, domain string, sshKey []byte) {

	wg := sync.WaitGroup{}

	_, sshPort, err := parseHostAddr(sshAddr)
	if err != nil {
		logger.Error("parse net.Addr failed", err, map[string]interface{}{
			"module": "serve",
		})
		return
	}

	ctx, cancelFunc := context.WithCancel(_ctx)

	server := newSshServer(sshAddr, sshKey, sshPort)

	server.Handler = func(session ssh.Session) {
		channel, err := newForwarder(session)
		if nil != err {
			logger.Error("create forward", err, map[string]interface{}{
				"module":     "serve",
				"remoteAddr": session.RemoteAddr().String(),
			})
			return
		}

	regenerating:

		id, err := generateAccessId(session.RemoteAddr())

		if nil != err {
			logger.Error("generating sessionId", err, map[string]interface{}{
				"module": "serve",
			})
			session.Write([]byte("generating request id error"))
			return
		}

		if _, found := sessionHub.Load(id); found {
			goto regenerating
		}
		sessionHub.Store(id, channel)
		session.Context().SetValue(sshAccessIdKey, id)
		logger.Debug("establishing ssh session", map[string]interface{}{
			"module":   "session",
			"accessId": id,
		})
		channel.serve() // blocked with loop
		sessionHub.Delete(id)
		logger.Debug("clean ssh session", map[string]interface{}{
			"module":   "session",
			"accessId": id,
		})
		session.Close()
	}
	wg.Add(1)
	go func() {
		wg.Done()
		facadeServe(ctx, facadeAddr, func(facadeId string, req *facadeRequest) bool {
			if value, found := sessionHub.Load(facadeId); found {
				channel := value.(*forwarder)
				channel.forward(req)
				return true
			}
			return false
		})
	}()
	logger.Info("started facade server", map[string]interface{}{
		"module":  "serve",
		"address": facadeAddr,
	})
	wg.Wait()

	wg.Add(1)
	go func() {
		wg.Done()
		logger.Info("starting ssh server", map[string]interface{}{
			"module":  "serve",
			"address": sshAddr,
		})
		err := server.ListenAndServe()
		logger.Error("starting ssh server", err, map[string]interface{}{
			"module":  "serve",
			"address": sshAddr,
		})
	}()
	wg.Wait()

	<-_ctx.Done()
	server.Shutdown(ctx)
	logger.Warn("starting ssh server", map[string]interface{}{
		"module": "serve",
	})
	cancelFunc()
}
