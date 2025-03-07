package analysis

import (
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/toheart/goanalysis/functrace"
	"github.com/toheart/goanalysis/internal/biz/entity"
	"github.com/toheart/goanalysis/internal/conf"
	"github.com/toheart/goanalysis/internal/data"
)

type AnalysisBiz struct {
	conf *conf.Biz
	data *data.Data
	log  *log.Helper

	currDB string
}

func (a *AnalysisBiz) GetTotalGIDs() (int, error) {
	traceDB, err := a.data.GetTraceDB(a.currDB)
	if err != nil {
		return 0, err
	}
	return traceDB.GetTotalGIDs()
}

func (a *AnalysisBiz) GetAllFunctionName() ([]string, error) {
	traceDB, err := a.data.GetTraceDB(a.currDB)
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

func (a *AnalysisBiz) GetTracesByGID(gid string) ([]entity.TraceData, error) {
	a.log.Infof("get traces by gid: %s", gid)
	traceDB, err := a.data.GetTraceDB(a.currDB)
	if err != nil {
		return nil, err
	}
	return traceDB.GetTracesByGID(gid)
}

func (a *AnalysisBiz) GetAllGIDs(page int, limit int) ([]uint64, error) {
	a.log.Infof("get all gids from db: %s", a.currDB)
	traceDB, err := a.data.GetTraceDB(a.currDB)
	if err != nil {
		return nil, err
	}
	return traceDB.GetAllGIDs(page, limit)
}

func (a *AnalysisBiz) GetInitialFunc(gid uint64) (string, error) {
	traceDB, err := a.data.GetTraceDB(a.currDB)
	if err != nil {
		return "", err
	}
	return traceDB.GetInitialFunc(gid)
}

func (a *AnalysisBiz) GetParamsByID(id int32) ([]functrace.TraceParams, error) {
	traceDB, err := a.data.GetTraceDB(a.currDB)
	if err != nil {
		return nil, err
	}
	return traceDB.GetParamsByID(id)
}

func (a *AnalysisBiz) GetGidsByFunctionName(functionName string) ([]string, error) {
	traceDB, err := a.data.GetTraceDB(a.currDB)
	if err != nil {
		return nil, err
	}
	return traceDB.GetGidsByFunctionName(functionName)
}

func (a *AnalysisBiz) SetCurrDB(dbPath string) {
	a.currDB = dbPath
}

func (a *AnalysisBiz) VerifyProjectPath(path string) bool {
	// 检查路径是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	a.log.Infof("set current db: %s", path)
	// 如果存在, 设置成当前数据库
	a.currDB = path
	return true
}

func (a *AnalysisBiz) CheckDatabase() bool {
	return a.VerifyProjectPath(a.conf.StaticDBpath)
}

// GetTracesByParentFunc 根据父函数名称获取函数调用
func (a *AnalysisBiz) GetTracesByParentFunc(parentFunc string) ([]entity.TraceData, error) {
	a.log.Infof("get traces by parent function: %s", parentFunc)
	traceDB, err := a.data.GetTraceDB(a.currDB)
	if err != nil {
		return nil, err
	}
	return traceDB.GetTracesByParentFunc(parentFunc)
}

// GetAllParentFuncNames 获取所有的父函数名称
func (a *AnalysisBiz) GetAllParentFuncNames() ([]string, error) {
	a.log.Infof("get all parent function names")
	traceDB, err := a.data.GetTraceDB(a.currDB)
	if err != nil {
		return nil, err
	}
	return traceDB.GetAllParentFuncNames()
}

// GetChildFunctions 获取函数的子函数
func (a *AnalysisBiz) GetChildFunctions(funcName string) ([]string, error) {
	a.log.Infof("get child functions of: %s", funcName)
	traceDB, err := a.data.GetTraceDB(a.currDB)
	if err != nil {
		return nil, err
	}
	return traceDB.GetChildFunctions(funcName)
}

// GetHotFunctions 获取热点函数分析数据
func (a *AnalysisBiz) GetHotFunctions(sortBy string) ([]entity.HotFunction, error) {
	a.log.Infof("get hot functions, sort by: %s", sortBy)
	traceDB, err := a.data.GetTraceDB(a.currDB)
	if err != nil {
		return nil, err
	}
	return traceDB.GetHotFunctions(sortBy)
}

// GetGoroutineStats 获取Goroutine统计信息
func (a *AnalysisBiz) GetGoroutineStats() (*entity.GoroutineStats, error) {
	a.log.Infof("get goroutine stats")
	traceDB, err := a.data.GetTraceDB(a.currDB)
	if err != nil {
		return nil, err
	}
	return traceDB.GetGoroutineStats()
}

// GetFunctionAnalysis 获取函数调用关系分析
func (a *AnalysisBiz) GetFunctionAnalysis(functionName string, queryType string) ([]entity.FunctionNode, error) {
	a.log.Infof("get function analysis, function: %s, type: %s", functionName, queryType)
	traceDB, err := a.data.GetTraceDB(a.currDB)
	if err != nil {
		return nil, err
	}
	return traceDB.GetFunctionAnalysis(functionName, queryType)
}

// GetFunctionCallGraph 获取函数调用关系图
func (a *AnalysisBiz) GetFunctionCallGraph(functionName string, depth int, direction string) (*entity.FunctionCallGraph, error) {
	a.log.Infof("get function call graph, function: %s, depth: %d, direction: %s", functionName, depth, direction)
	traceDB, err := a.data.GetTraceDB(a.currDB)
	if err != nil {
		return nil, err
	}
	return traceDB.GetFunctionCallGraph(functionName, depth, direction)
}

// GetGoroutineCallDepth 获取指定 Goroutine 的最大调用深度
func (a *AnalysisBiz) GetGoroutineCallDepth(gid uint64) (int, error) {
	a.log.Infof("get goroutine call depth for gid: %d", gid)
	traceDB, err := a.data.GetTraceDB(a.currDB)
	if err != nil {
		return 0, err
	}
	return traceDB.GetGoroutineCallDepth(gid)
}

// GetGoroutineExecutionTime 获取指定 Goroutine 的总执行时间
func (a *AnalysisBiz) GetGoroutineExecutionTime(gid uint64) (string, error) {
	a.log.Infof("get goroutine execution time for gid: %d", gid)
	traceDB, err := a.data.GetTraceDB(a.currDB)
	if err != nil {
		return "", err
	}
	return traceDB.GetGoroutineExecutionTime(gid)
}
