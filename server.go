package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	//在线用户列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	//消息广播的channel
	Message chan string
}

//创建一个server接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

//监听Message广播消息channel的goroutine，一旦有消息就发送给全部在线User
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message
		//将msg发送给全部在线user
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg

		}
		this.mapLock.Unlock()
	}
}

func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

func (this *Server) Handler(conn net.Conn) {
	//当前链接的业务'
	//fmt.Println("链接建立成功")
	user := NewUser(conn, this)
	user.Online()
	//接受客户端传递发送的消息

	//监听用户是否活跃的channel
	isLive := make(chan bool)

	go func() {
		buf := make([]byte, 4096)

		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return

			}
			if err != nil && err != io.EOF {
				return
			}
			//提取用户消息（去除'\n'
			msg := string(buf[:n-1])

			//用户针对msg进行消息处理
			user.DoMessage(msg)

			//用户的任意消息，代表当前用户是活跃的
			isLive <- true
		}
	}()

	//当前handle阻塞
	for {
		select {
		case <-isLive:
			//当前用户是活跃的，应重置计时器
			//不做任何事情，为了激活select，更新下面的定时器
		case <-time.After(time.Second * 600):
			//已经超时
			//将当前User强制关闭
			user.SendMsg("You are offline compulsory")

			//撤销使用的资源
			close(user.C)
			//关闭连接
			conn.Close()
			//退出当前handler
			return //runtime.Goexit()也可以
		}
	}

}

//启动服务器
func (this *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	//close listener socket
	defer listener.Close()
	//启动监听message的goroutine
	go this.ListenMessager()
	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}
		//do handler
		go this.Handler(conn)
	}
}
