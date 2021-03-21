# 类型

## 类型系统

~~~ go
type T struct {
	name string
}

func (t T) F1() {
	fmt.Println(t.name)
}

func main() {
	t := T{
		name: "eintr",
	}

	t.F1()
}
~~~

在go中这些是内置类型
- int8
- int16
- int32
- int64
- int
- byte
- string
- slice
- func
- map
...

而以下是自定义类型

~~~ go
type T int

type T struct {
    name string
}

type I interface {
    Name() string
}

~~~

> 注意

- 给内置类型递归伊方法是不被允许的，而接口类型是无效的方法接受者
- 数据类型虽然多，但是不管是内置类型还是自定义类型，都有对应的类型描述信息，称为`类型元数据`，每种类型元数据都是全局唯一的
- 这些类型元数据共同构成了go的`类型系统`

~~~ go
runtime._type

// 类型名称
// 类型大小
// 对齐边界
// 是否自定义

type _type struct {
	size       uintptr
	ptrdata    uintptr // size of memory prefix holding all pointers
	hash       uint32
	tflag      tflag
	align      uint8
	fieldAlign uint8
	kind       uint8
	// function for comparing objects of this type
	// (ptr to object A, ptr to object B) -> ==?
	equal func(unsafe.Pointer, unsafe.Pointer) bool
	// gcdata stores the GC type data for the garbage collector.
	// If the KindGCProg bit is set in kind, gcdata is a GC program.
	// Otherwise it is a ptrmask bitmap. See mbitmap.go for details.
	gcdata    *byte
	str       nameOff
	ptrToThis typeOff
}
~~~
`_type作为每个类型元数据的Header`，在`_type之后存储的各种类型额外需要描述的信息`

~~~ go
// []string

type slicetype struct {
    typ _type
    elem *_type//指向其存储的元素的类型元数据
}
~~~

如果是自定义类型的数据，折后免还会有一个`uncommontype`结构体

~~~ go

type uncommontype struct {
	pkgpath nameOff //包路径
	mcount  uint16 // number of methods
	xcount  uint16 // number of exported methods
	moff    uint32 // offset from this uncommontype to [mcount]method
	_       uint32 // unused
}
// 方法描述信息
type method struct {
	name nameOff
	mtyp typeOff
	ifn  textOff
	tfn  textOff
}
~~~

> 例子

~~~ go

type myslice []string

func (ms myslice) Len() {
	fmt.Println(len(ms))
}

func (ms myslice) Cap() {
	fmt.Println(cap(ms))
}
~~~

~~~ go
type MyType1 = int32 //给类型取别名

type MyType2 int32 //自定义类型
~~~

这两种类型的类型元数据是不同的

## 接口

### 空接口

~~~ go
//runtime.eface
type eface struct {
    _type *_type        //动态类型
    data unsafe.Pointer //动态值
}
~~~

~~~ go
var e interface{}

func main() {
	f, _ := os.Open("go.mod")
	e = f
}
~~~

### 非空接口

~~~ go
type iface struct {
    tab *itab
    data unsafe.Pointer
}

type itab struct {
    inter   *interfacetype
    _type   *_type      //动态类型
    hash    uint32      //类型哈希值
    _       [4]byte 
    func    [1]uintptr  //方法地址数组
}

type interfacetype struct {
    typ     _type
    pkgpath name
    mhdr    []imethod
}

~~~

**一旦接口类型确定了，动态类型确定了，`itab`就不会改变了，所以这个itab结构体是可复用的**
实际上Go会把用到的itab结构体缓存起来，用`<接口类型，动态类型的组合为key>`

## 类型断言

- 空接口.(具体类型)
~~~ go
var e interface{}

func main() {

	f, _ := os.Open("./go.mod")

	e = f

	r, ok := e.(*os.File)
	if !ok {
		fmt.Println("断言错误")
		return
	}

	fmt.Printf("%T\n", r)

}
~~~

- 非空接口.(具体类型)

~~~ go

func main() {

	var rw io.ReadWriter
	f, _ := os.Open("./go.mod")

	rw = f

	r, ok := rw.(*os.File)
	if !ok {
		fmt.Println("断言错误")
		return
	}

	fmt.Printf("%T\n", r)

}
~~~

