package main

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/rs/zerolog"
	"github.com/youkale/echogy"
	"github.com/youkale/echogy/logger"
	"github.com/youkale/echogy/pprof"
	"os"
	"os/signal"
	"syscall"
)

var _conf = flag.String("c", "config.json", "config file, format json")

type Config struct {
	LogLevel    string `json:"logLevel"`
	EnablePProf bool   `json:"pprof"`
	HttpAddr    string `json:"httpAddr"`
	SSHAddr     string `json:"SSHAddr"`
	Domain      string `json:"domain"`
	PrivateKey  string `json:"privateKey"`
}

func logLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	}
	return zerolog.WarnLevel
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

	logger.SetLogLevel(logLevel(config.LogLevel))

	ctx, cancelFunc := context.WithCancel(context.Background())

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	if config.EnablePProf {
		go func() {
			pprof.Serve()
		}()
	}

	go func() {
		echogy.Serve(ctx, config.SSHAddr, config.HttpAddr, config.Domain, []byte(config.PrivateKey))
	}()
	<-c
	logger.Warn("echogy will be shutdown", map[string]interface{}{})
	cancelFunc()
}
