package sqlx

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/lib/pq"
)

func ok(err error) {
	if err != nil {
		panic(err)
	}
}

type TableNamer interface {
	TableName() string
}

func GetTableName(iv interface{}) string {
	switch v := iv.(type) {
	case TableNamer:
		return v.TableName()
	default:
		return reflect.ValueOf(iv).Type().Name()
	}
}

var Quote = func(t string) string {
	return strings.ToLower(pq.QuoteIdentifier(t))
}

type Queryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type Q struct {
	Queryer
}

func Visit(pv interface{}, cs []string, proc func(f reflect.StructField, v interface{})) {
	rpv := reflect.ValueOf(pv)
	rv := rpv.Elem()
	tv := rv.Type()
	fs := Load(tv)
	for _, c := range cs {
		f, ok := fs[c]
		if !ok {
			log.Fatalf("cannot find %s\n", c)
		}
		proc(f, rv.FieldByIndex(f.Index).Interface())
	}
	return
}

func (q Q) Query(pvs interface{}, cmd string, args ...interface{}) (err error) {
	log.Println("QUERY", cmd, args)
	rows, err := q.Queryer.Query(cmd, args...)
	if err != nil {
		return
	}
	defer rows.Close()
	return Scan(pvs, rows)
}

func (q Q) Find(pvs interface{}, where string, args ...interface{}) error {
	tv := reflect.ValueOf(pvs).Elem().Type()
	if tv.Kind() == reflect.Slice {
		tv = tv.Elem()
	}
	tableName := Quote("bw" + tv.Name())
	if where != "" {
		where = " WHERE " + where
	}
	return q.Query(pvs, fmt.Sprintf("SELECT * FROM %s %s", tableName, where), args...)
}
func (q Q) Create(pv interface{}) (err error) {
	tv := reflect.TypeOf(pv).Elem()
	cs := []string{}
	for i, n := 0, tv.NumField(); i < n; i++ {
		f := tv.Field(i)
		if f.Name == "Id" {
			continue
		}
		cs = append(cs, f.Name)
	}
	return q.CreateEx(pv, cs...)
}

func (q Q) CreateEx(pv interface{}, cs ...string) (err error) {
	fs := []string{}
	ps := []string{}
	vs := []interface{}{}
	Visit(pv, cs, func(f reflect.StructField, v interface{}) {
		fs = append(fs, Quote(f.Name))
		ps = append(ps, fmt.Sprintf("$%d", len(fs)))
		vs = append(vs, v)
	})
	cmd := fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s) RETURNING *",
		Quote(GetTableName(pv)),
		strings.Join(fs, ","),
		strings.Join(ps, ","),
	)
	err = q.Query(pv, cmd, vs...)
	return
}

func (q Q) Update(pv interface{}, cs ...string) (err error) {
	rpv := reflect.ValueOf(pv)
	rv := rpv.Elem()
	tv := rv.Type()
	fs := Load(tv)

	i := 1
	vs := []interface{}{}
	ps := []string{}
	Visit(pv, cs, func(f reflect.StructField, v interface{}) {
		ps = append(ps, fmt.Sprintf("%s=$%d", Quote(f.Name), i))
		vs = append(vs, v)
		i++
	})
	cmd := fmt.Sprintf("UPDATE %s SET %s WHERE id=%d RETURNING *",
		Quote(GetTableName(pv)),
		strings.Join(ps, ","),
		rv.FieldByIndex(fs["id"].Index).Interface().(int64),
	)
	err = q.Query(pv, cmd, vs...)
	return
}

func (q Q) MustCreateEx(pv interface{}, cs ...string) {
	ok(q.CreateEx(pv, cs...))
}

func (q Q) MustCreate(pv interface{}) {
	ok(q.Create(pv))
}

func (q Q) MustUpdate(pv interface{}, cs ...string) {
	ok(q.Update(pv, cs...))
}

func (q Q) MustQuery(pvs interface{}, cmd string, args ...interface{}) {
	ok(q.Query(pvs, cmd, args...))
}

func (q Q) MustFind(pvs interface{}, where string, args ...interface{}) {
	ok(q.Find(pvs, where, args...))
}
