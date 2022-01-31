package main

import (
	"flag"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
)

var (
	addr      string
	port      int
	pprofPort int
	debug     bool
)

func parseConfig() {
	if env := os.Getenv("SHIBA_ADDR"); len(env) > 0 {
		addr = env
	}
	if env := os.Getenv("SHIBA_PORT"); len(env) > 0 {
		port, _ = strconv.Atoi(env)
	}
	if env := os.Getenv("SHIBA_PPROFPORT"); len(env) > 0 {
		pprofPort, _ = strconv.Atoi(env)
	}
	if env := os.Getenv("SHIBA_DEBUG"); len(env) > 0 {
		debug = true
	}
	if len(addr) == 0 {
		addr = "shiba-nat-gateway"
	}
	if port <= 0 {
		port = 7628
	}
	flag.StringVar(&addr, "addr", addr, "gateway address")
	flag.IntVar(&port, "port", port, "gateway port")
	flag.IntVar(&pprofPort, "pprof-port", pprofPort, "port to listen on for pprof")
	flag.BoolVar(&debug, "debug", debug, "enable debug mode")
	flag.Parse()
	if len(addr) == 0 {
		log.Error("unspecified gateway address")
		flag.PrintDefaults()
		os.Exit(2)
	}
	if port <= 0 || port >= 65535 {
		log.Errorf("invalid port [%d]", port)
		flag.PrintDefaults()
		os.Exit(2)
	}
}
