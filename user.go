package main

import "net"

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

// 用户发消息
func (this *User) DoMessage(msg string) {
	this.server.BoardCast(this, msg)
}

// 监听当前user channel的方法，一单有消息，就直接发送给对端客户端
// 这里是从服务端接收到消息，打印到客户端
func (this *User) ListerMessage() {
	for {
		msg := <-this.C

		this.conn.Write([]byte(msg + "\n"))
	}
}
