package main

//#include <stdio.h>
//static const char* cs = "hello!";
import "C"

type CChar C.char

func (p *CChar) GoString() string {
	return c.GoString((*C.char)(p))
}

func PrintCString(cs *C.char) {
	C.puts(cs)
}

func main() {
	PrintCString(C.cs)
}
