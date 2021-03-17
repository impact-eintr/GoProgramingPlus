package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var mutex sync.Mutex
	fmt.Println("Lock the lock[main]")

	mutex.Lock()

	fmt.Println("the lock is locked[main]")
	for i := 1; i <= 3; i++ {
		go func(i int) {
			fmt.Printf("Lock the lock[%d]\n", i)
			mutex.Lock()
			defer mutex.Unlock()
			fmt.Printf("The lock is locked[%d]\n", i)
		}(i)
	}

	time.Sleep(time.Second)
	fmt.Println("unlock the lock[main]")
	mutex.Unlock()

	fmt.Println("The lock is unlocked[main]")
	time.Sleep(time.Second)
}
