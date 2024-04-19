package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
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

func main() {
	fmt.Println("start a file transfer server")
	fts := new(FileTransService)
	rpc.Register(fts)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("listen error", err)
	}
	http.Serve(l, nil)
}
