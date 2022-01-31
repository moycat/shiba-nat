package main

import (
	"fmt"
	"net"
	"time"

	"github.com/moycat/shiba-nat/internal"
	log "github.com/sirupsen/logrus"
)

const (
	applyInterval     = 15 * time.Second
	heartbeatInterval = 3 * time.Second
	heartbeatTimeout  = time.Second
	heartbeatMaxError = 3
)

func nat(ip net.IP, port int) error {
	if err := applyNAT(ip); err != nil {
		return fmt.Errorf("failed to apply nat: %w", err)
	}
	errCh := make(chan error, 1)
	go heartbeatNAT(ip, port, errCh)
	ticker := time.NewTicker(applyInterval)
	defer ticker.Stop()
	for {
		select {
		case err := <-errCh:
			return fmt.Errorf("failed to heartbeat: %w", err)
		case <-ticker.C:
			if err := applyNAT(ip); err != nil {
				return fmt.Errorf("failed to apply nat: %w", err)
			}
		}
	}
}

func applyNAT(ip net.IP) error {
	link, err := findRoute(ip)
	if err != nil {
		return fmt.Errorf("failed to find route to [%s]: %w", ip.String(), err)
	}
	if err := setDefaultRoute(link); err != nil {
		return fmt.Errorf("failed to set default route via [%s]: %w", link.Attrs().Name, err)
	}
	return nil
}

func heartbeatNAT(ip net.IP, port int, errCh chan<- error) {
	defer close(errCh)
	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		errCh <- fmt.Errorf("failed to dial [%s] on [%d]: %w", ip.String(), port, err)
		return
	}
	defer func() { _ = conn.Close() }()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	remoteAddr := &net.UDPAddr{IP: ip, Port: port}
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()
	var continuousErrorCount int
	for range ticker.C {
		query := generateQuery(localAddr.Port)
		b, err := internal.Marshal(query)
		if err != nil {
			errCh <- fmt.Errorf("failed to encode the query: %w", err)
			return
		}
		if err := conn.SetReadDeadline(time.Now().Add(heartbeatTimeout)); err != nil {
			errCh <- fmt.Errorf("failed to set read deadline: %w", err)
			return
		}
		if _, err := conn.WriteTo(b, remoteAddr); err != nil {
			errCh <- fmt.Errorf("failed to write to [%s]: %w", remoteAddr, err)
			return
		}
		b = make([]byte, 256)
		reply := new(internal.QueryReply)
		n, err := conn.Read(b)
		if err != nil {
			err = fmt.Errorf("failed to read query reply: %w", err)
			goto countError
		}
		err = internal.Unmarshal(b[:n], reply)
		if err != nil {
			err = fmt.Errorf("failed to decode the query reply: %w", err)
			goto countError
		}
		if reply.Magic != internal.QueryMagic {
			err = fmt.Errorf("reply magic [%s] mismatches", query.Magic)
			goto countError
		}
		if reply.Token != query.Token {
			err = fmt.Errorf("reply token [%s] mismatches [%s]", reply.Token, query.Token)
			goto countError
		}
		continuousErrorCount = 0
		log.Debugf("successfully heartbeated")
		continue
	countError:
		continuousErrorCount++
		if continuousErrorCount > heartbeatMaxError {
			errCh <- fmt.Errorf("failed to heartbeat continuously, the last error: %w", err)
		}
	}
}
