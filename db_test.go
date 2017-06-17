package sqlx

import (
	"database/sql"
	"log"
	"sync"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

type Foo struct {
	Now    time.Time
	Int    int64
	Text   string
	Double float64
	Six    string
}

type Bar struct {
	Id        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func TestDB(tt *testing.T) {
	att := assert.New(tt)
	ok := att.NoError
	eq := att.Equal
	db, err := Open("postgres", "dbname=sqlx_test sslmode=disable")
	ok(err)
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(10)
	var foos []Foo
	err = db.Query(&foos, `
		(SELECT now() as now, 3 as "int", '4' as "text", 5::double precision as "double", 6::numeric as "six") UNION
		(SELECT now() as now, 33 as "int", '44' as "text", 55::double precision as "double", 66::numeric as "six")
	`)
	ok(err)

	_, err = db.Exec(`
		DROP TABLE IF EXISTS Bars;
		CREATE TABLE Bars(id bigint primary key, name Text not null default '', created_at timestamptz not null default 'now()');
		INSERT INTO Bars(id) VALUES(1);
		INSERT INTO Bars(id) VALUES(2);
	`)
	ok(err)
	wg := sync.WaitGroup{}
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			wg.Done()
			var bars []Bar
			err = db.Query(&bars, "SELECT * FROM Bars")
			ok(err)
		}()
	}
	wg.Wait()

	_ = ok
	_ = eq
}

func BenchmarkTypeNameField(b *testing.B) {
	db, _ := Open("postgres", "dbname=test sslmode=disable")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var bars []Bar
		err := db.Query(&bars, "SELECT* FROM Bars")
		if err != nil {
			log.Fatal(err)
		}
	}
}
func BenchmarkRaw(b *testing.B) {
	db, _ := sql.Open("postgres", "dbname=test sslmode=disable")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var bars []Bar
		rows, err := db.Query("SELECT id, name, created_at FROM Bars")
		if err != nil {
			log.Fatal(err)
		}
		for rows.Next() {
			var bar Bar
			err = rows.Scan(&bar.Id, &bar.Name, &bar.CreatedAt)
			if err != nil {
				log.Fatal(err)
			}
			bars = append(bars, bar)
		}
		rows.Close()
	}
}
