package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int //当前用户模式
}

func NewClient(serverIp string, serverPort int) *Client {
	//创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}
	//链接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}
	client.conn = conn
	//返回对象
	return client
}

//处理server回应的消息，显示到终端中
func (client *Client) DealResponse() {
	// 一旦client.conn有数据，就直接copy到stout标准输出
	// 永久阻塞，等同于下面for循环

	io.Copy(os.Stdout, client.conn)

}

func (client *Client) menu() bool {
	var flag int
	fmt.Println("1.Broad cast")
	fmt.Println("2.Private chat")
	fmt.Println("3.Update user name")
	fmt.Println("0.Exit")

	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println(">>>>>>>>>>Enter1,2or3<<<<<<<<<<<")
		return false
	}
}

func (client *Client) PublicChat() {
	//提示用户输入消息
	var chatMsg string
	fmt.Println("Please Enter the content,exit to exit")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn.Write  err: ", err)
				break
			}
		}
		chatMsg = ""
		fmt.Println("Please Enter the content,exit to exit")
		fmt.Scanln(&chatMsg)
	}
	//发给服务器
}

func (client *Client) SelectUsers() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn Write err: ", err)
		return
	}
}
func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string
	client.SelectUsers()
	fmt.Println("Enter the username you want to chat,exit to exit:")
	fmt.Scanln(&remoteName)

	for chatMsg != "exit" {
		fmt.Println("Enter the content of char,exit to exit:")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			if len(chatMsg) != 0 {

				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n\n"

				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn.Write  err: ", err)
					break
				}
			}
			chatMsg = "" //清空之前的消息
			fmt.Println("Enter the content of char,exit to exit:")
			fmt.Scanln(&chatMsg)
		}
		client.SelectUsers()
		fmt.Println("Enter the username you want to chat,exit to exit:")
		fmt.Scanln(&remoteName)

	}
}

func (client *Client) UpdateName() bool {
	fmt.Println(">>>>>>>>Please enter the username: ")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err: ", err)
		return false
	}
	return true
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {

		}
		//根据不同模式处理不同业务
		switch client.flag {
		case 1:
			//公聊模式
			fmt.Println("Broad cast choosing...")
			client.PublicChat()
			break
		case 2:
			//私聊模式
			fmt.Println("Private chat choosing...")
			client.PrivateChat()
			break

		case 3:
			//更新用户名
			fmt.Println("User name updating...")
			client.UpdateName()
			break

		}

	}
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "Setting ip address (127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "Port address (8888)")
}

func main() {
	//命令行解析
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>>>>server connect failed<<<<<<<<")
		return
	}
	go client.DealResponse()
	fmt.Println(">>>>>>>>connect server success<<<<<<<<<<<<<<")

	//启动客户端业务
	client.Run()
}
