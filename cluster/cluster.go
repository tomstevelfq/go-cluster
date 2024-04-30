package cluster

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

// the relationship between Cluster and ClusterNode is that the Cluster object abstractly encapsulates a cluster while a
// ClusterNode refers to a specific node within the cluster system
// there is a cluster process running in each host, while the communication between local host and others relies on
// rpc methods which provided by ClusterNode object
type MTYPE int32

const (
	STOP    MTYPE = 1
	RUNNING MTYPE = 2
	CALL    MTYPE = 3
)

type RpcFunc func(name string, args *interface{}, reply *interface{})

type Message struct {
	Mid           MTYPE
	TrigName      string //func name which need to call
	SourceAddress string
	DestAddress   string
}

type ClusterNode struct {
	Name      string
	address   string
	port      string
	State     bool
	rpcDone   chan *rpc.Call //store rpc finished message
	rpcServer *rpc.Server
	mesChan   chan *Message
}

type Cluster struct {
	state    bool //running or stopped
	Nodes    []ClusterNode
	NodesMap map[string]*ClusterNode
}

type Calculate struct{}

type CalArg struct {
	l int
	r int
}

func (cal *Calculate) Sum(arg *CalArg, reply *int) error {
	ret := 0
	l := arg.l
	r := arg.r
	for i := l; i <= r; i++ {
		ret += i
	}
	*reply = ret
	return nil
}

func GetHostAddr() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatal("failed to get host ip")
	}
	res := ""
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue // 如果无法转换，则忽略该地址
		}

		// 判断地址是否为回环地址和 IPv4 地址
		if ipNet.IP.IsLoopback() || ipNet.IP.To4() == nil {
			continue // 如果是回环地址或者不是 IPv4 地址，则忽略该地址
		}
		res = ipNet.IP.String()
	}
	return res
}

// 初始化ClusterNode
func InitNode(name string) *ClusterNode {
	addr := GetHostAddr()
	if addr == "" {
		log.Fatal("ipv4 addr is not found")
	}

	server := rpc.NewServer()
	server.Register(new(Calculate))

	return &ClusterNode{
		Name:      name,
		address:   addr,
		State:     false,
		rpcDone:   make(chan *rpc.Call, 10),
		rpcServer: rpc.NewServer(),
		mesChan:   make(chan *Message, 10),
	}
}

// describe running details of ClusterNode
func (node *ClusterNode) Run(port string) {
	go func() { // start rpc listener
		node.rpcServer.HandleHTTP("/rpc", "/debug/rpc")
		listener, err := net.Listen("tcp", node.address+":"+node.port)
		if err != nil {
			panic(err)
		}
		http.Serve(listener, nil)
	}()

	for { // start channel switch
		select {
		case <-node.rpcDone:
			fmt.Println("one rpc method finished")
		case mes := <-node.mesChan:
			fmt.Println("mes is coming", mes)
		}
	}
}

func (node *ClusterNode) RegisterObj(obj *interface{}) {
	node.rpcServer.Register(obj)
}

// 消息存入通道
func (node *ClusterNode) SendMessage(mes *Message) {
	node.mesChan<-
}

// 消息转rpc
func (node *ClusterNode) MessageToRpc(mes *Message) {

}

// 完成处理
func (node *ClusterNode) ProcessDone(done *rpc.Call) {

}

func (node *ClusterNode) MessageToNode(otherNode *ClusterNode, mes *Message) {

}

func Test() {
	fmt.Println("cluster")
}
