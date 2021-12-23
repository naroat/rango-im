package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// 创建结构体
type Server struct {
	Ip   string
	Port int
	//在线用户列表
	OnlineMap map[string]*User
	//读写锁定义，防止数据读写冲突
	maplock sync.RWMutex
	//消息广播channel
	Message chan string
}

//创建server接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server
}

//启动服务
func (this *Server) Start() {
	//监听服务
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println(err)
		return
	}

	//关闭服务
	defer listener.Close()

	//运行监听消息广播
	go this.ListenMessage()

	for {
		//接收客户连接
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		//处理请求
		go this.Handler(conn)
	}
}

//处理请求
func (this *Server) Handler(conn net.Conn) {
	fmt.Println("connect success!")

	//这个channel用于监听用户是否活跃
	isLive := make(chan bool)

	//创建用户
	user := NewUser(conn, this)

	//用户上线
	user.Online()

	//用户消息广播
	go user.UserBoradcast(conn, isLive)

	for {
		//select会随机执行case，解决多线程饥饿问题, 如果没有case可运行时会阻塞；
		//tips: 需要阻塞的场景可以直接使用select{}来实现
		select {
		case <-isLive:
			//活跃
			//不需要做任何事，让其执行另一个条件重置定时器
		case <-time.After(time.Second * 180):
			//重置定时器的方法就是重新执行一遍定时器，而上面isLive不满足条件时，也会执行该case条件重置定时器
			//超时下线问题提示
			user.DoMessage("超时下线！")

			//处理报错：panic: send on closed channel；阻塞2秒，让用户下线goroutine先执行
			time.Sleep(time.Second * 2)

			//关闭用户channel
			close(user.Channel)

			//关闭连接
			conn.Close()

			//退出当前handle
			return
		}
	}
}

//广播
func (this *Server) Broadcast(user *User, msg string) {
	//组合消息体
	sendMsg := "[" + user.Name + "] " + msg
	//消息添加到Message channel
	this.Message <- sendMsg
}

//监听消息广播，有消息时就发送给所有在线的user
func (this *Server) ListenMessage() {
	for {
		//接收Message channel
		msg := <-this.Message
		this.maplock.Lock()
		//循环添加到用户消息channel
		for _, userCli := range this.OnlineMap {
			userCli.Channel <- msg
		}
		this.maplock.Unlock()
	}
}
