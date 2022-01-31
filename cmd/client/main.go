package main

import (
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

const clientInterval = 2 * time.Second

func main() {
	parseConfig()
	if debug {
		log.SetLevel(log.DebugLevel)
	}
	if pprofPort > 0 {
		go servePprof(pprofPort)
	}

	for {
		time.Sleep(clientInterval)
		log.Infof("discovering server at [%s] on [%d]", addr, port)
		realIP, realPort, err := discover(addr, port)
		if err != nil {
			log.Errorf("failed to discover real address of gateway: %v", err)
			continue
		}
		err = nat(realIP, realPort)
		if err != nil {
			log.Errorf("failed to do nat: %v", err)
			continue
		}
	}
}

func servePprof(port int) {
	if port <= 0 {
		return
	}
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Errorf("pprof exited: %v", err)
	}
}
