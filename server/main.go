package main

func main() {
	//创建Server
	server := NewServer("127.0.0.1", 7701)
	server.Start()
}
