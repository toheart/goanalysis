package entity

/**
@file:
@author: levi.Tang
@time: 2024/11/4 17:09
@description:
**/

type Trace struct {
	ID       int64
	FuncName string
	Gid      int64
	Indent   int
	Params   string
}
