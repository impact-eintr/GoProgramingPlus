package main

import (
	"fmt"
	"io"
)

type eintr struct {
	name string
}

func (this *eintr) Write(p []byte) (n int, err error) {
	return 0, nil
}

func main() {
	var w io.Writer

	f := eintr{
		name: "eintr",
	}

	w = &f

	rw, ok := w.(io.ReadWriter)
	if !ok {
		fmt.Println("断言失败")
		return
	}

	fmt.Printf("%T\n", rw)

}
