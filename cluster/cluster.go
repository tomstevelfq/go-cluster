package cluster

import (
	"fmt"
	"net/rpc"
)

type ClusterNode struct {
	name      string
	address   string
	connected bool
	rpcMes    chan *rpc.Call
}

type Cluster struct {
	state bool
}

func Test() {
	fmt.Println("cluster")
}
