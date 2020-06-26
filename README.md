# CuteBi Network Server

网络代理服务端, 支持 IPV6，tcpFastOpen，UDP_Over_HttpTunnel(需要配合专门的客户端)

1.  普通的 CONNECT 代理服务器(暂时不考虑添加普通 http 支持)
2.  实现与 114DNS 以及腾讯的 dnsPod 一样的 httpDNS 服务端
3.  配合专门的客户端可以实现 TCP/UDP 全局代理, 目前只有: https://github.com/mmmdbybyd/CLNC

## 服务端

    1. 普通的CONNECT代理服务器(暂时不考虑添加普通http支持)
    2. 实现与114DNS以及腾讯的dnsPod一样的httpDNS服务端

## 服务端+客户端

    1. 使用自己的加密协议加密流量
    2. 可伪装为各种HTTP流量
    3. 支持UDP_Over_HttpTunnel

##### 配置文件(config.cfg)

    必选参数:
    proxyKey                    代理头域, 默认: 'Meng'
    udpFlag                     udp请求标识, 默认: 'httpUDP'
    listenAddr                  监听地址, 默认: ':80'
    udpTimeout                  udp连接超时, 默认: 30s
    tcpKeepAlive                tcp生存时间, 默认: 60s
    可选参数:
    password                    加密密码, 没有则不加密
    pidPath                     pid文件路径, 没有则不保存
    enableHttpDNS               httpDNS开关, 默认关闭
    enableTFO                   tcpFastOpen开关, 默认关闭
    enableDaemon                开启后台运行, 默认关闭

##### 命令行选项

    -h, --help                  显示帮助
    -config-file                指定`config-file.cfg`的文件路径

##### 编译

```
go build -o cns
```

##### 使用方法

1. 非 root 用户：

```
sudo cns
```

2. root 用户

```
./cns
```
