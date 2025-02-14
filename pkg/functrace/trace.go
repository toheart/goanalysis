package functrace

import (
	"bytes"
	"context"
	"entgo.io/ent/dialect"
	"fmt"
	"github.com/toheart/goanalysis/pkg/functrace/ent"
	"github.com/toheart/goanalysis/pkg/functrace/ent/trace"
	"gopkg.in/natefinch/lumberjack.v2"
	"log/slog"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
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
	log         *slog.Logger
)

func init() {
	NewTraceInstance()
}

// TraceInstance is a singleton structure that manages function tracing.
type TraceInstance struct {
	sync.Mutex
	indentations map[uint64]int
	enableDB     bool
	client       *ent.Client
	log          *slog.Logger
}

// NewTraceInstance initializes the singleton instance of TraceInstance.
func NewTraceInstance() *TraceInstance {
	once.Do(func() {
		initializeLogger()

		singleTrace = &TraceInstance{
			indentations: make(map[uint64]int),
			log:          log,
		}

		if os.Getenv(EnableFlag) != "" {
			initializeDatabase(singleTrace)
		}
	})
	return singleTrace
}

func initializeLogger() {
	log = slog.New(slog.NewTextHandler(&lumberjack.Logger{
		Filename:  "./trace.log",
		LocalTime: true,
		Compress:  true,
	}, nil))
}

func initializeDatabase(instance *TraceInstance) {
	instance.enableDB = true
	client, err := ent.Open(dialect.SQLite, "./trace.db?_fk=1")
	if err != nil {
		panic(err)
	}
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Error("failed creating schema resources: %v", err)
		panic(err)
	}
	instance.client = client
}

// enterTrace logs the entry of a function call and stores necessary trace details.
func (t *TraceInstance) enterTrace(id uint64, name string, params []interface{}) {
	t.Lock()
	indent := t.incrementIndent(id)
	t.Unlock()

	indents := generateIndentString(indent)
	output := prepareParamsOutput(params)

	if t.enableDB {
		t.logTraceInDB(id, indent, name, output)
	}
	t.log.Info(fmt.Sprintf("%s->%s", indents, name), "gid", id, "params", output)
}

func (t *TraceInstance) incrementIndent(id uint64) int {
	indent := t.indentations[id]
	t.indentations[id]++
	return indent
}

func generateIndentString(indent int) string {
	return strings.Repeat("**", indent)
}

func prepareParamsOutput(params []interface{}) string {
	var output string
	for i, item := range params {
		output += formatParam(i, item)
	}
	return output
}

func formatParam(index int, item interface{}) string {
	val := reflect.ValueOf(item)
	if !val.IsValid() {
		return fmt.Sprintf("#%d: nil, ", index)
	}
	return fmt.Sprintf("#%d: %s, ", index, Output(item, val))
}

func (t *TraceInstance) logTraceInDB(id uint64, indent int, name, output string) {
	_, err := t.client.Trace.Create().SetGid(int(id)).SetIndent(indent).SetFuncName(name).SetParams(output).Save(context.Background())
	if err != nil {
		t.log.Error("insert found err: %s", err)
	}
}

func (t *TraceInstance) GetTracesByGID(gid uint64) ([]*ent.Trace, error) {
	if t.client == nil {
		return nil, fmt.Errorf("database client is not initialized")
	}

	// 使用 ent 客户端查询数据库
	traces, err := t.client.Trace.Query().
		Where(trace.GidEQ(int(gid))).      // 查询条件是 gid
		Order(ent.Asc(trace.FieldIndent)). // 按照 Trace 的缩进层级排序
		All(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get traces by GID: %w", err)
	}

	return traces, nil
}

// exitTrace logs the exit of a function call and decrements the trace indentation.
func (t *TraceInstance) exitTrace(id uint64, name string) {
	t.Lock()
	indent := t.decrementIndent(id)
	t.Unlock()

	indents := generateIndentString(indent - 1)
	t.log.Info(fmt.Sprintf("%s<-%s", indents, name), "gid", id)
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

	singleTrace.enterTrace(id, name, params)
	return func() { singleTrace.exitTrace(id, name) }
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
