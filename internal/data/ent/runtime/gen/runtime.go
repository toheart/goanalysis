// Code generated by ent, DO NOT EDIT.

package gen

import (
	"github.com/toheart/goanalysis/internal/data/ent/runtime/gen/goroutinetrace"
	"github.com/toheart/goanalysis/internal/data/ent/runtime/gen/paramstoredata"
	"github.com/toheart/goanalysis/internal/data/ent/runtime/gen/tracedata"
	"github.com/toheart/goanalysis/internal/data/ent/runtime/schema"
)

// The init function reads all schema descriptors with runtime code
// (default values, validators, hooks and policies) and stitches it
// to their package variables.
func init() {
	goroutinetraceFields := schema.GoroutineTrace{}.Fields()
	_ = goroutinetraceFields
	// goroutinetraceDescOriginGid is the schema descriptor for originGid field.
	goroutinetraceDescOriginGid := goroutinetraceFields[1].Descriptor()
	// goroutinetrace.OriginGidValidator is a validator for the "originGid" field. It is called by the builders before save.
	goroutinetrace.OriginGidValidator = goroutinetraceDescOriginGid.Validators[0].(func(uint64) error)
	// goroutinetraceDescID is the schema descriptor for id field.
	goroutinetraceDescID := goroutinetraceFields[0].Descriptor()
	// goroutinetrace.IDValidator is a validator for the "id" field. It is called by the builders before save.
	goroutinetrace.IDValidator = goroutinetraceDescID.Validators[0].(func(int64) error)
	paramstoredataFields := schema.ParamStoreData{}.Fields()
	_ = paramstoredataFields
	// paramstoredataDescData is the schema descriptor for data field.
	paramstoredataDescData := paramstoredataFields[3].Descriptor()
	// paramstoredata.DefaultData holds the default value on creation for the data field.
	paramstoredata.DefaultData = paramstoredataDescData.Default.(string)
	// paramstoredataDescIsReceiver is the schema descriptor for isReceiver field.
	paramstoredataDescIsReceiver := paramstoredataFields[4].Descriptor()
	// paramstoredata.DefaultIsReceiver holds the default value on creation for the isReceiver field.
	paramstoredata.DefaultIsReceiver = paramstoredataDescIsReceiver.Default.(bool)
	tracedataFields := schema.TraceData{}.Fields()
	_ = tracedataFields
	// tracedataDescName is the schema descriptor for name field.
	tracedataDescName := tracedataFields[1].Descriptor()
	// tracedata.NameValidator is a validator for the "name" field. It is called by the builders before save.
	tracedata.NameValidator = tracedataDescName.Validators[0].(func(string) error)
	// tracedataDescIndent is the schema descriptor for indent field.
	tracedataDescIndent := tracedataFields[3].Descriptor()
	// tracedata.DefaultIndent holds the default value on creation for the indent field.
	tracedata.DefaultIndent = tracedataDescIndent.Default.(int)
	// tracedataDescParamsCount is the schema descriptor for paramsCount field.
	tracedataDescParamsCount := tracedataFields[4].Descriptor()
	// tracedata.DefaultParamsCount holds the default value on creation for the paramsCount field.
	tracedata.DefaultParamsCount = tracedataDescParamsCount.Default.(int)
	// tracedataDescID is the schema descriptor for id field.
	tracedataDescID := tracedataFields[0].Descriptor()
	// tracedata.IDValidator is a validator for the "id" field. It is called by the builders before save.
	tracedata.IDValidator = tracedataDescID.Validators[0].(func(int) error)
}
