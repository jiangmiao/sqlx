package sqlx

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type Table struct {
	SqlxQueryer
	Name string
}

var FieldPattern = regexp.MustCompile(`\w+`)

func (b *Table) Find(dest interface{}, fields string, src interface{}) error {
	rsrc := reflect.Indirect(reflect.ValueOf(src))
	nameField := Get(rsrc.Type())
	args := make([]interface{}, 0)
	where := FieldPattern.ReplaceAllStringFunc(fields, func(m string) string {
		if m == "or" || m == "and" {
			return m
		}
		name := m
		f, ok := nameField[name]
		if !ok {
			panic(fmt.Errorf("cannot find field %s", name))
		}
		args = append(args, rsrc.FieldByIndex(f.Index).Interface())
		return fmt.Sprintf("%s = $%d", Quote(name), len(args))
	})
	sqlcmd := fmt.Sprintf("SELECT * FROM %s WHERE %s", Quote(b.Name), where)
	return b.Query(dest, sqlcmd, args...)
}

func (b *Table) Insert(dest interface{}, fields string, src interface{}) error {
	rsrc := reflect.Indirect(reflect.ValueOf(src))
	nameField := Get(rsrc.Type())

	ms := FieldPattern.FindAllString(fields, -1)
	ps := make([]string, len(ms))
	fs := make([]string, len(ms))
	args := make([]interface{}, len(ms))
	for i, m := range ms {
		name := m
		fs[i] = Quote(name)
		ps[i] = fmt.Sprintf("$%d", i+1)
		f, ok := nameField[name]
		if !ok {
			panic(fmt.Errorf("cannot find field %s", name))
		}
		args[i] = rsrc.FieldByIndex(f.Index).Interface()
	}
	sqlcmd := fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s) RETURNING *", Quote(b.Name),
		strings.Join(fs, ","),
		strings.Join(ps, ","))
	return b.Query(dest, sqlcmd, args...)
}

func (t *Table) Save(src interface{}) (err error) {
	rsrc := reflect.Indirect(reflect.ValueOf(src))
	tsrc := rsrc.Type()
	n := tsrc.NumField()
	ps := make([]string, n)
	fs := make([]string, n)
	args := make([]interface{}, n)
	k := 0
	for i := 0; i < n; i++ {
		f := tsrc.Field(i)
		name := strings.ToLower(Underscore(f.Name))
		if name == "id" {
			continue
		}
		fs[k] = Quote(name)
		ps[k] = fmt.Sprintf("$%d", k+1)
		args[k] = rsrc.Field(i).Interface()
		k++
	}
	fs = fs[:k]
	ps = ps[:k]
	args = args[:k]
	sqlcmd := fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s) RETURNING *", Quote(t.Name),
		strings.Join(fs, ","),
		strings.Join(ps, ","))
	return t.Query(src, sqlcmd, args...)
}
