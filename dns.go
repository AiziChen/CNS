// dns.go
package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"
)

var dnsDomainRegexp = regexp.MustCompile(`\?dn=(.*)`)

func dns_tcpOverUdp(cConn *net.TCPConn, host string, buf []byte) {
	log.Println("Start dns_tcpOverUdp")
	var rlen, err = cConn.Read(buf)
	if err == nil {
		if CuteBi_XorCrypt_password != nil {
			CuteBi_XorCrypt(buf[:rlen], 0)
		}
		/* Connectting to the destination address */
		var sConn, err = net.Dial("udp", host)
		if err != nil {
			log.Println(err)
			cConn.Write([]byte("Proxy address [" + host + "] DNS Dial() error"))
		} else {
			var WLen, err = sConn.Write(buf[2:rlen])
			if WLen > 0 || err == nil {
				var rlen2, err = sConn.Read(buf[2:])
				if rlen2 > 0 || err == nil {
					buf[0] = byte(rlen2 >> 8)
					buf[1] = byte(rlen2)
					if CuteBi_XorCrypt_password != nil {
						CuteBi_XorCrypt(buf[:2+rlen2], 0)
					}
					cConn.Write(buf[:2+rlen2])
				}
			}
		}
		sConn.Close()
	}
}

func GetHttpdnsDomain(header []byte) (string, error) {
	var domain = dnsDomainRegexp.FindSubmatch(header)
	if len(domain) < 2 {
		return "", errors.New("get http DNS domain error")
	} else {
		return string(domain[1]), nil
	}
}

func RespondHttpdns(cConn *net.TCPConn, domain string) {
	log.Println("httpDNS domain: [" + domain + "]")
	var ips, err = net.LookupHost(domain)
	if err != nil {
		cConn.Write([]byte("HTTP/1.0 404 Not Found\r\nConnection: Close\r\nServer: CuteBi Linux Network httpDNS, (%>w<%)\r\nContent-type: charset=utf-8\r\n\r\n<html><head><title>HTTP DNS Server</title></head><body>查询域名失败<br/><br/>By: 萌萌萌得不要不要哒<br/>E-mail: 915445800@qq.com</body></html>"))
		log.Println("httpDNS domain: [" + domain + "], Lookup failed")
	} else {
		for _, ip := range ips {
			// 跳过ipv6
			if !strings.Contains(ip, ":") {
				fmt.Fprintf(cConn, "HTTP/1.0 200 OK\r\nConnection: Close\r\nServer: CuteBi Linux Network httpDNS, (%%>w<%%)\r\nContent-Length: %d\r\n\r\n%s", len(string(ip)), string(ip))
				break
			}
		}
		log.Println("httpDNS domain: ["+domain+"], IPS: ", ips)
	}
	cConn.Close()
}
