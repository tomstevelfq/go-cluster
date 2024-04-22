package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan int, 3) // create a buffer channel with size of three Integers

	// send data to channel
	ch <- 1
	ch <- 2
	ch <- 3

	close(ch)

	for {
		select {
		case value, ok := <-ch:
			if !ok {
				fmt.Println("closed")
			}
			fmt.Println("recv value:", value)
		default:
			fmt.Println("no value in channel")
			time.Sleep(500 * time.Millisecond)
		}
	}
}
