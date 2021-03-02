package main

import "fmt"

func A(i int) {
	i++
	fmt.Println(i)
}

func B() {
	f1 := A
	f1(1)
}

func C() {
	f2 := A
	f2(1)
}

func create() func() int {
	c := 666
	return func() int {
		return c
	}
}

func create2() (fs [2]func()) {
	for i := 0; i < 2; i++ {
		fs[i] = func() {
			fmt.Println(i)
		}
	}
	return
}

func main() {
	fs := create2()
	for i := 0; i < len(fs); i++ {
		fs[i]()
	}
}
