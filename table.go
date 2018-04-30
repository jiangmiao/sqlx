package sqlx

import (
	"database/sql"
	"fmt"
	"log"
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
	fields = Underscore(fields)
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
	sqlcmd := fmt.Sprintf("SELECT * FROM %s WHERE %s", b.Name, where)
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
		name := Underscore(m)
		fs[i] = Quote(name)
		ps[i] = fmt.Sprintf("$%d", i+1)
		f, ok := nameField[name]
		if !ok {
			panic(fmt.Errorf("cannot find field %s", name))
		}
		args[i] = rsrc.FieldByIndex(f.Index).Interface()
	}
	sqlcmd := fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s) RETURNING *", b.Name,
		strings.Join(fs, ","),
		strings.Join(ps, ","))
	return b.Query(dest, sqlcmd, args...)
}

func (t *Table) FindOrCreate(ob interface{}, fields string) (err error) {
	err = t.Find(ob, fields, ob)
	if err == sql.ErrNoRows {
		rpob := reflect.ValueOf(ob)
		rob := rpob.Elem()
		f, ok := Get(rob.Type())["id"]
		if !ok {
			panic("cannot find id")
		}
		id := rob.FieldByIndex(f.Index).Interface().(int64)
		if id == 0 {
			err = t.Save(ob)
		}
	}
	return
}

func (t *Table) Save(src interface{}) (err error) {
	rsrc := reflect.Indirect(reflect.ValueOf(src))
	tsrc := rsrc.Type()
	n := tsrc.NumField()
	ps := make([]string, n)
	fs := make([]string, n)
	args := make([]interface{}, n)
	k := 0
	var id int64
	for i := 0; i < n; i++ {
		f := tsrc.Field(i)
		name := Underscore(f.Name)
		if name == "id" {
			id = rsrc.Field(i).Interface().(int64)
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
	var sqlcmd string
	if id != 0 {
		del := fmt.Sprintf("DELETE FROM %s WHERE id = %d;", t.Name, id)
		_, err = t.Exec(del)
		if err != nil {
			log.Fatal(err)
		}
	}
	sqlcmd += fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s) RETURNING *", t.Name,
		strings.Join(fs, ","),
		strings.Join(ps, ","))
	return t.Query(src, sqlcmd, args...)
}
