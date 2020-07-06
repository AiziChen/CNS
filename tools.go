package main

import (
	"net"
	"strings"
)

func toAddrs(addrs string) []string {
	addrs = strings.ReplaceAll(addrs, " ", "")
	return strings.Split(addrs, ",")
}

const EACH_SIZE = 1024

func readLine(conn *net.TCPConn) []byte {
	var buff = make([]byte, EACH_SIZE)
	l, err := conn.Read(buff)
	if err != nil || l <= 0 {
		return nil
	}
	if l < EACH_SIZE {
		return buff[:l]
	} else {
		return append(buff, readLine(conn)...)
	}
}
