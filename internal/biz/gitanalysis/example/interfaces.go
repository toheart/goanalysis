package main

// I 是一个简单接口
type I interface {
	Do() int
}

// impl 实现 I，间接调用 A
type impl struct{}

func (impl) Do() int { return A() }

// UseInterface 通过接口调用实现，验证间接调用
func UseInterface(x I) int { return x.Do() }
