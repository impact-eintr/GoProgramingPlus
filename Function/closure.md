# 闭包

## 函数的特殊地位
函数在go中是头等对象
~~~ go
func A(){
    ...
}
~~~
==作为参数传递==
~~~ go
func B(f func()){
    ...
}
~~~
==作为返回值==
~~~ go
func C() func() {
    return A
}
~~~

==绑定到变量==
~~~ go
var f func() = C()
~~~
> go称这样的参数、返回值或变量为`function value`

函数的指令在编译期间生成，而`function value`本质上是一个指针,**但是并不直接指向函数指令入口，而是指向一个`runtime.funcval`结构体**,里面存着这个函数指令的入口地址
~~~ go
type funcval struct {
    fn uintprt
}
~~~

> example
~~~ go
func A(i int) {
	i++
	fmt.Println(i)
}

func B() {
	f1 := A
	f1(1)
}

func C() {
	f2 := A
	f2(1)
}

func main() {
	B()
	C()
}
~~~
这种情况编译器会做出优化，让f1 f2 共用一个`funcval`结构体

## 闭包

为什么要通过`funcval`结构体包装这个地址，然后使用一个二级指针来调用呢？这里主要时为了处理闭包的情况

- 必须有函数外部定义，但在函数内部使用的`自由变量`
- 脱离了形成闭包的上下文，闭包也能照常使用这些变量
> example
~~~ go
func create() func() int {
	c := 666
	return func() int {
		return c
	}
}

func main() {
	f1 := create()
	f2 := create()
	fmt.Println(f1())
	fmt.Println(f2())
}
~~~
通常称`c` 为`捕获变量`
闭包函数的指令自然也在编译阶段生成，但因为每个闭包对象都要保存自己的捕获变量，所以要到之心阶段才创建对应的闭包对象



- 因为每个闭包对象调用时使用的捕获列表不同，所以称闭包为有状态的函数

### 捕获列表
go中通过一个`function value`调用函数时，会把对应的`funcval`结构体地址存入特定的寄存器中(例如amd64平台的DX寄存器)，这样在闭包函数中，就可以通过寄存器取出funcval结构体的地址，在加上偏移取出每一个被捕获的变量
**go汇总闭包就是有捕获列表的function value**,没有捕获列表的fuction value直接忽略这个寄存器的值就可以

> 捕获列表不简单 不是拷贝值那么简单
**被闭包捕获的变量，要在外层函数和壁报函数中表现一致,好像在使用同一个变量一样**

~~~ go
func create2() (fs [2]func()) {
	for i := 0; i < 2; i++ {
		fs[i] = func() {
			fmt.Println(i)
		}
	}
	return
}

func main() {
	fs := create2()
	for i := 0; i < len(fs); i++ {
		fs[i]()
	}
}

~~~

闭包导致的局部变量堆分配，也是变量逃逸的一种场景

如果被捕获的是函数参数，涉及到函数原型

参数仍然通过调用者栈帧传入，但时编译器会把栈上这个参数拷贝到堆上一份,然后外层函数和壁报函数都使用堆上分配的这一个

如果被捕获的时返回值，处理方式又有些不同，调用者栈帧上依然会分配返回值空间，不过壁报的外层函数会在对上也分配一个，外层函数和闭包函数都使用堆上这一个,但是到外层函数返回前，需要把堆上的返回值拷贝到栈上的返回值空间


