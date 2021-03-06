package main

import (
	"fmt"
	"net"
	"strings"
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

// SendMsg 给当前User对应的用户发消息
// Deprecated: Just use channel this.C
func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

// DoMessage 处理用户发消息的业务
func (this *User) DoMessage(msg string) {
	if msg == "who" { // 当前用户查询有哪些用户在线
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + " is online."
			this.C <- onlineMsg
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" { // 修改用户名
		// 截取要修改到的新用户名
		newName := strings.Split(msg, "|")[1]
		// 判断newName是否已经被用户占用了
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.C <- "The new name has been used."
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.C <- fmt.Sprintf("Update user name to %v.", newName)
		}
	} else if len(msg) > 4 && msg[:3] == "to|" { // to|name|私聊消息
		spList := strings.Split(msg, "|")
		if len(spList) != 3 {
			this.C <- "Msg format error, please use to|name|content"
			return
		}

		recvName := spList[1]
		recvUser, ok := this.server.OnlineMap[recvName]
		if !ok {
			this.C <- fmt.Sprintf("User %v is not exist.", recvName)
			return
		}

		msgContent := spList[2]
		if msgContent == "" {
			this.C <- "Msg is empty."
			return
		}

		recvUser.C <- fmt.Sprintf("%v say: %v", this.Name, msgContent)
	} else { // 其它输入，进行消息广播
		this.server.BroadCast(this, msg)
	}
}
