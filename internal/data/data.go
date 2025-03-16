package data

import (
	"database/sql"
	"sync"

	"github.com/toheart/goanalysis/internal/data/sqlite"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	msqlite "modernc.org/sqlite"
)

func init() {
	sql.Register("sqlite3", &msqlite.Driver{})
}

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, sqlite.NewTraceEntDB, sqlite.NewStaticEntDBImpl, sqlite.NewFileEntDB)

// Data .
type Data struct {
	sync.RWMutex
	traceDB    map[string]*sqlite.TraceEntDB
	funcNodeDB map[string]*sqlite.StaticEntDBImpl
	log        *log.Helper
}

// NewData .
func NewData(logger log.Logger) *Data {
	return &Data{
		traceDB:    make(map[string]*sqlite.TraceEntDB),
		funcNodeDB: make(map[string]*sqlite.StaticEntDBImpl),
		log:        log.NewHelper(logger),
	}
}

func (d *Data) GetTraceDB(dbPath string) (*sqlite.TraceEntDB, error) {
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
		traceDB, err := sqlite.NewTraceEntDB(dbPath)
		if err != nil {
			return nil, err
		}
		d.traceDB[dbPath] = traceDB
	}
	return traceDB, nil
}

func (d *Data) GetFuncNodeDB(dbPath string) (*sqlite.StaticEntDBImpl, error) {
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
		funcNodeDB, err = sqlite.NewStaticEntDBImpl(dbPath)
		if err != nil {
			d.log.Errorf("get func node db failed: %s", err)
			return nil, err
		}
		d.funcNodeDB[dbPath] = funcNodeDB
	}
	return funcNodeDB, nil
}
