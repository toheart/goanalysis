package main

// GenericFn 是一个使用泛型的函数，内部调用 B
func GenericFn[T any](v T) int {
	return B()
}
