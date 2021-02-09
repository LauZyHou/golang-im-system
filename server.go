package main

import (
	"fmt"
	"net"
)

// Server 服务器
type Server struct {
	Ip   string // ip
	Port int // 端口
}

// NewServer 创建新Server对象的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:   ip,
		Port: port,
	}
	return server
}

// Start 启动Server的方法
func (this *Server) Start() {
	// socket listen and close
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("[server start] net listen error:", err)
		return
	}
	defer listener.Close()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("[server run] listener accept error:", err)
			continue
		}
		// do handler 开一个go程处理业务回调
		go this.Handle(conn)
	}

}

// Handle 连接建立以后的业务回调方法
func (this *Server) Handle(conn net.Conn) {
	fmt.Println("连接建立成功，处理业务回调")

}