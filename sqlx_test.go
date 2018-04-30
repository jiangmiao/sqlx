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
		Amount Integer NOT NULL DEFAULT 123,
		UNIQUE(Name)
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

	user.Id = 0
	user.Name = "Miao2"
	q.MustCreate(&user)

	var users []User
	q.MustFind(&users, "")

	eq(users[0].Id, int64(1))
	eq(users[1].Id, int64(2))

	tx.Rollback()

	_ = ok
	_ = eq
	_ = att
}
