package main

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/youkale/webssh"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var sshKey = []byte(`-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAABFwAAAAdzc2gtcn
NhAAAAAwEAAQAAAQEA2IhRhR8Be4ZiiRcv6IdzP3/yAkrNHEihDRMkkg6oua6Hj8614Rlo
VUANRLz1vu5MDzBnDvmHVvrm1HWF5Dy1G7arAiXrHqgSm1e0zk/F5pC5B/QHbczQcUtD6a
s9HCE/rz7FpM/XcXKCxDZ9gNnZysyxSmYHkg0BdA5a2zj1lyQ2b4d9YmA4udqelhBGa8rE
B55zaIkScFp3mrRoglS7Xn2xHAbJJYzTQxvixMaE46PUM5kjKYGoPGm7pC7P2AQyCMcW/+
q3ZW4WvvoLn56/YQR72GCnUopvagKUslZ/VFkKcLG4y5ziheKu/a6PMELXtcNghyxoqmUH
zIIHNZKoWQAAA9BlRNbeZUTW3gAAAAdzc2gtcnNhAAABAQDYiFGFHwF7hmKJFy/oh3M/f/
ICSs0cSKENEySSDqi5roePzrXhGWhVQA1EvPW+7kwPMGcO+YdW+ubUdYXkPLUbtqsCJese
qBKbV7TOT8XmkLkH9AdtzNBxS0Ppqz0cIT+vPsWkz9dxcoLENn2A2dnKzLFKZgeSDQF0Dl
rbOPWXJDZvh31iYDi52p6WEEZrysQHnnNoiRJwWneatGiCVLtefbEcBskljNNDG+LExoTj
o9QzmSMpgag8abukLs/YBDIIxxb/6rdlbha++gufnr9hBHvYYKdSim9qApSyVn9UWQpwsb
jLnOKF4q79ro8wQte1w2CHLGiqZQfMggc1kqhZAAAAAwEAAQAAAQEAqTaGdkCPuQeA019S
aiYH01TaPB5WgcbkTMJr7tQT2N9iQuioS9u+I/jlJZWBeg7hU3Fg6Fvp/vgeEWQyGPW0Fo
8+vnQBdLilqc31ltDSd+cbIfL7JzxKnG7UCLRwEh6NlRa5/50I4Tg6prlqhJo6T/h8iAaJ
3gHZ4+cf63dsvQn31Tl/7YW8GUV0OuoFfoiP9qq6X0/E4KB0MUm8WBFhH0r2HznWVF7Ou1
lk6/ciZRo0oYreLyoCfeXsbDj8wAJ9GGXBwo4Sl+L9yeExQv3jn3EtZlBNx+vEDgREd84D
fQDKoj9Wi1/S9cAQocgUyGpSXLZdGdwK/7WYuftrWBzXXQAAAIEAmtbf/QAj6X3D95qKVa
vqP4smGBwuqzGdDYB2k4yJu4zxU+y1UPmuUAS0hH3GuHSnLlxiXGOz0Ethf7oxwmchOAhu
CnmF5JGcUIWZ094zKKp/w4ZeLz9DMvU/JfTaDstGuluyCInUOAKZhDVWmBgbuBDeqUZ7/5
o94exkTA3VMqkAAACBAPtsafa4ao3sLrAoFsmVPZAKJroFSXTSMu8CbOCAF3P+50yf8W19
/IB/xloQCr8x80qlupLgmNkaqscylhuCdLPBEo8fM4Q8WXv4CRrep5rtLa/21LAoxtruQU
XfHwm4cYLOa8uSZ6HZQacooklyEPtLLRePUl0owwxTQmI6qJWzAAAAgQDceVGszDD4wrJy
a0i6Nlu0gZ2RYRQT11dM7HHzYNaJtzD+wjFumRPLVPiHp41UPHS4nyRPPmLzNjKmPGUhMo
UEDS2W6rVid4EspmfiiDI/O6WBs0gih4D4OvPV72QXAbK8srk1e9YvDn7HWhSVCYg8n3zo
jOCajkLVQH6YKUZbwwAAABJzZWFuQHlvdWthbGUubG9jYWwBAgMEBQYH
-----END OPENSSH PRIVATE KEY-----`)

var _conf = flag.String("c", "config.json", "config file, format json")

type Config struct {
	HttpAddr   string `json:"httpAddr"`
	SSHAddr    string `json:"SSHAddr"`
	Domain     string `json:"domain"`
	PrivateKey string `json:"privateKey"`
}

func main() {

	flag.Parse()

	f, err := os.ReadFile(*_conf)

	if nil != err {
		panic(err)
	}
	config := &Config{}
	err = json.Unmarshal(f, config)
	if nil != err {
		panic(err)
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	go func() {
		webssh.Serve(ctx, config.SSHAddr, config.HttpAddr, config.Domain, []byte(config.PrivateKey))
	}()
	<-c
	log.Print("(SSH Short)link will be shutdown")
	cancelFunc()
}
