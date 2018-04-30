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

func Visit(pv interface{}, cs []string, proc func(c string, v interface{})) {
	rpv := reflect.ValueOf(pv)
	rv := rpv.Elem()
	tv := rv.Type()
	fs := Load(tv)
	for _, c := range cs {
		f, ok := fs[c]
		if !ok {
			log.Fatalf("cannot find %s\n", c)
		}
		proc(c, rv.FieldByIndex(f.Index).Interface())
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
	rpvs := reflect.ValueOf(pvs)
	rvs := rpvs.Elem()
	tvs := rvs.Type()
	tv := tvs
	if tvs.Kind() == reflect.Slice {
		tv = tvs.Elem()
	}
	tableName := Quote(GetTableName(tv))
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
		cs = append(cs, strings.ToLower(f.Name))
	}
	return q.CreateEx(pv, cs...)
}

func (q Q) CreateEx(pv interface{}, cs ...string) (err error) {
	fs := []string{}
	ps := []string{}
	vs := []interface{}{}
	Visit(pv, cs, func(c string, v interface{}) {
		if c == "id" {
			if v.(int64) != 0 {
				panic("create id must be zero")
			}
			return
		}
		fs = append(fs, Quote(c))
		ps = append(ps, fmt.Sprintf("$%d", len(fs)))
		vs = append(vs, v)
	})
	tv := reflect.TypeOf(pv).Elem()
	cmd := fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s) RETURNING *",
		Quote(GetTableName(tv)),
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

	vs := []interface{}{}
	ps := []string{}
	Visit(pv, cs, func(c string, v interface{}) {
		vs = append(vs, v)
		ps = append(ps, fmt.Sprintf("%s=$%d", Quote(c), len(vs)))
	})
	cmd := fmt.Sprintf("UPDATE %s SET %s WHERE id=%d RETURNING *",
		Quote(GetTableName(tv)),
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
