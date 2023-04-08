// tcp.go
package main

import (
	"errors"
	"log"
	"net"
	"strings"
)

func tcpForward(fromConn, toConn *net.TCPConn) {
	var CuteBi_XorCrypt_passwordSub int = 0
	payload := make([]byte, 65536)
	for {
		len, err := fromConn.Read(payload)
		if err != nil {
			log.Println("tcp-forward read failed.")
			break
		}
		if CuteBi_XorCrypt_password != nil {
			CuteBi_XorCrypt_passwordSub = CuteBi_XorCrypt(payload[:len], CuteBi_XorCrypt_passwordSub)
		}
		if _, err := toConn.Write(payload[:len]); err != nil {
			log.Println("tcp-forward write failed.")
			break
		}
	}
	toConn.Close()
	fromConn.Close()
}

func getProxyHost(header []byte) (string, error) {
	found := hostRegex.FindSubmatch(header)
	if len(found) < 2 {
		return "", errors.New("not found host in header")
	}

	if CuteBi_XorCrypt_password != nil {
		host, err := CuteBi_decrypt_host(found[1])
		if err != nil {
			log.Println(err)
			return "", err
		}
		return string(host), nil
	} else {
		return string(found[1]), nil
	}
}

func handleTcpSession(cConn *net.TCPConn, header []byte) {
	host, err := getProxyHost(header)
	if err != nil {
		log.Println("No proxy host: {" + string(header) + "}")
		cConn.Write([]byte("No proxy host"))
		cConn.Close()
		return
	}
	log.Println("proxyHost: " + host)

	if enable_dns_tcpOverUdp && strings.HasSuffix(host, ":53") {
		// tcpDNS over udpDNS
		dns_tcpOverUdp(cConn, host, header)
		cConn.Close()
		return
	}

	/* connecting to the destination host */
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
	cConn.SetKeepAlive(true)
	/* starting forward */
	log.Println("Start tcpForward")
	go tcpForward(cConn, sConn)
	tcpForward(sConn, cConn)

	log.Println("A tcp client close")
}
