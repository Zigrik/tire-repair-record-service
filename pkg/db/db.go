package db

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

const schema string = `
CREATE TABLE tire_service (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	title VARCHAR NOT NULL DEFAULT "",
	record DATETIME,
	comment VARCHAR(128),
	status VARCHAR(32)
);`

var db *sql.DB

var (
	StartTime   = time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)  // 09:00
	FinishTime  = time.Date(0, 1, 1, 18, 0, 0, 0, time.UTC) // 18:00
	Interval    = 30                                        // интервал в минутах
	MinLeadTime = time.Duration(Interval) * time.Minute     // минимальное время для записи от текущего момента
)

type Record struct {
	ID      int64
	Date    time.Time
	Title   string
	Record  *time.Time // может быть nil (текущая очередь)
	Comment string
	Status  string
}

func CloseDatabase() {
	db.Close()
}

func Init(dbFile string, logger *log.Logger) error {

	var err error
	db, err = sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}

	var install bool
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		install = true
	}

	if install {
		file, err := os.Create(dbFile)
		if err != nil {
			return err
		}
		file.Close()
		logger.Printf("INFO: the %s file has been created\n", dbFile)

		_, err = db.Exec(schema)
		if err != nil {
			return err
		}
	}

	logger.Printf("INFO: the %s database is ready for use\n", dbFile)
	return nil
}
