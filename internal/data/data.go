package data

import (
	"sync"

	"github.com/toheart/goanalysis/internal/data/sqllite"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	_ "github.com/mattn/go-sqlite3" // 引入 sqlite3 驱动
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, sqllite.NewTraceDB, sqllite.NewFuncNodeDB, sqllite.NewFileDB)

// Data .
type Data struct {
	sync.RWMutex
	traceDB    map[string]*sqllite.TraceDB
	funcNodeDB map[string]*sqllite.StaticDBImpl
	log        *log.Helper
}

// NewData .
func NewData(logger log.Logger) *Data {
	return &Data{
		traceDB:    make(map[string]*sqllite.TraceDB),
		funcNodeDB: make(map[string]*sqllite.StaticDBImpl),
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
		// 双重检查
		if traceDB, ok := d.traceDB[dbPath]; ok {
			return traceDB, nil
		}
		d.log.Infof("create trace db: %s", dbPath)
		traceDB, err := sqllite.NewTraceDB(dbPath)
		if err != nil {
			return nil, err
		}
		d.traceDB[dbPath] = traceDB
	}
	return traceDB, nil
}

func (d *Data) GetFuncNodeDB(dbPath string) (*sqllite.StaticDBImpl, error) {
	d.RLock()
	funcNodeDB := d.funcNodeDB[dbPath]
	d.RUnlock()
	if funcNodeDB == nil {
		d.Lock()
		defer d.Unlock()
		// 双重检查
		if funcNodeDB, ok := d.funcNodeDB[dbPath]; ok {
			return funcNodeDB, nil
		}
		d.log.Infof("create func node db: %s", dbPath)
		var err error
		funcNodeDB, err = sqllite.NewFuncNodeDB(dbPath)
		if err != nil {
			d.log.Errorf("get func node db failed: %s", err)
			return nil, err
		}
		d.funcNodeDB[dbPath] = funcNodeDB
	}
	return funcNodeDB, nil
}
