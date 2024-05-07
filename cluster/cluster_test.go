package cluster

import (
	"log"
	"os"
	"runtime/debug"
	"testing"
	"time"
)

func EnableLog() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	file, err := os.OpenFile("test.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	log.SetOutput(file)
}

func TestCluster(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("panic")
			debug.PrintStack()
		}
	}()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	file, err := os.OpenFile("test.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		t.Error("cannot open the log file:", err)
	}
	defer file.Close()
	log.SetOutput(file)
	addrs := GetAddrs()
	var clustList []*Cluster
	for _, port := range addrs {
		clust := InitCluster(port)
		clust.EnableDebug() //enable debug log
		clustList = append(clustList, clust)
		go clust.Run()
	}
	time.Sleep(time.Second * 100)
}

func TestGetAddr(t *testing.T) {
	EnableLog()
	addr := GetHostAddr()
	log.Println("addr-test", addr)
}

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
	addrs := GetAddrs()
	var clustList []*Cluster
	for _, port := range addrs {
		clust := InitCluster(port)
		clust.EnableDebug() //enable debug log
		clustList = append(clustList, clust)
		go clust.Run()
	}
	time.Sleep(time.Second * 1000)
}