- 空接口.(非空接口)

~~~ go
func main() {

	var e interface{}
	f, _ := os.Open("./go.mod")

	e = f

	r, ok := e.(io.ReadWriter)
	if !ok {
		fmt.Println("断言错误")
		return
	}

	fmt.Printf("%T\n", r)

}
~~~

~~~ go
func main() {

	var e interface{}

	e = "eintr"

	r, ok := e.(io.ReadWriter)
	if !ok {
		fmt.Println("断言错误")
		return
	}

	fmt.Printf("%T\n", r)

}
~~~

- 非空接口.(非空接口)

~~~ go
type eintr struct {
	name string
}

func (this *eintr) Write(p []byte) (n int, err error) {
	return 0, nil
}

func main() {
	var w io.Writer

	f := eintr{
		name: "eintr",
	}

	w = &f

	rw, ok := w.(io.ReadWriter)
	if !ok {
		fmt.Println("断言失败")
		return
	}

	fmt.Printf("%T\n", rw)

}

~~~

### 小结

**类型断言的关键在于，明确接口的动态类型，以及对应的类型实现了哪些方法，而明确这些的关键还是`类型元数据`以及空接口与非空接口的数据结构**

## 反射

反射的作用就是把类型元数据暴露给用户使用

~~~ go

func TypeOf(i interface{}) Type {
	eface := *(*emptyInterface)(unsafe.Pointer(&i))
	return toType(eface.typ)
}
~~~


