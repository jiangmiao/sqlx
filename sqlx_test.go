package sqlx

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

type User struct {
	Id        int64
	Name      string
	CreatedAt time.Time
	Amount    int64
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
		l.Fatal(err)
	}

	tx, err := db.Begin()
	if err != nil {
		l.Fatal(err)
	}

	q := Q{tx}
	_, err = q.Exec(`
	DROP TABLE IF EXISTS sqlx_test;
	CREATE TABLE sqlx_test(
		Id Serial PRIMARY KEY,
		Name Text,
		Amount Integer NOT NULL DEFAULT 123,
		CreatedAt Timestamptz NOT NULL DEFAULT now(),
		UNIQUE(Name)
	);
	`)
	ok(err)

	var user = User{
		Name: "Miao",
	}

	l.Info("CREATE")
	q.Create(&user, "name")
	eq(user.Amount, int64(123))
	eq(user.Id, int64(1))
	l.Info(user)

	user.Amount = 234
	q.Update(&user, "amount")
	eq(user.Amount, int64(234))

	user.Id = 0
	user.Name = "Miao2"
	q.Create(&user)
	eq(user.Amount, int64(234))

	var users []User
	q.Find(&users, "")

	eq(users[0].Id, int64(1))
	eq(users[1].Id, int64(2))

	tx.Rollback()

	_ = ok
	_ = eq
	_ = att
}
