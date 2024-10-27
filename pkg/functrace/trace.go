package functrace

import (
	"bytes"
	"fmt"
	"log/slog"
	"runtime"
	"strconv"
	"sync"

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

func (t *TraceInstance) enterTrace(id uint64, name string, params ...interface{}) {
	// 先设置当前缩进
	t.Lock()
	indent := t.m[id]
	t.m[id] = indent + 1
	t.Unlock()

	indents := ""
	for i := 0; i < indent; i++ {
		indents += "\t"
	}
	var output string
	for _, item := range params {
		output += fmt.Sprintf("%s, ", item)
	}
	t.log.Info("enterTrace", "gid", id, "link", fmt.Sprintf("%s->%s", indents, name), "params", output)
}

func (t *TraceInstance) exitTrace(id uint64, name string) {
	// 先减去当前缩进
	t.Lock()
	indent := t.m[id]
	t.m[id] = indent - 1
	t.Unlock()

	indents := ""
	for i := 0; i < indent; i++ {
		indents += "\t"
	}
	t.log.Info("enterTrace", "gid", id, "link", fmt.Sprintf("%s<-%s", indents, name))
}

func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

func Trace(params ...interface{}) func() {
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
