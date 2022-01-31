package internal

import "net"

const QueryMagic = "猹Gateway🐳"

type Query struct {
	Magic string
	Token string
	Port  int
}

type QueryReply struct {
	Magic    string
	Token    string
	ClientIP net.IP
	Port     int
}
