package main

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/rs/zerolog"
	"github.com/youkale/webssh"
	"github.com/youkale/webssh/logger"
	"os"
	"os/signal"
	"syscall"
)

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

	logger.SetLogLevel(zerolog.DebugLevel)

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
	logger.Warn("webssh will be shutdown", map[string]interface{}{})
	cancelFunc()
}
