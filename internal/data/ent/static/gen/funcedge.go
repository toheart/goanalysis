// Code generated by ent, DO NOT EDIT.

package gen

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/toheart/goanalysis/internal/data/ent/static/gen/funcedge"
)

// FuncEdge is the model entity for the FuncEdge schema.
type FuncEdge struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// CreatedAt holds the value of the "CreatedAt" field.
	CreatedAt time.Time `json:"CreatedAt,omitempty"`
	// UpdatedAt holds the value of the "UpdatedAt" field.
	UpdatedAt time.Time `json:"UpdatedAt,omitempty"`
	// CallerKey holds the value of the "CallerKey" field.
	CallerKey string `json:"CallerKey,omitempty"`
	// CalleeKey holds the value of the "CalleeKey" field.
	CalleeKey    string `json:"CalleeKey,omitempty"`
	selectValues sql.SelectValues
}

// scanValues returns the types for scanning values from sql.Rows.
func (*FuncEdge) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case funcedge.FieldID:
			values[i] = new(sql.NullInt64)
		case funcedge.FieldCallerKey, funcedge.FieldCalleeKey:
			values[i] = new(sql.NullString)
		case funcedge.FieldCreatedAt, funcedge.FieldUpdatedAt:
			values[i] = new(sql.NullTime)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the FuncEdge fields.
func (fe *FuncEdge) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case funcedge.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			fe.ID = int(value.Int64)
		case funcedge.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field CreatedAt", values[i])
			} else if value.Valid {
				fe.CreatedAt = value.Time
			}
		case funcedge.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field UpdatedAt", values[i])
			} else if value.Valid {
				fe.UpdatedAt = value.Time
			}
		case funcedge.FieldCallerKey:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field CallerKey", values[i])
			} else if value.Valid {
				fe.CallerKey = value.String
			}
		case funcedge.FieldCalleeKey:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field CalleeKey", values[i])
			} else if value.Valid {
				fe.CalleeKey = value.String
			}
		default:
			fe.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the FuncEdge.
// This includes values selected through modifiers, order, etc.
func (fe *FuncEdge) Value(name string) (ent.Value, error) {
	return fe.selectValues.Get(name)
}

// Update returns a builder for updating this FuncEdge.
// Note that you need to call FuncEdge.Unwrap() before calling this method if this FuncEdge
// was returned from a transaction, and the transaction was committed or rolled back.
func (fe *FuncEdge) Update() *FuncEdgeUpdateOne {
	return NewFuncEdgeClient(fe.config).UpdateOne(fe)
}

// Unwrap unwraps the FuncEdge entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (fe *FuncEdge) Unwrap() *FuncEdge {
	_tx, ok := fe.config.driver.(*txDriver)
	if !ok {
		panic("gen: FuncEdge is not a transactional entity")
	}
	fe.config.driver = _tx.drv
	return fe
}

// String implements the fmt.Stringer.
func (fe *FuncEdge) String() string {
	var builder strings.Builder
	builder.WriteString("FuncEdge(")
	builder.WriteString(fmt.Sprintf("id=%v, ", fe.ID))
	builder.WriteString("CreatedAt=")
	builder.WriteString(fe.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("UpdatedAt=")
	builder.WriteString(fe.UpdatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("CallerKey=")
	builder.WriteString(fe.CallerKey)
	builder.WriteString(", ")
	builder.WriteString("CalleeKey=")
	builder.WriteString(fe.CalleeKey)
	builder.WriteByte(')')
	return builder.String()
}

// FuncEdges is a parsable slice of FuncEdge.
type FuncEdges []*FuncEdge
