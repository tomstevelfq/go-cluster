package cluster

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"sync"
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
	FuncName      string //func name which need to call
	SourceAddress string
	DestAddress   string
	args          interface{}
	reply         interface{}
}

type ClusterNode struct {
	mutex   sync.Mutex
	Name    string
	address string
	port    string
	State   bool
	mesDone chan *Message
	Client  *rpc.Client
}

type Cluster struct {
	mutex    sync.Mutex
	State    bool //running or stopped
	Nodes    []ClusterNode
	NodesMap map[string]*ClusterNode
	server   *rpc.Server
	addr     string
}

type Calculate struct{}

type CalArg struct {
	l int
	r int
}

func MessageCallback(mes *Message) {
	fmt.Printf("message callback for message: %d, called func name: %s", mes.Mid, mes.FuncName)
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
func InitNode(name string, addr string) *ClusterNode {
	cli, err := rpc.DialHTTP("tcp", addr)
	if err != nil {
		panic(err)
	}

	return &ClusterNode{
		Name:    name,
		address: addr,
		State:   false,
		Client:  cli,
		mesDone: make(chan *Message, 10),
	}
}

// describe running details of ClusterNode
func (node *ClusterNode) Run(port string) {
	node.mutex.Lock()
	node.State = true
	node.mutex.Unlock()
	for {
		if !node.State {
			break
		}

		msg := <-node.mesDone
		MessageCallback(msg)
	}
}

func (node *ClusterNode) Stop() {
	node.mutex.Lock()
	node.State = false
	node.mutex.Unlock()
}

// 异步调用，返回一个通道done
func (node *ClusterNode) CallAsync(addr string, method string, arg interface{}, reply interface{}) chan *rpc.Call {
	client, err := rpc.DialHTTP("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer client.Close()
	done := make(chan *rpc.Call)
	client.Go(method, arg, reply, done)
	return done
}

// 消息存入通道
func (node *ClusterNode) MessageToNode(mes *Message, other *ClusterNode) {
	done := node.CallAsync(other.address+":"+other.port, mes.FuncName, mes.args, mes.reply)
	go func() {
		<-done
		node.mesDone <- mes
	}()
}

func InitCluster(port string) *Cluster {
	addr := GetHostAddr()
	if addr == "" {
		log.Fatal("can't bind ip addr")
	}
	if port == "" {
		addr = addr + ":8080"
	} else {
		addr = addr + ":" + port
	}

	clust := new(Cluster)

	clust.State = false
	clust.addr = addr
	clust.server = rpc.NewServer()
	return clust
}

func (clust *Cluster) Run() {
	clust.mutex.Lock()
	clust.State = true
	clust.mutex.Unlock()
	clust.server.HandleHTTP("/rpc", "/debug/rpc")

	listener, err := net.Listen("tcp", clust.addr)
	if err != nil {
		panic(err)
	}
	http.Serve(listener, nil)
}

func (node *Cluster) RegisterObj(obj *interface{}) {
	node.server.Register(obj)
}

func Test() {
	fmt.Println("cluster")
}
