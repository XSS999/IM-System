package main

import "net"

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn // 跟客户端去通信的连接
}

//创建一个用户的API
func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,
	}

	//启动监听当前user channel消息的goroutine
	go user.ListerMessage()

	return user
}

// 监听当前user channel的方法，一单有消息，就直接发送给对端客户端
// 这里是从服务端接收到消息，打印到客户端
func (this *User) ListerMessage() {
	for {
		msg := <-this.C

		this.conn.Write([]byte(msg + "\n"))
	}
}
