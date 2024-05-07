package main

import (
	"fmt"
	"net/rpc"
)

type ClustArg struct {
	L int
	R int
}

func main() {
	// 连接到 RPC 服务器
	client, err := rpc.Dial("tcp", "192.168.3.11:9999")
	if err != nil {
		fmt.Println("Failed to connect:", err)
		return
	}
	defer client.Close()

	// 调用远程方法
	var reply int
	err = client.Call("Calculate.Sum", &ClustArg{1, 1000}, &reply)
	if err != nil {
		fmt.Println("RPC call failed:", err)
		return
	}

	// 打印结果
	fmt.Println("Server response:", reply)
}
