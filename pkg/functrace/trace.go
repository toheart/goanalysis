package functrace

import (
	"bytes"
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"log/slog"
	"reflect"
	"runtime"
	"strconv"
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

var once sync.Once
var singleTrace *TraceInstance

func init() {
	NewTraceInstance()
}

type TraceInstance struct {
	sync.Mutex
	m map[uint64]int

	log *slog.Logger
}

func NewTraceInstance() *TraceInstance {
	once.Do(func() {
		log := slog.New(slog.NewTextHandler(&lumberjack.Logger{
			Filename:  "./trace.log",
			LocalTime: true,
			Compress:  true,
		}, nil))
		singleTrace = &TraceInstance{
			m:   make(map[uint64]int),
			log: log,
		}
	})
	return singleTrace
}

func (t *TraceInstance) enterTrace(id uint64, name string, params []interface{}) {
	// 先设置当前缩进
	t.Lock()
	indent := t.m[id]
	t.m[id] = indent + 1
	t.Unlock()

	indents := ""
	for i := 0; i < indent; i++ {
		indents += "**"
	}
	var output string
	for i, item := range params {
		// 获取 interface{} 的值
		val := reflect.ValueOf(item)

		// 处理 nil 值
		if !val.IsValid() {
			output += fmt.Sprintf("#%d: nil, ", i)
			continue
		}

		switch val.Kind() {
		case reflect.String:
			// 如果值是字符串类型，直接返回
			output += fmt.Sprintf("#%d: %s, ", i, val.String())
		case reflect.Ptr, reflect.Interface:
			// 如果值是指针或接口，尝试解引用
			output += fmt.Sprintf("#%d: %+v, ", i, item)
		default:
			output += fmt.Sprintf("#%d: %s, ", i, item)
		}
	}
	t.log.Info(fmt.Sprintf("%s->%s", indents, name), "gid", id, "params", output)
}

func (t *TraceInstance) exitTrace(id uint64, name string) {
	// 先减去当前缩进
	t.Lock()
	indent := t.m[id]
	t.m[id] = indent - 1
	t.Unlock()

	indents := ""
	for i := 0; i < indent-1; i++ {
		indents += "**"
	}
	t.log.Info(fmt.Sprintf("%s<-%s", indents, name), "gid", id)
}

func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

func Trace(params []interface{}) func() {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		panic("not found caller")
	}

	id := getGID()
	fn := runtime.FuncForPC(pc)
	name := fn.Name()

	singleTrace.enterTrace(id, name, params)
	return func() {
		singleTrace.exitTrace(id, name)
	}
}
