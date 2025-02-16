package data

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/toheart/goanalysis/functrace"
	"github.com/toheart/goanalysis/internal/biz/entity"
	"github.com/toheart/goanalysis/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	_ "github.com/mattn/go-sqlite3" // 引入 sqlite3 驱动
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData)

// Data .
type Data struct {
	conf *conf.Data
	db   *sql.DB // 修改为 sql.DB
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("Closing the data resources")
	}
	fmt.Println(c.Database.Source)
	// Check if trace.db exists, if not panic
	if _, err := os.Stat(c.Database.Source); os.IsNotExist(err) {
		panic("trace.db not found")
	}
	db, err := sql.Open("sqlite3", c.Database.Source) // 使用 sqlite3
	if err != nil {
		return nil, nil, err
	}
	return &Data{conf: c, db: db}, cleanup, nil
}

func (d *Data) GetTracesByGID(gid string) ([]entity.TraceData, error) {
	var traces []entity.TraceData
	rows, err := d.db.Query("SELECT * FROM TraceData WHERE gid = ?", gid) // 使用 sqlite3 查询
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var trace entity.TraceData
		var paramsJSON string
		if err := rows.Scan(&trace.ID, &trace.Name, &trace.GID, &trace.Indent, &paramsJSON, &trace.TimeCost); err != nil {
			return nil, err
		}

		// 将 JSON 字符串解析为列表
		var params []functrace.TraceParams
		if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
			return nil, err
		}
		trace.Params = params // 假设 TraceData 结构体中有 Params 字段
		traces = append(traces, trace)
	}
	return traces, nil
}

func (d *Data) GetAllGIDs() ([]uint64, error) {
	var gids []uint64
	rows, err := d.db.Query("SELECT DISTINCT gid FROM TraceData") // 查询所有不同的 GID
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var gid uint64
		if err := rows.Scan(&gid); err != nil {
			return nil, err
		}
		gids = append(gids, gid)
	}
	return gids, nil
}

func (d *Data) GetAllFunctionName() ([]string, error) {
	var functionNames []string
	rows, err := d.db.Query("SELECT DISTINCT name FROM TraceData") // 查询所有不同的函数名
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var functionName string
		if err := rows.Scan(&functionName); err != nil {
			return nil, err
		}
		functionNames = append(functionNames, functionName)
	}
	return functionNames, nil
}

func (d *Data) GetParamsByID(id int32) ([]functrace.TraceParams, error) {
	var params []functrace.TraceParams
	var paramsJSON string
	rows, err := d.db.Query("SELECT params FROM TraceData WHERE id = ?", id) // 使用 sqlite3 查询
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&paramsJSON); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
			return nil, err
		}
	}
	return params, nil
}

func (d *Data) GetGidsByFunctionName(functionName string) ([]string, error) {
	var gids []string
	rows, err := d.db.Query("SELECT DISTINCT gid FROM TraceData WHERE name = ?", functionName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var gid string
		if err := rows.Scan(&gid); err != nil {
			return nil, err
		}
		gids = append(gids, gid)
	}
	return gids, nil
}
