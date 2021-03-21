# 方法

~~~ go
type A struct {
	name string
}

func (a A) Name() string {
	a.name = "Hi! " + a.name
	return a.name
}

func main() {
	a := A{name: "eintr"}
	fmt.Println(a.Name())
	fmt.Println(A.Name(a))
}
~~~
变量`a`就是所谓的`方法接收者`，他会作为方法Name的第一个参数传入

go中函数类型只和参数与返回值有关，方法本质上就是普通的函数

~~~ go
func NameOfA(a A) string {
	a.name = "Hi! " + a.name
	return a.name
}

func main() {

	t1 := reflect.TypeOf(A.Name)
	t2 := reflect.TypeOf(NameOfA)

	fmt.Println(t1 == t2)
}
~~~

> 值接收者与指针接收者

~~~ go
func (a A) Name() string {
	a.name = "Hi! " + a.name
	return a.name
}

func (a *A) NameP() string {
	a.name = "Hi! " + a.name
	return a.name
}

func NameOfA(a A) string {
	a.name = "Hi!" + a.name
	return a.name
}

func main() {

	a := A{name: "eintr"}

	fmt.Println(a.Name(), a.name)

	pa := &a
	fmt.Println(pa.NameP(), a.name)
}
~~~

> 其他语法糖

~~~ go

type A struct {
	name string
}
func (a A) GetName() string {
	return a.name
}

func (pa *A) SetName() string {
	pa.name = "Hi! " + pa.name
	return pa.name
}

func main() {

	a := A{name: "eintr"}
	pa := &a

	fmt.Println(a.SetName(), a.name)
	fmt.Println(pa.GetName(), a.name)
}
~~~

## 将方法复制给变量

~~~ go

type A struct {
	name string
}
func (a A) GetName() string {
	return a.name
}

func main() {
	a := A{name: "eintr"}

	f1 := A.GetName//方法表达式
	f1(a)
}
~~~

相当于
~~~ go
func GetName(a A) string {
	return a.name
}

func main() {
	a := A{name: "eintr"}

	f1 := GetName
	f1(a)
}

~~~
