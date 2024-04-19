package main

import (
	"fmt"
	"log"
	"net/rpc"
	"os"

	"example.com/global"
)

func client() {
	fmt.Println("start a file transfer client")
	client, err := rpc.DialHTTP("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("client connect error", err)
	}
	fileContent, err := os.ReadFile("test.png")
	if err != nil {
		log.Fatal("read file error")
	}
	req := global.UploadRequest{
		FileName:    "test.png",
		FileContent: fileContent,
	}
	var res global.UploadResponse
	err = client.Call("FileTransService.UploadFile", req, &res)
	if err != nil {
		log.Fatal("cal file Upload error:", err)
	}
	if res.Success {
		fmt.Println("file upload success")
	} else {
		fmt.Println("file upload error")
	}
}

func main() {
	client()
}
