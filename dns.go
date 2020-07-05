// dns.go
package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

func dns_tcpOverUdp(cConn *net.TCPConn, host string, buffer []byte) {
	log.Println("Start dns_tcpOverUdp")
	defer cConn.Close()

	var err error
	var WLen, RLen, payloadLen, CuteBi_XorCrypt_passwordSub int
	var pkgLen uint16
	for {
		//cConn.SetReadDeadline(time.Now().Add(tcp_timeout))
		RLen, err = cConn.Read(buffer[payloadLen:])
		if RLen <= 0 || err != nil {
			return
		}
		//解密
		if CuteBi_XorCrypt_password != nil {
			CuteBi_XorCrypt_passwordSub = CuteBi_XorCrypt(buffer[payloadLen:payloadLen+RLen], CuteBi_XorCrypt_passwordSub)
		}
		payloadLen += RLen
		if payloadLen > 2 {
			pkgLen = (uint16(buffer[0]) << 8) | (uint16(buffer[1])) //包长度转换
			//防止访问非法数据
			if int(pkgLen)+2 > len(buffer) {
				return
			}
			//如果读取到了一个完整的包，就跳出循环
			if int(pkgLen)+2 <= payloadLen {
				break
			}
		}
	}
	/* 连接目标地址 */
	sConn, dialErr := net.Dial("udp", host)
	if dialErr != nil {
		log.Println(dialErr)
		cConn.Write([]byte("Proxy address [" + host + "] DNS Dial() error"))
		return
	}
	defer sConn.Close()
	if WLen, err = sConn.Write(buffer[2:payloadLen]); WLen <= 0 || err != nil {
		return
	}
	sConn.SetReadDeadline(time.Now().Add(udp_timeout))
	if RLen, err = sConn.Read(buffer[2:]); RLen <= 0 || err != nil {
		return
	}
	//包长度转换
	buffer[0] = byte(RLen >> 8)
	buffer[1] = byte(RLen)
	//加密
	if CuteBi_XorCrypt_password != nil {
		CuteBi_XorCrypt(buffer[:2+RLen], 0)
	}
	cConn.Write(buffer[:2+RLen])
}

func RespondHttpdns(cConn *net.TCPConn, header []byte) bool {
	var domain string
	httpdnsDomainsub := bytes.Index(header[:], []byte("?dn="))
	if httpdnsDomainsub < 0 {
		return false
	}
	if _, err := fmt.Sscanf(string(header[httpdnsDomainsub+4:]), "%s", &domain); err != nil {
		log.Println(err)
		return false
	}

	log.Println("httpDNS domain: [" + domain + "]")
	ips, err := net.LookupHost(domain)
	if err != nil {
		cConn.Write([]byte("HTTP/1.0 404 Not Found\r\nConnection: Close\r\nServer: CuteBi Linux Network httpDNS, (%>w<%)\r\nContent-type: charset=utf-8\r\n\r\n<html><head><title>HTTP DNS Server</title></head><body>查询域名失败<br/><br/>By: 萌萌萌得不要不要哒<br/>E-mail: 915445800@qq.com</body></html>"))
		log.Println("httpDNS domain: [" + domain + "], Lookup failed")
	} else {
		for i := 0; i < len(ips); i++ {
			if !strings.Contains(ips[i], ":") { //跳过ipv6
				fmt.Fprintf(cConn, "HTTP/1.0 200 OK\r\nConnection: Close\r\nServer: CuteBi Linux Network httpDNS, (%%>w<%%)\r\nContent-Length: %d\r\n\r\n%s", len(string(ips[i])), string(ips[i]))
				break
			}
		}
		log.Println("httpDNS domain: ["+domain+"], IPS: ", ips)
	}
	cConn.Close()
	return true
}
