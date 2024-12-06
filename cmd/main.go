package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/youkale/echogy"
	"github.com/youkale/echogy/logger"
	"github.com/youkale/echogy/pprof"
)

var _conf = flag.String("c", "config.json", "config file, format json")
var _pidFile = flag.String("pid", "", "pid file path (default: executable directory)")

type Config struct {
	LogLevel    string `json:"logLevel"`
	LogFile     string `json:"logFile"` // Path to log file
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

func setOSEnv() {
	os.Setenv("COLORTERM", "truecolor")
	os.Setenv("TERM", "xterm-256color")
	os.Setenv("CLICOLOR", "1")
	os.Setenv("CLICOLOR_FORCE", "1")
	os.Setenv("FORCE_COLOR", "1")
	os.Setenv("TERM_PROGRAM", "xterm")
}

func main() {

	flag.Parse()

	// Create PID file
	pidPath := *_pidFile
	if pidPath == "" {
		execPath, err := os.Executable()
		if err != nil {
			panic(err)
		}
		pidPath = execPath + ".pid"
	}

	setOSEnv()

	// Write PID to file
	pid := os.Getpid()
	if err := os.WriteFile(pidPath, []byte(fmt.Sprint(pid)), 0644); err != nil {
		panic(err)
	}
	defer os.Remove(pidPath)

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

	// Setup log file output if configured
	if config.LogFile != "" {
		if err := logger.AddFileOutput(config.LogFile); err != nil {
			panic(fmt.Sprintf("Failed to setup log file: %v", err))
		}
		logger.Info("Log file output enabled", logger.Fields{"path": config.LogFile})
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	c := make(chan os.Signal, 1)
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
