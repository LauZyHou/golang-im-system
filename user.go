package main

import (
	"net"
)

// User 表达用户在服务器上的实例
type User struct {
	Name   string      // 用户唯一标识，不需要是业务角度的用户名
	Addr   string      // ip地址
	C      chan string // 用来给用户client发消息的channel
	conn   net.Conn    // 连接
	server *Server     // 当前User是属于哪个Server的
}

// NewUser 创建一个User的接口
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	// 创建后就启动监听当前User channel的go程
	go user.ListenMessage()

	return user
}

// ListenMessage 监听当前User channel的方法，一旦有消息，就发送给对端客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		_, _ = this.conn.Write([]byte(msg + "\n"))
	}
}

// Online 用户的上线业务
func (this *User) Online() {
	// 当前User加入到Server的OnlineMap
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	// 用户上线成功，广播这条上线消息
	this.server.BroadCast(this, "is online")
}

// Offline 用户的下线业务
func (this *User) Offline() {
	// 当前User从Server的OnlineMap中移除
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	// 用户上线成功，广播这条上线消息
	this.server.BroadCast(this, "is offline")
}

// DoMessage 处理用户发消息的业务
func (this *User) DoMessage(msg string) {
	// 将当前用户发送的msg广播
	this.server.BroadCast(this, msg)
}
