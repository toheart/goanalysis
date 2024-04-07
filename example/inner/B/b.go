package B

import "fmt"

/**
@file:
@author: levi.Tang
@time: 2024/4/4 16:12
@description:
**/

func init() {
	fmt.Println("b init")
}

type CallB struct {
	name string
}

func NewCallB(name string) *CallB {
	return &CallB{name: name}
}

func (b CallB) PrintB() {
	fmt.Printf("CallB print : %s \n", b.name)
}

func CalledA() {
	fmt.Printf("a init call b \n")
}
