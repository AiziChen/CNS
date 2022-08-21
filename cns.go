// cns.go
package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"time"
)

var (
	udpFlag                                           string
	enable_dns_tcpOverUdp, enable_httpDNS, enable_TFO bool
	listenAddrs                                       []string
	udp_timeout                                       time.Duration
	hostRegex                                         *regexp.Regexp
)

var METHODS [][]byte = [][]byte{
	[]byte("GET"), []byte("POST"), []byte("HEAD"), []byte("PUT"), []byte("COPY"), []byte("DELETE"), []byte("MOVE"), []byte("OPTOINS"), []byte("LINK"), []byte("UNLINK"), []byte("TRACE"),
	[]byte("PATCH"), []byte("WRAPPED"),
}

func isHttpHeader(header []byte) bool {
	for i := 0; i < len(METHODS); i++ {
		if bytes.HasPrefix(header, METHODS[i]) {
			return true
		}
	}
	return false
}

func rspHeader(header []byte) []byte {
	if bytes.Contains(header, []byte("WebSocket")) {
		return []byte("HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: CuteBi Network Tunnel, (%>w<%)\r\n\r\n")
	} else if bytes.HasPrefix(header, []byte("CON")) {
		return []byte("HTTP/1.1 200 Connection established\r\nServer: CuteBi Network Tunnel, (%>w<%)\r\nConnection: keep-alive\r\n\r\n")
	} else {
		return []byte("HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\nServer: CuteBi Network Tunnel, (%>w<%)\r\nConnection: keep-alive\r\n\r\n")
	}
}

func handleConn(cConn *net.TCPConn) {
	var payload = make([]byte, 65536)
	len, err := cConn.Read(payload)
	if err != nil {
		cConn.Close()
		return
	}
	payload = payload[:len]
	if isHttpHeader(payload) {
		/* handle http requests */
		// process httpDNS request first
		if enable_httpDNS {
			if domain, err := GetHttpdnsDomain(payload); err == nil {
				RespondHttpdns(cConn, domain)
				return
			}
		}
		// process TCP & UDP request next
		if WLen, err := cConn.Write(rspHeader(payload)); err != nil || WLen <= 0 {
			cConn.Close()
			return
		}
		if bytes.Contains(payload, []byte(udpFlag)) {
			// 丢弃含有udpFlag标识的请求
			handleConn(cConn)
		} else {
			handleTcpSession(cConn, payload)
		}
	} else {
		// handle udp requests
		handleUdpSession(cConn, payload)
	}
}

func pidSaveToFile(pidPath string) {
	fp, err := os.Create(pidPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = fp.WriteString(fmt.Sprintf("%d", os.Getpid()))
	if err != nil {
		fmt.Println(err)
	}
	fp.Close()
}

func initConfig() {
	var CuteBi_XorCrypt_passwordStr, pidPath string
	var isHelp, enable_daemon bool
	var configFile string

	flag.BoolVar(&enable_daemon, "daemon", true, "daemon mode switch(开启后台运行)")
	flag.StringVar(&configFile, "config-file", "config.cfg", "set configuration file, default `config.cfg`(指定配置文件，不指定时默认为`config.cfg`)")
	flag.BoolVar(&isHelp, "help", false, "display this message(显示此帮助信息)")
	flag.Parse()

	configMap := InitConfig(configFile)
	if proxyKey := configMap["proxyKey"]; proxyKey != "" {
		hostRegex = regexp.MustCompile(proxyKey + ":\\s*(.*)\r")
	} else {
		fmt.Fprintf(os.Stderr, "proxyKey为必填项，请先在配置文件中设置它")
		return
	}
	udpFlag = configMap["udpFlag"]
	listenAddrs = toAddrs(configMap["listenAddr"])
	CuteBi_XorCrypt_passwordStr = configMap["password"]
	/* udp timeout */
	udpTimeout, err := strconv.ParseInt(configMap["udpTimeout"], 10, 64)
	if err != nil {
		fmt.Printf("udpTimeout参数错误：%v，将使用默认值30", configMap["udpTimeout"])
		udp_timeout = 30
	} else {
		udp_timeout = time.Duration(udpTimeout)
	}
	udp_timeout *= time.Second

	pidPath = configMap["pidPath"]
	if configMap["enableDnsTcpOverUdp"] == "#t" {
		enable_dns_tcpOverUdp = true
	} else {
		enable_dns_tcpOverUdp = false
	}
	enable_httpDNS = configMap["enableHttpDNS"] == "#t"

	enable_TFO = configMap["enableTFO"] == "#t"

	if isHelp {
		fmt.Println("CuteBi Network Server v0.2.4")
		flag.Usage()
		os.Exit(0)
	}
	if enable_daemon {
		exec.Command(os.Args[0], []string(append(os.Args[1:], "-daemon=false"))...).Start()
		os.Exit(0)
	}

	if CuteBi_XorCrypt_passwordStr == "" {
		CuteBi_XorCrypt_password = nil
	} else {
		CuteBi_XorCrypt_password = []byte(CuteBi_XorCrypt_passwordStr)
	}

	if pidPath != "" {
		pidSaveToFile(pidPath)
	}
	setsid()
	setMaxNofile()
}

func handling(listener *net.TCPListener) {
	for {
		conn, err := listener.AcceptTCP()
		if err == nil {
			go handleConn(conn)
		} else {
			log.Println(err)
			time.Sleep(3 * time.Second)
		}
	}
	// listener.Close()
}

func initListener(listenAddr string) *net.TCPListener {
	addr, _ := net.ResolveTCPAddr("tcp", listenAddr)
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	if enable_TFO {
		enableTcpFastOpen(listener)
	}
	return listener
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	initConfig()
	addrsLen := len(listenAddrs)
	for i := 0; i < addrsLen-1; i++ {
		listener := initListener(listenAddrs[i])
		go handling(listener)
	}
	listener := initListener(listenAddrs[addrsLen-1])
	handling(listener)
}
