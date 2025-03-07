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
	indentations map[uint64]*TraceIndent
	log          *slog.Logger
	db           *sql.DB          // 修改为 sql.DB
	closed       bool             // 添加标志位表示是否已关闭
	updateChan   chan dbOperation // 添加通道用于异步数据库操作
}

// TraceIndent 存储函数调用的缩进信息和父函数名称
type TraceIndent struct {
	Indent      int            // 当前缩进级别
	ParentFuncs map[int]string // 每一层的父函数名称
}

// dbOperation 定义数据库操作
type dbOperation struct {
	query string
	args  []interface{}
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
			indentations: make(map[uint64]*TraceIndent),
			log:          initializeLogger(),
			closed:       false,
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

		_, err = singleTrace.db.Exec("CREATE TABLE IF NOT EXISTS TraceData (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, gid INTEGER, indent INTEGER, params TEXT, timeCost TEXT, parentFuncname TEXT)") // 创建表，添加 parentFuncname 字段
		if err != nil {
			singleTrace.log.Error("Unable to create table", "error", err)
			return
		}
		_, err = singleTrace.db.Exec("CREATE INDEX IF NOT EXISTS idx_gid ON TraceData (gid)") // 对Gid创建索引
		if err != nil {
			singleTrace.log.Error("Unable to create index on gid", "error", err)
			return
		}
		// 添加对 parentFuncname 的索引
		_, err = singleTrace.db.Exec("CREATE INDEX IF NOT EXISTS idx_parent ON TraceData (parentFuncname)")
		if err != nil {
			singleTrace.log.Error("Unable to create index on parentFuncname", "error", err)
			return
		}

		// 启动异步处理数据库操作的协程
		go singleTrace.processDBUpdate()
	})
	return singleTrace
}

