package main

import (
	"fmt"
)

type A struct {
	name string
}

func (a A) Name() string {
	a.name = "Hi! " + a.name
	return a.name
}

func (a *A) NameP() string {
	a.name = "Hi! " + a.name
	return a.name
}

func NameOfA(a A) string {
	a.name = "Hi!" + a.name
	return a.name
}

func (a A) GetName() string {
	return a.name
}

func (pa *A) SetName() string {
	pa.name = "Hi! " + pa.name
	return pa.name
}

func main() {
	a := A{name: "eintr"}

	f1 := A.GetName
	f1(a)
}
