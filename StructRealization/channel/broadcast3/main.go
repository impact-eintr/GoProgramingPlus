package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type threadSafeSlice struct {
	sync.Mutex
	workers []*worker
}

func (slice *threadSafeSlice) Push(w *worker) {
	slice.Lock()
	defer slice.Unlock()

	slice.workers = append(slice.workers, w)
}

func (slice *threadSafeSlice) Iter(routine func(*worker)) {
	slice.Lock()
	defer slice.Unlock()

	for _, worker := range slice.workers {
		routine(worker)
	}
}

type worker struct {
	name   string
	source chan interface{}
	quit   chan struct{}
}

func (w *worker) Start() {
	w.source = make(chan interface{})
	go func() {
		for {
			select {
			case msg := <-w.source:
				fmt.Println("==========>> ", w.name, msg)
			case <-w.quit: // 后面解释
				fmt.Println(w.name, " quit!")
				return
			}
		}
	}()
}

func main() {
	globalQuit := make(chan struct{})

	tss := &threadSafeSlice{}

	// 1秒钟添加一个新的worker至slice中
	go func() {
		name := "worker"
		for i := 0; i < 10; i++ {
			time.Sleep(1 * time.Second)
			w := &worker{
				name: fmt.Sprintf("%s%d", name, i),
				quit: globalQuit,
			}
			w.Start()
			tss.Push(w)
		}
	}()

	// 派发消息
	go func() {
		msg := "test"
		count := 0
		var sendMsg string
		for {
			select {
			case <-globalQuit:
				fmt.Println("Stop send message")
				return
			case <-time.Tick(500 * time.Millisecond):
				count++
				sendMsg = fmt.Sprintf("%s-%d", msg, count)
				fmt.Println("Send message is ", sendMsg)
				tss.Iter(func(w *worker) { w.source <- sendMsg })
			}
		}
	}()

	// 截获退出信号
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	for sig := range c {
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM: // 获取退出信号时，关闭globalQuit, 让所有监听者退出
			close(globalQuit)
			time.Sleep(1 * time.Second)
			return
		}
	}
}
