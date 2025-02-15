package functrace

import (
	"bytes"
	"fmt"
	"log/slog"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"go.etcd.io/bbolt"
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

type TraceData struct {
	Name   string
	GID    uint64
	Indent int
	Params string
}

func init() {
	NewTraceInstance()
}

// TraceInstance is a singleton structure that manages function tracing.
type TraceInstance struct {
	sync.Mutex
	indentations map[uint64]int
	enableDB     bool
	log          *slog.Logger
	db           *bbolt.DB
}

// NewTraceInstance initializes the singleton instance of TraceInstance.
func NewTraceInstance() *TraceInstance {
	once.Do(func() {
		var err error
		singleTrace = &TraceInstance{
			indentations: make(map[uint64]int),
			log:          log,
		}
		singleTrace.db, err = bbolt.Open("trace.db", 0600, nil)
		if err != nil {
			log.Error("无法打开数据库", "error", err)
		}
		singleTrace.db.Update(func(tx *bbolt.Tx) error {
			_, err := tx.CreateBucket([]byte("TraceData"))
			return err
		})
	})
	return singleTrace
}

// enterTrace logs the entry of a function call and stores necessary trace details.
func (t *TraceInstance) enterTrace(id uint64, name string, params []interface{}) {
	t.Lock()
	indent := t.incrementIndent(id)
	t.Unlock()

	indents := generateIndentString(indent)
	output := prepareParamsOutput(params)

	t.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("TraceData"))
		traceData := TraceData{
			Name:   name,
			GID:    id,
			Indent: indent,
			Params: output,
		}
		return b.Put([]byte(fmt.Sprintf("%d", id)), []byte(fmt.Sprintf("%+v", traceData)))
	})

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
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	return fmt.Sprintf("#%d: %v, ", index, val.Interface())
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
