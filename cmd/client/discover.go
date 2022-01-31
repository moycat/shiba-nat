package main

import (
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/moycat/shiba-nat/internal"
	log "github.com/sirupsen/logrus"
)

const (
	discoverPacketNumber = 3
	discoverWaitTime     = time.Second
)

func discover(addr string, port int) (net.IP, int, error) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv6unspecified,
		Port: 0,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to listen for replies: %w", err)
	}
	defer func() { _ = conn.Close() }()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	log.Debugf("listening on [%s] at [%d] for replies", localAddr.IP.String(), localAddr.Port)

	gatewayIP := net.ParseIP(addr)
	switch {
	case gatewayIP.To4() != nil, gatewayIP.To16() != nil:
	default:
		ips, err := net.LookupIP(addr)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to resolve address [%s]: %w", addr, err)
		}
		if len(ips) == 0 {
			return nil, 0, fmt.Errorf("address [%s] has no resolved ips", addr)
		}
		gatewayIP = ips[rand.Intn(len(ips))]
	}
	gatewayAddr := &net.UDPAddr{
		IP:   gatewayIP,
		Port: port,
	}
	log.Debugf("sending queries to [%s]", gatewayAddr.String())

	query := generateQuery(localAddr.Port)
	b, err := internal.Marshal(query)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to encode query: %w", err)
	}
	for i := 0; i < discoverPacketNumber; i++ {
		_, err := conn.WriteToUDP(b, gatewayAddr)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to send query: %w", err)
		}
	}

	if err := conn.SetReadDeadline(time.Now().Add(discoverWaitTime)); err != nil {
		return nil, 0, fmt.Errorf("failed to set read deadline for replies: %w", err)
	}
	buf := make([]byte, 256)
	n, realAddr, err := conn.ReadFromUDP(buf)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read response: %w", err)
	}
	reply := new(internal.QueryReply)
	if err := internal.Unmarshal(buf[:n], reply); err != nil {
		return nil, 0, fmt.Errorf("failed to decode reply: %w", err)
	}
	if reply.Magic != internal.QueryMagic {
		return nil, 0, fmt.Errorf("reply magic [%s] mismatches", query.Magic)
	}
	if reply.Token != query.Token {
		return nil, 0, fmt.Errorf("reply token [%s] mismatches [%s]", reply.Token, query.Token)
	}
	log.Debugf("received a valid reply from [%s]", realAddr.String())
	return realAddr.IP, reply.Port, nil
}
