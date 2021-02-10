package main

import (
	"net"
)

// User 表达用户在服务器上的实例
type User struct {
	Name string      // 用户唯一标识，不需要是业务角度的用户名
	Addr string      // ip地址
	C    chan string // 用来给用户client发消息的channel
	conn net.Conn    // 连接
}

// NewUser 创建一个User的接口
func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,
	}

	// 创建后就启动监听当前User channel的go程
	go user.ListenMessage()

	return user
}

// 监听当前User channel的方法，一旦有消息，就发送给对端客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		_, _ = this.conn.Write([]byte(msg + "\n"))
	}
}
