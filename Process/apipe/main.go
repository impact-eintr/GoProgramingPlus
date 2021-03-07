package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

func main() {
	//runCmd() //不使用管道
	fmt.Println()
	runCmdWithPipe() //使用管道
}

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

func runCmd() {

}
