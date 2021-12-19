// tcp.go
package main

import (
	"log"
	"net"
	"regexp"
	"strings"
)

func tcpForward(fromConn, toConn *net.TCPConn) {
	var CuteBi_XorCrypt_passwordSub int = 0
	var data = make([]byte, 65536)
	for {
		len, err := fromConn.Read(data)
		if err != nil {
			log.Println("tcp-forward read failed.")
			break
		}
		if CuteBi_XorCrypt_password != nil {
			CuteBi_XorCrypt_passwordSub = CuteBi_XorCrypt(data[:len], CuteBi_XorCrypt_passwordSub)
		}
		if _, err := toConn.Write(data[:len]); err != nil {
			log.Println("tcp-forward write failed.")
			break
		}
	}
	toConn.Close()
	fromConn.Close()
}

func getProxyHost(header []byte) string {
	re := regexp.MustCompile(proxyKey + ":\\s*(.*)\r")
	found := re.FindSubmatch(header)
	if len(found) < 2 {
		return ""
	}

	if CuteBi_XorCrypt_password != nil {
		host, err := CuteBi_decrypt_host(found[1])
		if err != nil {
			log.Println(err)
			return ""
		}
		return string(host)
	} else {
		return string(found[1])
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

	if strings.HasSuffix(host, ":53") {
		// tcpDNS over udpDNS
		if enable_dns_tcpOverUdp {
			dns_tcpOverUdp(cConn, host, header)
		}
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
	sConn.SetKeepAlivePeriod(tcp_keepAlive)
	/* start forward */
	log.Println("Start tcpForward")
	go tcpForward(cConn, sConn)
	tcpForward(sConn, cConn)

	log.Println("A tcp client close")
}
