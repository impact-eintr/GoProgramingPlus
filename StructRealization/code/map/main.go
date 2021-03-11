package main

import "fmt"

func main() {
	map1 := make(map[int]string, 5)
	map1[1] = "hhh"
	for i, v := range map1 {
		fmt.Println("i:", i, "v:", v)
	}
}
