package main

import (
	"flag"
	"fmt"
	"net"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
}

func NewClient(serverIp string, serverPort int) *Client {
	// 创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
	}

	// 连接Server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}

	client.conn = conn

	return client
}

// 从命令行中解析的服务器ip和端口存在这里
var serverIp string
var serverPort int

// init 在main之前执行，在这里绑定命令行用法
func init() {
	// client -ip 127.0.0.1 -port 8888
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "server ip (default is \"127.0.0.1\")")
	flag.IntVar(&serverPort, "port", 8888, "server port (default is 8888)")
}

func main() {
	// 解析命令行，解析结果在绑定的变量里
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println("Link server failed.")
		return
	}
	fmt.Println("Link server success.")

	// 启动客户端的业务
	select {}
}
