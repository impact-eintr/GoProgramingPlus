package main

import "fmt"

type Animal struct {
	Name string
}

type Cat struct {
	Animal
}

func (c *Cat) SetName(name string) {
	c.Name = name
}

func main() {
	mimi := new(Cat)
	mimi.SetName("mimi")

	haha := Cat{}
	haha.Name = "haha"
	haha.SetName("HaHa")
	fmt.Println(haha.Name)

}
