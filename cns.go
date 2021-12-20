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
	"runtime"
	"strconv"
	"time"
)

var (
	udpFlag                                           string
	proxyKey                                          string
	udp_timeout                                       time.Duration
	enable_dns_tcpOverUdp, enable_httpDNS, enable_TFO bool
	listenAddrs                                       []string
)

func isHttpHeader(header []byte) bool {
	return bytes.HasPrefix(header, []byte("CONNECT")) ||
		bytes.HasPrefix(header, []byte("GET")) ||
		bytes.HasPrefix(header, []byte("POST")) ||
		bytes.HasPrefix(header, []byte("HEAD")) ||
		bytes.HasPrefix(header, []byte("PUT")) ||
		bytes.HasPrefix(header, []byte("COPY")) ||
		bytes.HasPrefix(header, []byte("DELETE")) ||
		bytes.HasPrefix(header, []byte("MOVE")) ||
		bytes.HasPrefix(header, []byte("OPTIONS")) ||
		bytes.HasPrefix(header, []byte("LINK")) ||
		bytes.HasPrefix(header, []byte("UNLINK")) ||
		bytes.HasPrefix(header, []byte("TRACE")) ||
		bytes.HasPrefix(header, []byte("PATCH")) ||
		bytes.HasPrefix(header, []byte("WRAPPED"))
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
	// data := readLine(cConn)
	var data = make([]byte, 65536)
	len, err := cConn.Read(data)
	if err != nil {
		cConn.Close()
		return
	}
	data = data[:len]
	if isHttpHeader(data) {
		// handle http requests
		if !enable_httpDNS || !RespondHttpdns(cConn, data) { /*先处理httpDNS请求*/
			if WLen, err := cConn.Write(rspHeader(data)); err != nil || WLen <= 0 {
				cConn.Close()
				return
			}
			if bytes.Contains(data, []byte(udpFlag)) {
				// 丢弃含有udpFlag标识的请求
				handleConn(cConn)
			} else {
				handleTcpSession(cConn, data)
			}
		}
	} else {
		// handle udp requests
		handleUdpSession(cConn, data)
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
	proxyKey = configMap["proxyKey"]
	udpFlag = configMap["udpFlag"]
	listenAddrs = toAddrs(configMap["listenAddr"])
	CuteBi_XorCrypt_passwordStr = configMap["password"]
	udpTimeout, err := strconv.ParseInt(configMap["udpTimeout"], 10, 64)
	if err != nil {
		fmt.Printf("udpTimeout参数指定错误：%v，将使用默认值30", configMap["udpTimeout"])
		udp_timeout = 30
	} else {
		udp_timeout = time.Duration(udpTimeout)
	}
	pidPath = configMap["pidPath"]
	if rs := configMap["enableDnsTcpOverUdp"]; rs == "#t" {
		enable_dns_tcpOverUdp = true
	} else {
		enable_dns_tcpOverUdp = false
	}
	if configMap["enableHttpDNS"] == "#t" {
		enable_httpDNS = true
	} else {
		enable_httpDNS = false
	}
	if configMap["enableTFO"] == "#t" {
		enable_TFO = true
	} else {
		enable_TFO = false
	}

	if isHelp {
		fmt.Println("CuteBi Network Server v0.2.1")
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
	udp_timeout *= time.Second

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
			conn.SetKeepAlive(true)
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
