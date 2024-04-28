package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type Postgres struct {
	Db *sql.DB
}

func New() (*Postgres, error) {

	// db connection
	// var err error

	// err = godotenv.Load()
	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// }

	psqlInfo := os.Getenv("DATABASE_URL")

	db, err := sql.Open("postgres", psqlInfo)

	if err != nil {
		log.Fatal("Connect to database error ", err)
		return nil, err
	}
	// defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("cannot Ping ", err)
	}
	return &Postgres{Db: db}, nil
}
