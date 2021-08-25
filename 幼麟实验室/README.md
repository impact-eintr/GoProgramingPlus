# string

### 从编码说起
- 一个字节由8个比特组成，当比特位全为0时代表数字0，全为1时代表数字255，一个字节可以表示256个数字，2个字节可以表示65536个数字。
- 而字符的表示方法与之不同，是通过将字符进行编号，比如将`A`编号为`65`，对应二进制的`01000001`是这个字符的编码，通过这种映射关系可以将字符以比特的形式存起来，而这种映射关系就是字符
集。
- 常见的字符集有： ASCII GB18030 GBK Unicode等等
**字符集促成了字符与二进制的合作，但如何表示字符串呢？**

`hello世界`

|字符|编号|二进制|
|--|--|--|
|h|104|0110 1000|
|e|101|0110 0101|
|l|108|0110 1100|
|l|108|0110 1100|
|o|111|0110 0111|
|世|19990|01001110 00010110|
|界|30028|01110101 01001100|

如果不加处理直接存放就是`01101000|01100101|01101100|01101100|01100111|0100111000010110|0111010101001100`，可以发现如果将分隔符去掉我们完全不知道这些比特串将要表达什么意思。
#### 解决方案
- 定长编码
无论字符的二进制表示原本有多长，全部按字符集最长的编码来比如 `e`:`00000000 00000000 00000000 01100101`
但这样明显是在浪费内存
- 变长编码
小编号字符少占用字节，大编号字符多占用字节

|编号|编码模板|
|---------|----------|
|`[0,127]`|`0???????`|
|`[128,2047]`|`110????? 10??????`|
|`[2048,65536]`|`1110???? 10?????? 10??????`|

比如对于`世` 编号`19990` 二进制表示`01001110 00010110` 变长编码 1110 `0100` 10`111000` 10`010110`

**以上编码方式就是所谓的`UTF-8`编码，也就是GO语言默认的编码方式**

### string in golang
`var str string = "hello"`
在c中字符串变量存放着一块以`\0`结尾的连续内存的起始地址，在go中字符串变量不光存放这个地址，还存放着这块连续内存用多少个字节(源码包 `src/runtime/string/string.go:stringStruct`)
~~~ go
type stringStruct struct {
	str unsafe.Pointer
	len int // go的int 8Byte
}
~~~

string的数据结构跟切片有些类似，只不过切片还有一个表示容量的成员，事实上string和切片准确来说是`[]byte`经常发生交换

~~~ go
var str string
str = "Hello 世界"
~~~
字符串生成时，会先构建stringStruct对象，在转换成string，源码：
~~~ go
func gostringnocopy(str *byte) string {
	ss := stringStruct{
		str: unsafe.Pointer(str),
		len: findnull(str),
		}
	s := *(*string)(unsafe.Pointer(&ss))
	return s
}
~~~
#### 字符串拼接
字符串可以方便地拼接`str := "Str1" + "Str2`
即便有非常多的字符串需要拼接，性能上也有比较好的保证，因为新字符串的内存空间是**一次分配**完成的，所以性能消耗主要在拷贝数据上
在`runtime`包中，使用concatstrings()函数来拼接字符串。在一个拼接语句中，所有待拼接的字符串都被编译器组织到一个切片中传入concatstrings函数，拼接的过程需要遍历两次切片 ，第一次获取长度来申请内存，第二次将字符逐个拷贝过去（以下是伪代码）
~~~ go
func concatstrings(a []string) string{
	length := 0
	for _,str := range a {
		length += len(str)
	}
	//分配内存 返回一个string 和 slice 它们共享内存
	s,b := rawstring(length)
	//string无法修改 可以通过切片修改
	for _,str := range a{
		copy(b,str)
		b = b[len(str):]
	}
	return s
}
~~~
rawstring() 的源码，初始化一个指定大小string同时返回一个切片二者共享同一块内存空间，后面向切片中拷贝数据，就间接地修改了string
~~~ go
func rawstring(size int) (s string, b []byte) {
	p := mallocgc(uintptr(size), nil, false)

	stringStructOf(&s).str = p
	stringStructOf(&s).len = size

	*(*slice)(unsafe.Pointer(&b)) = slice{p, size, size}

	return
}
~~~

