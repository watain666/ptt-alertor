package connections

import (
	"database/sql"
	"os"

	log "github.com/Ptt-Alertor/logrus"
	_ "github.com/lib/pq"
)

// var db = newDB()

func newDB() *sql.DB {
	d, err := sql.Open("postgres", os.Getenv("DB_CONNECTION"))
	if err != nil {
		log.Fatal(err)
	}
	defer d.Close()

	return d

	// d, err := sql.Open("postgres", conn)
	// if err != nil {
	// 	panic(err)
	// }
	// defer d.Close()
	// var version string
	// if err := d.QueryRow("select version()").Scan(&version); err != nil {
	// 	panic(err)
	// }

	// return d
}

// func DB() *sql.DB {
// 	return db
// }
