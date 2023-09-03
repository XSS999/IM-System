package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

/**
首先，必须要知道 咱们写这个项目是为了完成什么功能
只是写Server去监听客户端，建立链接
1. 把main和server分开写：
	为了模块基础划分，也就是解耦
2. conn是什么意思，套接字又是什么
	conn是指connection，通常表示一个网络连接，可以是客户端与服务端之间的连接通道
	套接字（socket）是一种在网络编程中使用的抽象概念，用于在不同计算机之间进行数据传输。
	套接字可以用来创建、维护和关闭网络连接，使计算机之间能够进行数据交换。

3. 通过fmt.Sprint()就可以拼接成字符串

4. 目前打印日志是通过fmt.Println，golang中常用什么打印日志
	log.Println()

5. 理解一下go关键字
	使用 go 关键字启动一个新的 Go 协程（goroutine），以便异步处理每个客户端连接。
	这样可以同时处理多个客户端连接而不阻塞主循环。

6. nc 127.0.0.1 8888
	netcat 用于创建和管理网络连接，测试网络服务，传输文件，以及进行端口扫描

7. io.EOF是什么

8. 写代码的时候，如何写出来易于拓展的代码
	模块要划分清楚，用户上、下线和用户处理消息都应该放到User模块中，写成单独的方法。不要写到其他地方，比如Server
*/

type Server struct {
	Ip   string
	Port int

	//在线用户列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	//消息广播的channel
	Message chan string
}

//创建一个Server的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// 监听Message广播消息channel的goroutine，一但有消息就发送给全部的在线User
func (this *Server) ListerMessage() {
	for {
		//取到消息
		msg := <-this.Message

		//把消息发送给所有的在线User
		this.mapLock.Lock()
		//TODO: 当用户数量很大时，循环效率肯定低，怎么解决
		for _, cli := range this.OnlineMap {
			cli.C <- msg
			//这时，客户端的channel就已经有消息了，User中的ListerMessage就会监听到这个channel，然后打印
		}
		this.mapLock.Unlock()
	}
}

// 广播消息，
func (this *Server) BoardCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	this.Message <- sendMsg
	//这里还需要一个方法，不断的从Message中读数据，如果有数据，就广播给所有在线user
}

func (this *Server) Handler(conn net.Conn) {
	//log.Println("log:net listen err:")
	//fmt.Println("链接建立成功")

	//当前客户端 --> User
	user := NewUser(conn)

	//把用户加入到OnlineMap中
	this.mapLock.Lock()
	this.OnlineMap[user.Name] = user
	this.mapLock.Unlock()

	//广播当前用户上线消息
	this.BoardCast(user, "已上线")

	// 每开一个客户端之后，就会开一个goroutine去监听当前客户端有没有发送消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)

			//下线逻辑
			if n == 0 {
				this.BoardCast(user, "已下线")
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			//提取当前用户的消息(去掉\n)
			msg := string(buf[:n-1])

			//进行广播
			this.BoardCast(user, msg)
		}
	}()

	//当前handler阻塞
	//select {}
}

func (this *Server) Start() {
	//socket listen
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net listen err:", err)
	}
	//listen close
	defer listen.Close()
	//启动监听Message的goroutine
	go this.ListerMessage()

	//不断监听有没有新客户端
	for {
		//accept
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("listen accept err:", err)
			continue
		}

		//do handler
		go this.Handler(conn)
	}
}