~~~ go
type Type interface {
    // Methods applicable to all types.

    // Align returns the alignment in bytes of a value of
    // this type when allocated in memory.
    Align() int

    // FieldAlign returns the alignment in bytes of a value of
    // this type when used as a field in a struct.
    FieldAlign() int

    // Method returns the i'th method in the type's method set.
    // It panics if i is not in the range [0, NumMethod()).
    //
    // For a non-interface type T or *T, the returned Method's Type and Func
    // fields describe a function whose first argument is the receiver.
    //
    // For an interface type, the returned Method's Type field gives the
    // method signature, without a receiver, and the Func field is nil.
    //
    // Only exported methods are accessible and they are sorted in
    // lexicographic order.
    Method(int) Method

    // MethodByName returns the method with that name in the type's
    // method set and a boolean indicating if the method was found.
    //
    // For a non-interface type T or *T, the returned Method's Type and Func
    // fields describe a function whose first argument is the receiver.
    //
    // For an interface type, the returned Method's Type field gives the
    // method signature, without a receiver, and the Func field is nil.
    MethodByName(string) (Method, bool)

    // NumMethod returns the number of exported methods in the type's method set.
    NumMethod() int

    // Name returns the type's name within its package for a defined type.
    // For other (non-defined) types it returns the empty string.
    Name() string

    // PkgPath returns a defined type's package path, that is, the import path
    // that uniquely identifies the package, such as "encoding/base64".
    // If the type was predeclared (string, error) or not defined (*T, struct{},
    // []int, or A where A is an alias for a non-defined type), the package path
    // will be the empty string.
    PkgPath() string

    // Size returns the number of bytes needed to store
    // a value of the given type; it is analogous to unsafe.Sizeof.
    Size() uintptr

    // String returns a string representation of the type.
    // The string representation may use shortened package names
    // (e.g., base64 instead of "encoding/base64") and is not
    // guaranteed to be unique among types. To test for type identity,
    // compare the Types directly.
    String() string

    // Kind returns the specific kind of this type.
    Kind() Kind

    // Implements reports whether the type implements the interface type u.
    Implements(u Type) bool

    // AssignableTo reports whether a value of the type is assignable to type u.
    AssignableTo(u Type) bool

    // ConvertibleTo reports whether a value of the type is convertible to type u.
    ConvertibleTo(u Type) bool

    // Comparable reports whether values of this type are comparable.
    Comparable() bool

    // Methods applicable only to some types, depending on Kind.
    // The methods allowed for each kind are:
    //
    //	Int*, Uint*, Float*, Complex*: Bits
    //	Array: Elem, Len
    //	Chan: ChanDir, Elem
    //	Func: In, NumIn, Out, NumOut, IsVariadic.
    //	Map: Key, Elem
    //	Ptr: Elem
    //	Slice: Elem
    //	Struct: Field, FieldByIndex, FieldByName, FieldByNameFunc, NumField

    // Bits returns the size of the type in bits.
    // It panics if the type's Kind is not one of the
    // sized or unsized Int, Uint, Float, or Complex kinds.
    Bits() int

    // ChanDir returns a channel type's direction.
    // It panics if the type's Kind is not Chan.
    ChanDir() ChanDir

    // IsVariadic reports whether a function type's final input parameter
    // is a "..." parameter. If so, t.In(t.NumIn() - 1) returns the parameter's
    // implicit actual type []T.
    //
    // For concreteness, if t represents func(x int, y ... float64), then
    //
    //	t.NumIn() == 2
    //	t.In(0) is the reflect.Type for "int"
    //	t.In(1) is the reflect.Type for "[]float64"
    //	t.IsVariadic() == true
    //
    // IsVariadic panics if the type's Kind is not Func.
    IsVariadic() bool

    // Elem returns a type's element type.
    // It panics if the type's Kind is not Array, Chan, Map, Ptr, or Slice.
    Elem() Type

    // Field returns a struct type's i'th field.
    // It panics if the type's Kind is not Struct.
    // It panics if i is not in the range [0, NumField()).
    Field(i int) StructField

    // FieldByIndex returns the nested field corresponding
    // to the index sequence. It is equivalent to calling Field
    // successively for each index i.
    // It panics if the type's Kind is not Struct.
    FieldByIndex(index []int) StructField

    // FieldByName returns the struct field with the given name
    // and a boolean indicating if the field was found.
    FieldByName(name string) (StructField, bool)

    // FieldByNameFunc returns the struct field with a name
    // that satisfies the match function and a boolean indicating if
    // the field was found.
    //
    // FieldByNameFunc considers the fields in the struct itself
    // and then the fields in any embedded structs, in breadth first order,
    // stopping at the shallowest nesting depth containing one or more
    // fields satisfying the match function. If multiple fields at that depth
    // satisfy the match function, they cancel each other
    // and FieldByNameFunc returns no match.
    // This behavior mirrors Go's handling of name lookup in
    // structs containing embedded fields.
    FieldByNameFunc(match func(string) bool) (StructField, bool)

    // In returns the type of a function type's i'th input parameter.
    // It panics if the type's Kind is not Func.
    // It panics if i is not in the range [0, NumIn()).
    In(i int) Type

    // Key returns a map type's key type.
    // It panics if the type's Kind is not Map.
    Key() Type

    // Len returns an array type's length.
    // It panics if the type's Kind is not Array.
    Len() int

    // NumField returns a struct type's field count.
    // It panics if the type's Kind is not Struct.
    NumField() int

    // NumIn returns a function type's input parameter count.
    // It panics if the type's Kind is not Func.
    NumIn() int

    // NumOut returns a function type's output parameter count.
    // It panics if the type's Kind is not Func.
    NumOut() int

    // Out returns the type of a function type's i'th output parameter.
    // It panics if the type's Kind is not Func.
    // It panics if i is not in the range [0, NumOut()).
    Out(i int) Type

    common() *rtype
    uncommon() *uncommonType
}

~~~

> 实例

~~~ go
package main

import (
	"TypeSystem/Reflect/eintr"
	"fmt"
	"reflect"
)

func main() {
	u1 := &eintr.Eintr{
		Name: "eintr",
	}

	u1.Setname("szc")

	fmt.Println(u1.Getname())

	t := reflect.TypeOf(*u1)

	fmt.Println(t.Name(), t.NumMethod())
}
~~~

~~~ go
package eintr

type Eintr struct {
	Name string
}

func (e *Eintr) Getname() (name string) {
	name = e.Name
	return
}

func (e *Eintr) Setname(name string) {
	e.Name = name
	return
}
~~~
