package main

import (
	"log"
	"os"
	"runtime/debug"
	"time"

	"example.com/cluster"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Println("panic")
			debug.PrintStack()
		}
	}()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	file, err := os.OpenFile("test.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Println("cannot open the log file:", err)
	}
	defer file.Close()
	log.SetOutput(file)
	addrs := cluster.GetAddrs()
	var clustList []*cluster.Cluster
	for _, port := range addrs {
		clust := cluster.InitCluster(port)
		clust.EnableDebug() //enable debug log
		clustList = append(clustList, clust)
		go clust.Run()
	}
	time.Sleep(time.Second * 6000)
}
