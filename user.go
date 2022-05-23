package main

import (
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

//创建一个用户API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,

		server: server,
	}
	//启动监听当前user channel消息的goroutine
	go user.ListenMessage()
	return user
}

//用户上线业务
func (this *User) Online() {
	//用户上线，将用户加入到onlinemap中
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()
	//广播当前用户上线消息
	this.server.BroadCast(this, "Online")
}

//用户下线业务
func (this *User) Offline() {
	//用户下线，将用户从onlinemap中删除
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()
	//广播当前用户上线消息
	this.server.BroadCast(this, "Offline")
}

func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

//用户处理消息业务
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		//查询当前在线用户
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "online...\n"
			this.SendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()

	} else if len(msg) > 7 && msg[:7] == "rename|" {
		newName := strings.Split(msg, "|")[1]

		//判断名字是否存在
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.SendMsg("Exist username\n")

		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.SendMsg("Username change to " + newName + " successfully\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		//1.获取对方用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.SendMsg("Message format is wrong")
			return
		}
		//2.根据用户名得到对方user对象
		remoteUser, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.SendMsg("This username doesn't exist\n")
			return
		}
		//3.获取消息内容，通过对方的user对象将消息发送过去
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.SendMsg("content is empty \n")
			return
		}
		remoteUser.SendMsg("[" + this.Name + "]" + " to you " + content)
	} else {
		this.server.BroadCast(this, msg)
	}

}

//监听当前user channel的方法，一旦有消息,就发送给对端客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C

		this.conn.Write([]byte(msg + "\n"))
	}

}
