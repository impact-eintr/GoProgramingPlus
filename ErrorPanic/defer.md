# defer

~~~ go
func A(){
    defer B()
    //code to do sth
}

~~~

~~~ go
func A(){
    r = deferproc(8,B)//defer注册

    //与panic recover有关先忽略
    //if r > 0 {
    //    goto ret
    //}

    // code to do sth

    runtime.deferreturn()
    return

    //与panic recover有关先忽略
//ret:
//    runtime.deferreturn
}
~~~

~~~ go
func A(){
    //defer注册
    r = deferproc(8,B)

    // code to do sth

    //defer执行
    runtime.deferreturn()
    return
}
~~~

**正是defer的先注册后调用才表现出defer延时执行的效果**

> 原理


defer信息会先注册到一个链表，**而当前执行的goroutine持有这个链表的头指针**
每个goroutine在运行是都有一个对应的结构体`g`,其中有一个字段指向defer链表头


~~~ go
func A1(a int) {
	fmt.Println(a)
}

func main() {
	a, b := 1, 2
	defer A1(a)

	a = a + b
	fmt.Println(a, b)
}
~~~
> deferproc 原型

~~~ go
func deferproc(siz int32,fn *funcval)
~~~

第一个参数是defer函数A1的参数加上返回值共占用多大空间,第二个参数是*funcval,没有捕获列表的funcval会优化，会在只读数据段分配一个共用的duncval结构体


> defer结构体

~~~ go
type _defer struct{
    siz int32   //参数和返回值共占用多少字节
    started bool//是否已经执行
    sp uintptr  //调用者栈指针
    pc uintptr  //返回地址
    fn *funval  //注册的函数
    _panic *_panic
    link *defer //next _defer
}
~~~

~~~ go
func main(){
    a, b := 1, 2
    //defer注册
    r = deferproc(8,B)

    a = a + b
    fmt.Println(a, b)

    //defer执行
    runtime.deferreturn()
    return
}

~~~

