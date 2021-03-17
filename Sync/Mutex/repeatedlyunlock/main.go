package main

import (
	"fmt"
	"sync"
)

func main() {
	defer func() {
		fmt.Println("Try to recover the panic")
		if p := recover(); p != nil {
			fmt.Println("Mutex ERROR:", p)
		}
	}()

	var mutex sync.Mutex

	mutex.Lock()

	fmt.Println("The lock is locked")
	fmt.Println("Unlock the lock")

	mutex.Unlock()

	fmt.Println("The lock is unlocked")
	fmt.Println("Unlock the lock again")

	mutex.Unlock()
}
