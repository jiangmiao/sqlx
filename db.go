package sqlx

import (
	"database/sql"
)

type DB struct {
	*sql.DB
}

func Open(driver, source string) (*DB, error) {
	db, err := sql.Open(driver, source)
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

func (db *DB) Query(dest interface{}, cmd string, args ...interface{}) (err error) {
	return query(db.DB, dest, cmd, args...)
}

func (db *DB) Begin() (*Tx, error) {
	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{tx}, err
}
