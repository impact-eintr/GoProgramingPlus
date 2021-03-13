package main

import "fmt"

func main() {
	SliceCap()
	fmt.Println(1 << 3)
}
func SlicePrint() {
	fmt.Println("slice 共享内存")
	s1 := []int{1, 2}
	s2 := s1
	s3 := s2[:]
	fmt.Printf("%p %p %p\n", &s1[0], &s2[0], &s3[0])

}

func SliceCap() {
	var array [10]int
	var slice = array[5:10]
	var slice2 = array[8:9]
	fmt.Printf("len(slice) = %d\n", len(slice))
	fmt.Printf("cap(slice) = %d\n", cap(slice))
	slice2 = append(slice2, 10)
	fmt.Println(array)
}

func SliceExpress() {
	orderLen := 5
	order := make([]uint16, 2*orderLen)

	pollorder := order[:orderLen:orderLen]
	lockorder := order[orderLen:][:orderLen:orderLen]

	fmt.Println(pollorder)
	fmt.Println(lockorder)
}
