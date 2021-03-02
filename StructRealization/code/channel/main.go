package main

import "fmt"

var counter int = 0

func main() {
	var ch = make(chan int)
	var exit_ch = make(chan struct{}, 1)
	go testlock(ch, exit_ch)
	go testunlock(ch, exit_ch)
	for _, ok := <-exit_ch; ok; {
		fmt.Println(counter)
	}
}

func testlock(ch chan int, exit_ch chan struct{}) {
	ch <- 1
	counter++
	exit_ch <- struct{}{}
}

func testunlock(ch chan int, exit_ch chan struct{}) {
	<-ch
	counter++
	exit_ch <- struct{}{}
}
