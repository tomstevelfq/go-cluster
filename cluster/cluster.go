package cluster

import (
	"fmt"
	"net/rpc"
)

// the relationship between Cluster and ClusterNode is that the Cluster object abstractly encapsulates a cluster while a
// ClusterNode refers to a specific node within the cluster system
// there is a cluster process running in each host, while the communication between local host and others relies on
// rpc methods which provided by ClusterNode object
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
