# 数据结构

## 数组
数组是由相同类型元素的集合组成的数据结构，计算机会为数组分配一块连续的内存来保存其中的元素，我们可以利用数组中元素的索引快速访问特定元素，常见的数组大多都是一维的线性数组，而多维数组在数值和图形计算领域却有比较常见的应用


数组作为一种基本的数据类型，我们通常会从两个维度描述数组，也就是数组中存储的元素类型和数组最大能存储的元素个数，在 Go 语言中我们往往会使用如下所示的方式来表示数组类型：

``` go
[10]int
[200]interface{}
```

Go 语言数组在初始化之后大小就无法改变，存储元素类型相同、但是大小不同的数组类型在 Go 语言看来也是完全不同的，只有两个条件都相同才是同一类型。

``` go
func NewArray(elem *Type, bound int64) *Type {
    if bound < 0 {
        Fatalf("非法的数组大小%v", bound)
    }
    t := New(TARRAY)
    t.Extra = &Array{Elem: elem, Bound: bound}
    t.SetNotInHeap(elem.NotInHeap())
    return t
}
```

编译期间的数组类型是由上述的 cmd/compile/internal/types.NewArray 函数生成的，该类型包含两个字段，分别是元素类型 Elem 和数组的大小 Bound，这两个字段共同构成了数组类型，而当前数组是否应该在堆栈中初始化也在编译期就确定了

### 初始化
Go 语言的数组有两种不同的创建方式，一种是显式的指定数组大小，另一种是使用 [...]T 声明数组，Go 语言会在编译期间通过源代码推导数组的大小：

``` go

arr1 := [3]int{1, 2, 3}
arr2 := [...]int{1, 2, 3}
```

上述两种声明方式在运行期间得到的结果是完全相同的，后一种声明方式在编译期间就会被转换成前一种，这也就是编译器对数组大小的推导，下面我们来介绍编译器的推导过程。

#### 上限推导
两种不同的声明方式会导致编译器做出完全不同的处理，如果我们使用第一种方式 [10]T，那么变量的类型在编译进行到类型检查阶段就会被提取出来，随后使用 cmd/compile/internal/types.NewArray创建包含数组大小的 cmd/compile/internal/types.Array 结构体。

当我们使用 [...]T 的方式声明数组时，编译器会在的 cmd/compile/internal/gc.typecheckcomplit 函数中对该数组的大小进行推导：

``` go
func typecheckcomplit(n *Node) (res *Node) {
	...
	if n.Right.Op == OTARRAY && n.Right.Left != nil && n.Right.Left.Op == ODDD {
		n.Right.Right = typecheck(n.Right.Right, ctxType)
		if n.Right.Right.Type == nil {
			n.Type = nil
			return n
		}
		elemType := n.Right.Right.Type

		length := typecheckarraylit(elemType, -1, n.List.Slice(), "array literal")

		n.Op = OARRAYLIT
		n.Type = types.NewArray(elemType, length)
		n.Right = nil
		return n
	}
	...

	switch t.Etype {
	case TARRAY:
		typecheckarraylit(t.Elem(), t.NumElem(), n.List.Slice(), "array literal")
		n.Op = OARRAYLIT
		n.Right = nil
	}
}
```
这个删减后的 cmd/compile/internal/gc.typecheckcomplit 会调用 cmd/compile/internal/gc.typecheckarraylit 通过遍历元素的方式来计算数组中元素的数量。

所以我们可以看出 [...]T{1, 2, 3} 和 [3]T{1, 2, 3} 在运行时是完全等价的，[...]T 这种初始化方式也只是 Go 语言为我们提供的一种语法糖，当我们不想计算数组中的元素个数时可以通过这种方法减少一些工作量。

#### 语句转换
对于一个由字面量组成的数组，根据数组元素数量的不同，编译器会在负责初始化字面量的 cmd/compile/internal/gc.anylit 函数中做两种不同的优化：

- 当元素数量小于或者等于 4 个时，会直接将数组中的元素放置在栈上；
- 当元素数量大于 4 个时，会将数组中的元素放置到静态区并在运行时取出；

``` go
func anylit(n *Node, var_ *Node, init *Nodes) {
	t := n.Type
	switch n.Op {
	case OSTRUCTLIT, OARRAYLIT:
		if n.List.Len() > 4 {
			...
		}

		fixedlit(inInitFunction, initKindLocalCode, n, var_, init)
	...
	}
}
```

