package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Func func(string)

type Task struct {
	Name         string
	Cost         int
	ConsumerCost int
	f            Func
}

var wg sync.WaitGroup

func TaskDone(name string) {
	fmt.Println("task ", name, "hash been done")
}

var tasks [2]Task
var buffer chan *Task

func producer(buffer chan *Task) {
	defer wg.Done()
	for i := 1; i < 10; i++ {
		task := &tasks[rand.Intn(2)]
		buffer <- task
		fmt.Println(task.Name, "was created")
		time.Sleep(time.Duration(task.Cost) * time.Second)
	}
}

func consumer(buffer chan *Task) {
	defer wg.Done()
	for i := 1; i < 10; i++ {
		task := <-buffer
		task.f(task.Name)
		time.Sleep(time.Duration(task.ConsumerCost) * time.Second)
	}
}

func main() {
	buffer = make(chan *Task, 5)
	tasks = [2]Task{
		{"task1", 1, 3, TaskDone},
		{"task2", 1, 2, TaskDone},
	}
	wg.Add(2)
	go producer(buffer)
	go consumer(buffer)
	wg.Wait()
	fmt.Println("All finish")
}
