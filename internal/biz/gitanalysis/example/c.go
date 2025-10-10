package main

// C 调用 B，形成更长调用链 C->B->A
func C() int {
	return B()
}