#### 类型转换
##### []byte 转 string
~~~ go
func slicebytetostring(buf *tmpBuf, b []byte) (str string) {
	
	//非主要代码
	
	var p unsafe.Pointer
	if buf != nil && len(b) <= len(buf) {
		//预留空间够就用预留空间
		p = unsafe.Pointer(buf)
	} else {
		//预留空间不够就申请新内存
		p = mallocgc(uintptr(len(b)), nil, false)
	}
	//构建字符串
	stringStructOf(&str).str = p
	stringStructOf(&str).len = len(b)
	//将切片底层数组拷贝到字符串
	memmove(p, (*(*slice)(unsafe.Pointer(&b))).array, uintptr(len(b)))
	return
}

~~~
##### string 转 []byte

~~~ go
func stringtoslicebyte(buf *tmpBuf, s string) []byte {
	var b []byte
	if buf != nil && len(s) <= len(buf) {
		*buf = tmpBuf{}
		b = buf[:len(s)]
	} else {
		//申请未经初始化的切片（string的内容会完全覆盖切片）
		b = rawbyteslice(len(s))
	}
	copy(b, s)
	return b
}
~~~
# 总结
<font color=#999AAA >
golang的string不包含内存空间，只有一个内存的指针，这样的好处是String变得非常轻量，可以方便地进行传递而不用担心内存拷贝

# slice

## slice的特性
又称动态数组，依托数组实现，可以方便地进行扩容和传递，实际使用时比数组更灵活。
~~~ go
type slice struct {
	array unsafe.Pointer
	len   int
	cap   int
}
~~~
以上是go中slice的声明。
### 初始化
- `var ints []int`
- `ints := make([]int , 2 ,5)`  // make 会分配底层数组
- `ps := new([]string)`  //注意这里的ps是一个指针 new不会分配底层数组
- `array := [5]int{1,2,3,4,5} 	s1 := array[0:3]  s2 := s1[0:1]`
### 切片操作与切片表达式
##### 简单表达式
`a[low : hihg]`
如果 a 为数组或者切片，则该表达式将切取 `a[ low , high ) `的元素，如果 a  为string 该表达式将会生成一个string,而不是slice
~~~ go
a := [5]int{1,2,3,4,5}
b := a[1:4]
b[0] = 2
c := b[1:2]

~~~
根据之前切片结构的声明，我们知道 slice 有三个元素，对于array(底层数组地址)，着重强调，使用简单表达式生成的slice将与原数组或slice共享底层数组，新切片的生成逻辑可以理解为
~~~ go
b.array = &a[low]
b.len = heigh - low
b.cap = len(a) - low //注意 b 的 cap 不是 len(b)
~~~
大家可以试试以下代码的输出
~~~ go
func SlicePrint() {
        s1 := []int{1, 2}
        s2 := s1
        s3 := s2[:]
        fmt.Printf("%p %p %p\n", &s1[0], &s2[0], &s3[0])
}

func SliceCap() {
        var array [10]int{1,2,3,4,5,6,7,8,9}
        var slice = array[5:6]
		var slice2 = array[9:10]
		
        fmt.Printf("len(slice) = %d\n", len(slice))
        fmt.Printf("cap(slice) = %d\n", cap(slice))
        
        fmt.Printf("len(slice) = %d\n", len(slice2))
        fmt.Printf("cap(slice) = %d\n", cap(slice2))  
}
~~~
在上面的例子中如果给slice2添加新元素 `slice2 = append(slice2,10)`，原本的底层数组就不够用了，这时go会分配一段的内存空间作为底层数组，并将slice中的元素拷贝到新的数组中然后将新添加的元素加到数组中，而这段新的内存有多大呢，这在一会儿的实现原理中说。

另外，需要注意，如果简单表达式的对象是slice，**那么表达式a[low : high]中 low 和 high 的最大值可以是 a 的容量，而不是 a 的长度**。

##### 扩展表达式
`a[low : high : max]`
简单表达是生成的新slice与原数组共享底层数组避免了拷贝元素，节约内存空间的同时可能会带来一定的风险。
新slice （`b := a[low:high]`）不仅仅可以读写 a[low] 到 a[high-1] 的元素，而且在使用`append(a,x)`添加新的元素还会覆盖掉 a[high]以及后面的元素
~~~ go
a  := [5]slice{1,2,3,4,5}
b := a[1:4]
b = append(b,0)
fmt.Println(a[4]) //0
~~~
而扩展表达式就是解决这个问题的机制 ，low high max 满足 `0 <= low <= high <= max <= cap(a)` `max`用于限制新生成切片的容量，新切片的容量为 `max - low`

~~~ go
array := [10]int
a := array[5:7] //cap = 5
b := array[5:7:7] //cap = 2
~~~

