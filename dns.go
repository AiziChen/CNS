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

	RLen, err := cConn.Read(buf)
	if err != nil {
		return
	}
	if CuteBi_XorCrypt_password != nil {
		CuteBi_XorCrypt(buf[:RLen], 0)
	}

	/* Connectting to the destination address */
	sConn, dialErr := net.Dial("udp", host)
	if dialErr != nil {
		log.Println(dialErr)
		cConn.Write([]byte("Proxy address [" + host + "] DNS Dial() error"))
		return
	}
	defer sConn.Close()
	if WLen, err := sConn.Write(buf[2:RLen]); WLen <= 0 || err != nil {
		return
	}

	RLen, err = sConn.Read(buf[2:])
	if RLen <= 0 || err != nil {
		return
	}
	buf[0] = byte(RLen >> 8)
	buf[1] = byte(RLen)
	if CuteBi_XorCrypt_password != nil {
		CuteBi_XorCrypt(buf[:2+RLen], 0)
	}
	cConn.Write(buf[:2+RLen])
}

func GetHttpdnsDomain(header []byte) (string, error) {
	domain := dnsDomainRegexp.FindSubmatch(header)
	if len(domain) < 2 {
		return "", errors.New("get http DNS domain error")
	} else {
		return string(domain[1]), nil
	}
}

func RespondHttpdns(cConn *net.TCPConn, domain string) {
	log.Println("httpDNS domain: [" + domain + "]")
	ips, err := net.LookupHost(domain)
	if err != nil {
		cConn.Write([]byte("HTTP/1.0 404 Not Found\r\nConnection: Close\r\nServer: CuteBi Linux Network httpDNS, (%>w<%)\r\nContent-type: charset=utf-8\r\n\r\n<html><head><title>HTTP DNS Server</title></head><body>查询域名失败<br/><br/>By: 萌萌萌得不要不要哒<br/>E-mail: 915445800@qq.com</body></html>"))
		log.Println("httpDNS domain: [" + domain + "], Lookup failed")
	} else {
		for i := 0; i < len(ips); i++ {
			if !strings.Contains(ips[i], ":") { // 跳过ipv6
				fmt.Fprintf(cConn, "HTTP/1.0 200 OK\r\nConnection: Close\r\nServer: CuteBi Linux Network httpDNS, (%%>w<%%)\r\nContent-Length: %d\r\n\r\n%s", len(string(ips[i])), string(ips[i]))
				break
			}
		}
		log.Println("httpDNS domain: ["+domain+"], IPS: ", ips)
	}
	cConn.Close()
}
