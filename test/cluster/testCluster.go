package main

import "fmt"

func main() {
	s := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	n := 0
	for n = 0; n < len(s); n++ {
		if s[n]%2 == 0 {
			s = append(s[:n], s[n+1:]...)
			n--
		}
	}
	fmt.Println(n)
	fmt.Println(s)
}
