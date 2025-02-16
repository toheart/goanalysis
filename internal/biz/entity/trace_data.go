package entity

import (
	"github.com/toheart/goanalysis/functrace"
)

type TraceData struct {
	ID       int
	Name     string
	GID      int
	Indent   int
	Params   []functrace.TraceParams
	TimeCost string
}
