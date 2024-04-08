package A

import (
	"example/inner/B"
	"fmt"
)

/*
*
@file:
@author: levi.Tang
@time: 2024/4/4 16:11
@description:
*
*/
func init() {
	fmt.Println("a init")
	B.CalledA()
}

type CallA struct {
	name string
}

func NewCallA(name string) *CallA {
	return &CallA{name: name}
}

func (a CallA) PrintB() {
	b := B.NewCallB("levi")
	b.PrintB()
	fmt.Printf("CallA call B : %s \n", a.name)
}

// 递归测试
func RecursionA(i int, n int) int {
	if i == n {
		return i
	}

	return i + RecursionA(i+1, n)
}