func (t *TraceInstance) processDBUpdate() {
	p := pool.New().WithMaxGoroutines(10)
	for op := range t.updateChan {
		p.Go(func() {
			result, err := t.db.Exec(op.query, op.args...)
			if err != nil {
				t.log.Error("Failed to execute update query", "error", err)
				return
			}
			t.log.Info("update query success", "result", result)
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
func (t *TraceInstance) enterTrace(id uint64, name string, params []interface{}) (lastInsertId int64, startTime time.Time) {
	startTime = time.Now() // 记录开始时间
	t.Lock()

	// 获取或初始化 TraceIndent
	traceIndent, exists := t.indentations[id]
	if !exists {
		traceIndent = &TraceIndent{
			Indent:      0,
			ParentFuncs: make(map[int]string),
		}
		t.indentations[id] = traceIndent
	}

	// 获取当前缩进和父函数名称
	indent := traceIndent.Indent
	parentFunc := traceIndent.ParentFuncs[indent-1] // 获取上一层的函数名称作为父函数

	// 更新缩进和父函数名称
	traceIndent.ParentFuncs[indent] = name // 当前层的函数名称
	traceIndent.Indent++

	t.Unlock()

	// 生成缩进字符串
	indents := generateIndentString(indent)

	// 准备参数输出
	traceParams := prepareParamsOutput(params)

	// 将 traceParams 转换为 JSON 字符串
	paramsJSON, err := json.Marshal(traceParams)
	if err != nil {
		t.log.Error("无法将参数转换为JSON", "错误", err)
		return 0, startTime
	}

	// 在插入数据时包含父函数名称
	result, err := t.db.Exec("INSERT INTO TraceData (name, gid, indent, params, parentFuncname) VALUES (?, ?, ?, ?, ?)",
		name, id, indent, paramsJSON, parentFunc)
	if err != nil {
		t.log.Error("执行插入查询失败", "错误", err)
		return
	}

	lastInsertId, err = result.LastInsertId()
	if err != nil {
		t.log.Error("获取最后插入ID失败", "错误", err)
		return
	}

	// 构建更美观的日志输出
	var logBuilder strings.Builder
	logBuilder.WriteString(fmt.Sprintf("%s➡️ %s", indents, name))

	if parentFunc != "" {
		logBuilder.WriteString(fmt.Sprintf(" (父函数: %s)", parentFunc))
	}

	// 记录日志，包含更多有用信息
	t.log.Info(logBuilder.String(),
		"goroutine", id,
		"参数", string(paramsJSON),
		"时间", startTime.Format("15:04:05.000"))

	return lastInsertId, startTime
}

func generateIndentString(indent int) string {
	return strings.Repeat("**", indent)
}

func prepareParamsOutput(params []interface{}) []*TraceParams {
	var traceParams []*TraceParams

	// 如果没有参数，返回一个特殊标记
	if len(params) == 0 {
		traceParams = append(traceParams, &TraceParams{
			Pos:   -1,
			Param: "无参数",
		})
		return traceParams
	}

	// 处理参数
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
		return fmt.Sprintf("#%d: nil", index)
	}

	// 获取参数类型名称
	typeName := val.Type().String()

	// 使用 Output 函数获取格式化后的值
	formattedValue := Output(item, val)

	return fmt.Sprintf("#%d(%s): %s", index, typeName, formattedValue)
}

// exitTrace logs the exit of a function call and decrements the trace indentation.
func (t *TraceInstance) exitTrace(id uint64, name string, startTime time.Time, lastInsertId int64) {
	t.Lock()

	// 获取 TraceIndent
	traceIndent, exists := t.indentations[id]
	if !exists {
		t.log.Error("找不到goroutine的TraceIndent", "goroutine", id)
		t.Unlock()
		return
	}

	// 获取当前缩进
	indent := traceIndent.Indent

	// 删除当前层的父函数名称
	delete(traceIndent.ParentFuncs, indent-1)

	// 更新缩进
	traceIndent.Indent--

	// 如果缩进小于等于0，清除所有父函数名称
	if traceIndent.Indent <= 0 {
		traceIndent.ParentFuncs = make(map[int]string)
	}

	t.Unlock()

	// 计算函数执行时间
	duration := time.Since(startTime)

	// 生成缩进字符串
	indents := generateIndentString(indent - 1)

	// 构建更美观的日志输出
	var logBuilder strings.Builder
	logBuilder.WriteString(fmt.Sprintf("%s⬅️ %s", indents, name))

	// 格式化执行时间
	durationStr := formatDuration(duration)

	// 记录日志，包含更多有用信息
	t.log.Info(logBuilder.String(),
		"goroutine", id,
		"耗时", durationStr,
		"时间", time.Now().Format("15:04:05.000"))

	// 异步更新数据库
	t.updateChan <- dbOperation{
		query: "UPDATE TraceData SET timeCost = ? WHERE id = ?",
		args:  []interface{}{duration.String(), lastInsertId},
	}
}

// formatDuration 格式化持续时间，使其更易读
func formatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%d ns", d.Nanoseconds())
	} else if d < time.Millisecond {
		return fmt.Sprintf("%.2f µs", float64(d.Nanoseconds())/1000)
	} else if d < time.Second {
		return fmt.Sprintf("%.2f ms", float64(d.Nanoseconds())/1000000)
	} else {
		return fmt.Sprintf("%.2f s", d.Seconds())
	}
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

	// 确保 TraceIndent 已初始化
	instance.Lock()
	if _, exists := instance.indentations[id]; !exists {
		instance.indentations[id] = &TraceIndent{
			Indent:      0,
			ParentFuncs: make(map[int]string),
		}
	}
	instance.Unlock()

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
	if !val.IsValid() {
		return "nil"
	}

	switch val.Kind() {
	case reflect.Func:
		return fmt.Sprintf("func(%s)", runtime.FuncForPC(val.Pointer()).Name())
	case reflect.String:
		// 不再限制字符串长度，完整输出
		return fmt.Sprintf("\"%s\"", val.String())
	case reflect.Ptr:
		if val.IsNil() {
			return "nil"
		}
		return fmt.Sprintf("&%s", Output(val.Elem().Interface(), val.Elem()))
	case reflect.Interface:
		if val.IsNil() {
			return "nil"
		}
		return Output(val.Elem().Interface(), val.Elem())
	case reflect.Struct:
		// 完整输出结构体内容，并添加类型信息
		typeName := val.Type().String()
		return fmt.Sprintf("%s: %+v", typeName, item)
	case reflect.Map:
		// 完整输出 Map 内容，并添加类型信息
		if val.IsNil() {
			return "nil"
		}
		typeName := val.Type().String()
		return fmt.Sprintf("%s: %+v", typeName, item)
	case reflect.Slice, reflect.Array:
		// 完整输出切片/数组内容，并添加类型信息
		if val.Kind() == reflect.Slice && val.IsNil() {
			return "nil"
		}
		typeName := val.Type().String()
		return fmt.Sprintf("%s: %+v", typeName, item)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", val.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", val.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%.4f", val.Float())
	case reflect.Bool:
		return fmt.Sprintf("%v", val.Bool())
	case reflect.Chan:
		if val.IsNil() {
			return "nil"
		}
		typeName := val.Type().String()
		return fmt.Sprintf("%s: (chan)", typeName)
	case reflect.Complex64, reflect.Complex128:
		return fmt.Sprintf("%v", val.Complex())
	default:
		typeName := val.Type().String()
		return fmt.Sprintf("%s: %+v", typeName, item)
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
