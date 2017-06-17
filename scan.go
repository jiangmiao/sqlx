package sqlx

import (
	"database/sql"
	"fmt"
	. "reflect"
	"strings"
	"sync"
)

type TypeNameField struct {
	Map map[Type]map[string]StructField
	sync.RWMutex
}

func NewTypeNameField() *TypeNameField {
	return &TypeNameField{
		Map: make(map[Type]map[string]StructField),
	}
}

func (ob TypeNameField) Get(to Type) map[string]StructField {
	ob.RLock()
	nameField, ok := ob.Map[to]
	ob.RUnlock()
	if !ok {
		nameField = ob.Register(to)
	}
	return nameField
}

func (ob TypeNameField) Register(to Type) map[string]StructField {
	ob.Lock()
	defer ob.Unlock()
	nameField, found := ob.Map[to]
	if found {
		return nameField
	}
	nameField = make(map[string]StructField)
	n := to.NumField()
	for i := 0; i < n; i++ {
		f := to.Field(i)
		nameField[strings.ToLower(f.Name)] = f
		nameField[strings.ToLower(Underscore(f.Name))] = f
	}
	ob.Map[to] = nameField
	return nameField
}

var typeNameField = NewTypeNameField()

func register(to Type) map[string]StructField {
	return typeNameField.Register(to)
}

func Scan(rs *sql.Rows, pos interface{}) (err error) {
	// r reflect
	// p pointer
	// o object
	// s plural
	// t type
	rpos := ValueOf(pos)
	ros := rpos.Elem()
	tos := ros.Type()
	var to Type
	var isSlice bool
	if tos.Kind() == Slice {
		// tos = []Foo
		ros = New(tos).Elem()
		to = tos.Elem()
		isSlice = true
	} else {
		// tos = Foo
		to = tos
	}
	nameField := typeNameField.Get(to)

	ks, err := rs.Columns()
	if err != nil {
		return
	}
	var vs = make([]interface{}, len(ks))
	var pvs = make([]interface{}, len(ks))

	for i, _ := range ks {
		pvs[i] = &vs[i]
	}

	for rs.Next() {
		err = rs.Scan(pvs...)
		if err != nil {
			return err
		}
		ro := New(to).Elem()
		for i, k := range ks {
			f, found := nameField[k]
			if !found {
				continue
			}
			idx := f.Index
			var iv interface{}
			switch v := vs[i].(type) {
			case []byte:
				iv = string(v)
			default:
				iv = v
			}
			if iv == nil {
				continue
			}
			riv := ValueOf(iv)
			if f.Type != riv.Type() {
				return fmt.Errorf("%v != %v", f.Type, riv.Type())
			}
			ro.FieldByIndex(idx).Set(riv)
		}
		if isSlice {
			ros = Append(ros, ro)
		} else {
			ros = ro
			break
		}
	}
	rpos.Elem().Set(ros)
	return nil
}
