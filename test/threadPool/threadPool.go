package main

import (
	"fmt"
	"log"
	"reflect"
	"runtime/debug"
	"sync"
	"time"
)

type Job struct {
	id       int
	callback func()
}

type Func func(values ...interface{})

type Worker struct {
	JobChannel chan *Job
	id         int
	quit       chan *sync.WaitGroup
}

type Pool struct {
	Workers    []*Worker
	JobChannel chan *Job
	wg         sync.WaitGroup
	methods    map[string]interface{}
}

func AddNum(a int, b int) {
	a = a + b
	fmt.Printf("sum is %d\n", a)
}

func DoubleStr(a string) {
	a = a + a
	fmt.Printf("double string a is %s\n", a)
}

func WrapThreadFunc(f interface{}, args ...interface{}) func() {
	fmt.Println("para", args)
	vf := reflect.ValueOf(f)
	if vf.Kind() != reflect.Func {
		log.Fatal("the first para is not func type")
	}
	vargs := make([]reflect.Value, len(args))
	for idx, val := range args {
		vargs[idx] = reflect.ValueOf(val)
	}
	return func() {
		vf.Call(vargs)
	}
}

func (w *Worker) Process(job *Job) {
	fmt.Println("worker ", w.id, "is processing job ", job.id)
	if job.callback == nil {
		log.Fatal("callback is nil")
	}
	job.callback() //do the job
}

func (w *Worker) Stop(wg *sync.WaitGroup) {
	defer func() {
		if err := recover(); err != nil {
			log.Fatal("wrong Stop:", err)
		}
	}()
	defer wg.Done()
Loop:
	for {
		select {
		//process remained jobs
		case job := <-w.JobChannel:
			w.Process(job)
		default:
			break Loop
		}
	}
	fmt.Println("worker ", w.id, " stopped")
}

func (w *Worker) Run() {
	for {
		select {
		case job := <-w.JobChannel:
			if job == nil {
				fmt.Println("job is nil")
			}
			w.Process(job)
		case wg := <-w.quit:
			if wg == nil {
				fmt.Println("wg is null")
			}
			fmt.Println("worker ", w.id, " quit")
			w.Stop(wg)
			return
		}
	}
}

func NewWorker(id int, jobChannel chan *Job) *Worker {
	return &Worker{
		id:         id,
		JobChannel: jobChannel,
		quit:       make(chan *sync.WaitGroup),
	}
}

// Initialize a thread pool
func InitPool(workerSize int, jobSize int) *Pool {
	var pool Pool
	pool.methods = make(map[string]interface{})
	pool.JobChannel = make(chan *Job, jobSize)
	for i := 1; i <= workerSize; i++ {
		worker := NewWorker(i, pool.JobChannel)
		pool.Workers = append(pool.Workers, worker)
		pool.wg.Add(1)
		//if there is not go, it will be a deadlock
		go worker.Run()
	}
	return &pool
}

func (pool *Pool) Terminate() {
	//close(pool.JobChannel)
	//can not use close because it will conflicts with quit channel
	for i := 0; i < len(pool.Workers); i++ {
		pool.Workers[i].quit <- &pool.wg
	}
	// wait for all threads done
	pool.wg.Wait()
}

func (pool *Pool) AddJob(job *Job) {
	pool.JobChannel <- job
	fmt.Println("job ", job.id, " was allocated")
}

func (p *Pool) Register(name string, fn interface{}) {
	p.methods[name] = fn
}

func (p *Pool) Call(name string, args ...interface{}) {
	fn, ok := p.methods[name]
	if !ok {
		log.Fatal("function name is not found")
	}
	j := &Job{id: 0}
	j.callback = WrapThreadFunc(fn, args...) //we need to use ... to expand args or else it will be a double level slice
	p.AddJob(j)
}

func hello(name string) {
	fmt.Printf("hello my name is %s\n", name)
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Fatal("wrong main:", err)
			debug.PrintStack()
		}
	}()
	fmt.Println("start")
	pool := InitPool(5, 10)
	for i := 0; i < 10; i++ {
		j := &Job{id: i}
		if i%2 == 0 {
			j.callback = WrapThreadFunc(AddNum, 1, 2)
		} else {
			j.callback = WrapThreadFunc(DoubleStr, "123")
		}
		pool.AddJob(j)
	}
	fmt.Println("start to end")

	fmt.Println("sleep for 2 seconds")
	time.Sleep(time.Second * 2)
	pool.Register("hello", hello)
	pool.Call("hello", "tom")
	pool.Terminate()
}

// Add the following extensions,which allow you to register a function with name to thread pool
// call them with name by user,instead of excute them immediately
