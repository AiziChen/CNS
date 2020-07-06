package main

import (
	"net"
	"strings"
)

func toAddrs(addrs string) []string {
	addrs = strings.ReplaceAll(addrs, " ", "")
	return strings.Split(addrs, ",")
}

const EACH_SIZE = 2048

func readLine(conn *net.TCPConn) []byte {
	var buff = make([]byte, EACH_SIZE)
	l, err := conn.Read(buff)
	if err != nil || l <= 0 {
		return nil
	}
	if l <= EACH_SIZE {
		return buff[:l]
	} else {
		return append(buff, readLine(conn)...)
	}
}

func readLine2(conn *net.TCPConn) []byte {
	var buff = make([]byte, EACH_SIZE*2)
	l, err := conn.Read(buff)
	if err != nil || l <= 0 {
		return nil
	}
	if l <= EACH_SIZE {
		return buff[:l]
	} else {
		return append(buff, readLine(conn)...)
	}
}

func readLineFromUdp(conn *net.UDPConn) ([]byte, *net.UDPAddr) {
	var buff = make([]byte, EACH_SIZE*2)
	l, addr, err := conn.ReadFromUDP(buff)
	if err != nil || l <= 0 {
		return nil, nil
	}
	if l <= EACH_SIZE {
		return buff[:l], addr
	} else {
		b, addr := readLineFromUdp(conn)
		return append(buff, b...), addr
	}
}
