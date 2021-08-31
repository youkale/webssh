package main

import (
	"encoding/json"
	"flag"
	"github.com/4chain/edge"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
)

var _conf = flag.String("c", "config.json", "config file, format json")

func main() {
	flag.Parse()

	f, err := ioutil.ReadFile(*_conf)

	if nil != err {
		panic(err)
	}
	config := &edge.Config{}
	err = json.Unmarshal(f, config)
	if nil != err {
		panic(err)
	}

	c := make(chan os.Signal, 0)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	edge.Start(config)

	<- c
}
