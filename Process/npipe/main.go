package main

import (
	"fmt"
	"io"
	"os"
	"time"
)

func main() {
	fileBasePipe() //基于命名管道文件的实现
}

func fileBasePipe() {
	reader, writer, err := os.Pipe()
	if err != nil {
		fmt.Printf("Error: Could't create the named pipe: %s\n", err)
	}

	go func() { //命名管道默认会在其中一段还未就绪的时候阻塞另一端的进程 必须并行执行
		output := make([]byte, 100)
		n, err := reader.Read(output)
		if err != nil {
			fmt.Printf("Error: read err %s\n", err)
		}
		fmt.Printf("Read %d byte(s)\n", n)
		for i := range output {
			fmt.Printf("%c", i)
		}
	}()

	input := make([]byte, 26)
	for i := 65; i <= 90; i++ {
		input[i-65] = byte(i)
	}
	n, err := writer.Write(input)
	if err != nil {
		fmt.Printf("Error: write data err %s\n", err)
	}

	fmt.Printf("Written %d byte(s)\n", n)

	time.Sleep(time.Millisecond)
}

func inMemorySyncPipe() {
	reader, writer := io.Pipe()
	go func() {
		output := make([]byte, 100)
		n, err := reader.Read(output)
		if err != nil {
			fmt.Printf("Error: Couldn't read data from the named pipe: %s\n", err)

		}
		fmt.Printf("Read %d byte(s). [in-memory pipe]\n", n)

	}()
	input := make([]byte, 26)
	for i := 65; i <= 90; i++ {
		input[i-65] = byte(i)

	}
	n, err := writer.Write(input)
	if err != nil {
		fmt.Printf("Error: Couldn't write data to the named pipe: %s\n", err)

	}
	fmt.Printf("Written %d byte(s). [in-memory pipe]\n", n)
	time.Sleep(200 * time.Millisecond)

}