## slice的实现原理
slice的使用很灵活，但是想要正确使用它，就要了解它的实现原理。
- `var ints []int`   slice{array : nil , len : 0 , cap : 0}
- `ints := make([]int , 2 ,5)` slice {array : 一段连续内存的起始地址(同时将元素全部初始化为整型的默认值 0 ) , len : 2 , cap : 5}  

	- slice元素的访问
		~~~ go
		ints := make([]int,2,5)
		fmt.Pritln(ints[0])
		ints = append(ints , 3)
		// ints 的底层数组变化为 0 0 3 0 0 但只可以访问前三个元素 之后的属于越界访问
		~~~
- `ps := new([]string)` slice{array : nil , len : 0 , cap : 0}
	- 这里`ps` 就是slice结构体的**起始地址**，这时slice还没有分配底层数组，如果想要向slice中添加元素需要使用内置函数`append()`  `*ps = append(*ps , "hello世界")` 这样 *ps.array 指向的就是一个 `stringStruct` stringStruct{ str : 底层数组起始地址, len : 11}

slice 依托数组实现，底层数组对用户屏蔽在底层数组容量不足时可以实现自动分配并生成新的slice
#### 扩容
扩容容量的选择遵守以下规则
- 如果原slice的容量翻倍后仍然小于最低容量需求 cap ，就直接扩容到 cap
- 否则，如果原slice的容量小于1024，则新slice的容量将扩大为原来的2倍
- 如果原slice的容量大于等于1024，则新slice的容量将扩大为原来的1.25倍
`src/runtime/slice.go:growslice`
~~~ go
	newcap := old.cap
	doublecap := newcap + newcap
	if cap > doublecap {
		newcap = cap
	} else {
		if old.len < 1024 {
			newcap = doublecap
		} else {
			// Check 0 < newcap to detect overflow
			// and prevent an infinite loop.
			for 0 < newcap && newcap < cap {
				newcap += newcap / 4
			}
			// Set newcap to the requested cap when
			// the newcap calculation overflowed.
			if newcap <= 0 {
				newcap = cap
			}
		}
	}

~~~

在此规则之上，还会考虑元素类型与内存分配规则，对实际扩张值做一些微调。比如，os常常将内存切分为 64、80、96、112 等常用的大小 go的内存分配机制会根据预估大小匹配合适的内存块分配给新slice。


# Map

