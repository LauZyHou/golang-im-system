package main

import (
	"fmt"
	"net"
	"sync"
)

// Server 服务器
type Server struct {
	Ip        string           // ip
	Port      int              // 端口
	OnlineMap map[string]*User // 在线用户表
	mapLock   sync.RWMutex     // 在线用户表的读写锁
	Message   chan string      // 消息广播的channel
}

// NewServer 创建新Server对象的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
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

	// 启动监听Message的go程
	go this.ListenMessage()

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
	// 用户上线，创建User并加入到OnlineMap中
	user := NewUser(conn)
	// 要添加元素所以要加锁
	this.mapLock.Lock()
	this.OnlineMap[user.Name] = user
	this.mapLock.Unlock()

	// 用户上线成功，广播这条上线消息
	this.BroadCast(user, "is online")

	// 让当前go程阻塞，否则子go程都会死掉
	select {}
}

// BroadCast 向所有User广播消息的方法
// user: 由哪个用户发起
// msg:  消息内容
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

// 监听Message一旦有就给所有User的channel发消息的go程
func (this *Server) ListenMessage() {
	for {
		msg := <-this.Message
		// 将msg发送给全部的在线User，要遍历所以也要加锁
		this.mapLock.Lock()
		for _, user := range this.OnlineMap {
			user.C <- msg
		}
		this.mapLock.Unlock()
	}
}
