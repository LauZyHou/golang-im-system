package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

// Client 客户端
type Client struct {
	ServerIp   string   // 对端服务器ip
	ServerPort int      // 对端服务器端口
	Name       string   // 用户名
	conn       net.Conn // 和对端服务器建立的连接句柄
	flag       int      // 当前用户从菜单选择的操作模式
}

func NewClient(serverIp string, serverPort int) *Client {
	// 创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999, // 默认值不为0是因为0表示退出
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

// DealResponse 是用来处理Server回应的消息的goroutine
func (this *Client) DealResponse() {
	// 只要从this.conn里Read到东西就从os.Stdout打印出来，永久阻塞监听
	// 手动写一个for来不停conn.Read+fmt.Println处理也可以
	io.Copy(os.Stdout, this.conn)
}

// menu 打印菜单
func (this *Client) menu() bool {
	fmt.Println("1. Public chat")
	fmt.Println("2. Secret chat")
	fmt.Println("3. Rename")
	fmt.Println("0. Exit")

	var flag int
	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		this.flag = flag
		return true
	} else {
		fmt.Println("Undefined input flag")
		return false
	}
}

// SelectUsers 查询在线用户
func (this *Client) SelectUsers() {
	// 发一个who指令过去，回执由回执处理的goroutine处理并打印出来
	sendMsg := "who\n"
	_, err := this.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write error: ", err)
		return
	}
}

// PrivateChat 私聊模式业务
func (this *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	// 先查询当前用户
	this.SelectUsers()
	// 用户输入要通信的用户名
	fmt.Println("Please input recv user name. \"exit\" to exit.")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		// 用户输入私聊消息内容
		fmt.Println("Please input chat content. \"exit\" to exit.")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := fmt.Sprintf("to|%v|%v\n", remoteName, chatMsg)
				_, err := this.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn.Write error: ", err)
					break
				}
			}

			chatMsg = ""
			fmt.Println("Please input chat content. \"exit\" to exit.")
			fmt.Scanln(&chatMsg)
		}

		this.SelectUsers()
		fmt.Println("Please input recv user name. \"exit\" to exit.")
		fmt.Scanln(&remoteName)
	}
}

// PublicChat 公聊模式业务
func (this *Client) PublicChat() {
	// 提示用户输入消息
	fmt.Println("Please input chat content, \"exit\" to exit.")
	var chatMsg string
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		// 消息不为空则发给服务器
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := this.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn.Write error: ", err)
				break
			}
		}
		// 发完一轮发下一轮，也是提示+等待输入
		chatMsg = ""
		fmt.Println("Please input chat content, \"exit\" to exit.")
		fmt.Scanln(&chatMsg)
	}
}

// UpdateName update current user name
func (this *Client) UpdateName() bool {
	fmt.Println("Please input user name")
	fmt.Scanln(&this.Name)

	sendMsg := fmt.Sprintf("rename|%v\n", this.Name)
	_, err := this.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write error: ", err)
		return false
	}
	return true
}

// Run 是 Client 主业务
func (this *Client) Run() {
	// =0时才退出
	for this.flag != 0 {
		// 输入不是合法数字就一直菜单提示用户输入
		for this.menu() != true {
		}
		// 根据不同模式处理不同业务
		switch this.flag {
		case 1:
			// 公聊
			this.PublicChat()
		case 2:
			//私聊
			this.PrivateChat()
		case 3:
			// 改名
			this.UpdateName()
		}
	}
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

	// 单独开一个go程处理go回执的消息，不影响Run
	go client.DealResponse()

	// 启动客户端的业务
	client.Run()
}
