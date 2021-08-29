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

### 范围循环
与简单的经典循环相比，范围循环在 Go 语言中更常见、实现也更复杂。这种循环同时使用 for 和 range 两个关键字，编译器会在编译期间将所有 for-range 循环变成经典循环。从编译器的视角来看，就是将 ORANGE 类型的节点转换成 OFOR 节点:

![](https://img.draveness.me/2020-01-17-15792766926441-Golang-For-Range-Loop.png)

节点类型的转换过程都发生在中间代码生成阶段，所有的 for-range 循环都会被 cmd/compile/internal/gc.walkrange 转换成不包含复杂结构、只包含基本表达式的语句。接下来，我们按照循环遍历的元素类型依次介绍遍历数组和切片、哈希表、字符串以及管道时的过程。

#### 数组和切片
对于数组和切片来说，Go 语言有三种不同的遍历方式，这三种不同的遍历方式分别对应着代码中的不同条件，它们会在 cmd/compile/internal/gc.walkrange 函数中转换成不同的控制逻辑，我们会分成几种情况分析该函数的逻辑：

- 分析遍历数组和切片清空元素的情况；
- 分析使用 for range a {} 遍历数组和切片，不关心索引和数据的情况；
- 分析使用 for i := range a {} 遍历数组和切片，只关心索引的情况；
- 分析使用 for i, elem := range a {} 遍历数组和切片，关心索引和数据的情况；

``` go
func walkrange(n *Node) *Node {
	switch t.Etype {
	case TARRAY, TSLICE:
		if arrayClear(n, v1, v2, a) {
			return n
		}
```
cmd/compile/internal/gc.arrayClear 是一个非常有趣的优化，它会优化 Go 语言遍历数组或者切片并删除全部元素的逻辑：

``` go
// 原代码
for i := range a {
	a[i] = zero
}

// 优化后
if len(a) != 0 {
	hp = &a[0]
	hn = len(a)*sizeof(elem(a))
	memclrNoHeapPointers(hp, hn)
	i = len(a) - 1
}
```

相比于依次清除数组或者切片中的数据，Go 语言会直接使用 runtime.memclrNoHeapPointers 或者 runtime.memclrHasPointers 清除目标数组内存空间中的全部数据，并在执行完成后更新遍历数组的索引，这也印证了我们在遍历清空数组一节中观察到的现象。
处理了这种特殊的情况之后，我们可以回到 ORANGE 节点的处理过程了。这里会设置 for 循环的 Left 和 Right 字段，也就是终止条件和循环体每次执行结束后运行的代码：

``` go
ha := a

hv1 := temp(types.Types[TINT])
hn := temp(types.Types[TINT])

init = append(init, nod(OAS, hv1, nil))
init = append(init, nod(OAS, hn, nod(OLEN, ha, nil)))

n.Left = nod(OLT, hv1, hn)
n.Right = nod(OAS, hv1, nod(OADD, hv1, nodintconst(1)))

if v1 == nil {
	break
}
```

如果循环是 for range a {}，那么就满足了上述代码中的条件 v1 == nil，即循环不关心数组的索引和数据，这种循环会被编译器转换成如下形式：

``` go
ha := a
hv1 := 0
hn := len(ha)
v1 := hv1
for ; hv1 < hn; hv1++ {
    ...
}
```

这是 ORANGE 结构在编译期间被转换的最简单形式，由于原代码不需要获取数组的索引和元素，只需要使用数组或者切片的数量执行对应次数的循环，所以会生成一个最简单的 for 循环。

如果我们在遍历数组时需要使用索引 for i := range a {}，那么编译器会继续会执行下面的代码：

``` go
		if v2 == nil {
			body = []*Node{nod(OAS, v1, hv1)}
			break
		}
```


v2 == nil 意味着调用方不关心数组的元素，只关心遍历数组使用的索引。它会将 for i := range a {} 转换成下面的逻辑，与第一种循环相比，这种循环在循环体中添加了 v1 := hv1 语句，传递遍历数组时的索引：

``` go
ha := a
hv1 := 0
hn := len(ha)
v1 := hv1
for ; hv1 < hn; hv1++ {
    v1 = hv1
    ...
}
```

上面两种情况虽然也是使用 range 会经常遇到的情况，但是同时去遍历索引和元素也很常见。处理这种情况会使用下面这段的代码：

``` go
		tmp := nod(OINDEX, ha, hv1)
		tmp.SetBounded(true)
		a := nod(OAS2, nil, nil)
		a.List.Set2(v1, v2)
		a.Rlist.Set2(hv1, tmp)
		body = []*Node{a}
	}
	n.Ninit.Append(init...)
	n.Nbody.Prepend(body...)

	return n
}
```

**对于所有的 range 循环，Go 语言都会在编译期将原切片或者数组赋值给一个新变量 ha，在赋值的过程中就发生了拷贝，而我们又通过 len 关键字预先获取了切片的长度，所以在循环中追加新的元素也不会改变循环执行的次数，这也就解释了循环永动机一节提到的现象。**

而遇到这种同时遍历索引和元素的 range 循环时，**Go 语言会额外创建一个新的 v2 变量存储切片中的元素*，循环中使用的这个变量 v2 会在每一次迭代被重新赋值而覆盖，赋值时也会触发拷贝。**

``` go
func main() {
	arr := []int{1, 2, 3}
	newArr := []*int{}
	for i, _ := range arr {
		newArr = append(newArr, &arr[i])
	}
	for _, v := range newArr {
		fmt.Println(*v)
	}
}
```

因为在循环中获取返回变量的地址都完全相同，所以会发生神奇的指针一节中的现象。因此当我们想要访问数组中元素所在的地址时，不应该直接获取 range 返回的变量地址 &v2，而应该使用 &a[index] 这种形式。

``` go
func main() {
	arr := []int{1, 2, 3}
	newArr := []*int{}
	for _, v := range arr {
		fmt.Printf("%p %v\n", &v, v)
		newArr = append(newArr, &v)
	}
	for _, v := range newArr {
		fmt.Println(*v)
	}
}

```

#### 哈希表
在遍历哈希表时，编译器会使用 runtime.mapiterinit 和 runtime.mapiternext 两个运行时函数重写原始的 for-range 循环：

``` go
ha := a
hit := hiter(n.Type)
th := hit.Type
mapiterinit(typename(t), ha, &hit)
for ; hit.key != nil; mapiternext(&hit) {
    key := *hit.key
    val := *hit.val
}
```

上述代码是展开 for key, val := range hash {} 后的结果，在 cmd/compile/internal/gc.walkrange 处理 TMAP 节点时，编译器会根据 range 返回值的数量在循环体中插入需要的赋值语句：
![](https://img.draveness.me/2020-01-17-15792766877639-golang-range-map.png)
这三种不同的情况分别向循环体插入了不同的赋值语句。遍历哈希表时会使用 runtime.mapiterinit 函数初始化遍历开始的元素：

``` go
func mapiterinit(t *maptype, h *hmap, it *hiter) {
	it.t = t
	it.h = h
	it.B = h.B
	it.buckets = h.buckets

	r := uintptr(fastrand())
	it.startBucket = r & bucketMask(h.B)
	it.offset = uint8(r >> h.B & (bucketCnt - 1))
	it.bucket = it.startBucket
	mapiternext(it)
}
```

该函数会初始化 runtime.hiter 结构体中的字段，并通过 runtime.fastrand 生成一个随机数帮助我们随机选择一个遍历桶的起始位置。Go 团队在设计哈希表的遍历时就不想让使用者依赖固定的遍历顺序，所以引入了随机数保证遍历的随机性。

遍历哈希会使用 runtime.mapiternext，我们在这里简化了很多逻辑，省去了一些边界条件以及哈希表扩容时的兼容操作，这里只需要关注处理遍历逻辑的核心代码，我们会将该函数分成桶的选择和桶内元素的遍历两部分，首先是桶的选择过程：

``` go
func mapiternext(it *hiter) {
	h := it.h
	t := it.t
	bucket := it.bucket
	b := it.bptr
	i := it.i
	alg := t.key.alg

next:
	if b == nil {
		if bucket == it.startBucket && it.wrapped {
			it.key = nil
			it.value = nil
			return
		}
		b = (*bmap)(add(it.buckets, bucket*uintptr(t.bucketsize)))
		bucket++
		if bucket == bucketShift(it.B) {
			bucket = 0
			it.wrapped = true
		}
		i = 0
	}
```

这段代码主要有两个作用：
- 在待遍历的桶为空时，选择需要遍历的新桶；
- 在不存在待遍历的桶时。返回 (nil, nil) 键值对并中止遍历；

runtime.mapiternext 剩余代码的作用是从桶中找到下一个遍历的元素，在大多数情况下都会直接操作内存获取目标键值的内存地址，不过如果哈希表处于扩容期间就会调用 runtime.mapaccessK 获取键值对：

``` go
	for ; i < bucketCnt; i++ {
		offi := (i + it.offset) & (bucketCnt - 1)
		k := add(unsafe.Pointer(b), dataOffset+uintptr(offi)*uintptr(t.keysize))
		v := add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.keysize)+uintptr(offi)*uintptr(t.valuesize))
		if (b.tophash[offi] != evacuatedX && b.tophash[offi] != evacuatedY) ||
			!(t.reflexivekey() || alg.equal(k, k)) {
			it.key = k
			it.value = v
		} else {
			rk, rv := mapaccessK(t, h, k)
			it.key = rk
			it.value = rv
		}
		it.bucket = bucket
		it.i = i + 1
		return
	}
	b = b.overflow(t)
	i = 0
	goto next
}
```
当上述函数已经遍历了正常桶后，会通过 runtime.bmap.overflow 遍历哈希中的溢出桶。
![](https://img.draveness.me/2020-01-17-15792766877646-golang-range-map-and-buckets.png)
简单总结一下哈希表遍历的顺序，首先会选出一个绿色的正常桶开始遍历，随后遍历所有黄色的溢出桶，最后依次按照索引顺序遍历哈希表中其他的桶，直到所有的桶都被遍历完成。

#### 字符串
字符串 #
遍历字符串的过程与数组、切片和哈希表非常相似，只是在遍历时会获取字符串中索引对应的字节并将字节转换成 rune。我们在遍历字符串时拿到的值都是 rune 类型的变量，for i, r := range s {} 的结构都会被转换成如下所示的形式：

``` go
ha := s
for hv1 := 0; hv1 < len(ha); {
    hv1t := hv1
    hv2 := rune(ha[hv1])
    if hv2 < utf8.RuneSelf {
        hv1++
    } else {
        hv2, hv1 = decoderune(ha, hv1)
    }
    v1, v2 = hv1t, hv2
}
```


在前面的字符串一节中我们曾经介绍过字符串是一个只读的字节数组切片，所以范围循环在编译期间生成的框架与切片非常类似，只是细节有一些不同。

使用下标访问字符串中的元素时得到的就是字节，但是这段代码会将当前的字节转换成 rune 类型。如果当前的 rune 是 ASCII 的，那么只会占用一个字节长度，每次循环体运行之后只需要将索引加一，但是如果当前 rune 占用了多个字节就会使用 runtime.decoderune 函数解码，具体的过程就不在这里详细介绍了。

### 通道
使用 range 遍历 Channel 也是比较常见的做法，一个形如 for v := range ch {} 的语句最终会被转换成如下的格式：

``` go
ha := a
hv1, hb := <-ha
for ; hb != false; hv1, hb = <-ha {
    v1 := hv1
    hv1 = nil
    ...
}
```


这里的代码可能与编译器生成的稍微有一些出入，但是结构和效果是完全相同的。该循环会使用 <-ch 从管道中取出等待处理的值，这个操作会调用 runtime.chanrecv2 并阻塞当前的协程，当 runtime.chanrecv2 返回时会根据布尔值 hb 判断当前的值是否存在：

如果不存在当前值，意味着当前的管道已经被关闭；
如果存在当前值，会为 v1 赋值并清除 hv1 变量中的数据，然后重新陷入阻塞等待新数据；
### 小结
这一节介绍的两个关键字 for 和 range 都是我们在学习和使用 Go 语言中无法绕开的，通过分析和研究它们的底层原理，让我们对实现细节有了更清楚的认识，包括 Go 语言遍历数组和切片时会复用变量、哈希表的随机遍历原理以及底层的一些优化，这都能帮助我们更好地理解和使用 Go 语言。


## select
select 是操作系统中的系统调用，我们经常会使用 select、poll 和 epoll 等函数构建 I/O 多路复用模型提升程序的性能。Go 语言的 select 与操作系统中的 select 比较相似，本节会介绍 Go 语言 select 关键字常见的现象、数据结构以及实现原理。

C 语言的 select 系统调用可以同时监听多个文件描述符的可读或者可写的状态，Go 语言中的 select 也能够让 Goroutine 同时等待多个 Channel 可读或者可写，在多个文件或者 Channel状态改变之前，select 会一直阻塞当前线程或者 Goroutine。

![](https://img.draveness.me/2020-01-19-15794018429532-Golang-Select-Channels.png)

select 是与 switch 相似的控制结构，与 switch 不同的是，select 中虽然也有多个 case，但是这些 case 中的表达式必须都是 Channel 的收发操作。下面的代码就展示了一个包含 Channel 收发操作的 select 结构：

``` go
func fibonacci(c, quit chan int) {
	x, y := 0, 1
	for {
		select {
		case c <- x:
			x, y = y, x+y
		case <-quit:
			fmt.Println("quit")
			return
		}
	}
}
```
上述控制结构会等待 c <- x 或者 <-quit 两个表达式中任意一个返回。无论哪一个表达式返回都会立刻执行 case 中的代码，当 select 中的两个 case 同时被触发时，会随机执行其中的一个。

### 现象
当我们在 Go 语言中使用 select 控制结构时，会遇到两个有趣的现象：
- select 能在 Channel 上进行非阻塞的收发操作；
- select 在遇到多个 Channel 同时响应时，会随机执行一种情况；

#### 非阻塞的收发

``` go
func main() {
	ch := make(chan int)
	select {
	case i := <-ch:
		println(i)

	default:
		println("default")
	}
}

$ go run main.go
default
```

只要我们稍微想一下，就会发现 Go 语言设计的这个现象很合理。select 的作用是同时监听多个 case 是否可以执行，如果多个 Channel 都不能执行，那么运行 default 也是理所当然的。

非阻塞的 Channel 发送和接收操作还是很有必要的，在很多场景下我们不希望 Channel 操作阻塞当前 Goroutine，只是想看看 Channel 的可读或者可写状态，如下所示：

``` go
errCh := make(chan error, len(tasks))
wg := sync.WaitGroup{}
wg.Add(len(tasks))
for i := range tasks {
    go func() {
        defer wg.Done()
        if err := tasks[i].Run(); err != nil {
            errCh <- err
        }
    }()
}
wg.Wait()

select {
case err := <-errCh:
    return err
default:
    return nil
}
```
在上面这段代码中，我们不关心到底多少个任务执行失败了，只关心是否存在返回错误的任务，最后的 select 语句能很好地完成这个任务

#### 随机执行
另一个使用 select 遇到的情况是同时有多个 case 就绪时，select 会选择哪个 case 执行的问题，我们通过下面的代码可以简单了解一下：

``` go

func main() {
	ch := make(chan int)
	go func() {
		for range time.Tick(1 * time.Second) {
			ch <- 0
		}
	}()

	for {
		select {
		case <-ch:
			println("case1")
		case <-ch:
			println("case2")
		}
	}
}

$ go run main.go
case1
case2
case1
...
```
从上述代码输出的结果中我们可以看到，select 在遇到多个 <-ch 同时满足可读或者可写条件时会随机选择一个 case 执行其中的代码。

这个设计是在十多年前被 select 提交5引入并一直保留到现在的，虽然中间经历过一些修改6，但是语义一直都没有改变。在上面的代码中，两个 case 都是同时满足执行条件的，如果我们按照顺序依次判断，那么后面的条件永远都会得不到执行，而随机的引入就是为了**避免饥饿问题**的发生。

### 数据结构
select 在 Go 语言的源代码中不存在对应的结构体，但是我们使用 runtime.scase 结构体表示 select 控制结构中的 case：

``` go
type scase struct {
	c    *hchan         // chan
	elem unsafe.Pointer // data element
}
```

因为非默认的 case 中都与 Channel 的发送和接收有关，所以 runtime.scase 结构体中也包含一个 runtime.hchan 类型的字段存储 case 中使用的 Channel。

### 实现原理
select 语句在编译期间会被转换成 OSELECT 节点。每个 OSELECT 节点都会持有一组 OCASE 节点，如果 OCASE 的执行条件是空，那就意味着这是一个 default 节点。
![](https://img.draveness.me/2020-01-18-15793463657473-golang-oselect-and-ocases.png)
append上图展示的就是 select 语句在编译期间的结构，每一个 OCASE 既包含执行条件也包含满足条件后执行的代码。

编译器在中间代码生成期间会根据 select 中 case 的不同对控制语句进行优化，这一过程都发生在 cmd/compile/internal/gc.walkselectcases 函数中，我们在这里会分四种情况介绍处理的过程和结果：

- select 不存在任何的 case；
- select 只存在一个 case；
- select 存在两个 case，其中一个 case 是 default；
- select 存在多个 case；
#### 直接阻塞
简单总结一下，空的 select 语句会直接阻塞当前 Goroutine，导致 Goroutine 进入无法被唤醒的永久休眠状态。
#### 单一管道
如果当前的 select 条件只包含一个 case，那么编译器会将 select 改写成 if 条件语句。下面对比了改写前后的代码：

``` go
// 改写前
select {
case v, ok <-ch: // case ch <- v
    ...    
}

// 改写后
if ch == nil {
    block()
}
v, ok := <-ch // case ch <- v
...

```

cmd/compile/internal/gc.walkselectcases 在处理单操作 select 语句时，会根据 Channel 的收发情况生成不同的语句。当 case 中的 Channel 是空指针时，会直接挂起当前 Goroutine 并陷入永久休眠。

#### 非阻塞操作
当 select 中仅包含两个 case，并且其中一个是 default 时，Go 语言的编译器就会认为这是一次非阻塞的收发操作。cmd/compile/internal/gc.walkselectcases 会对这种情况单独处理。不过在正式优化之前，该函数会将 case 中的所有 Channel 都转换成指向 Channel 的地址，我们会分别介绍非阻塞发送和非阻塞接收时，编译器进行的不同优化。


> 发送

首先是 Channel 的发送过程，当 case 中表达式的类型是 OSEND 时，编译器会使用条件语句和 runtime.selectnbsend 函数改写代码：

``` go
select {
case ch <- i:
    ...
default:
    ...
}

if selectnbsend(ch, i) {
    ...
} else {
    ...
}
```


这段代码中最重要的就是 runtime.selectnbsend，它为我们提供了向 Channel 非阻塞地发送数据的能力。我们在 Channel 一节介绍了向 Channel 发送数据的 runtime.chansend 函数包含一个 block 参数，该参数会决定这一次的发送是不是阻塞的：

``` go

func selectnbsend(c *hchan, elem unsafe.Pointer) (selected bool) {
	return chansend(c, elem, false, getcallerpc())
}
```

由于我们向 runtime.chansend 函数传入了非阻塞，所以在不存在接收方或者缓冲区空间不足时，当前 Goroutine 都不会阻塞而是会直接返回。

> 接收

由于从 Channel 中接收数据可能会返回一个或者两个值，所以接收数据的情况会比发送稍显复杂，不过改写的套路是差不多的：

``` go
// 改写前
select {
case v <- ch: // case v, ok <- ch:
    ......
default:
    ......
}

// 改写后
if selectnbrecv(&v, ch) { // if selectnbrecv2(&v, &ok, ch) {
    ...
} else {
    ...
}
```


返回值数量不同会导致使用函数的不同，两个用于非阻塞接收消息的函数 runtime.selectnbrecv 和 runtime.selectnbrecv2 只是对 runtime.chanrecv 返回值的处理稍有不同：

``` go
func selectnbrecv(elem unsafe.Pointer, c *hchan) (selected bool) {
	selected, _ = chanrecv(c, elem, false)
	return
}

func selectnbrecv2(elem unsafe.Pointer, received *bool, c *hchan) (selected bool) {
	selected, *received = chanrecv(c, elem, false)
	return
}
```


因为接收方不需要，所以 runtime.selectnbrecv 会直接忽略返回的布尔值，而 runtime.selectnbrecv2 会将布尔值回传给调用方。与 runtime.chansend 一样，runtime.chanrecv 也提供了一个 block 参数用于控制这次接收是否阻塞。
#### 常见流程
在默认的情况下，编译器会使用如下的流程处理 select 语句：

- 将所有的 case 转换成包含 Channel 以及类型等信息的 runtime.scase 结构体；
- 调用运行时函数 runtime.selectgo 从多个准备就绪的 Channel 中选择一个可执行的 runtime.scase 结构体；
- 通过 for 循环生成一组 if 语句，在语句中判断自己是不是被选中的 case；

一个包含三个 case 的正常 select 语句其实会被展开成如下所示的逻辑，我们可以看到其中处理的三个部分：

``` go
selv := [3]scase{}
order := [6]uint16
for i, cas := range cases {
    c := scase{}
    c.kind = ...
    c.elem = ...
    c.c = ...
}
chosen, revcOK := selectgo(selv, order, 3)
if chosen == 0 {
    ...
    break
}
if chosen == 1 {
    ...
    break
}
if chosen == 2 {
    ...
    break
}
```

展开后的代码片段中最重要的就是用于选择待执行 case 的运行时函数 runtime.selectgo，这也是我们要关注的重点。因为这个函数的实现比较复杂， 所以这里分两部分分析它的执行过程：

- 执行一些必要的初始化操作并确定 case 的处理顺序；
- 在循环中根据 case 的类型做出不同的处理；




### 小结
我们简单总结一下 select 结构的执行过程与实现原理，首先在编译期间，Go 语言会对 select 语句进行优化，它会根据 select 中 case 的不同选择不同的优化路径：

- 空的 select 语句会被转换成调用 runtime.block 直接挂起当前 Goroutine；
- 如果 select 语句中只包含一个 case，编译器会将其转换成 if ch == nil { block }; n; 表达式；
  - 首先判断操作的 Channel 是不是空的；
  - 然后执行 case 结构中的内容；
- 如果 select 语句中只包含两个 case 并且其中一个是 default，那么会使用 runtime.selectnbrecv 和 runtime.selectnbsend 非阻塞地执行收发操作；
- 在默认情况下会通过 runtime.selectgo 获取执行 case 的索引，并通过多个 if 语句执行对应 case 中的代码；

在编译器已经对 select 语句进行优化之后，Go 语言会在运行时执行编译期间展开的 runtime.selectgo 函数，该函数会按照以下的流程执行：

- 随机生成一个遍历的轮询顺序 pollOrder 并根据 Channel 地址生成锁定顺序 lockOrder；
- 根据 pollOrder 遍历所有的 case 查看是否有可以立刻处理的 Channel；
  - 如果存在，直接获取 case 对应的索引并返回；
  - 如果不存在，创建 runtime.sudog 结构体，将当前 Goroutine 加入到所有相关 Channel 的收发队列，并调用 runtime.gopark 挂起当前 Goroutine 等待调度器的唤醒；
- 当调度器唤醒当前 Goroutine 时，会再次按照 lockOrder 遍历所有的 case，从中查找需要被处理的 runtime.sudog 对应的索引；

select 关键字是 Go 语言特有的控制结构，它的实现原理比较复杂，需要编译器和运行时函数的通力合作。


## defer 

## panic recover

## make new
当我们想要在 Go 语言中初始化一个结构时，可能会用到两个不同的关键字 — make 和 new。因为它们的功能相似，所以初学者可能会对这两个关键字的作用感到困惑1，但是它们两者能够初始化的变量却有较大的不同

- make 的作用是**初始化内置的数据结构**，也就是我们在前面提到的切片、哈希表和 Channel；
- new 的作用是**根据传入的类型分配一片内存空间并返回指向这片内存空间的指针**；
我们在代码中往往都会使用如下所示的语句初始化这三类基本类型，这三个语句分别返回了不同类型的数据结构：

``` go
slice := make([]int, 0, 100)
hash := make(map[int]bool, 10)
ch := make(chan int, 5)
```

- slice 是一个包含 data、cap 和 len 的结构体 reflect.SliceHeader；
- hash 是一个指向 runtime.hmap 结构体的指针；
- ch 是一个指向 runtime.hchan 结构体的指针；

相比与复杂的 make 关键字，new 的功能就简单多了，它只能接收类型作为参数然后返回一个指向该类型的指针：

``` go
i := new(int)

var v int
i := &v
```

上述代码片段中的两种不同初始化方法是等价的，它们都会创建一个指向 int 零值的指针。
![](https://img.draveness.me/golang-make-and-new.png)

接下来我们将分别介绍 make 和 new 初始化不同数据结构的过程，我们会从编译期间和运行时两个不同阶段理解这两个关键字的原理，不过由于前面的章节已经详细地分析过 make 的原理，所以这里会将重点放在另一个关键字 new 上。

### make
![](https://img.draveness.me/golang-make-typecheck.png)
在编译期间的类型检查阶段，Go 语言会将代表 make 关键字的 OMAKE 节点根据参数类型的不同转换成了 OMAKESLICE、OMAKEMAP 和 OMAKECHAN 三种不同类型的节点，这些节点会调用不同的运行时函数来初始化相应的数据结构。

### new

编译器会在中间代码生成阶段通过以下两个函数处理该关键字：

- cmd/compile/internal/gc.callnew 会将关键字转换成 ONEWOBJ 类型的节点；
- cmd/compile/internal/gc.state.expr 会根据申请空间的大小分两种情况处理：
    - 如果申请的空间为 0，就会返回一个表示空指针的 zerobase 变量；
    - 在遇到其他情况时会将关键字转换成 runtime.newobject 函数：

``` go
func callnew(t *types.Type) *Node {
	...
	n := nod(ONEWOBJ, typename(t), nil)
	...
	return n
}

func (s *state) expr(n *Node) *ssa.Value {
	switch n.Op {
	case ONEWOBJ:
		if n.Type.Elem().Size() == 0 {
			return s.newValue1A(ssa.OpAddr, n.Type, zerobaseSym, s.sb)
		}
		typ := s.expr(n.Left)
		vv := s.rtcall(newobject, true, []*types.Type{n.Type}, typ)
		return vv[0]
	}
}
```

需要注意的是，无论是直接使用 new，还是使用 var 初始化变量，它们在编译器看来都是 ONEW 和 ODCL 节点。如果变量会逃逸到堆上，这些节点在这一阶段都会被 cmd/compile/internal/gc.walkstmt 转换成通过 runtime.newobject 函数并在堆上申请内存：

``` go
func walkstmt(n *Node) *Node {
	switch n.Op {
	case ODCL:
		v := n.Left
		if v.Class() == PAUTOHEAP {
			if prealloc[v] == nil {
				prealloc[v] = callnew(v.Type)
			}
			nn := nod(OAS, v.Name.Param.Heapaddr, prealloc[v])
			nn.SetColas(true)
			nn = typecheck(nn, ctxStmt)
			return walkstmt(nn)
		}
	case ONEW:
		if n.Esc == EscNone {
			r := temp(n.Type.Elem())
			r = nod(OAS, r, nil)
			r = typecheck(r, ctxStmt)
			init.Append(r)
			r = nod(OADDR, r.Left, nil)
			r = typecheck(r, ctxExpr)
			n = r
		} else {
			n = callnew(n.Type.Elem())
		}
	}
}
```

不过这也不是绝对的，如果通过 var 或者 new 创建的变量不需要在当前作用域外生存(闭包)，例如不用作为返回值返回给调用方，那么就不需要初始化在堆上。

runtime.newobject 函数会获取传入类型占用空间的大小，调用 runtime.mallocgc 在堆上申请一片内存空间并返回指向这片内存空间的指针：

``` go
func newobject(typ *_type) unsafe.Pointer {
	return mallocgc(typ.size, typ, true)
}
```

### 小结
这里我们简单总结一下 Go 语言中 make 和 new 关键字的实现原理，make 关键字的作用是创建切片、哈希表和 Channel 等内置的数据结构，而 new 的作用是为类型申请一片内存空间，并返回指向这片内存的指针。
