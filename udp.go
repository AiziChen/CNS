// udp.go
package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"time"
)

type UdpSession struct {
	cConn                           *net.TCPConn
	udpSConn                        *net.UDPConn
	c2s_CuteBi_XorCrypt_passwordSub int
	s2c_CuteBi_XorCrypt_passwordSub int
}

func (udpSess *UdpSession) udpServerToClient() {
	var ignore_head_len int = 0
	payload := make([]byte, 65536)
	udpSess.cConn.SetDeadline(time.Now().Add(udp_timeout))
	udpSess.udpSConn.SetDeadline(time.Now().Add(udp_timeout))
	for {
		payload_len, RAddr, err := udpSess.udpSConn.ReadFromUDP(payload[24:] /*24为httpUDP协议头保留使用*/)
		if err != nil || payload_len <= 0 {
			break
		}
		fmt.Println("readUdpServerLen: ", payload_len, "RAddr: ", RAddr.String())
		if bytes.HasPrefix(RAddr.IP, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff}) {
			/* ipv4 */
			// 忽略数组前面的12个字节
			ignore_head_len = 12
			// 从第13个字节开始设置协议头
			payload[12] = byte(payload_len + 10)
			payload[13] = byte((payload_len + 10) >> 8)
			copy(payload[14:18], []byte{0, 0, 0, 1})
			copy(payload[18:22], []byte(RAddr.IP)[12:16])
		} else {
			/* ipv6 */
			ignore_head_len = 0
			payload[0] = byte(payload_len + 22)
			payload[1] = byte((payload_len + 22) >> 8)
			copy(payload[2:6], []byte{0, 0, 0, 3})
			copy(payload[6:22], []byte(RAddr.IP))
		}
		payload[22] = byte(RAddr.Port >> 8)
		payload[23] = byte(RAddr.Port)
		if CuteBi_XorCrypt_password != nil {
			udpSess.s2c_CuteBi_XorCrypt_passwordSub = CuteBi_XorCrypt(payload[ignore_head_len:24+payload_len], udpSess.s2c_CuteBi_XorCrypt_passwordSub)
		}
		if WLen, err := udpSess.cConn.Write(payload[ignore_head_len : 24+payload_len]); err != nil || WLen <= 0 {
			break
		}
	}
	udpSess.udpSConn.Close()
	udpSess.cConn.Close()
}

func (udpSess *UdpSession) writeToServer(httpUDP_data []byte) int {
	var (
		udpAddr                           net.UDPAddr
		err                               error
		WLen                              int
		pkgSub, httpUDP_protocol_head_len int
		pkgLen                            uint16
	)
	for pkgSub = 0; pkgSub+2 < len(httpUDP_data); pkgSub += 2 + int(pkgLen) {
		pkgLen = uint16(httpUDP_data[pkgSub]) | (uint16(httpUDP_data[pkgSub+1]) << 8) //2字节储存包的长度，包括socks5头
		//log.Println("pkgSub: ", pkgSub, ", pkgLen: ", pkgLen, "  ", uint16(len(httpUDP_data)))
		if pkgSub+2+int(pkgLen) > len(httpUDP_data) || pkgLen <= 10 {
			return 0
		}
		if !bytes.HasPrefix(httpUDP_data[pkgSub+3:pkgSub+5], []byte{0, 0}) {
			return 1
		}
		if httpUDP_data[5] == 1 {
			/* ipv4 */
			udpAddr.IP = net.IPv4(httpUDP_data[pkgSub+6], httpUDP_data[pkgSub+7], httpUDP_data[pkgSub+8], httpUDP_data[pkgSub+9])
			udpAddr.Port = int((uint16(httpUDP_data[pkgSub+10]) << 8) | uint16(httpUDP_data[pkgSub+11]))
			httpUDP_protocol_head_len = 12
		} else {
			if pkgLen <= 24 {
				return 0
			}
			/* ipv6 */
			udpAddr.IP = net.IP(httpUDP_data[pkgSub+6 : pkgSub+22])
			udpAddr.Port = int((uint16(httpUDP_data[pkgSub+22]) << 8) | uint16(httpUDP_data[pkgSub+23]))
			httpUDP_protocol_head_len = 24
		}
		//log.Println("WriteToUdpAddr: ", udpAddr.String())
		if WLen, err = udpSess.udpSConn.WriteToUDP(httpUDP_data[pkgSub+httpUDP_protocol_head_len:pkgSub+2+int(pkgLen)], &udpAddr); err != nil || WLen <= 0 {
			return -1
		}
	}
	return int(pkgSub)
}

func (udpSess *UdpSession) udpClientToServer(httpUDP_data []byte) {
	var payload_len, RLen, WLen int
	var err error
	WLen = udpSess.writeToServer(httpUDP_data)
	if WLen == -1 {
		udpSess.udpSConn.Close()
		udpSess.cConn.Close()
		return
	}
	payload := make([]byte, 65536)
	if WLen < len(httpUDP_data) {
		payload_len = copy(payload, httpUDP_data[WLen:])
	}
	udpSess.cConn.SetDeadline(time.Now().Add(udp_timeout))
	udpSess.udpSConn.SetReadDeadline(time.Now().Add(udp_timeout))
	for {
		RLen, err = udpSess.cConn.Read(payload[payload_len:])
		if err != nil || RLen <= 0 {
			break
		}
		if CuteBi_XorCrypt_password != nil {
			udpSess.c2s_CuteBi_XorCrypt_passwordSub = CuteBi_XorCrypt(payload[payload_len:payload_len+RLen], udpSess.c2s_CuteBi_XorCrypt_passwordSub)
		}
		payload_len += RLen
		//log.Println("Read Client: ", payload_len)
		WLen = udpSess.writeToServer(payload[:payload_len])
		if WLen == -1 {
			break
		} else if WLen < payload_len {
			payload_len = copy(payload, payload[WLen:payload_len])
		} else {
			payload_len = 0
		}
	}
	udpSess.udpSConn.Close()
	udpSess.cConn.Close()
}

func (udpSess *UdpSession) initUdp(httpUDP_data []byte) bool {
	if CuteBi_XorCrypt_password != nil {
		de := make([]byte, 5)
		copy(de, httpUDP_data[0:5])
		CuteBi_XorCrypt(de, 0)
		if de[2] != 0 || de[3] != 0 || de[4] != 0 {
			return false
		}
		udpSess.c2s_CuteBi_XorCrypt_passwordSub = CuteBi_XorCrypt(httpUDP_data, 0)
	}
	var err error
	udpSess.udpSConn, err = net.ListenUDP("udp", nil)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

func handleUdpSession(cConn *net.TCPConn, httpUDP_data []byte) {
	udpSess := new(UdpSession)
	udpSess.cConn = cConn
	if !udpSess.initUdp(httpUDP_data) {
		cConn.Close()
		log.Println("Is not httpUDP protocol or Decrypt failed")
		return
	}
	log.Println("Start udpForward")
	go udpSess.udpClientToServer(httpUDP_data)
	udpSess.udpServerToClient()
	log.Println("A udp client close")
}
