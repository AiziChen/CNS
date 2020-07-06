// tcp.go
package main

import (
	"bytes"
	"log"
	"net"
	"strings"
)

func tcpForward(fromConn, toConn *net.TCPConn) {
	var CuteBi_XorCrypt_passwordSub int = 0
	for {
		data := readLine(fromConn)
		if data == nil {
			break
		}
		if CuteBi_XorCrypt_password != nil {
			CuteBi_XorCrypt_passwordSub = CuteBi_XorCrypt(data, CuteBi_XorCrypt_passwordSub)
		}
		if _, err := toConn.Write(data); err != nil {
			break
		}
	}
	toConn.Close()
	fromConn.Close()
}

func getProxyHost(header []byte) string {
	var hostSub int = bytes.Index(header, proxyKey)
	if hostSub < 0 {
		return ""
	}
	hostSub += len(proxyKey)
	hostEndSub := bytes.IndexByte(header[hostSub:], '\r')
	if hostEndSub < 0 {
		return ""
	}
	hostEndSub += hostSub
	if CuteBi_XorCrypt_password != nil {
		host, err := CuteBi_decrypt_host(header[hostSub:hostEndSub])
		if err != nil {
			log.Println(err)
			return ""
		}
		return string(host)
	} else {
		return string(header[hostSub:hostEndSub])
	}
}

func handleTcpSession(cConn *net.TCPConn, header []byte) {
	host := getProxyHost(header)
	if host == "" {
		log.Println("No proxy host: {" + string(header) + "}")
		cConn.Write([]byte("No proxy host"))
		cConn.Close()
		return
	}
	log.Println("proxyHost: " + host)
	//tcpDNS over udpDNS
	if enable_dns_tcpOverUdp && strings.HasSuffix(host, ":53") {
		dns_tcpOverUdp(cConn, host, header)
		return
	}

	/* 连接目标地址 */
	if !strings.Contains(host, ":") {
		host += ":80"
	}
	sAddr, resErr := net.ResolveTCPAddr("tcp", host)
	if resErr != nil {
		log.Println(resErr)
		cConn.Write([]byte("Proxy address [" + host + "] ResolveTCP() error"))
		cConn.Close()
		return
	}
	sConn, dialErr := net.DialTCP("tcp", nil, sAddr)
	if dialErr != nil {
		log.Println(dialErr)
		cConn.Write([]byte("Proxy address [" + host + "] DialTCP() error"))
		cConn.Close()
		return
	}
	sConn.SetKeepAlive(true)
	sConn.SetKeepAlivePeriod(tcp_keepAlive)
	/* start forward */
	log.Println("Start tcpForward")
	go tcpForward(cConn, sConn)
	tcpForward(sConn, cConn)

	log.Println("A tcp client close")
}
