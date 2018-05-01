package sqlx

import (
	"database/sql"
	"reflect"
	"strings"
	"sync"
)

// Type - ColumnName - Field
var tv2cn2f = sync.Map{}

// ColumnName - Field
type CN2F map[string]reflect.StructField

// Type - TableName
var tv2tn = make(map[reflect.Type]string)

type TableNamer interface {
	TableName() string
}

func GetTableName(tv reflect.Type) string {
	name, ok := tv2tn[tv]
	if !ok {
		panic("cannot find table name " + tv.Name())
	}
	return name
}

// init table info when load
func Load(tv reflect.Type) CN2F {
	var cn2f CN2F
	icn2f, ok := tv2cn2f.Load(tv)
	if ok {
		cn2f = icn2f.(CN2F)
	} else {
		l.Debugln("Register", tv.Name())
		cn2f = CN2F{}
		for i, n := 0, tv.NumField(); i < n; i++ {
			f := tv.Field(i)
			cn := strings.ToLower(f.Name)
			cn2f[cn] = f
		}
		switch v := reflect.New(tv).Interface().(type) {
		case TableNamer:
			tv2tn[tv] = v.TableName()
		default:
			tv2tn[tv] = strings.ToLower(tv.Name())
		}
		tv2cn2f.Store(tv, cn2f)
	}
	return cn2f
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
	cn2f := Load(tv)
	csIdx := make([]*reflect.StructField, len(cs))
	rs := make([]interface{}, len(cs))
	prs := make([]interface{}, len(cs))
	for i, _ := range rs {
		prs[i] = &rs[i]
	}
	for i, c := range cs {
		f, ok := cn2f[c]
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
