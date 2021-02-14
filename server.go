package main

import (
	"fmt"
	"io"
	"net"
	"runtime"
	"sync"
	"time"
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
	// 用户上线，创建User并调用User的业务方上线处理
	user := NewUser(conn, this)
	user.Online()

	// 监听用户是否活跃的channel
	isLive := make(chan bool)

	// 接收客户端发来的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			// 读取的字节数为0表示客户端合法关闭
			if n == 0 {
				// User下线业务
				user.Offline()
				return
			}
			// err = io.EOF表示到了文件尾
			if err != nil && err != io.EOF {
				fmt.Println("[server handle] conn read error:", err)
				return
			}
			// 提取用户的消息，去除最后的"\n"
			msg := string(buf[:n-1])
			// 处理User发消息的业务
			user.DoMessage(msg)

			// 读到用户的任意消息，接收到认为用户是活跃的
			isLive <- true
		}
	}()

	for {
		select {
		// 读到用户活跃消息，重置定时器
		case <-isLive:
			//time.After(time.Second * 10)
			// 这里不用显式调一下，case进来之后还是会去匹配下面的case，然后就更新了定时器
		// 十秒超时，将当前用户强制下线
		case <-time.After(time.Second * 10):
			// 发踢出去的消息
			user.C <- "You will be offline."
			// 销毁channel资源
			close(user.C)
			// 关闭连接
			conn.Close()
			// 退出当前的handler
			runtime.Goexit() // return也行
		}
	}
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
