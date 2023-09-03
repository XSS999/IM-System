package main

import (
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn // 跟客户端去通信的连接

	server *Server //当前在的服务端
}

//创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	//启动监听当前user channel消息的goroutine
	go user.ListerMessage()

	return user
}

// 用户上线
func (this *User) Online() {
	//把用户加入到OnlineMap中
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	//广播当前用户上线消息
	this.server.BoardCast(this, "已上线")
}

// 用户下线
func (this *User) Offline() {
	//把用户从OnlineMap中 删除
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	//广播当前用户上线消息
	this.server.BoardCast(this, "已下线")
}

func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

// 用户发消息
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		// 查询当前在线用户都有哪些
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线...\n"
			this.SendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// 重命名
		newName := msg[7:]

		// 判断name是否存在
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.SendMsg("当前用户名已被使用")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.Name = newName
			this.server.OnlineMap[this.Name] = this
			this.server.mapLock.Unlock()

			this.SendMsg("您已更新用户名：" + newName + "\n")
		}

	} else if len(msg) > 4 && msg[:3] == "to|" { //私聊
		// 1 取到remoteName
		remoteUser := strings.Split(msg, "|")[1]
		if remoteUser == "" {
			this.SendMsg("请输入正确的格式：\"to|张三|消息\"")
		}
		// 2 根据用户名，获取remoteUser
		user, ok := this.server.OnlineMap[remoteUser]
		if !ok {
			this.SendMsg(remoteUser + "is null")
		}
		// 3 将消息内容发送给remoteUse
		context := strings.Split(msg, "|")[2]
		user.SendMsg(this.Name + "对您说：" + context)
	} else {
		this.server.BoardCast(this, msg)
	}
}

// 监听当前user channel的方法，一单有消息，就直接发送给对端客户端
// 这里是从服务端接收到消息，打印到客户端
func (this *User) ListerMessage() {
	for {
		msg := <-this.C

		this.conn.Write([]byte(msg + "\n"))
	}
}
