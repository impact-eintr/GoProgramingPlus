package main

import "fmt"

func main() {
	fmt.Println("vim-go")
	SlicePrint()
	SliceCap()
	SliceExpress()
}
func SlicePrint() {
	fmt.Println("slice 共享内存")
	s1 := []int{1, 2}
	s2 := s1
	s3 := s2[:]
	fmt.Printf("%p %p %p", &s1[0], &s2[0], &s3[0])

}

func SliceCap() {
	var array [10]int
	var slice = array[5:6]

	fmt.Printf("len(slice) = %d\n", len(slice))
	fmt.Printf("cap(slice) = %d\n", cap(slice))
}

func SliceExpress() {
	orderLen := 5
	order := make([]uint16, 2*orderLen)

	pollorder := order[:orderLen:orderLen]
	lockorder := order[orderLen:][:orderLen:orderLen]

	fmt.Println(pollorder)
	fmt.Println(lockorder)
}
