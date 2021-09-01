package edge

import (
	"github.com/caddyserver/certmagic"
	"github.com/gliderlabs/ssh"
	"github.com/libdns/cloudflare"
	gossh "golang.org/x/crypto/ssh"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Config struct {
	Addr        string `json:"addr"`
	Email       string `json:"email"`
	SSHAddr     string `json:"ssh_addr"`
	Domain      string `json:"domain"`
	CfToken     string `json:"cf_token"`
	IdleTimeout int    `json:"idle_timeout"`
	Key         string `json:"key"`
}

func storageDir() (string, error) {
	dir, err := os.UserHomeDir()
	if nil != err {
		return "", err
	}
	p := strings.Join([]string{dir, ".edge"}, string(os.PathSeparator))
	err = os.MkdirAll(p, 0700)
	return p, err
}

func Start(config *Config) {
	key, err := gossh.ParseRawPrivateKey([]byte(config.Key))
	if nil != err {
		panic(err)
	}
	signer, err := gossh.NewSignerFromKey(key)
	if nil != err {
		panic(err)
	}

	h := newHandle(config)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		wg.Done()
		listen, err := net.Listen("tcp", config.SSHAddr)
		if nil != err {
			panic(err)
		}
		log.Printf(`start ssh server on %s`, config.SSHAddr)
		server := newSSHServer(config, h, signer)
		log.Fatal(server.Serve(listen))
	}()
	wg.Wait()

	wg.Add(1)
	go func() {

		dir, err := storageDir()
		if nil != err {
			panic(err)
		}
		storage := &certmagic.FileStorage{Path: dir}

		certmagicConfig := certmagic.NewDefault()
		certmagicConfig.Storage = storage

		acmeManager := certmagic.NewACMEManager(certmagicConfig, certmagic.ACMEManager{
			Email: config.Email,
			Agreed: true,
			DNS01Solver: &certmagic.DNS01Solver{
				DNSProvider: &cloudflare.Provider{
					APIToken: config.CfToken,
				},
			},
		})

		certmagicConfig.Issuers = append(certmagicConfig.Issuers, acmeManager)

		tlsConfig := certmagicConfig.TLSConfig()

		go func() {
			wg.Done()
			if er := certmagicConfig.ManageSync([]string{"*." + config.Domain}); nil != er {
				panic(er)
			}
		}()

		svr := &http.Server{
			Addr:      config.Addr,
			TLSConfig: tlsConfig,
			Handler:   h,
		}
		log.Printf(`start http server on %s`, config.Addr)

		if err := svr.ListenAndServeTLS("", ""); nil != err {
			panic(err)
		}
	}()
	wg.Wait()
}

func newSSHServer(config *Config,hand ForwardHandler, hostkey ssh.Signer) *ssh.Server {

	return &ssh.Server{
		IdleTimeout: time.Duration(config.IdleTimeout) * time.Second,
		PtyCallback: func(ctx ssh.Context, pty ssh.Pty) bool {
			return false
		},
		LocalPortForwardingCallback: func(ctx ssh.Context, destinationHost string, destinationPort uint32) bool {
			return true
		},
		HostSigners: []ssh.Signer{hostkey},
		Handler:     hand.SessionHandler,
		ReversePortForwardingCallback: func(ctx ssh.Context, bindHost string, bindPort uint32) bool {
			return true
		},
		RequestHandlers: map[string]ssh.RequestHandler{
			"tcpip-forward":        hand.TCPIPForward,
			"cancel-tcpip-forward": hand.CancelTCPIPForward,
		},
	}
}
