package pprof

import (
	"github.com/youkale/echogy/logger"
	"net/http"
	_ "net/http/pprof"
)

const addr = "localhost:9191"

var fields = map[string]interface{}{
	"module": "pprof",
	"listen": addr,
}

func Serve() {
	logger.Warn("starting pprof server", fields)
	if err := http.ListenAndServe(addr, nil); err != nil {
		logger.Error("pprof server failed", err, fields)
	}
}
