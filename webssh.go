package webssh

import (
	"context"
	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
	"log"
	"sync"
	"time"
)

func newSshServer(sshAddr string, sshKey []byte, fwd *forwarder) *ssh.Server {
	key, _ := gossh.ParseRawPrivateKey(sshKey)
	signer, _ := gossh.NewSignerFromKey(key)
	return &ssh.Server{
		IdleTimeout: 300 * time.Second,
		HostSigners: []ssh.Signer{signer},
		Addr:        sshAddr,
		PtyCallback: func(ctx ssh.Context, pty ssh.Pty) bool {
			return true
		},
		Handler: fwd.sessionHandle,
		ReversePortForwardingCallback: func(ctx ssh.Context, bindHost string, bindPort uint32) bool {
			return true
		},
		RequestHandlers: map[string]ssh.RequestHandler{
			sshRequestTypeForward:       fwd.handleRequest,
			sshRequestTypeCancelForward: fwd.handleRequest,
		},
	}
}

func Serve(_ctx context.Context, sshAddr, inboundAddr, domain string, sshKey []byte) {

	wg := sync.WaitGroup{}

	ctx, cancelFunc := context.WithCancel(_ctx)

	fwd, err := newForwarder(ctx, inboundAddr, domain)
	if nil != err {
		log.Printf("starting ssh server on port %s", sshAddr)
	}

	in, err := newInbound(ctx, inboundAddr, fwd.forward)

	wg.Add(1)
	go func() {
		wg.Done()
		log.Printf("starting inbound server on port %s", inboundAddr)
		in.startServe()
	}()
	wg.Wait()

	server := newSshServer(sshAddr, sshKey, fwd)

	wg.Add(1)
	go func() {
		wg.Done()
		log.Printf("starting ssh server on port %s", sshAddr)
		log.Fatal(server.ListenAndServe())
	}()
	wg.Wait()

	<-_ctx.Done()
	server.Shutdown(ctx)
	log.Println("server will be shutdown !!!")
	cancelFunc()
}
