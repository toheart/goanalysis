package functrace

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3" // 引入 sqlite3 驱动
	"gopkg.in/natefinch/lumberjack.v2"
)

/*
*
@file:
@author: levi.Tang
@time: 2024/10/27 22:45
@description:
*
*/
// EnableFlag defines the environment variable to enable trace functionality with optional database logging.
const EnableFlag = "EnableFlag"

var (
	once        sync.Once
	singleTrace *TraceInstance
)

func init() {
	NewTraceInstance()
}

// TraceInstance is a singleton structure that manages function tracing.
type TraceInstance struct {
	sync.Mutex
	indentations map[uint64]int
	enableDB     bool
	log          *slog.Logger
	db           *sql.DB // 修改为 sql.DB
}

type TraceParams struct {
	Pos   int    // 记录参数的位置
	Param string // 记录函数参数
}

// NewTraceInstance initializes the singleton instance of TraceInstance.
func NewTraceInstance() *TraceInstance {
	once.Do(func() {
		var err error
		singleTrace = &TraceInstance{
			indentations: make(map[uint64]int),
			log:          initializeLogger(),
		}
		var dbName string
		index := 0
		for {
			dbName = fmt.Sprintf("trace_%d.db", index)
			if _, err := os.Stat(dbName); os.IsNotExist(err) {
				var err error
				singleTrace.db, err = sql.Open("sqlite3", dbName) // 使用 sqlite3
				if err == nil {
					break
				}
			}
			index++
		}
		if err != nil {
			singleTrace.log.Error("Unable to open database", "error", err)
		}
		_, err = singleTrace.db.Exec("CREATE TABLE IF NOT EXISTS TraceData (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, gid INTEGER, indent INTEGER, params TEXT, timeCost TEXT)") // 创建表
		if err != nil {
			singleTrace.log.Error("Unable to create table", "error", err)
		}
		_, err = singleTrace.db.Exec("CREATE INDEX IF NOT EXISTS idx_gid ON TraceData (gid)") // 对Gid创建索引
		if err != nil {
			singleTrace.log.Error("Unable to create index on gid", "error", err)
		}
	})
	return singleTrace
}

func initializeLogger() *slog.Logger {
	log := slog.New(slog.NewTextHandler(&lumberjack.Logger{
		Filename:  "./trace.log",
		LocalTime: true,
		Compress:  true,
	}, nil))
	return log
}

// enterTrace logs the entry of a function call and stores necessary trace details.
func (t *TraceInstance) enterTrace(id uint64, name string, params []interface{}) int64 {
	t.Lock()
	indent := t.incrementIndent(id)
	t.Unlock()

	indents := generateIndentString(indent)
	traceParams := prepareParamsOutput(params)

	// 将 traceParams 转换为 JSON 字符串
	paramsJSON, err := json.Marshal(traceParams)
	if err != nil {
		t.log.Error("Unable to marshal params to JSON", "error", err)
		return 0
	}

	// 使用 sqlite3 插入数据
	var lastInsertId int64
	res, err := t.db.Exec("INSERT INTO TraceData (name, gid, indent, params) VALUES (?, ?, ?, ?)", name, id, indent, paramsJSON)
	if err != nil {
		t.log.Error("Unable to insert data", "error", err)
		return 0
	}
	lastInsertId, err = res.LastInsertId() // 获取刚才插入的自增id
	if err != nil {
		t.log.Error("Unable to get last insert id", "error", err)
		return 0
	}

	t.log.Info(fmt.Sprintf("%s->%s", indents, name), "gid", id, "params", string(paramsJSON), "lastInsertId", lastInsertId)
	return lastInsertId
}

func (t *TraceInstance) incrementIndent(id uint64) int {
	indent := t.indentations[id]
	t.indentations[id]++
	return indent
}

func generateIndentString(indent int) string {
	return strings.Repeat("**", indent)
}

func prepareParamsOutput(params []interface{}) []TraceParams {
	var traceParams []TraceParams
	for i, item := range params {
		traceParams = append(traceParams, TraceParams{
			Pos:   i,
			Param: formatParam(i, item),
		})
	}
	return traceParams
}

func formatParam(index int, item interface{}) string {
	val := reflect.ValueOf(item)
	if !val.IsValid() {
		return fmt.Sprintf("#%d: nil, ", index)
	}
	return fmt.Sprintf("%s, ", Output(item, val))
}

// exitTrace logs the exit of a function call and decrements the trace indentation.
func (t *TraceInstance) exitTrace(id uint64, name string, startTime time.Time, lastInsertId int64) {
	t.Lock()
	indent := t.decrementIndent(id)
	t.Unlock()

	indents := generateIndentString(indent - 1)
	t.log.Info(fmt.Sprintf("%s<-%s", indents, name), "gid", id, "timeCost(ms)", time.Since(startTime).String())
	// 更新时间
	_, err := t.db.Exec("UPDATE TraceData SET timeCost = ? WHERE id = ?", time.Since(startTime).String(), lastInsertId)
	if err != nil {
		t.log.Error("Unable to update time cost", "error", err)
	}
}

func (t *TraceInstance) decrementIndent(id uint64) int {
	indent := t.indentations[id]
	t.indentations[id]--
	return indent
}

// getGID retrieves the current goroutine's ID.
func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

// Trace is a decorator that traces function entry and exit.
func Trace(params []interface{}) func() {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		panic("not found caller")
	}

	id := getGID()
	fn := runtime.FuncForPC(pc)
	name := fn.Name()

	if skipFunction(name) {
		return func() {}
	}

	lastInsertId := singleTrace.enterTrace(id, name, params)
	return func() { singleTrace.exitTrace(id, name, time.Now(), lastInsertId) }
}

func skipFunction(name string) bool {
	return strings.Contains(name, "Config")
}

// Output formats trace parameters based on their type for logging.
func Output(item interface{}, val reflect.Value) string {
	switch val.Kind() {
	case reflect.Func:
		return runtime.FuncForPC(val.Pointer()).Name()
	case reflect.String:
		return val.String()
	case reflect.Ptr, reflect.Interface:
		return fmt.Sprintf("%+v", item)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", item)
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%.4f", item)
	default:
		return fmt.Sprintf("%s", item)
	}
}
