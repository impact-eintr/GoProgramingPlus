package main

import "fmt"

type T struct {
	name string
}

func (t T) F1() {
	fmt.Println(t.name)
}

type myslice []string

func (ms myslice) Len() {
	fmt.Println(len(ms))
}

func (ms myslice) Cap() {
	fmt.Println(cap(ms))
}

type MyType1 = int32 //给类型取别名

type MyType2 int32 //自定义类型

func main() {
	t := T{
		name: "eintr",
	}

	t.F1()
}
