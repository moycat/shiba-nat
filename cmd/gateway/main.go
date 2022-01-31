package main

import (
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"

	"github.com/moycat/shiba-nat/internal"
	log "github.com/sirupsen/logrus"
)

func main() {
	parseConfig()
	if debug {
		log.SetLevel(log.DebugLevel)
	}
	if pprofPort > 0 {
		go servePprof(pprofPort)
	}

	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv6unspecified,
		Port: port,
	})
	if err != nil {
		log.Fatalf("failed to listen on [%d]: %v", port, err)
	}
	log.Infof("listening on [%d]", port)

	for {
		buf := make([]byte, 256)
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Fatalf("failed to read from socket: %v", err)
		}
		log.Debugf("received a packet from [%s]", addr.String())
		query := new(internal.Query)
		if err := internal.Unmarshal(buf[:n], query); err != nil {
			log.Warningf("failed to decode the query message: %v", err)
			continue
		}
		if query.Magic != internal.QueryMagic {
			log.Warningf("query magic [%s] mismatches", query.Magic)
			continue
		}
		reply := &internal.QueryReply{
			Magic:    internal.QueryMagic,
			Token:    query.Token,
			ClientIP: addr.IP,
			Port:     port,
		}
		b, err := internal.Marshal(reply)
		if err != nil {
			log.Warningf("failed to encode the reply message: %v", err)
			continue
		}
		// Don't use the read source port as they may be SNAT-ed.
		addr.Port = query.Port
		// Don't reuse the connection or we get SNAT-ed.
		conn, err := net.DialUDP("udp", nil, addr)
		if err != nil {
			log.Warningf("failed to dial [%s]: %v", addr.String(), err)
			continue
		}
		_, err = conn.Write(b)
		if err != nil {
			_ = conn.Close()
			log.Warningf("failed to reply the query: %v", err)
			continue
		}
		_ = conn.Close()
		log.Debug("sent a reply")
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
