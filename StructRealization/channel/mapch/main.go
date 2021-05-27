package main

import (
	"fmt"
)

type Ch chan string

func main() {
	var test = make(map[string]chan string, 1)

	test["test"] = make(chan string, 1)
	test["test"] <- "测试"
	fmt.Println(<-test["test"])
}
