# 并发编程

## 上下文Context
上下文 context.Context Go 语言中用来设置截止日期、同步信号，传递请求相关值的结构体。上下文与 Goroutine 有比较密切的关系，是 Go 语言中独特的设计，在其他编程语言中我们很少见到类似的概念。:

context.Context 是 Go 语言在 1.7 版本中引入标准库的接口，该接口定义了四个需要实现的方法，其中包括：

1. Deadline 返回context.Context 被取消的时间，也就是完成工作的截止时间
2. Done 返回一个Channel，这个Channel会在当前工作完成或者上下文被取消后关闭，多次调用Done方法会返回同一个Channel
3. Err 返回context.Context 结束的原因，它只会在Done方法对应的Channel关闭时返回非空的值
   - 如果context.Context被取消，会返回Canceled错误
   - 如果context.Context超时，会返回DeadlineExceeded
4. Value — 从 context.Context 中获取键对应的值，对于同一个上下文来说，多次调用 Value 并传入相同的 Key 会返回相同的结果，该方法可以用来传递请求特定的数据

``` go
type Context interface {
	Deadline() (deadline time.Time, ok bool)
	Done() <-chan struct{}
	Err() error
	Value(key interface{}) interface{}
}
```

context 包中提供的 context.Background、context.TODO、context.WithDeadline 和 context.WithValue 函数会返回实现该接口的私有结构体，我们会在后面详细介绍它们的工作原理。

### 设计原理
在 Goroutine 构成的树形结构中对信号进行同步以减少计算资源的浪费是 context.Context 的最大作用。Go 服务的每一个请求都是通过单独的 Goroutine 处理的，HTTP/RPC 请求的处理器会启动新的 Goroutine 访问数据库和其他服务。

