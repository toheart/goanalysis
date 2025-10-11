package main

// G 启动 goroutine 调用 A
func G() int {
	ch := make(chan int, 1)
	go func() { ch <- A() }()
	return <-ch
}
