# 常用关键字
## for range 
循环是所有编程语言都有的控制结构，除了使用经典的三段式循环之外，Go 语言还引入了另一个关键字 range 帮助我们快速遍历数组、切片、哈希表以及 Channel 等集合类型。本节将深入分析 Go 语言的两种循环，也就是 for 循环和 for-range 循环，我们会分析这两种循环的运行时结构以及它们的实现原理，

for 循环能够将代码中的数据和逻辑分离，让同一份代码能够多次复用相同的处理逻辑。我们先来看一下 Go 语言 for 循环对应的汇编代码，下面是一段经典的三段式循环的代码，我们将它编译成汇编指令：

``` go
package main

func main() {
	for i := 0; i < 10; i++ {
		println(i)
	}
}

"".main STEXT size=98 args=0x0 locals=0x18
	00000 (main.go:3)	TEXT	"".main(SB), $24-0
	...
	00029 (main.go:3)	XORL	AX, AX                   ;; i := 0
	00031 (main.go:4)	JMP	75
	00033 (main.go:4)	MOVQ	AX, "".i+8(SP)
	00038 (main.go:5)	CALL	runtime.printlock(SB)
	00043 (main.go:5)	MOVQ	"".i+8(SP), AX
	00048 (main.go:5)	MOVQ	AX, (SP)
	00052 (main.go:5)	CALL	runtime.printint(SB)
	00057 (main.go:5)	CALL	runtime.printnl(SB)
	00062 (main.go:5)	CALL	runtime.printunlock(SB)
	00067 (main.go:4)	MOVQ	"".i+8(SP), AX
	00072 (main.go:4)	INCQ	AX                       ;; i++
	00075 (main.go:4)	CMPQ	AX, $10                  ;; 比较变量 i 和 10
	00079 (main.go:4)	JLT	33                           ;; 跳转到 33 行如果 i < 10
	...
```

这里将上述汇编指令的执行过程分成三个部分进行分析：

0029 ~ 0031 行负责循环的初始化；
对寄存器 AX 中的变量 i 进行初始化并执行 JMP 75 指令跳转到 0075 行；
0075 ~ 0079 行负责检查循环的终止条件，将寄存器中存储的数据 i 与 10 比较；
JLT 33 命令会在变量的值小于 10 时跳转到 0033 行执行循环主体；
JLT 33 命令会在变量的值大于 10 时跳出循环体执行下面的代码；
0033 ~ 0072 行是循环内部的语句；
通过多个汇编指令打印变量中的内容；
INCQ AX 指令会将变量加一，然后再与 10 进行比较，回到第二步；

经过优化的 for-range 循环的汇编代码有着相同的结构。无论是变量的初始化、循环体的执行还是最后的条件判断都是完全一样的，所以这里也就不展开分析对应的汇编指令了

``` go
package main

func main() {
	arr := []int{1, 2, 3}
	for i, _ := range arr {
		println(i)
	}
}
```

在汇编语言中，无论是经典的 for 循环还是 for-range 循环都会使用 JMP 等命令跳回循环体的开始位置复用代码。从不同循环具有相同的汇编代码可以猜到，使用 for-range 的控制结构最终也会被 Go 语言编译器转换成普通的 for 循环，后面的分析也会印证这一点。


#### 循环永动机 
如果我们在遍历数组的同时修改数组的元素，能否得到一个永远都不会停止的循环呢？你可以尝试运行下面的代码：

``` go
func main() {
	arr := []int{1, 2, 3}
	for _, v := range arr {
		arr = append(arr, v)
	}
	fmt.Println(arr)
}

$ go run main.go
1 2 3 1 2 3
```

上述代码的输出意味着循环只遍历了原始切片中的三个元素，我们在遍历切片时追加的元素不会增加循环的执行次数，所以循环最终还是停了下来。

#### 神奇的指针
第二个例子是使用 Go 语言经常会犯的错误1。当我们在遍历一个数组时，如果获取 range 返回变量的地址并保存到另一个数组或者哈希时，会遇到令人困惑的现象，下面的代码会输出 “3 3 3”：

``` go
func main() {
	arr := []int{1, 2, 3}
	newArr := []*int{}
	for _, v := range arr {
		newArr = append(newArr, &v)
	}
	for _, v := range newArr {
		fmt.Println(*v)
	}
}

$ go run main.go
3 3 3
```

#### 遍历清空数组
当我们想要在 Go 语言中清空一个切片或者哈希时，一般都会使用以下的方法将切片中的元素置零：

``` go
func main() {
	arr := []int{1, 2, 3}
	for i, _ := range arr {
		arr[i] = 0
	}
}
```

依次遍历切片和哈希看起来是非常耗费性能的，因为数组、切片和哈希占用的内存空间都是连续的，所以最快的方法是直接清空这片内存中的内容，当我们编译上述代码时会得到以下的汇编指令：

``` go
"".main STEXT size=93 args=0x0 locals=0x30
	0x0000 00000 (main.go:3)	TEXT	"".main(SB), $48-0
	...
	0x001d 00029 (main.go:4)	MOVQ	"".statictmp_0(SB), AX
	0x0024 00036 (main.go:4)	MOVQ	AX, ""..autotmp_3+16(SP)
	0x0029 00041 (main.go:4)	MOVUPS	"".statictmp_0+8(SB), X0
	0x0030 00048 (main.go:4)	MOVUPS	X0, ""..autotmp_3+24(SP)
	0x0035 00053 (main.go:5)	PCDATA	$2, $1
	0x0035 00053 (main.go:5)	LEAQ	""..autotmp_3+16(SP), AX
	0x003a 00058 (main.go:5)	PCDATA	$2, $0
	0x003a 00058 (main.go:5)	MOVQ	AX, (SP)
	0x003e 00062 (main.go:5)	MOVQ	$24, 8(SP)
	0x0047 00071 (main.go:5)	CALL	runtime.memclrNoHeapPointers(SB)
	...
```

#### 随机遍历

当我们在 Go 语言中使用 range 遍历哈希表时，往往都会使用如下的代码结构，但是这段代码在每次运行时都会打印出不同的结果：

``` go
func main() {
	hash := map[string]int{
		"1": 1,
		"2": 2,
		"3": 3,
	}
	for k, v := range hash {
		println(k, v)
	}
}
```

两次运行上述代码可能得到，第一次会打印 2 3 1，第二次 1 2 3,如果我们运行的次数足够多，最后会得到几个不同的遍历顺序。

``` go
$ go run main.go
2 2
3 3
1 1

$ go run main.go
1 1
2 2
3 3
```

Go 语言在运行时为哈希表的遍历引入了不确定性，也是告诉所有 Go 语言的使用者，程序不要依赖于哈希表的稳定遍历，我们在下面的小节会介绍在遍历的过程是如何引入不确定性的。

### 经典循环
Go 语言中的经典循环在编译器看来是一个 OFOR 类型的节点，这个节点由以下四个部分组成：
- 初始化循环的 Ninit；
- 循环的继续条件 Left；
- 循环体结束时执行的 Right；
- 循环体 NBody：

``` go
for Ninit; Left; Right {
    NBody
}
```

在生成 SSA 中间代码的阶段，cmd/compile/internal/gc.state.stmt 方法在发现传入的节点类型是 OFOR 时会执行以下的代码块，这段代码会将循环中的代码分成不同的块：


## select



## defer 

## panic recover

## make new


