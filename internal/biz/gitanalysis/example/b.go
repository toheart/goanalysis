package main

// B 调用 A，形成简单调用链
func B() int {
	return A()
}
