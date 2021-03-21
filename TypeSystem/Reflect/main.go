package main

import (
	"TypeSystem/Reflect/eintr"
	"fmt"
	"reflect"
)

func main() {
	u1 := &eintr.Eintr{
		Name: "eintr",
	}

	u1.Setname("szc")

	fmt.Println(u1.Getname())

	t := reflect.TypeOf(*u1)

	fmt.Println(t.Name(), t.NumMethod())
}