总结起来，在不考虑逃逸分析的情况下，如果数组中元素的个数小于或者等于 4 个，那么所有的变量会直接在栈上初始化，如果数组元素大于 4 个，变量就会在静态存储区初始化然后拷贝到栈上，这些转换后的代码才会继续进入中间代码生成和机器码生成两个阶段，最后生成可以执行的二进制文件。

### 访问与赋值

无论是在栈上还是静态存储区，数组在内存中都是一连串的内存空间，我们通过指向数组开头的指针、元素的数量以及元素类型占的空间大小表示数组。如果我们不知道数组中元素的数量，访问时可能发生越界；而如果不知道数组中元素类型的大小，就没有办法知道应该一次取出多少字节的数据，无论丢失了哪个信息，我们都无法知道这片连续的内存空间到底存储了什么数据。

数组访问越界是非常严重的错误，Go 语言中可以在编译期间的静态类型检查判断数组越界，cmd/compile/internal/gc.typecheck1 会验证访问数组的索引

``` go
func typecheck1(n *Node, top int) (res *Node) {
	switch n.Op {
	case OINDEX:
		ok |= ctxExpr
		l := n.Left  // array
		r := n.Right // index
		switch n.Left.Type.Etype {
		case TSTRING, TARRAY, TSLICE:
			...
			if n.Right.Type != nil && !n.Right.Type.IsInteger() {
				yyerror("non-integer array index %v", n.Right)
				break
			}
			if !n.Bounded() && Isconst(n.Right, CTINT) {
				x := n.Right.Int64()
				if x < 0 {
					yyerror("invalid array index %v (index must be non-negative)", n.Right)
				} else if n.Left.Type.IsArray() && x >= n.Left.Type.NumElem() {
					yyerror("invalid array index %v (out of bounds for %d-element array)", n.Right, n.Left.Type.NumElem())
				}
			}
		}
	...
	}
}
```
1. 访问数组的索引是非整数时，报错 “non-integer array index %v”；
2. 访问数组的索引是负数时，报错 “invalid array index %v (index must be non-negative)"；
3. 访问数组的索引越界时，报错 “invalid array index %v (out of bounds for %d-element array)"；

数组和字符串的一些简单越界错误都会在编译期间发现，例如：直接使用整数或者常量访问数组；但是如果使用变量去访问数组或者字符串时，编译器就无法提前发现错误，我们需要 Go 语言运行时阻止不合法的访问
Go 语言运行时在发现数组、切片和字符串的越界操作会由运行时的 runtime.panicIndex 和 runtime.goPanicIndex 触发程序的运行时错误并导致崩溃退出：

``` go
TEXT runtime·panicIndex(SB),NOSPLIT,$0-8
	MOVL	AX, x+0(FP)
	MOVL	CX, y+4(FP)
	JMP	runtime·goPanicIndex(SB)

func goPanicIndex(x int, y int) {
	panicCheck1(getcallerpc(), "index out of range")
	panic(boundsError{x: int64(x), signed: true, y: y, code: boundsIndex})
}
```

Go 语言对于数组的访问还是有着比较多的检查的，它不仅会在编译期间提前发现一些简单的越界错误并插入用于检测数组上限的函数调用，还会在运行期间通过插入的函数保证不会发生越界。

### 小结
数组是 Go 语言中重要的数据结构，了解它的实现能够帮助我们更好地理解这门语言，通过对其实现的分析，我们知道了对数组的访问和赋值需要同时依赖编译器和运行时，它的大多数操作在编译期间都会转换成直接读写内存，在中间代码生成期间，编译器还会插入运行时方法 runtime.panicIndex 调用防止发生越界错误。

## 切片
上一节介绍的数组在 Go 语言中没那么常用，更常用的数据结构是切片，即动态数组，其长度并不固定，我们可以向切片中追加元素，它会在容量不足时自动扩容。

在 Go 语言中，切片类型的声明方式与数组有一些相似，不过由于切片的长度是动态的，所以声明时只需要指定切片中的元素类型：

``` go
[]int
[]interface{}
[]type // 这个类型是在编译时期确定的
```

从切片的定义我们能推测出，切片在编译期间的生成的类型只会包含切片中的元素类型，即 int 或者 interface{} 等。cmd/compile/internal/types.NewSlice 就是编译期间用于创建切片类型的函数：

