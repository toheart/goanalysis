package main

func main() {
	_ = C()
	_ = UseInterface(impl{})
	_ = GenericFn(1)
	_ = G()
	var s S
	_ = s.M()
	_ = s.N()
}
