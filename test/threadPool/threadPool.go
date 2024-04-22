package main

import (
	"fmt"
	"log"
	"runtime/debug"
	"sync"
)

type Job struct {
	id int
}

type Worker struct {
	JobChannel chan *Job
	id         int
	quit       chan *sync.WaitGroup
}

type Pool struct {
	Workers    []*Worker
	JobChannel chan *Job
	wg         sync.WaitGroup
}

func (w *Worker) Process(job *Job) {
	fmt.Println("worker ", w.id, "is processing job ", job.id)
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
		pool.AddJob(&Job{id: i})
	}
	fmt.Println("start to end")
	pool.Terminate()
}
