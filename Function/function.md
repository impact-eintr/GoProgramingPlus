# 函数
## 函数的一生

我们按照编程语言的语法定义的函数，会被编译器编译为一对对机器指令，写入可执行文件，程序执行时，可执行文件被加载到内存，这些机器指令对应到虚拟空间地址中，位于代码段,如果在一个函数中调用另一个函数，编译器就会对应生成一条`call`指令，程序执行到这一条指令时，就会跳到函数如后出开始执行，而每个函数的最后都有一条`ret`指令，负责在函数结束后条会到调用处继续执行。

运行时内存的布局如图

在golang中，函数栈帧布局是这样的

`call`指令只做两件事
- 将下一条指令的地址入栈,也就是返回地址
- 跳转到被调用的函数入口处执行
`call`指令执行后进入被调用函数的栈帧，所有的函数栈帧布局都遵循统一的约定,所以被调用函数时通过`sp+偏移量`来定位到每个参数和返回值的

> go函数栈帧的扩张

一次性扩张到所需最大栈空间的位置，然后通过`sp+偏移值`使用函数栈帧

这样是为了避免`栈访问越界`

由于函数栈帧的大小，可以砸死编译时期确定，对于栈消耗较大的函数，go的编译器会在函数头部插入检测代码，检测到需要`栈增长`，就会另外分配一段足够大的栈空间，并把原来栈上的数据拷过来,原来的栈空间就被释放了


> call与ret

函数通过call实现跳转，而每个函数开始时会分配栈帧名结束前又会释放自己的栈帧,ret指令又会把栈恢复到call之前的样子,通过这些指令的配合能够实现函数层层嵌套

## 参数与返回值

~~~ go
func swap1(a, b int) {
	a, b = b, a
}

func swap2(a, b *int) {
	*a, *b = *b, *a
}

func main() {
	a := 1
	b := 3
	swap1(a, b)
	fmt.Println(a, b)
	swap2(&a, &b)
	fmt.Println(a, b)
}
~~~

~~~ go
func incr(a int) int {
	var b int

	defer func() {
		a++
		b++
		fmt.Println("in defer ", a, b)
	}()
	a++
	b = a
	fmt.Println("in inc ", a, b)
	return b
}

func main() {
	var a, b int
	b = incr(a)
	fmt.Println(a, b)
}

~~~





~~~ go
func incr1(a int) (b int) {
	defer func() {
		a++
		b++
	}()
	a++
	b = a
	return
}

func main() {
	var a, b int
	b = incr1(a)
	fmt.Println(a, b)
}
~~~

























