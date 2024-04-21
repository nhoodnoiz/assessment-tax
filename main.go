package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
)

// type Taxes struct{
// 	totalIncome float64 `json:"totalIncome"`
// 	wht float64 `json:"wht"`
// 	allowances []
// }

var db *sql.DB

func main() {

	// db connection
	var err error

	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	psqlInfo := os.Getenv("DATABASE_URL")

	// psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable", host, port, username, password, databaseName)

	db, err = sql.Open("postgres", psqlInfo)

	if err != nil {
		log.Fatal("Connect to database error ", err)
	}
	defer db.Close()

	// Run only first time
	// createTb := `
	// CREATE TABLE IF NOT EXISTS taxes (
	// 	id SERIAL PRIMARY KEY, title TEXT,
	// 	amount FLOAT,
	// 	note TEXT,
	// 	tags TEXT[]);
	// `
	// createTb := `
	// CREATE TABLE IF NOT EXISTS taxes(
	// 	totalIncome FLOAT,
	// 	wht FLOAT,

	// )
	// `

	// _, err = db.Exec(createTb)

	// if err != nil {
	// 	log.Fatal("can't create table", err)
	// }
	// fmt.Println("create table success")

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	e := echo.New()
	// e.GET("/health", healthHandler)

	adminUsername := os.Getenv("ADMIN_USERNAME")
	adminPassword := os.Getenv("ADMIN_PASSWORD")

	e.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if username == adminUsername && password == adminPassword {
			return true, nil
		}
		return false, nil
	}))

	e.GET("/health", healthHandler)
	// e.GET("/users", getUsersHandler)

	// Start http server
	port := os.Getenv("PORT")
	e.Logger.Fatal(e.Start(":" + port))
}

func healthHandler(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, Go Bootcamp!")
}

// func getUsersHandler(c echo.Context) error {

// }