``` go
func NewSlice(elem *Type) *Type {
	if t := elem.Cache.slice; t != nil {
		if t.Elem() != elem {
			Fatalf("elem mismatch")
		}
		return t
	}

	t := New(TSLICE)
	t.Extra = Slice{Elem: elem}
	elem.Cache.slice = t
	return t
}
```
上述方法返回结构体中的 Extra 字段是一个只包含切片内元素类型的结构，也就是说切片内元素的类型都是在**编译期间**确定的，编译器确定了类型之后，会将类型存储在 Extra 字段中帮助程序在运行时动态获取。

### 数据结构
编译期间的切片是 cmd/compile/internal/types.Slice 类型的，但是在运行时切片可以由如下的 reflect.SliceHeader 结构体表示，其中:

- Data 是指向数组的指针;
- Len 是当前切片的长度；
- Cap 是当前切片的容量，即 Data 数组的大小：

``` go
type SliceHeader struct {
	Data uintptr
	Len  int
	Cap  int
}
```

Data 是一片连续的内存空间，这片内存空间可以用于存储切片中的全部元素，数组中的元素只是逻辑上的概念，底层存储其实都是连续的，所以我们可以将切片理解成一片连续的内存空间加上**长度与容量**的标识。
![](https://img.draveness.me/2019-02-20-golang-slice-struct.png)

从上图中，我们会发现切片与数组的关系非常密切，切片引入了一个抽象层，提供了对数组中部分连续片段的引用，而作为数组的引用，我们可以在运行区间可以修改它的长度和范围。当切片底层的数组长度不足时就会触发扩容，切片指向的数组可能会发生变化，不过在上层看来切片是没有变化的，上层只需要与切片打交道不需要关心数组的变化。

我们在上一节介绍过编译器在编译期间简化了获取数组大小、读写数组中的元素等操作：因为数组的内存固定且连续，多数操作都会直接读写内存的特定位置。但是切片是运行时才会确定内容的结构，所有操作还需要依赖 Go 语言的运行时
### 初始化
Go 语言中包含三种初始化切片的方式：

1. 通过下标的方式获得数组或者切片的一部分；
2. 使用字面量初始化新的切片；
3. 使用关键字 make 创建切片：

``` go
arr[0:3] or slice[0:3]
slice := []int{1, 2, 3}
slice := make([]int, 10)
```

#### 使用下标 
使用下标创建切片是最原始也最接近汇编语言的方式，它是所有方法中最为底层的一种，编译器会将 arr[0:3] 或者 slice[0:3] 等语句转换成 OpSliceMake 操作，我们可以通过下面的代码来验证一下：

``` go
package main

func newSlice() []int {
	arr := [3]int{1, 2, 3}
	slice := arr[0:1]
	return slice
}

func main() {
	s := newSlice()
	print(s)
}

```

通过 GOSSAFUNC 变量编译上述代码可以得到一系列 SSA 中间代码，其中 slice := arr[0:1] 语句在 “decompose builtin” 阶段对应的代码如下所示：
``` bash
GOSSAFUNC=main go build main.go
```


``` go
v27 (+5) = SliceMake <[]int> v11 v14 v17

name &arr[*[3]int]: v11
name slice.ptr[*int]: v11
name slice.len[int]: v14
name slice.cap[int]: v17
```
SliceMake 操作会接受四个参数创建新的切片，**元素类型**、**数组指针**、**切片大小**和**切片容量**，这也是我们在数据结构一节中提到的切片的几个字段 ，需要注意的是使用下标初始化切片不会拷贝原数组或者原切片中的数据，它只会创建一个指向原数组的切片结构体，所以修改新切片的数据也会修改原切片。

#### 字面量
当我们使用字面量 []int{1, 2, 3} 创建新的切片时，cmd/compile/internal/gc.slicelit 函数会在编译期间将它展开成如下所示的代码片段：

``` go
var vstat [3]int
vstat[0] = 1
vstat[1] = 2
vstat[2] = 3
var vauto *[3]int = new([3]int)
*vauto = vstat
slice := vauto[:]
```

- 根据切片中的元素数量对底层数组的大小进行推断并创建一个数组；
- 将这些字面量元素存储到初始化的数组中；
- 创建一个同样指向 [3]int 类型的数组指针；
- 将静态存储区的数组 vstat 赋值给 vauto 指针所在的地址；
- 通过 [:] 操作获取一个底层使用 vauto 的切片；

第 5 步中的 [:] 就是使用下标创建切片的方法，从这一点我们也能看出 [:] 操作是创建切片最底层的一种方法。

#### 关键字
如果使用字面量的方式创建切片，大部分的工作都会在编译期间完成。但是当我们使用 make 关键字创建切片时，很多工作都需要运行时的参与；调用方必须向 make 函数传入切片的大小以及可选的容量，类型检查期间的 cmd/compile/internal/gc.typecheck1 函数会校验入参：

``` go
	switch n.Op {
	...
	case OMAKE:
		args := n.List.Slice()

		i := 1
		switch t.Etype {
		case TSLICE:
			if i >= len(args) {
				yyerror("missing len argument to make(%v)", t)
				return n
			}

			l = args[i]
			i++
			var r *Node
			if i < len(args) {
				r = args[i]
			}
			...
			if Isconst(l, CTINT) && r != nil && Isconst(r, CTINT) && l.Val().U.(*Mpint).Cmp(r.Val().U.(*Mpint)) > 0 {
				yyerror("len larger than cap in make(%v)", t)
				return n
			}

			n.Left = l
			n.Right = r
			n.Op = OMAKESLICE
		}
	...
	}
}
```

上述函数不仅会检查 len 是否传入，还会保证传入的容量 cap 一定大于或者等于 len。除了校验参数之外，当前函数会将 OMAKE 节点转换成 OMAKESLICE，中间代码生成的 cmd/compile/internal/gc.walkexpr 函数会依据下面两个条件转换 OMAKESLICE 类型的节点：
- 切片的大小和容量是否足够小；
- 切片是否发生了逃逸，最终在堆上初始化
当切片发生逃逸或者非常大时，运行时需要 runtime.makeslice 在堆上初始化切片，如果当前的切片不会发生逃逸并且切片非常小的时候，make([]int, 3, 4) 会被直接转换成如下所示的代码：

``` go
var arr [4]int
n := arr[:3]
```

上述代码会初始化数组并通过下标 [:3] 得到数组对应的切片，这两部分操作都会在编译阶段完成，编译器会在栈上或者静态存储区创建数组并将 [:3] 转换成上一节提到的 OpSliceMake 操作。

分析了主要由编译器处理的分支之后，我们回到用于创建切片的运行时函数 runtime.makeslice，这个函数的实现很简单：

``` go
func makeslice(et *_type, len, cap int) unsafe.Pointer {
	mem, overflow := math.MulUintptr(et.size, uintptr(cap))
	if overflow || mem > maxAlloc || len < 0 || len > cap {
		mem, overflow := math.MulUintptr(et.size, uintptr(len))
		if overflow || mem > maxAlloc || len < 0 {
			panicmakeslicelen()
		}
		panicmakeslicecap()
	}

	return mallocgc(mem, et, true)
}
```
上述函数的主要工作是计算切片占用的内存空间并在堆上申请一片连续的内存，它使用如下的方式计算占用的内存：

`内存空间=切片中元素大小×切片容量`

虽然编译期间可以检查出很多错误，但是在创建切片的过程中如果发生了以下错误会直接触发运行时错误并崩溃：

- 内存空间的大小发生了溢出；
- 申请的内存大于最大可分配的内存；
- 传入的长度小于 0 或者长度大于容量；

`runtime.makeslice` 在最后调用的 `runtime.mallocgc` 是用于申请内存的函数，这个函数的实现还是比较复杂，如果遇到了比较小的对象会直接初始化在 Go 语言调度器里面的 P 结构中，而大于 32KB 的对象会在堆上初始化，我们会在后面的章节中详细介绍 Go 语言的内存分配器，这里就不展开分析了。

在之前版本的 Go 语言中，数组指针、长度和容量会被合成一个 runtime.slice 结构，但是从 cmd/compile: move slice construction to callers of makeslice 提交之后，构建结构体 reflect.SliceHeader 的工作就都交给了 runtime.makeslice 的调用方，该函数仅会返回指向底层数组的指针，调用方会在编译期间构建切片结构体：

``` go
func typecheck1(n *Node, top int) (res *Node) {
	switch n.Op {
	...
	case OSLICEHEADER:
	switch 
		t := n.Type
		n.Left = typecheck(n.Left, ctxExpr)
		l := typecheck(n.List.First(), ctxExpr)
		c := typecheck(n.List.Second(), ctxExpr)
		l = defaultlit(l, types.Types[TINT])
		c = defaultlit(c, types.Types[TINT])

		n.List.SetFirst(l)
		n.List.SetSecond(c)
	...
	}
}
```


OSLICEHEADER 操作会创建我们在上面介绍过的结构体 reflect.SliceHeader，其中包含数组指针、切片长度和容量，它是切片在运行时的表示：

``` go
type SliceHeader struct {
	Data uintptr
	Len  int
	Cap  int
}
```
正是因为大多数对切片类型的操作并不需要直接操作原来的 runtime.slice 结构体，所以 reflect.SliceHeader 的引入能够减少切片初始化时的少量开销，该改动不仅能够减少 ~0.2% 的 Go 语言包大小，还能够减少 92 个 runtime.panicIndex 的调用，占 Go 语言二进制的 ~3.5%1。

### 访问元素
使用 len 和 cap 获取长度或者容量是切片最常见的操作，编译器将这它们看成两种特殊操作，即 OLEN 和 OCAP，cmd/compile/internal/gc.state.expr 函数会在 SSA 生成阶段阶段将它们分别转换成 OpSliceLen 和 OpSliceCap：

``` go
func (s *state) expr(n *Node) *ssa.Value {
	switch n.Op {
	case OLEN, OCAP:
		switch {
		case n.Left.Type.IsSlice():
			op := ssa.OpSliceLen
			if n.Op == OCAP {
				op = ssa.OpSliceCap
			}
			return s.newValue1(op, types.Types[TINT], s.expr(n.Left))
		...
		}
	...
	}
}
```


切片的操作基本都是在编译期间完成的，除了访问切片的长度、容量或者其中的元素之外，编译期间也会将包含 range 关键字的遍历转换成形式更简单的循环，我们会在后面的章节中介绍使用 range 遍历切片的过程。

### 追加和扩容
使用 append 关键字向切片中追加元素也是常见的切片操作，中间代码生成阶段的 cmd/compile/internal/gc.state.append 方法会根据返回值是否会覆盖原变量，选择进入两种流程，如果 append 返回的新切片不需要赋值回原有的变量，就会进入如下的处理流程：

``` go
// append(slice, 1, 2, 3)
ptr, len, cap := slice
newlen := len + 3
if newlen > cap {
    ptr, len, cap = growslice(slice, newlen)
    newlen = len + 3
}
*(ptr+len) = 1
*(ptr+len+1) = 2
*(ptr+len+2) = 3
return makeslice(ptr, newlen, cap)
```

我们会先解构切片结构体获取它的数组指针、大小和容量，如果在追加元素后切片的大小大于容量，那么就会调用 runtime.growslice 对切片进行扩容并将新的元素依次加入切片。

如果使用 slice = append(slice, 1, 2, 3) 语句，那么 append 后的切片会覆盖原切片，这时 cmd/compile/internal/gc.state.append 方法会使用另一种方式展开关键字：

``` go
// slice = append(slice, 1, 2, 3)
a := &slice
ptr, len, cap := slice
newlen := len + 3
if uint(newlen) > uint(cap) {
   newptr, len, newcap = growslice(slice, newlen)
   vardef(a)
   *a.cap = newcap
   *a.ptr = newptr
}
newlen = len + 3
*a.len = newlen
*(ptr+len) = 1
*(ptr+len+1) = 2
*(ptr+len+2) = 3
```

是否覆盖原变量的逻辑其实差不多，最大的区别在于得到的新切片是否会赋值回原变量。如果我们选择覆盖原有的变量，就不需要担心切片发生拷贝影响性能，因为 Go 语言编译器已经对这种常见的情况做出了优化。
![](https://img.draveness.me/2020-03-12-15839729948451-golang-slice-append.png)

到这里我们已经清楚了 Go 语言如何在切片容量足够时向切片中追加元素，不过仍然需要研究切片容量不足时的处理流程。当切片的容量不足时，我们会调用 runtime.growslice 函数为切片扩容，扩容是为切片分配新的内存空间并拷贝原切片中元素的过程，我们先来看新切片的容量是如何确定的

``` go
func growslice(et *_type, old slice, cap int) slice {
	newcap := old.cap
	doublecap := newcap + newcap
	if cap > doublecap {
		newcap = cap
	} else {
		if old.len < 1024 {
			newcap = doublecap
		} else {
			for 0 < newcap && newcap < cap {
				newcap += newcap / 4
			}
			if newcap <= 0 {
				newcap = cap
			}
		}
	}
}
```

在分配内存空间之前需要先确定新的切片容量，运行时根据切片的当前容量选择不同的策略进行扩容：

- 如果期望容量大于当前容量的两倍就会使用期望容量；
- 如果当前切片的长度小于 1024 就会将容量翻倍；
- 如果当前切片的长度大于 1024 就会每次增加 25% 的容量，直到新容量大于期望容量；

上述代码片段仅会确定切片的大致容量，下面还需要根据切片中的元素大小对齐内存，当数组中元素所占的字节大小为 1、8 或者 2 的倍数时，运行时会使用如下所示的代码对齐内存：

``` go
	var overflow bool
	var lenmem, newlenmem, capmem uintptr
	switch {
	case et.size == 1:
		lenmem = uintptr(old.len)
		newlenmem = uintptr(cap)
		capmem = roundupsize(uintptr(newcap))
		overflow = uintptr(newcap) > maxAlloc
		newcap = int(capmem)
	case et.size == sys.PtrSize:
		lenmem = uintptr(old.len) * sys.PtrSize
		newlenmem = uintptr(cap) * sys.PtrSize
		capmem = roundupsize(uintptr(newcap) * sys.PtrSize)
		overflow = uintptr(newcap) > maxAlloc/sys.PtrSize
		newcap = int(capmem / sys.PtrSize)
	case isPowerOfTwo(et.size):
		...
	default:
		...
	}
```

runtime.roundupsize 函数会将待申请的内存向上取整，取整时会使用 runtime.class_to_size 数组，使用该数组中的整数可以提高内存的分配效率并减少碎片，我们会在内存分配一节详细介绍该数组的作用：

``` go
var class_to_size = [_NumSizeClasses]uint16{
    0,
    8,
    16,
    32,
    48,
    64,
    80,
    ...,
}
```
在默认情况下，我们会将目标容量和元素大小相乘得到占用的内存。如果计算新容量时发生了内存溢出或者请求内存超过上限，就会直接崩溃退出程序，不过这里为了减少理解的成本，将相关的代码省略了。


``` go
var arr []int64
arr = append(arr, 1, 2, 3, 4, 5)
```

简单总结一下扩容的过程，当我们执行上述代码时，会触发 runtime.growslice 函数扩容 arr 切片并传入期望的新容量 5，这时期望分配的内存大小为 40 字节；不过因为切片中的元素大小等于 sys.PtrSize，所以运行时会调用 runtime.roundupsize 向上取整内存的大小到 48 字节，所以新切片的容量为 48 / 8 = 6。

总之，切片扩容就是翻倍后还不够，需要多少扩多少，要是翻倍够而且原切片的大小小于1024就直接翻倍，要是大于1024就每次增长1/4，之后在进行内存对齐

### 拷贝切片
切片的拷贝虽然不是常见的操作，但是却是我们学习切片实现原理必须要涉及的。当我们使用 copy(a, b) 的形式对切片进行拷贝时，编译期间的 cmd/compile/internal/gc.copyany 也会分两种情况进行处理拷贝操作，如果当前 copy 不是在运行时调用的，copy(a, b) 会被直接转换成下面的代码：

``` go
n := len(a)
if n > len(b) {
    n = len(b)
}
if a.ptr != b.ptr {
    memmove(a.ptr, b.ptr, n*sizeof(elem(a))) 
}
```


上述代码中的 runtime.memmove 会负责拷贝内存。而如果拷贝是在运行时发生的，例如：go copy(a, b)，编译器会使用 runtime.slicecopy 替换运行期间调用的 copy，该函数的实现很简单：

``` go
func slicecopy(to, fm slice, width uintptr) int {
	if fm.len == 0 || to.len == 0 {
		return 0
	}
	n := fm.len
	if to.len < n {
		n = to.len
	}
	if width == 0 {
		return n
	}
	...

	size := uintptr(n) * width
	if size == 1 {
		*(*byte)(to.array) = *(*byte)(fm.array)
	} else {
		memmove(to.array, fm.array, size)
	}
	return n
}
```



## 哈希表

## 字符串



