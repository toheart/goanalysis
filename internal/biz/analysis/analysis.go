package analysis

import (
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/toheart/functrace"
	"github.com/toheart/goanalysis/internal/biz/entity"
	"github.com/toheart/goanalysis/internal/conf"
	"github.com/toheart/goanalysis/internal/data"
)

type AnalysisBiz struct {
	conf *conf.Biz
	data *data.Data
	log  *log.Helper
}

func (a *AnalysisBiz) GetTotalGIDs(dbpath string) (int, error) {
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return 0, err
	}
	return traceDB.GetTotalGIDs()
}

func (a *AnalysisBiz) GetAllFunctionName(dbpath string) ([]string, error) {
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}
	return traceDB.GetAllFunctionName()
}

func (a *AnalysisBiz) GetStaticDBPath() string {
	return a.conf.StaticDBpath
}

func NewAnalysisBiz(conf *conf.Biz, data *data.Data, logger log.Logger) *AnalysisBiz {
	return &AnalysisBiz{conf: conf, data: data, log: log.NewHelper(logger)}
}

func (a *AnalysisBiz) GetTracesByGID(dbpath string, gid string) ([]functrace.TraceData, error) {
	a.log.Infof("get traces by gid: %s from db: %s", gid, dbpath)
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}
	return traceDB.GetTracesByGID(gid)
}

func (a *AnalysisBiz) GetAllGIDs(dbpath string, page int, limit int) ([]uint64, error) {
	a.log.Infof("get all gids from db: %s", dbpath)
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}
	return traceDB.GetAllGIDs(page, limit)
}

func (a *AnalysisBiz) GetInitialFunc(dbpath string, gid uint64) (string, error) {
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return "", err
	}
	return traceDB.GetInitialFunc(gid)
}

func (a *AnalysisBiz) GetParamsByID(dbpath string, id int32) ([]functrace.TraceParams, error) {
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}
	return traceDB.GetParamsByID(id)
}

func (a *AnalysisBiz) GetGidsByFunctionName(dbpath string, functionName string) ([]string, error) {
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}
	return traceDB.GetGidsByFunctionName(functionName)
}

func (a *AnalysisBiz) VerifyProjectPath(path string) bool {
	// 检查路径是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	a.log.Infof("verified db path: %s", path)
	return true
}

// GetTracesByParentFunc 根据父函数名称获取函数调用
func (a *AnalysisBiz) GetTracesByParentFunc(dbpath string, parentFunc string) ([]functrace.TraceData, error) {
	a.log.Infof("get traces by parent function: %s from db: %s", parentFunc, dbpath)
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}
	return traceDB.GetTracesByParentFunc(parentFunc)
}

// GetAllParentFuncNames 获取所有的父函数名称
func (a *AnalysisBiz) GetAllParentFuncNames(dbpath string) ([]string, error) {
	a.log.Infof("get all parent function names from db: %s", dbpath)
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}
	return traceDB.GetAllParentFuncNames()
}

// GetChildFunctions 获取函数的子函数
func (a *AnalysisBiz) GetChildFunctions(dbpath string, funcName string) ([]string, error) {
	a.log.Infof("get child functions of: %s from db: %s", funcName, dbpath)
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}
	return traceDB.GetChildFunctions(funcName)
}

// GetHotFunctions 获取热点函数分析数据
func (a *AnalysisBiz) GetHotFunctions(dbpath string, sortBy string) ([]entity.HotFunction, error) {
	a.log.Infof("get hot functions, sort by: %s from db: %s", sortBy, dbpath)
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}
	return traceDB.GetHotFunctions(sortBy)
}

// GetGoroutineStats 获取Goroutine统计信息
func (a *AnalysisBiz) GetGoroutineStats(dbpath string) (*entity.GoroutineStats, error) {
	a.log.Infof("get goroutine stats from db: %s", dbpath)
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}
	return traceDB.GetGoroutineStats()
}

// GetFunctionAnalysis 获取函数调用关系分析
func (a *AnalysisBiz) GetFunctionAnalysis(dbpath string, functionName string, queryType string) ([]entity.FunctionNode, error) {
	a.log.Infof("get function analysis, function: %s, type: %s from db: %s", functionName, queryType, dbpath)
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}
	return traceDB.GetFunctionAnalysis(functionName, queryType)
}

// GetFunctionCallGraph 获取函数调用关系图
func (a *AnalysisBiz) GetFunctionCallGraph(dbpath string, functionName string, depth int, direction string) (*entity.FunctionCallGraph, error) {
	a.log.Infof("get function call graph, function: %s, depth: %d, direction: %s from db: %s", functionName, depth, direction, dbpath)
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return nil, err
	}
	return traceDB.GetFunctionCallGraph(functionName, depth, direction)
}

// GetGoroutineCallDepth 获取指定 Goroutine 的最大调用深度
func (a *AnalysisBiz) GetGoroutineCallDepth(dbpath string, gid uint64) (int, error) {
	a.log.Infof("get goroutine call depth for gid: %d from db: %s", gid, dbpath)
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return 0, err
	}
	return traceDB.GetGoroutineCallDepth(gid)
}

// GetGoroutineExecutionTime 获取指定 Goroutine 的总执行时间
func (a *AnalysisBiz) GetGoroutineExecutionTime(dbpath string, gid uint64) (string, error) {
	a.log.Infof("get goroutine execution time for gid: %d from db: %s", gid, dbpath)
	traceDB, err := a.data.GetTraceDB(dbpath)
	if err != nil {
		return "", err
	}
	return traceDB.GetGoroutineExecutionTime(gid)
}
