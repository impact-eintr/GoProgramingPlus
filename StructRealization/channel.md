# 管道
## 题目
1. 写关闭的管道会触发panic
2. 
~~~go
func ChanCap(){
    ch := make(chan int,10)
    ch <-1
    ch <- 2
    fmt.Println(len(ch))//2
    fmt.Println(lcap(ch))//10
}
~~~
3.
**互斥锁**
~~~go
var counter int = 0
var ch make(chan int,1)

func Worker(){
    <- ch
    counter++
    ch <-1
}
~~~

~~~go

~~~
