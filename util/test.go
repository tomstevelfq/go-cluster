package main

import (
	"fmt"
	"log"
)

func main() {
	fmt.Println("start")
	log.Fatal("hello", nil)
	fmt.Println("haha")
	return
}
