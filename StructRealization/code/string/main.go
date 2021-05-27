package main

import "fmt"

func main() {
	m := make(map[string]int, 1)
	m["test"] = 0
	fmt.Println(m)
	test(m)
	fmt.Println(m)
}

func test(m map[string]int) {
	v := m["test"]
	m["test"] = 1
	m["test"] = v
}
