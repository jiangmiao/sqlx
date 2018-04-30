package sqlx

import "database/sql"

type Tx struct {
	*sql.Tx
}

func (tx *Tx) Query(dest interface{}, cmd string, args ...interface{}) (err error) {
	return query(tx.Tx, dest, cmd, args...)
}

func (tx *Tx) End(err *error) {
	if *err != nil {
		tx.Rollback()
	} else {
		tx.Commit()
	}
}

func (tx *Tx) Table(tableName string) *Table {
	return &Table{
		Name:        tableName,
		SqlxQueryer: tx,
	}
}
