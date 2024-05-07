package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"time"

	"example.com/global"
)

type FileTransService struct{}

func (fts *FileTransService) UploadFile(req global.UploadRequest, res *global.UploadResponse) error {
	fileContent := req.FileContent
	names := strings.Split(req.FileName, ".")
	timestr := strconv.FormatInt(time.Now().Unix(), 10)
	name := timestr + "." + names[len(names)-1]
	err := os.WriteFile(name, fileContent, 0644)
	if err != nil {
		fmt.Println("file write error")
		*res = global.UploadResponse{Success: false}
		return err
	}
	*res = global.UploadResponse{Success: true}
	return nil
}

type ClustArg struct {
	L int
	R int
}

func (cal *FileTransService) Sum(arg *ClustArg, reply *int) error {
	ret := 0
	l := arg.L
	r := arg.R
	for i := l; i <= r; i++ {
		ret += i
	}
	*reply = ret
	return nil
}

func main() {
	fmt.Println("start a file transfer server")
	fts := new(FileTransService)
	rpc.Register(fts)

	listener, err := net.Listen("tcp", "192.168.3.11:9999")
	if err != nil {
		fmt.Println("Failed to listen:", err)
		return
	}
	defer listener.Close()
	log.Println("RPC server is running on", "192.168.3.11:9999")

	// 接受连接并为每个连接启动一个 goroutine 来处理请求
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection:", err)
			continue
		}
		go rpc.ServeConn(conn)
	}
}
