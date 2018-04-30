package sqlx

import (
	"database/sql"
	"log"
)

type Queryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type SqlxQueryer interface {
	Query(dest interface{}, query string, args ...interface{}) error
	Exec(query string, args ...interface{}) (sql.Result, error)
	End(err *error)
}

var Debug bool

func query(q Queryer, dest interface{}, cmd string, args ...interface{}) error {
	if Debug {
		log.Println(cmd, args)
	}
	rows, err := q.Query(cmd, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	return Scan(rows, dest)
}
