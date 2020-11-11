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
### 管道数据读写
- 管道没有缓存区时，从管道读取数据会堵塞，直到有协程向管道中写入数据
- 管道有缓存去但缓存区没有数据时，从管道读取数据时也会阻塞，直到有协程写入数据
- 管道读取表达式最多可以给两个变量赋值
~~~go
value,ok := <- ch
//ok 只有在管道已经关闭且缓存区中没有数据时 才代表管道已经关闭
~~~

