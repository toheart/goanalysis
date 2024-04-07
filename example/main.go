package main

import "example/inner/A"

/**
@file:
@author: levi.Tang
@time: 2024/4/4 16:11
@description:
**/

func main() {
	// callA
	a := A.NewCallA("tly")
	a.PrintB()
}
