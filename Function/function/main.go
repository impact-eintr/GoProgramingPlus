package main

import "fmt"

func swap1(a, b int) {
	a, b = b, a
}

func swap2(a, b *int) {
	*a, *b = *b, *a
}

func incr(a int) int {
	var b int

	defer func() {
		a++
		b++
		fmt.Println("in defer ", a, b)
	}()
	a++
	b = a
	fmt.Println("in inc ", a, b)
	return b
}

func incr1(a int) (b int) {
	defer func() {
		a++
		b++
	}()
	a++
	b = a
	return
}

func main() {
	var a, b int
	b = incr1(a)
	fmt.Println(a, b)
}
