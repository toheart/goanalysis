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