如下图所示，我们可能会创建多个 Goroutine 来处理一次请求，而 context.Context 的作用是在不同 Goroutine 之间同步请求特定数据、取消信号以及处理请求的截止日期。
![](https://img.draveness.me/golang-context-usage.png)

每一个 context.Context 都会从最顶层的 Goroutine 一层一层传递到最下层。context.Context 可以在上层 Goroutine 执行出现错误时，将信号及时同步给下层

![](https://img.draveness.me/golang-without-context.png)
a如上图所示，当最上层的 Goroutine 因为某些原因执行失败时，下层的 Goroutine 由于没有接收到这个信号所以会继续工作；但是当我们正确地使用 context.Context 时，就可以在下层及时停掉无用的工作以减少额外资源的消耗：
![](https://img.draveness.me/golang-with-context.png)
我们可以通过一个代码片段了解 context.Context 是如何对信号进行同步的。在这段代码中，我们创建了一个过期时间为 1s 的上下文，并向上下文传入 handle 函数，该方法会使用 500ms 的时间处理传入的请求：

``` go
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	go handle(ctx, 500*time.Millisecond)
	select {
	case <-ctx.Done():
		fmt.Println("main", ctx.Err())
	}
}

func handle(ctx context.Context, duration time.Duration) {
	select {
	case <-ctx.Done():
		fmt.Println("handle", ctx.Err())
	case <-time.After(duration):
		fmt.Println("process request with", duration)
	}
}
```

因为过期时间大于处理时间，所以我们有足够的时间处理该请求，运行上述代码会打印出下面的内容：

``` go
$ go run context.go
process request with 500ms
main context deadline exceeded
```

handle 函数没有进入超时的 select 分支，但是 main 函数的 select 却会等待 context.Context 超时并打印出 main context deadline exceeded。

如果我们将处理请求时间增加至 1500ms，整个程序都会因为上下文的过期而被中止，：

``` go
$ go run context.go
main context deadline exceeded
handle context deadline exceeded
```

相信这两个例子能够帮助各位读者理解 context.Context 的使用方法和设计原理 — 多个 Goroutine 同时订阅 ctx.Done() 管道中的消息，一旦接收到取消信号就立刻停止当前正在执行的工作。

### 默认上下文

context 包中最常用的方法还是 context.Background、context.TODO，这两个方法都会返回预先初始化好的私有变量 background 和 todo，它们会在同一个 Go 程序中被复用：

``` go
func Background() Context {
	return background
}

func TODO() Context {
	return todo
}
```

这两个私有变量都是通过 new(emptyCtx) 语句初始化的，它们是指向私有结构体 context.emptyCtx 的指针，这是最简单、最常用的上下文类型：

``` go
type emptyCtx int

func (*emptyCtx) Deadline() (deadline time.Time, ok bool) {
	return
}

func (*emptyCtx) Done() <-chan struct{} {
	return nil
}

func (*emptyCtx) Err() error {
	return nil
}

func (*emptyCtx) Value(key interface{}) interface{} {
	return nil
}
```

从上述代码中，我们不难发现 context.emptyCtx 通过空方法实现了 context.Context 接口中的所有方法，它没有任何功能。

从源代码来看，context.Background 和 context.TODO 也只是互为别名，没有太大的差别，只是在使用和语义上稍有不同：
- context.Background 是上下文的默认值，所有其他的上下文都应该从它衍生出来；
- context.TODO 应该仅在不确定应该使用哪种上下文时使用；

在多数情况下，如果当前函数没有上下文作为入参，我们都会使用 context.Background 作为起始的上下文向下传递。 

### 取消信号





### 传值方法




### 小结


## 同步原语与锁




## 定时器




## Channel



## GMP模型
### 概览
- G goroutine 协程
- P Processor 处理器
- M Thread 内核线程

> 全局队列
存放等待执行的G

> P的本地队列
- 存放等待运行的G
- 数量限制 不超过256个
- 优先将新创建的G 放到本地队列中，如果满了会放在全局队列中

> P列表(P的队列集合)
- 程序启动时创建
- 最多有GOMAXPROCS个

> M列表
当前操作系统分配到当前Go程序的内核线程数

> P 和 M 的数量
- P
  - $GOMAXPROCS
  - runtime.GOMAXPROCS()
- M 
  - Go语言本身限制10000
  - runtime.debug 中SetMaxThreads()
  - 有一个M阻塞就会创建一个新的M
  - 如果有M空闲，就会回收或者睡眠
### 调度策略
- 复用线程
- 利用并行
- 抢占
- 全局G队列

> 复用线程 
- work stealing 机制
比如有两个M和两个P，其中一个P中有多个G，而另一个P中没有G，为了让M不空闲，平会从另一个P的runq的队尾获取一个G。
- hand off 机制
如果当前在M运行的G阻塞了，调度器会创建或者唤醒一个thread作为新的M,将当前M上的P移动到新的M上，而阻塞的G仍然占用M,并进入休眠/销毁

> 利用并行
GOMAXPROCS

> 抢占 
- co-routine 一个c(co-routine)除非主动释放cpu，否则其他等待的c永远无法获取cpu
- goroutine 一个G在cpu上最多运行10ms,到时间就会被调度器踢开，其他G会抢占cpu

> 全局G队列
先从全局队列中获取，没有的话再从其他P的runq中steal


### 流程(重要)
1. 我们通过go func() 来创建一个goroutine
2. 有两个存储G的队列，一个是局部调度器P的本地队列，一个是全局G队列，新创建的G会先保存在P的本地队列中，如果P的本地队列已经满了就会保存在全局的队列中
3. G只能运行在M中，一个M必须持有一个P，M与P是1:1的关系，M会从P的本地队列中弹出一个可执行状态的G来执行，如果P的本地队列为空，先获取全局队列的G，然后从其他P中steal
4. 一个M调度G执行的过程是一个循环机制
5. 当M执行某个G的时候如果发生了syscall或者其他阻塞操作，M会阻塞，如果当前有一些G在执行，runtime会把这个线程M从P中摘除，然后再创建一个新的操作系统线程(如果有空闲的线程可以使用空闲的线程来服务这个P)
6. 当M系统调用结束的时候，这个G会尝试获取一个空闲的P来执行，并放到这个P的本地队列中，如果获取不到P那么这个线程M会变成休眠状态，加入空闲线程中，然后这个G会被放到全局队列中

### 生命周期
#### M0
- 启动程序后的编号为0的主主线程
- 在全局变量runtime.M0中，不需要在堆上分配
- 负责执行和初始化操作的启动第一个G
- 启动第一个G之后，M0就和其他的M一样了

#### G0
- 每次启动一个M,都会第一个创建的goroutine,就是G0
- G0仅用于负责调度的G
- G0不指向任何可执行的函数
- 每个M都会有一个自己的G0
- 在调度或系统调用时M会切换到G0，来调度goroutine
- M的G0会放在全局空间中

### 场景分析

#### 创建G
两个M执行G1,G2,现在要创建G3,根据局部性，G2优先加入G1所在本地队列(自产自销)

#### G1执行完毕
G1调用goexit()，M1切换为自己的G0(G0负责调度协程切换schdule()函数)，从P的本地队列中获取G2，并开始运行G2函数，实现了线程M1的复用

#### 连续创建多个G,导致队列满
G2创建6个G，假设本地队列最多为4个已满，接下来g2创建G7，将本地队列的前一半G打乱顺序和新创建的G7一同放到全局队列中，G2创建G8，G8加入到P1的本地队列

#### 唤醒正在休眠的M
在创建G时，运行的G会尝试唤醒其他空闲的P和M组合去执行
假定G2唤醒了M2，M2绑定了P2，并运行了G0，但P2本地队列中没有G，M2现在被称为自选线程(没有G但为运行状态的线程，不断寻找G)

#### 被唤醒的M从全局取得G
M2尝试从全局队列中取一批G放到P2的本地队列，M2从全局队列取出G的数量符合下面的公式

``` go
n = min(len(GQ)/GOMAXPROCS+1, len(GQ/2))
```
1. M2自选线程会寻找可运行的G(优先全局，然后steal)
2. M2从全局队列中获取G3
3. M2完成从G0切换到G3，不再是自旋状态

#### 偷取G的情况
全局队列中已经没有G，那m就要执行work stealing 从其他有G的P那里偷取一半G过来，放在自己的P本地队列中。

#### 自选线程的最大限制
自旋线程+执行线程 <= GOMAXPROCS

#### 发生阻塞调用
G8创建G9 并且G8进行系统调用
M2系统调用中
P2空闲的M5绑定通过G0切换到G9并执行

#### G的阻塞结束
G8创建完G9后阻塞，M与P分离，等阻塞结束后，寻找P，先找原来的，再找全局空闲的，最后放弃，G和M回到全局队列


## 调度器
Go 语言在并发编程方面有强大的能力，这离不开语言层面对并发编程的支持。本节会介绍 Go 语言运行时调度器的实现原理，其中包含调度器的设计与实现原理、演变过程以及与运行时调度相关的数据结构。

谈到 Go 语言调度器，我们绕不开的是操作系统、进程与线程这些概念，线程是操作系统调度时的最基本单元，而 Linux 在调度器并不区分进程和线程的调度，它们在不同操作系统上也有不同的实现，但是在大多数的实现中线程都属于进程：

![](https://img.draveness.me/2020-02-05-15808864354570-process-and-threads.pnghttps://img.draveness.me/2020-02-05-15808864354570-process-and-threads.png)



## 网络轮询器




## 系统监视





