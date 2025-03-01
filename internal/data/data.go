package data

import (
	"github.com/toheart/goanalysis/internal/data/sqllite"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	_ "github.com/mattn/go-sqlite3" // 引入 sqlite3 驱动
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData)

// Data .
type Data struct {
	traceDB    map[string]*sqllite.TraceDB
	funcNodeDB *sqllite.FuncTree
	log        *log.Helper
}

// NewData .
func NewData(logger log.Logger) *Data {
	return &Data{
		traceDB:    make(map[string]*sqllite.TraceDB),
		funcNodeDB: nil,
		log:        log.NewHelper(logger),
	}
}

func (d *Data) GetTraceDB(dbPath string) (*sqllite.TraceDB, error) {
	if d.traceDB[dbPath] == nil {
		d.log.Infof("get trace db: %s", dbPath)
		traceDB, err := sqllite.NewTraceDB(dbPath)
		if err != nil {
			return nil, err
		}
		d.traceDB[dbPath] = traceDB
	}
	return d.traceDB[dbPath], nil
}

func (d *Data) GetFuncNodeDB(dbPath string) (*sqllite.FuncTree, error) {
	if d.funcNodeDB == nil {
		funcNodeDB, err := sqllite.NewFuncNodeDB(dbPath)
		if err != nil {
			return nil, err
		}
		d.funcNodeDB = funcNodeDB
	}
	return d.funcNodeDB, nil
}