## hash与buckets
说到键值对的存储，我们就会想到哈希表，哈希表通常会有一堆桶来存储键值对，一个键值对来了自然要存到一个桶中。首先将key通过hash()处理一下得到一个hash值，现在要利用这个hash值从m个桶中选择一个，桶的编号区间 [0,m-1] 。
![在这里插入图片描述](https://img-blog.csdnimg.cn/20210129173159537.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FxXzQ0ODA3Nzk2,size_16,color_FFFFFF,t_70#pic_center)
### 取模法 
`hash%m`
### 与运算法 
`hash & (m-1)`

**想要使用与运算法就要限制桶的个数 m 必须是2 的整数次幂，这样 m 的二进制表示一定只有一位为1，(m-1) 就是除了最高为其他均为 1 ，避免一些桶绝对不会被选中的情况**

如果之后有其他键值对也选择了同一个桶，就是发生了**哈希冲突**，解决方案一般是 **开放地址法** 和 **拉链法**
- 开放地址法 就是在发生冲突的桶后找没有被占用的桶来存放键值对，查找时定位到桶后hash不匹配，就继续向下找直到遇到空桶。
- 如图
![在这里插入图片描述](https://img-blog.csdnimg.cn/20210129175642621.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FxXzQ0ODA3Nzk2,size_16,color_FFFFFF,t_70#pic_center)
### 解决hash冲突
- 发生hash冲突会影响hash表的读写效率，选择散列均匀的 hash() 可以减少hash冲突的发生
- 适时对hash表进行扩容意识保障读写效率的有效手段，通常会把`（count of key : value） / m`作为是否需要扩容的判断依据，这个比值被称作`负载因子`

当需要扩容时，就要分配更多的桶，将旧桶中的数据迁移时，为了避免一次性迁移大量数据带来的性能损耗，通常会在hash表扩容时先分配足够多的新桶，然后用一个字段记录旧桶的位置，再加一个字段记录旧桶迁移的进度(记录下一个要迁移的旧桶编号)，在每次hash表读写操作时，如果检测到当前正处于扩容阶段就完成一部分键值对迁移任务，直到所有的键值对迁移完成，旧桶不再使用，算完成了一次扩容操作，这就是所谓`渐进式扩容`

## map in golang
`src/runtime/map.go:hmap`
~~~ go
type hmap struct {
	count     int // 已经存储的键值对数目
	flags     uint8
	B         uint8  // 桶的数目是 2^B golang使用了与运算法
	noverflow uint16 // 溢出桶数量
	hash0     uint32 // hash seed

	buckets    unsafe.Pointer // 桶在哪儿
	oldbuckets unsafe.Pointer // 扩容阶段的旧桶位置记录
	nevacuate  uintptr        // 扩容时，下一个要迁移的旧桶编号

	extra *mapextra // optional fields
}
~~~
再来看看map用的桶长什么样。也就是 bmap 
- 一个桶中可以存放8个键值对，但是为了让内存排列更加紧凑，采用 8个键+8个值 的存放方式，在键值对之上是8个 `tophash` 每个tophash都是对应hash值的高8位，最后的 `voerflow`bmap指针指向`溢出桶`，`溢出桶`的结构与常规桶相同，是为了减少扩容次数而引入的。当一个桶存满了，还有可用的溢出桶是时，就会在同种桶后链接一个溢出桶，实际上，如果哈希表要分配的桶的数目大于`2^4`时就认为使用到溢出桶的几率较大，就会预分配 `2^(B-4)`个溢出桶备用，这些溢出桶和常规桶在内存中是连续的 ，前 `2^B`用作常规桶，后面的用于溢出桶


~~~ go
type mapextra struct {

	overflow    *[]*bmap   // 目前已经被使用的溢出桶的地址
	oldoverflow *[]*bmap // 扩容阶段存储存储旧桶用到的那些溢出桶的地址

	nextOverflow *bmap  //下一个空闲溢出桶
}
~~~
![在这里插入图片描述](https://img-blog.csdnimg.cn/20210130103253374.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FxXzQ0ODA3Nzk2,size_16,color_FFFFFF,t_70#pic_center)



![在这里插入图片描述](https://img-blog.csdnimg.cn/20210130114105143.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FxXzQ0ODA3Nzk2,size_16,color_FFFFFF,t_70#pic_center)
如果将这个桶存满的话，接下来再继续存储新的键值对时，这个hash表是会创建溢出桶还是会发生扩容呢?
### map的扩容规则
Go语言的map默认负载因子(count / (2^B))是 6.5 （即平均每个buckets存储的键值对达到6.5个以上）,超过这个值就会发生翻倍扩容，分配新桶的数目是旧桶的两倍 
- `hamp`中的`buckets`指向新分配的两个桶
- `oldbuckets`指向旧桶，`bevacuate`为0表示接下来要迁移编号为0的旧桶
- 过程如图所示
![在这里插入图片描述](https://img-blog.csdnimg.cn/20210201202628820.jpg?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FxXzQ0ODA3Nzk2,size_16,color_FFFFFF,t_70#pic_center)
还有一种情况也会触发扩容，就是负载因子没有超标，但是溢出桶使用较多。
- `B` <= 15 `noveflow` > 2^B
- `B` > 15   `neverflow` >= 2^15

这种情况针对的是在某些极端情况下，例如经过大量的元素增删后，兼职对刚好集中在一小部分buckets中，这时会发生**等量扩容**，即buckets不变，重新做一次类似增量扩容的迁移动作，将松散的键值对重新排列一次。


# 函数调用栈

# 闭包

``` go
func A() {

}

func B(f func()) {

}

func C() func() {
    return A
}

var f func() = C()
```

上面的函数 函数参数 函数返回值 变量都是`function value`

`function value` 本质上是一个指向代码段函数指令入口的`指针` `runtime.funcval`

``` go
type funcval {
    fn uinptr
}
```

``` go
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
```

可以直接使用函数入口地址，却要使用二级指针传递，为的是处理`闭包`

``` go
func create() func() int {
        c := 2 // 捕获变量
        return func() int {
                c++
                return c
        }
}

func create2() (fs [2]func()) {
        for i := 0; i < 2; i++ {
                fs[i] = func() {
                        fmt.Println(i)
                }
        }
        return
}

func main() {
        f1 := create()
        f2 := create()
        fmt.Println(f1())
        fmt.Println(f1())
        fmt.Println(f1())
        fmt.Println(f2())
        f3 := create2()
        f3[0]()
        f3[1]()
}

```

闭包对象的指令自然在编译阶段生成，但因为每个闭包对象都要保存自己的捕获变量，所以要到执行阶段才会创建对应的闭包对象
通过f1和f2调用闭包函数，就会找到各自对应的funcval结构体，拿到同一个函数入口，但是通过不同的函数变量调用时会使用对应的捕获列表, 这就是称闭包为有状态的函数的原因了
go通过一个`function value`调用函数时，会把对应的`funcval`结构体地址存入`DX`寄存器, 然后加上相应的偏移来找到每一个被捕获的变量，所以go中`闭包`就是有捕获列表的`function value`

被闭包捕获的变量，要在外层函数和闭包函数中表现一致，好像他们在使用同一个变量

#### 值拷贝
被捕获的变量除了初始化赋值外，没有被修改过，则直接拷贝

#### 返回值
调用这栈帧上依然会分配返回值空间，不过闭包的外层函数会在堆上也分配一个，外层函数和闭包函数都使用堆上这一个，但是到外层函数返回前吗，需要吧堆上的返回值拷贝到栈上的返回值空间

#### 参数堆分配
参数依然通过调用者栈帧传入，但是编译器会把栈上的这个参数拷贝到堆上一份，然后外层函数和闭包函数都使用堆上分配的这一个

#### 变量逃逸
局部变量i 改为堆分配，在栈上只存一个地址，这样闭包函数就和外层函数操作同一个变量了

# 方法

``` go
type A struct {
	name string
}

func (a A) Name() string {
	a.name = "Hi! " + a.name
	return a.name
}

func main() {
	a := A{name: "eintr"}
	fmt.Println(a.Name()) // 是语法糖 == A.Name(a)
	fmt.Println(A.Name(a))
}
```

这里的`a`就是所谓的方法接收者

``` go
func main() {
	t1 := reflect.TypeOf(A.Name)
	t2 := reflect.TypeOf(NameOfA)
	fmt.Println(t1 == t2)
}

type A struct {
	name string
}

func (a A) Name() string {
	a.name = "Hi! " + a.name
	return a.name
}

func NameOfA(a A) string {
	a.name = "Hi " + a.name
	return a.name
}

```

这里的pa 是指针接收者
``` go
type A struct {
	name string
}

func (a A) GetName() string {
	return a.name
}

func (pa *A) SetName() string {
	pa.name = "Hi " + pa.name
	return pa.name
}

func main() {
	a := A{name: "eintr"}
	pa := &a

	fmt.Println(pa.GetName()) // == (*pa).GetName()
	fmt.Println(a.SetName())  // == (&a).SetName()
}

```


``` go
type A struct {
	name string
}

func GetName(a A) string {
	return a.name
}

func main() {
	a := A{name: "eintr"}

	f1 := GetName
	f1(a)
}

```


``` go

type A struct {
	name string
}

func (a A) GetName() string {
	return a.name
}

func Getname(a A) string {
	return a.name
}

func main() {
	a := A{name: "eintr"}

	// func GetName(a A) string{} == func (a A) GetName() string {}
	f1 := A.GetName // 方法表达式 相当于方法原型
	f2 := a.GetName // 方法变量
	f1(a)
	f2()
	t1 := reflect.TypeOf(f1) // func(Main.A) string
	t2 := reflect.TypeOf(f2) // func() string
	fmt.Println(t1 == t2, t1, t2)
}

```


``` go
type A struct {
	name string
}

func (a A) GetName() string {
	return a.name
}

func GetFunc1() func() string {
	a := A{name: "eintr in GetFunc"}
	return a.GetName // 返回一个方法变量
}

// 以上函数等价于
func GetFunc2() func() string {
	a := A{name: "eintr in GetFunc"}

	return func() string { // 这里我们可以清晰地看到闭包是如何形成的
		return A.GetName(a) // 执行时 捕获了局部变量a 所以输出 eintr in GetFunc
	}
}

func Getname(a A) string {
	return a.name
}

func main() {
	a := A{name: "eintr in Main"}

	f2 := a.GetName // A.GetName(a)
	fmt.Println(f2())

	f3 := GetFunc1()
	fmt.Println(f3())

	f4 := GetFunc2()
	fmt.Println(f4())

}

```

# defer
关于defer 我们知道它会在函数返回之前 逆序执行

``` go
func main() {
	defer func() {
		fmt.Println(1)
	}()

	defer func() {
		fmt.Println(2)
	}()

	defer func() {
		fmt.Println(3)
	}()

}

```

# panic和recover

# 类型系统

# GPM

# GC

# Mutex

# 信号量

# 方法集

# 抢占式调度

