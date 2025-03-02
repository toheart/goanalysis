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
	"github.com/sourcegraph/conc/pool"
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
const (
	EnableFlag  = "EnableFlag"
	IgnoreNames = "log,context"
)

var (
	once        sync.Once
	singleTrace *TraceInstance
)

// TraceInstance is a singleton structure that manages function tracing.
type TraceInstance struct {
	sync.Mutex
	indentations map[uint64]int
	log          *slog.Logger
	db           *sql.DB          // 修改为 sql.DB
	closed       bool             // 添加标志位表示是否已关闭
	insertChan   chan dbOperation // 添加通道用于异步数据库操作
	updateChan   chan dbOperation // 添加通道用于异步数据库操作
}

// dbOperation 定义数据库操作
type dbOperation struct {
	query    string
	args     []interface{}
	resultCh chan int64 // 用于返回操作结果的通道
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
			closed:       false,
			insertChan:   make(chan dbOperation, 20), // 创建带缓冲的通道
			updateChan:   make(chan dbOperation, 20), // 创建带缓冲的通道
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
			return
		}

		// 测试数据库连接
		if err = singleTrace.db.Ping(); err != nil {
			singleTrace.log.Error("Unable to connect to database", "error", err)
			return
		}

		_, err = singleTrace.db.Exec("CREATE TABLE IF NOT EXISTS TraceData (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, gid INTEGER, indent INTEGER, params TEXT, timeCost TEXT)") // 创建表
		if err != nil {
			singleTrace.log.Error("Unable to create table", "error", err)
			return
		}
		_, err = singleTrace.db.Exec("CREATE INDEX IF NOT EXISTS idx_gid ON TraceData (gid)") // 对Gid创建索引
		if err != nil {
			singleTrace.log.Error("Unable to create index on gid", "error", err)
			return
		}

		// 启动异步处理数据库操作的协程
		go singleTrace.processDBInsert()
		go singleTrace.processDBUpdate()
	})
	return singleTrace
}

func (t *TraceInstance) processDBInsert() {
	p := pool.New().WithMaxGoroutines(10)
	for op := range t.insertChan {
		p.Go(func() {
			result, err := t.db.Exec(op.query, op.args...)
			if err != nil {
				t.log.Error("Failed to execute insert query", "error", err)
				return
			}
			lastInsertId, err := result.LastInsertId()
			if err != nil {
				t.log.Error("Failed to get last insert ID", "error", err)
				return
			}
			op.resultCh <- lastInsertId
		})
	}
	p.Wait()
}

func (t *TraceInstance) processDBUpdate() {
	p := pool.New().WithMaxGoroutines(10)
	for op := range t.updateChan {
		p.Go(func() {
			lastInsertId := <-op.resultCh
			op.args = append(op.args, lastInsertId)
			result, err := t.db.Exec(op.query, op.args...)
			if err != nil {
				t.log.Error("Failed to execute update query", "error", err)
				return
			}
			t.log.Info("update query success", "lastInsertId", lastInsertId, "result", result)
			close(op.resultCh)
		})
	}

	p.Wait()
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
func (t *TraceInstance) enterTrace(id uint64, name string, params []interface{}) (resultCh chan int64, startTime time.Time) {
	startTime = time.Now() // 记录开始时间
	t.log.Info("enterTrace start", "id", id, "name", name, "startTime", startTime)
	t.Lock()
	t.log.Info("enterTrace Lock", "id", id, "name", name, "startTime", startTime)
	indent := t.incrementIndent(id)
	t.Unlock()
	t.log.Info("enterTrace end", "id", id, "name", name, "startTime", startTime)
	indents := generateIndentString(indent)
	traceParams := prepareParamsOutput(params)
	// 创建一个通道来接收异步操作的结果
	resultCh = make(chan int64, 1)
	// 将 traceParams 转换为 JSON 字符串
	paramsJSON, err := json.Marshal(traceParams)
	if err != nil {
		t.log.Error("Unable to marshal params to JSON", "error", err)
		return nil, startTime
	}

	// 异步执行数据库插入
	t.insertChan <- dbOperation{
		query:    "INSERT INTO TraceData (name, gid, indent, params) VALUES (?, ?, ?, ?)",
		args:     []interface{}{name, id, indent, paramsJSON},
		resultCh: resultCh,
	}

	// 记录日志，但不等待数据库操作完成
	t.log.Info(fmt.Sprintf("%s->%s", indents, name), "gid", id, "params", string(paramsJSON))

	// 返回通道，让调用者决定是否等待结果
	return resultCh, startTime
}

func (t *TraceInstance) incrementIndent(id uint64) int {
	indent := t.indentations[id]
	t.indentations[id]++
	return indent
}

func generateIndentString(indent int) string {
	return strings.Repeat("**", indent)
}

func prepareParamsOutput(params []interface{}) []*TraceParams {
	var traceParams []*TraceParams
	for i, item := range params {
		traceParams = append(traceParams, &TraceParams{
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
func (t *TraceInstance) exitTrace(id uint64, name string, startTime time.Time, resultCh chan int64) {
	t.log.Info("exitTrace", "id", id, "name", name, "startTime", startTime)
	t.Lock()
	t.log.Info("exitTrace Lock", "id", id, "name", name, "startTime", startTime)
	indent := t.decrementIndent(id)
	t.Unlock()

	duration := time.Since(startTime)
	indents := generateIndentString(indent - 1)
	t.log.Info(fmt.Sprintf("%s<-%s", indents, name), "gid", id, "timeCost(ms)", duration.String())

	t.updateChan <- dbOperation{
		query:    "UPDATE TraceData SET timeCost = ? WHERE id = ?",
		args:     []interface{}{duration.String()},
		resultCh: resultCh,
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
	instance := NewTraceInstance()
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

	lastInsertId, startTime := instance.enterTrace(id, name, params)
	return func() { instance.exitTrace(id, name, startTime, lastInsertId) }
}

func skipFunction(name string) bool {
	ignoreEnv := os.Getenv("IgnoreNames")
	var ignoreNames []string
	if ignoreEnv != "" {
		ignoreNames = strings.Split(ignoreEnv, ",")
	} else {
		ignoreNames = strings.Split(IgnoreNames, ",")
	}
	for _, ignoreName := range ignoreNames {
		if strings.Contains(strings.ToLower(name), ignoreName) {
			return true
		}
	}
	return false
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

// Close closes the database connection and releases resources.
func (t *TraceInstance) Close() error {
	t.Lock()
	defer t.Unlock()

	if t.closed {
		return nil
	}

	t.closed = true

	// 关闭数据库操作通道
	close(t.insertChan)
	close(t.updateChan)

	if t.db != nil {
		return t.db.Close()
	}
	return nil
}

// CloseTraceInstance closes the singleton trace instance.
func CloseTraceInstance() error {
	if singleTrace != nil {
		return singleTrace.Close()
	}
	return nil
}
