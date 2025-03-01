package sqllite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3" // 引入 sqlite3 驱动

	"github.com/toheart/goanalysis/functrace"
	"github.com/toheart/goanalysis/internal/biz/entity"
)

type TraceDB struct {
	db *sql.DB
}

func NewTraceDB(dbPath string) (*TraceDB, error) {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("trace db file not found: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open trace db failed: %w", err)
	}

	return &TraceDB{db: db}, nil
}

func (d *TraceDB) GetTracesByGID(gid string) ([]entity.TraceData, error) {
	var traces []entity.TraceData
	rows, err := d.db.Query("SELECT * FROM TraceData WHERE gid = ?", gid) // 使用 sqlite3 查询
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var trace entity.TraceData
		var paramsJSON string
		var timeCost sql.NullString // 使用 sql.NullString 处理可能的 NULL 值
		if err := rows.Scan(&trace.ID, &trace.Name, &trace.GID, &trace.Indent, &paramsJSON, &timeCost); err != nil {
			return nil, err
		}

		// 将 JSON 字符串解析为列表
		var params []functrace.TraceParams
		if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
			return nil, err
		}
		trace.Params = params // 假设 TraceData 结构体中有 Params 字段

		// 处理 timeCost 的值
		if timeCost.Valid {
			trace.TimeCost = timeCost.String // 只有在有效时才赋值
		} else {
			trace.TimeCost = "" // 或者设置为默认值
		}

		traces = append(traces, trace)
	}
	return traces, nil
}

func (d *TraceDB) GetAllGIDs(page int, limit int) ([]uint64, error) {
	var gids []uint64
	offset := (page - 1) * limit // 计算偏移量
	query := "SELECT DISTINCT gid FROM TraceData LIMIT ? OFFSET ?"
	rows, err := d.db.Query(query, limit, offset) // 使用 LIMIT 和 OFFSET 进行分页
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

func (d *TraceDB) GetAllFunctionName() ([]string, error) {
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

func (d *TraceDB) GetParamsByID(id int32) ([]functrace.TraceParams, error) {
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

func (d *TraceDB) GetGidsByFunctionName(functionName string) ([]string, error) {
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

func (d *TraceDB) Close() error {
	return d.db.Close()
}

func (d *TraceDB) GetTotalGIDs() (int, error) {
	var total int
	query := "SELECT COUNT(DISTINCT gid) FROM TraceData"
	err := d.db.QueryRow(query).Scan(&total) // 查询总数
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (d *TraceDB) GetInitialFunc(gid uint64) (string, error) {
	var initialFunc string
	query := "SELECT name FROM TraceData WHERE gid = ? limit 1 offset 0"
	err := d.db.QueryRow(query, gid).Scan(&initialFunc)
	if err != nil {
		return "", err
	}
	return initialFunc, nil
}
