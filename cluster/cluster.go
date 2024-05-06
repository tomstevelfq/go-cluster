package cluster

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strings"
	"sync"
)

// the relationship between Cluster and ClientNode is that the Cluster object abstractly encapsulates a cluster while a
// ClientNode refers to a specific node within the cluster system
// there is a cluster process running in each host, while the communication between local host and others relies on
// rpc methods which provided by ClientNode object
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

type ClientNode struct {
	mutex   sync.Mutex
	Name    string
	address string
	State   bool
	mesDone chan *Message
	Client  *rpc.Client
}

type Cluster struct {
	mutex    sync.Mutex
	State    bool //running or stopped
	Nodes    []*ClientNode
	NodesMap map[string]*ClientNode
	addrList []string
	server   *rpc.Server
	addr     string
	Master   bool //is master node or not
}

type Calculate struct{}

type CalArg struct {
	l     int
	r     int
	clust *Cluster
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

func (cal *Calculate) DoCalculate(arg *CalArg, reply *int) error {
	//是主节点，直接开始计算
	if arg.clust.Master {
		*reply = arg.clust.CalCulateSum(arg.l, arg.r)
	} else {
		//否则转移至主节点
		masterNode := arg.clust.GetMasterNode()
		done := masterNode.CallAsync("Calculate.DoCalculate", arg, reply)
		<-done
	}
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

// 初始化ClientNode
func InitNode(name string, addr string) *ClientNode {
	cli, err := rpc.DialHTTP("tcp", addr)
	if err != nil {
		panic(err)
	}

	return &ClientNode{
		Name:    name,
		address: addr,
		State:   false,
		Client:  cli,
		mesDone: make(chan *Message, 10),
	}
}

// describe running details of ClientNode
func (node *ClientNode) Run() {
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

func (node *ClientNode) Stop() {
	node.mutex.Lock()
	node.State = false
	node.mutex.Unlock()
}

// 异步调用，返回一个通道done
func (node *ClientNode) CallAsync(method string, arg interface{}, reply interface{}) chan *rpc.Call {
	if node.Client == nil {
		log.Fatal("node client is nil")
	}

	done := make(chan *rpc.Call)
	node.Client.Go(method, arg, reply, done)
	return done
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

	if port == "9999" {
		clust.Master = true
	}
	clust.State = false
	clust.addr = addr
	clust.server = rpc.NewServer()
	clust.RegisterObj(new(Calculate))
	return clust
}

func (clust *Cluster) PingAdd() {
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

	for _, addr := range addrs {
		client, err := rpc.DialHTTP("tcp", addr)
		if err != nil {
			fmt.Println("client connected error")
			continue
		}
		defer client.Close()
		cli := &ClientNode{
			State:   false,
			Name:    "client-" + addr,
			address: addr,
			Client:  client,
			mesDone: make(chan *Message, 10),
		}
		clust.Nodes = append(clust.Nodes, cli)
	}

	for _, node := range clust.Nodes {
		go node.Run() //start all of the cluster nodes
	}
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

func (clust *Cluster) RegisterObj(obj interface{}) {
	clust.server.Register(obj)
}

// 消息存入通道
func (host *Cluster) MessageToNode(mes *Message, node *ClientNode) {
	done := node.CallAsync(mes.FuncName, mes.args, mes.reply)
	go func() {
		<-done
		node.mesDone <- mes
	}()
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

// get the master node
func (clust *Cluster) GetMasterNode() *ClientNode {
	res := (*ClientNode)(nil)
	for _, node := range clust.Nodes {
		if strings.Contains(node.address, "9999") {
			res = node
		}
	}
	return res
}

// test code for Cluster, it aims to calculate sum for range of numa to numb with the use of distributed nodes
func (clust *Cluster) CalCulateSum(numa int, numb int) int {
	if len(clust.Nodes) == 0 {
		log.Fatal("no nodes for calculating sum")
	}
	div := math.Ceil(float64(numb-numa) / float64(len(clust.Nodes)))

	left := 0
	right := int(div)
	var wg sync.WaitGroup
	res := 0
	for _, node := range clust.Nodes {
		reply := 0
		done := node.CallAsync("Calculate.Sum", &CalArg{left, min(right, numb), nil}, &reply)
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-done
			res = res + reply
		}()
	}
	wg.Wait()
	return res
}

func (clust *Cluster) CallMasterAsync(method string, arg interface{}, reply interface{}) chan *rpc.Call {
	masterNode := clust.GetMasterNode()
	done := masterNode.CallAsync(method, arg, reply)
	return done
}

func (clust *Cluster) CallMaster(method string, arg interface{}, reply interface{}) {
	<-clust.CallMasterAsync(method, arg, reply)
}
