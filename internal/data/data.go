package data

import (
	"sync"

	"github.com/toheart/goanalysis/internal/data/sqllite"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	_ "github.com/mattn/go-sqlite3" // 引入 sqlite3 驱动
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData)

// Data .
type Data struct {
	sync.RWMutex
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
	d.RLock()
	traceDB := d.traceDB[dbPath]
	d.RUnlock()
	if traceDB == nil {
		d.Lock()
		defer d.Unlock()
		d.log.Infof("get trace db: %s", dbPath)
		traceDB, err := sqllite.NewTraceDB(dbPath)
		if err != nil {
			return nil, err
		}
		d.traceDB[dbPath] = traceDB
	}
	return traceDB, nil
}

func (d *Data) GetFuncNodeDB(dbPath string) (*sqllite.FuncTree, error) {
	d.RLock()
	funcNodeDB := d.funcNodeDB
	d.RUnlock()
	if funcNodeDB == nil {
		d.Lock()
		defer d.Unlock()
		d.log.Infof("get func node db: %s", dbPath)
		funcNodeDB, err := sqllite.NewFuncNodeDB(dbPath)
		if err != nil {
			return nil, err
		}
		d.funcNodeDB = funcNodeDB
	}
	return funcNodeDB, nil
}
