package cluster

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"strings"
	"sync"
	"time"
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
	mutex    sync.Mutex
	Name     string
	address  string
	State    bool
	mesDone  chan *Message
	Client   *rpc.Client
	IsClosed bool
}

type Cluster struct {
	mutex     sync.Mutex
	NodesMtx  sync.Mutex
	State     bool //running or stopped
	Nodes     []*ClientNode
	NodesMap  map[string]*ClientNode
	addrList  []string
	server    *rpc.Server
	addr      string
	Master    bool //is master node or not
	DebugMode bool
}

type Calculate struct {
	clust *Cluster
}

type CalArg struct {
	l int
	r int
}

type ClustArg struct {
	L int
	R int
}

func MessageCallback(mes *Message) {
	fmt.Printf("message callback for message: %d, called func name: %s", mes.Mid, mes.FuncName)
}

func (cal *Calculate) Sum(arg *ClustArg, reply *int) error {
	time.Sleep(time.Second * 3)
	ret := 0
	l := arg.L
	r := arg.R
	for i := l; i <= r; i++ {
		ret += i
	}
	*reply = ret
	log.Println("calculate.sum finished", *reply, arg.L, arg.R)
	return nil
}

func (cal *Calculate) DoCalculate(arg *CalArg, reply *int) error {
	//是主节点，直接开始计算
	if cal.clust.Master {
		*reply = cal.clust.CalCulateSum(arg.l, arg.r)
	} else {
		//否则转移至主节点
		masterNode := cal.clust.GetMasterNode()
		done := masterNode.CallAsync("Calculate.DoCalculate", arg, reply)
		<-done
	}
	return nil
}

func (cal *Calculate) ClusterCalculateRpc(arg *ClustArg, reply *int) error {
	//是主节点，直接开始计算
	if cal.clust.Master {
		*reply = cal.clust.CalCulateSum(arg.L, arg.R)
	} else {
		//否则转移至主节点
		masterNode := cal.clust.GetMasterNode()
		done := masterNode.CallAsync("Calculate.ClusterCalculateRpc", arg, reply)
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

		log.Println("addr", ipNet.IP.String())
		// 判断地址是否为回环地址和 IPv4 地址
		if ipNet.IP.IsLoopback() || ipNet.IP.To4() == nil {
			continue // 如果是回环地址或者不是 IPv4 地址，则忽略该地址
		}
		res = ipNet.IP.String()
		if res[:3] == "192" || res[:3] == "172" {
			break
		}
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
	log.Println("node", node.address, "callasync method:", method)
	if node.Client == nil {
		log.Fatal("node client is nil")
	}

	done := make(chan *rpc.Call, 10)
	node.Client.Go(method, arg, reply, done)
	return done
}

func (node *ClientNode) Call(method string, arg interface{}, reply interface{}) error {
	log.Println("node", node.address, "calla method:", method, node.Client)
	if node.Client == nil {
		log.Fatal("node client is nil")
	}

	return node.Client.Call(method, arg, reply)
}

func (node *ClientNode) Reconnect() error {
	log.Println("cli node reconnnect", node.address)
	client, err := rpc.Dial("tcp", node.address)
	if err != nil {
		log.Println("client connected error", node.address)
		return err
	}

	log.Println("client connected success", node.address)
	node.Client = client
	return nil
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
	cal := new(Calculate)
	cal.clust = clust
	clust.RegisterObj(cal)
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
		addr = "192.168.3.11:" + addr
		clust.DebugLog("cluster", clust.addr, "--pingadd", addr)
		client, err := rpc.Dial("tcp", addr)
		if err != nil {
			log.Println("client connected error", clust.addr, "---", addr)
			continue
		}

		log.Println("client connected success", clust.addr, "---", addr, client)

		cli := &ClientNode{
			State:    false,
			Name:     "client-" + addr,
			address:  addr,
			Client:   client,
			mesDone:  make(chan *Message, 10),
			IsClosed: false,
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
	// 创建一个 TCP 监听器
	log.Println("clust.addr", clust.addr)
	listener, err := net.Listen("tcp", clust.addr)
	if err != nil {
		fmt.Println("Failed to listen:", err)
		return
	}
	defer listener.Close()
	log.Println("RPC server is running on", clust.addr)

	go func() {
		time.Sleep(time.Second * 5)
		clust.PingAdd()
	}()

	// 接受连接并为每个连接启动一个 goroutine 来处理请求
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection:", err)
			continue
		}
		go clust.server.ServeConn(conn)
	}
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

func (host *Cluster) EnableDebug() {
	host.DebugMode = true
}

func (host *Cluster) DisableDebug() {
	host.DebugMode = false
}

// only print in debug mode
func (host *Cluster) DebugLog(args ...interface{}) {
	if host.DebugMode {
		log.Println(args...)
	}
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

func (clust *Cluster) GetRandomNode() *ClientNode {
	l := len(clust.Nodes)
	p := rand.Intn(l)
	log.Println("get random cient node", clust.Nodes[p])
	return clust.Nodes[p]
}

func (clust *Cluster) RemoveNode(node *ClientNode) {
	log.Println("cluster", clust.addr, "remove node", node.address)
	for ind, n := range clust.Nodes {
		if node == n {
			clust.NodesMtx.Lock()
			clust.Nodes = append(clust.Nodes[:ind], clust.Nodes[ind+1:]...)
			clust.NodesMtx.Unlock()
			return
		}
	}
}

// This function is guaranteed to be successfully delivered
func (clust *Cluster) CallAsyncGuarantee(done chan *rpc.Call, dep *int, cli *ClientNode, ind int) {
	if *dep == 10 {
		return
	}
	(*dep)++
	call := <-done
	if call.Error != nil {
		log.Println("calculate error", call.Error)
		if cli.IsClosed {
			err := cli.Reconnect()
			if err != nil {
				clust.RemoveNode(cli)
			}
		}
		node := clust.GetRandomNode()
		done = node.CallAsync("Calculate.Sum", call.Args, call.Reply)
		clust.CallAsyncGuarantee(done, dep, node, ind)
	}
}

// test code for Cluster, it aims to calculate sum for range of numa to numb with the use of distributed nodes
func (clust *Cluster) CalCulateSum(numa int, numb int) int {
	fmt.Println("cluster", clust.addr, "--start calculate sum")
	if len(clust.Nodes) == 0 {
		log.Fatal("no nodes for calculating sum")
	}
	div := int(math.Ceil(float64(numb-numa+1) / float64(len(clust.Nodes))))
	left := numa
	right := div
	var wg sync.WaitGroup
	res := 0
	for ind, node := range clust.Nodes {
		reply := 0
		log.Println("cluster", clust.addr, "--call async calculate", node.address)
		done := node.CallAsync("Calculate.Sum", &ClustArg{left, min(right, numb)}, &reply)
		left += div
		right += div
		wg.Add(1)
		go func() {
			defer wg.Done()
			dep := 0
			clust.CallAsyncGuarantee(done, &dep, node, ind)
			log.Println("calculate over", reply, dep)
			res = res + reply
		}()
	}
	wg.Wait()
	return res
}

func (clust *Cluster) CallMasterAsync(method string, arg interface{}, reply interface{}) chan *rpc.Call {
	masterNode := clust.GetMasterNode()
	log.Println("cluster Node", clust.addr, "--call master Node", masterNode.address)
	done := masterNode.CallAsync(method, arg, reply)
	return done
}

func (clust *Cluster) CallMaster(method string, arg interface{}, reply interface{}) {
	<-clust.CallMasterAsync(method, arg, reply)
}

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
