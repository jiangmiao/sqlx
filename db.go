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
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(50)
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

func (db *DB) Commit() (err error) {
	return
}

func (db *DB) Rollback() (err error) {
	return
}

func (db *DB) End(err *error) {
	return
}

func (db *DB) Table(tableName string) *Table {
	return &Table{
		Name:        tableName,
		SqlxQueryer: db,
	}
}
func (db *DB) BeginTable(tableName string) (*Table, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	return &Table{
		Name:        tableName,
		SqlxQueryer: tx,
	}, nil
}
