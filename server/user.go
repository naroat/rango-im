package main

import (
	"fmt"
	"io"
	"net"
	"strings"
)

//定义结构体
type User struct {
	Name string
	Addr string
	//用户
	Channel chan string
	//连接流
	conn   net.Conn
	server *Server
}

//创建用户
func NewUser(conn net.Conn, server *Server) *User {
	//获取地址
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:    userAddr,
		Addr:    userAddr,
		Channel: make(chan string),
		conn:    conn,
		server:  server,
	}

	//启动消息监听服务
	go user.ListenMessage()

	return user
}

//监听当前用户channel,收到消息时及时通知客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.Channel
		//向连接写入数据
		this.conn.Write([]byte(msg + "\n"))
	}
}

//上线
func (this *User) Online() {
	this.server.maplock.Lock()
	//当前用户添加到在线集合
	this.server.OnlineMap[this.Name] = this
	this.server.maplock.Unlock()
	//上线消息广播
	this.DoMessage("上线...")
}

//下线
func (this *User) Offline() {
	this.server.maplock.Lock()
	//用户列表中删除该用户
	delete(this.server.OnlineMap, this.Name)
	this.server.maplock.Unlock()
	//下线消息广播
	this.DoMessage("下线...")
}

//处理数据
func (this *User) DoMessage(msg string) {
	fmt.Println(msg)
	if msg == "list" {
		//查看在线用户列表
		for _, user := range this.server.OnlineMap {
			onlineMsg := "当前 [" + user.Name + "] 在线 ...\n"

			this.conn.Write([]byte(onlineMsg))
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		//私聊
		newMsg := strings.Split(msg, "|")
		fmt.Println(newMsg)
		remoteName := newMsg[1]
		if remoteName == "" {
			this.conn.Write([]byte("消息格式错误\n"))
			return
		}

		remoteUser, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.conn.Write([]byte("用户不存在或已下线！\n"))
			return
		}

		if newMsg[2] == "" {
			this.conn.Write([]byte("发送内容不能为空！\n"))
			return
		}
		remoteUser.conn.Write([]byte(this.Name + "对您说：" + newMsg[2] + "\n"))
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//更新名称
		newName := msg[7:]
		this.server.maplock.Lock()
		delete(this.server.OnlineMap, this.Name)
		this.Name = newName
		this.server.OnlineMap[newName] = this
		this.server.maplock.Unlock()
	} else {
		this.server.Broadcast(this, msg)
	}
}

//用户消息广播
func (user *User) UserBoradcast(conn net.Conn, isLive chan bool) {
	userMsg := make([]byte, 4096)
	for {
		//读取用户输入数据
		n, err := conn.Read(userMsg)
		if n == 0 {
			//掉线/下线
			user.Offline()
			return
		}

		if err != nil && err != io.EOF {
			fmt.Println(err)
			return
		}

		//取消结尾"\n"
		msg := string(userMsg[:n-1])
		//注意：如果没有去掉"\n"这一步会导致转换字符串后的数据还携带切片的长度，导致字符串比较时一直false
		// msg := string(userMsg)	//错误例子
		user.DoMessage(msg)

		//用户有发送消息时表示是活跃用户
		isLive <- true
	}
}
