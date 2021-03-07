# 多进程编程

## 管道

- 手动实现
~~~ go
func runCmdWithPipe() {
	fmt.Println("Run command `ps aux | grep apipe ` : ")
	cmd1 := exec.Command("ps", "aux")
	cmd2 := exec.Command("grep", "apipe")

	var outputBuf1 bytes.Buffer
	cmd1.Stdout = &outputBuf1 //输出重定向

	err := cmd1.Start()
	if err != nil {
		fmt.Printf("Error: The first command can not be startup %s\n", err)
		return
	}

	err = cmd1.Wait()
	if err != nil {
		fmt.Printf("Error: Could wait for the second command: %s\n", err)
		return
	}

	cmd2.Stdin = &outputBuf1 //输入重定向
	var outputBuf2 bytes.Buffer
	cmd2.Stdout = &outputBuf2 //输出重定向

	err = cmd2.Start()
	if err != nil {
		fmt.Printf("Error: The first command can not be startup %s\n", err)
		return
	}
	err = cmd2.Wait()
	if err != nil {
		fmt.Printf("Error: Could wait for the second command: %s\n", err)
		return
	}

	fmt.Printf("%s\n", outputBuf2.Bytes())

}
~~~
- 使用标准库
~~~ go
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
~~~

- 使用io库
~~~ go
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
~~~

## 


## 


##


##


