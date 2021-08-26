package main

//void SayHello(char *s);
//void SayHi(_GoString_ s);
import "C"

import "fmt"

func main() {
	C.SayHello(C.CString("Hello world!"))
	C.SayHi("Hi World!")
}

//export SayHello
func SayHello(s *C.char) {
	fmt.Println(C.GoString(s))
}

//export SayHi
func SayHi(s string) {
	fmt.Println(s)
}
