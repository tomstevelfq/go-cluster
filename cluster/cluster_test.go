package cluster

import (
	"encoding/json"
	"log"
	"os"
	"testing"
	"time"
)

func GetAddrs() []string {
	//build a sequence of addrs list for nodes, test if the ping is ok
	file, err := os.Open("addrs.json")
	if err != nil {
		log.Fatal("file open error")
	}
	defer file.Close()
	var addrs []string
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&addrs)
	if err != nil {
		log.Fatal("decode error")
	}
	return addrs
}

func TestCluster(t *testing.T) {
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
		clustList = append(clustList, clust)
		go clust.Run()
	}
	time.Sleep(time.Second * 100)
}
