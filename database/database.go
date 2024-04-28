package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var db *sql.DB

func InitDB() {

	// db connection
	var err error

	// err = godotenv.Load()
	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// }

	psqlInfo := os.Getenv("DATABASE_URL")

	db, err = sql.Open("postgres", psqlInfo)

	if err != nil {
		log.Fatal("Connect to database error ", err)
	}
	// defer db.Close()

	// err = db.Ping()
	// if err != nil {
	// 	log.Fatal("cannot Ping ", err)
	// }
}
