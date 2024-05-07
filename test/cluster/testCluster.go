package main

import (
	"log"
	"os"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	file, err := os.OpenFile("test.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("cannot open the log file:", err)
	}
	defer file.Close()
	log.SetOutput(file)
	a := 20
	log.Println("haha", a)
}
