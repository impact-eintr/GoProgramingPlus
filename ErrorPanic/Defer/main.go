package main

import "fmt"

func A1(a int) {
	fmt.Println(a)
}

func main() {
	a, b := 1, 2
	defer A1(a)

	a = a + b
	fmt.Println(a, b)
}
