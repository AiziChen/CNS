// tcp.go
package main

import (
	"errors"
	"log"
	"net"
	"strings"
)

func forwardLoop(fromConn *net.TCPConn, toConn *net.TCPConn, payload []byte, sub int) {
	var len, err = fromConn.Read(payload)
	if err == nil {
		if CuteBi_XorCrypt_password != nil {
			sub = CuteBi_XorCrypt(payload[:len], sub)
		}
		var _, err = toConn.Write(payload[:len])
		if err == nil {
			forwardLoop(fromConn, toConn, payload, sub)
		} else {
			log.Println("tcp-forward write failed.")
		}
	} else {
		log.Println("tcp-forward read failed.")
	}
}

func tcpForward(fromConn *net.TCPConn, toConn *net.TCPConn) {
	var CuteBi_XorCrypt_passwordSub int = 0
	var payload = make([]byte, 65536)
	forwardLoop(fromConn, toConn, payload, CuteBi_XorCrypt_passwordSub)
	fromConn.Close();
	toConn.Close();
}

func getProxyHost(header []byte) (string, error) {
	var found = hostRegex.FindSubmatch(header)
	if len(found) >= 2 {
		if CuteBi_XorCrypt_password != nil {
			var host, err = CuteBi_decrypt_host(found[1])
			if err != nil {
				log.Println(err)
				return "", err
			} else {
				return string(host), nil
			}
		} else {
			return string(found[1]), nil
		}
	} else {
		return "", errors.New("not found host in header")
	}
}

func handleTcpSession(cConn *net.TCPConn, header []byte) {
	var host, err = getProxyHost(header)
	if err != nil {
		log.Println("No proxy host: {" + string(header) + "}")
		cConn.Write([]byte("No proxy host"))
	} else {
		log.Println("proxyHost: " + host)
		if !(enable_dns_tcpOverUdp && strings.HasSuffix(host, ":53")) {
			/* connecting to the destination host */
			if !strings.Contains(host, ":") {
				host += ":80"
			}
			var sAddr, resErr = net.ResolveTCPAddr("tcp", host)
			if resErr != nil {
				log.Println(resErr)
				cConn.Write([]byte("Proxy address [" + host + "] ResolveTCP() error"))
			} else {
				var sConn, dialErr = net.DialTCP("tcp", nil, sAddr)
				if dialErr != nil {
					log.Println(dialErr)
					cConn.Write([]byte("Proxy address [" + host + "] DialTCP() error"))
				} else {
					sConn.SetKeepAlive(true)
					cConn.SetKeepAlive(true)
					/* starting forward */
					log.Println("Start tcpForward")
					go tcpForward(cConn, sConn)
					tcpForward(sConn, cConn)
					sConn.Close()
				}
			}
		} else {
			// tcpDNS over udpDNS
			dns_tcpOverUdp(cConn, host, header)
		}
		cConn.Close()
		log.Println("A tcp client has been close")
	}
}
