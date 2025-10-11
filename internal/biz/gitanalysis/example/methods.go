package main

// S 是一个示例结构体
type S struct{}

// M 方法调用 A
func (s *S) M() int {
	return A()
}

// N 方法调用 B
func (s S) N() int {
	return B()
}
