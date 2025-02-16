package analysis

import (
	"github.com/toheart/goanalysis/functrace"
	"github.com/toheart/goanalysis/internal/biz/entity"
	"github.com/toheart/goanalysis/internal/data"
)

type AnalysisBiz struct {
	data *data.Data
}

func (a *AnalysisBiz) GetAllFunctionName() ([]string, error) {
	return a.data.GetAllFunctionName()
}

func NewAnalysisBiz(data *data.Data) *AnalysisBiz {
	return &AnalysisBiz{data: data}
}

func (a *AnalysisBiz) GetTracesByGID(gid string) ([]entity.TraceData, error) {
	return a.data.GetTracesByGID(gid)
}

func (a *AnalysisBiz) GetAllGIDs() ([]uint64, error) {
	return a.data.GetAllGIDs()
}

func (a *AnalysisBiz) GetParamsByID(id int32) ([]functrace.TraceParams, error) {
	return a.data.GetParamsByID(id)
}

func (a *AnalysisBiz) GetGidsByFunctionName(functionName string) ([]string, error) {
	return a.data.GetGidsByFunctionName(functionName)
}
