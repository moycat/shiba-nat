package main

import (
	"flag"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
)

var (
	port      int
	pprofPort int
	debug     bool
)

func parseConfig() {
	if env := os.Getenv("SHIBA_PORT"); len(env) > 0 {
		port, _ = strconv.Atoi(env)
	}
	if env := os.Getenv("SHIBA_PPROFPORT"); len(env) > 0 {
		pprofPort, _ = strconv.Atoi(env)
	}
	if env := os.Getenv("SHIBA_DEBUG"); len(env) > 0 {
		debug = true
	}
	if port <= 0 {
		port = 7628
	}
	flag.IntVar(&port, "port", port, "port to listen on")
	flag.IntVar(&pprofPort, "pprof-port", pprofPort, "port to listen on for pprof")
	flag.BoolVar(&debug, "debug", debug, "enable debug mode")
	flag.Parse()
	if port <= 0 || port >= 65535 {
		log.Errorf("invalid port [%d]", port)
		flag.PrintDefaults()
		os.Exit(2)
	}
}
