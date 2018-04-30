package sqlx

import (
	"database/sql"
	"log"
	"reflect"
	"strings"
	"sync"
)

var mp = sync.Map{}

type Fields map[string]reflect.StructField

func Load(tv reflect.Type) Fields {
	var fields Fields
	fieldsPtr, ok := mp.Load(tv)
	if ok {
		fields = fieldsPtr.(Fields)
	} else {
		log.Println("Register", tv.Name())
		fields = Fields{}
		for i, n := 0, tv.NumField(); i < n; i++ {
			f := tv.Field(i)
			name := strings.ToLower(f.Name)
			fields[name] = f
		}
		mp.Store(tv, fields)
	}
	return fields
}

func Scan(pvs interface{}, rows *sql.Rows) (err error) {
	var isSlice bool
	rpvs := reflect.ValueOf(pvs)
	rvs := rpvs.Elem()
	tvs := rvs.Type()
	if tvs.Kind() == reflect.Slice {
		isSlice = true
		rvs = reflect.MakeSlice(tvs, 0, 16)
	} else {
		tvs = reflect.SliceOf(tvs)
	}
	cs, err := rows.Columns()
	if err != nil {
		return
	}
	tv := tvs.Elem()
	fields := Load(tv)
	csIdx := make([]*reflect.StructField, len(cs))
	rs := make([]interface{}, len(cs))
	prs := make([]interface{}, len(cs))
	for i, _ := range rs {
		prs[i] = &rs[i]
	}
	for i, c := range cs {
		f, ok := fields[c]
		if !ok {
			continue
		}
		csIdx[i] = &f
	}
	for rows.Next() {
		rows.Scan(prs...)
		rpv := reflect.New(tv)
		rv := rpv.Elem()
		for i, f := range csIdx {
			if f == nil {
				continue
			}
			var iv interface{}
			var fv = rv.FieldByIndex(f.Index)
			switch v := rs[i].(type) {
			case []byte:
				iv = string(v)
			default:
				iv = rs[i]
			}
			fv.Set(reflect.ValueOf(iv))
		}
		if isSlice {
			rvs = reflect.Append(rvs, rv)
		} else {
			rvs = rv
			break
		}
	}
	rpvs.Elem().Set(rvs)
	return
}
