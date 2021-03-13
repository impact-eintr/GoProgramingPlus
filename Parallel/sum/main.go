package main

import (
	"fmt"
)

func main() {

	var res int
	const N int = 1000000

	numch := make(chan int, 30)
	resch := make(chan int, 8)

	for i := 0; i < 8; i++ {
		go func() {
			res := 0
			for num := range numch {
				res += num
			}
			resch <- res
		}()
	}

	for i := 0; i < N; i++ {
		numch <- i
	}

	close(numch)

	for i := 0; i < 8; i++ {
		res += <-resch
	}

	fmt.Println(res)
}
