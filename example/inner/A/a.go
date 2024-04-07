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
