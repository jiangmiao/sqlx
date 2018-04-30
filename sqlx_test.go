package sqlx

import (
	"database/sql"
	"log"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

type User struct {
	Id     int64
	Name   string
	Amount int64
}

func (s User) TableName() string {
	return "sqlx_test"
}

func TestUser(tt *testing.T) {
	att := assert.New(tt)
	ok := att.NoError
	eq := att.Equal

	db, err := sql.Open("postgres", "dbname=test sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	q := Q{tx}
	_, err = q.Exec(`
	DROP TABLE IF EXISTS sqlx_test;
	CREATE TABLE sqlx_test(
		Id Serial PRIMARY KEY,
		Name Text,
		Amount Integer NOT NULL DEFAULT 123
	);
	`)
	ok(err)

	var user = User{
		Name: "Miao",
	}

	log.Println("CREATE")
	q.MustCreateEx(&user, "name")
	eq(user.Amount, int64(123))
	eq(user.Id, int64(1))

	user.Amount = 234
	q.MustUpdate(&user, "amount")
	eq(user.Amount, int64(234))

	rs, err := q.Exec("SELECT 1 FROM sqlx_test WHERE 0 = 0")
	log.Println(rs)
	log.Println(rs.LastInsertId())

	tx.Rollback()

	_ = ok
	_ = eq
	_ = att
}
