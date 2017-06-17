package sqlx

import "database/sql"

type Queryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

func query(q Queryer, dest interface{}, cmd string, args ...interface{}) error {
	rows, err := q.Query(cmd, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	return Scan(rows, dest)
}
