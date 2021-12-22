package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

//定义结构体
type Client struct {
	//服务器ip
	ServerIp string
	//服务器端口
	ServerPort int
	//连接流
	conn net.Conn
}

//服务器ip
var ip string

//服务器端口
var port int

//init函数会在每个包完成初始化后自动执行，优先级高于main
func init() {
	//关于flag.string,int的参数解释
	//第一个参数：存储该命令参数值得地址
	//第二个参数：命令行参数的名称
	//第三个参数：默认值
	//第四个参数：描述，help命令时会显示
	flag.StringVar(&ip, "ip", "127.0.0.1", "需要连接的ip地址")
	flag.IntVar(&port, "port", 7701, "需要连接的端口")
}

func main() {
	//解析命令:传参方式： clinet -ip 127.0.0.1 -port 7701
	flag.Parse()

	//创建Client
	client := NewClient(ip, port)
	if client == nil {
		fmt.Println(">>> connect fail...")
		return
	}
	fmt.Println(">>> connect success...")

	//处理服务端回调消息
	go client.DealResponse()

	//推迟关闭client.conn
	defer client.conn.Close()

	//执行操作
	client.Run()
}

//创建对象
func NewClient(ip string, port int) *Client {
	//创建client对象
	client := &Client{
		ServerIp:   ip,
		ServerPort: port,
	}
	//连接服务器
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		fmt.Println(">>> connect fail...")
		return nil
	}

	client.conn = conn

	//返回对象
	return client
}

//处理服务端回调消息
func (Client *Client) DealResponse() {
	//client.conn有数据时，直接copy到标准输出，永久阻塞
	io.Copy(os.Stdout, Client.conn)
}

//执行操作
func (client *Client) Run() {
	//消息
	var data string

	//操作说明
	fmt.Println(">>> 操作提示：")
	fmt.Println(">>> 直接输入发送全局消息")
	fmt.Println(">>> 通过`to|xxx|xxx`用户私聊， 比如： to|zhangsan|你好啊")
	fmt.Println(">>> 输入`list`在线用户")
	fmt.Println(">>> 输入`exit`退出")
	fmt.Println(">>> 请输入内容...")

	//fmt.Scanln()从标准输入扫描文本，遇到换行后才停止扫描
	fmt.Scanln(&data)

	for data != "exit" {
		if len(data) != 0 {
			//向服务端发送消息
			_, err := client.conn.Write([]byte(data + "\n"))
			if err != nil {
				fmt.Println("发送消息失败：", err)
				break
			}
		} else {
			fmt.Println("请输入内容...")
		}
		//初始化输入的值
		data = ""
		fmt.Scanln(&data)
	}
}
