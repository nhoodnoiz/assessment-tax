package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
)

type Allowance struct {
	AllowanceType string  `json:"allowanceType"`
	Amount        float64 `json:"amount"`
}

type IncomeData struct {
	TotalIncome float64     `json:"totalIncome"`
	Wht         float64     `json:"wht"`
	Allowances  []Allowance `json:"allowances"`
}

type Tax struct {
	Tax float64 `json:"tax"`
}

type TaxRefund struct {
	Tax       float64 `json:"tax"`
	TaxRefund float64 `json:"taxRefund"`
}

// Tax Level declaring

type TaxLevel struct {
	Level string  `json:"level"`
	Tax   float64 `json:"tax"`
}

type TaxData struct {
	Tax      float64    `json:"tax"`
	TaxLevel []TaxLevel `json:"taxLevel"`
}

type Amount struct {
	Amount float64 `json:"amount"`
}

type PersonalDeduction struct {
	PersonalDeduction float64 `json:"personalDeduction"`
}

type KReceipt struct {
	KReceipt float64 `json:"kReceipt"`
}

// starting value for personalDeduction = 60000
var personalDeduction float64 = 60000

var donation float64
var kReceipt float64

// starting value for kReceiptRange = 50000
var kReceiptRange float64 = 50000

var db *sql.DB

type Err struct {
	Message string `json:"message"`
}

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
	e.POST("/tax/calculations", getTaxHandler)
	e.POST("/admin/deductions/personal", setPersonaldeductionHandler)
	e.POST("/admin/deductions/k-receipt", setKreceiptHandler)

	// Start http server
	port := os.Getenv("PORT")
	e.Logger.Fatal(e.Start(":" + port))
}

func healthHandler(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, Go Bootcamp!")
}

func getTaxHandler(c echo.Context) error {

	var incomeData IncomeData
	// tax := new(Tax)
	err := c.Bind(&incomeData)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: err.Error()})
	}
	if checkNegativeNumber(incomeData) {
		return c.JSON(http.StatusBadRequest, "The input must be positive number")
	}
	donation, kReceipt = checkValue(incomeData)

	// implement story: EXP04 and EXP07; provide the tax level detail
	_, response, responseRefund := calculateTax(incomeData, donation, kReceipt)

	// fmt.Println("(From calculateTax)tx =", tx)

	// fmt.Println("tax =", tax)
	// fmt.Println("taxRefund =", taxRefund)
	// if err != nil {
	// 	return c.JSON(http.StatusBadRequest, Err{Message: err})
	// }

	// implement story: EXP01,EXP02,EXP03; not provide the tax level detail

	// if responseRefund.TaxRefund != 0 {
	// 	return c.JSON(http.StatusCreated, responseRefund)
	// } else {
	// 	return c.JSON(http.StatusCreated, tx)
	// }

	// implement story: EXP04,EXP07; provide the tax level detail

	if responseRefund.TaxRefund != 0 {
		return c.JSON(http.StatusCreated, responseRefund)
	} else {
		return c.JSON(http.StatusCreated, response)
	}
}
func checkNegativeNumber(data IncomeData) bool {
	// check negative number
	if data.TotalIncome < 0 {
		return true
	}
	if data.Wht < 0 {
		return true
	}
	for _, value := range data.Allowances {
		if value.Amount < 0 {
			return true
		}
	}
	return false
}

func checkValue(data IncomeData) (donation, kReceipt float64) {

	for _, value := range data.Allowances {

		// check donation value
		// starting value; donation = 100,000
		// starting value; kReceipt = 50,000
		if value.AllowanceType == "donation" {
			if value.Amount <= 100000 {
				donation = value.Amount
			} else if value.Amount > 100000 {
				donation = 100000
			}
		} else if value.AllowanceType == "k-receipt" {

			if value.Amount <= kReceiptRange && value.Amount > 0 {
				kReceipt = value.Amount
			} else if value.Amount > kReceiptRange {
				kReceipt = kReceiptRange
			}
		}
	}
	return donation, kReceipt

}

func calculateTax(data IncomeData, donation, kReceipt float64) (tx Tax, response TaxData, responseRefund TaxRefund) {

	// for _, t := range data.Allowances {
	// 	if t.AllowanceType == "donation" {
	// 		if t.Amount <= 100000 {
	// 			donation = t.Amount
	// 		} else if t.Amount > 100000 {
	// 			donation = 100000
	// 		}
	// 	}
	// 	if t.AllowanceType == "k-receipt" {

	// 		if t.Amount > 0 && t.Amount <= 50000 {
	// 			kReceipt = t.Amount
	// 		}
	// 	}
	// }

	var tax float64

	tax = data.TotalIncome - personalDeduction - donation - kReceipt

	fmt.Printf("totalIncome = %v, personalDeduction = %v, donation = %v, kReceipt = %v\n", data.TotalIncome, personalDeduction, donation, kReceipt)

	txLevel := 0
	if tax > 2000000 {
		tax = tax - 2000000
		// fmt.Println("tax after =", tax)
		tax = 0.35 * tax
		// fmt.Println("tax afer multiply =", tax)
		tax = tax + (1000000 * 0.2) + (500000 * 0.15) + (350000 * 0.1)
		// fmt.Println("tax after include =", tax)
		txLevel = 4
	} else if tax > 1000000 && tax <= 2000000 {
		tax = tax - 1000000
		tax = 0.2 * tax
		tax = tax + (500000 * 0.15) + (350000 * 0.1)
		txLevel = 3
	} else if tax > 500000 && tax <= 1000000 {
		tax = tax - 500000
		tax = 0.15 * tax
		tax = tax + (350000 * 0.1)
		txLevel = 2
	} else if tax > 150000 && tax <= 500000 {
		tax = tax - 150000
		tax = 0.1 * tax
		txLevel = 1
	} else {
		tax = 0
		txLevel = 0
	}
	tax = tax - data.Wht

	// in case when there's tax refund
	var taxRefund float64

	if tax < 0 {
		taxRefund = (-1) * tax
		tax = 0
		fmt.Println("taxRefund =", taxRefund)
	} else {
		tx = Tax{Tax: tax}
		// fmt.Println("tx =", tx)
	}
	// fmt.Println("(outside if-else)tx =", tx)

	fmt.Println("tax =", tax)
	fmt.Println("wht =", data.Wht)
	fmt.Println("tax level =", txLevel)
	fmt.Println("<<<<<End of Request>>>>>")

	if taxRefund != 0 {

		responseRefund = TaxRefund{
			Tax:       tax,
			TaxRefund: taxRefund,
		}
	} else {
		// with tax level detail
		response = TaxData{
			Tax: tax,
			TaxLevel: []TaxLevel{

				{
					Level: "0-150,000",
					Tax:   0.0,
				},
				{
					Level: "150,001-500,000",
					Tax:   0.0,
				},
				{
					Level: "500,001-1,000,000",
					Tax:   0.0,
				},
				{
					Level: "1,000,001-2,000,000",
					Tax:   0.0,
				},
				{
					Level: "2,000,001 ขึ้นไป",
					Tax:   0.0,
				},
			},
		}

		response.TaxLevel[txLevel].Tax = tax
	}

	return tx, response, responseRefund

}

// func calculateTax(data IncomeData) []float64 {
// 	var results []float64
// 	for _, allowance := range data.Allowances {
// 		results = append(results, allowance.Amount)
// 	}
// 	return results

// }

func setPersonaldeductionHandler(c echo.Context) error {

	var amount Amount
	err := c.Bind(&amount)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: err.Error()})
	}

	if amount.Amount > 10000 && amount.Amount <= 100000 {

		personalDeduction = amount.Amount

		var personal PersonalDeduction
		personal.PersonalDeduction = personalDeduction

		return c.JSON(http.StatusCreated, personal)

	} else {
		return c.JSON(http.StatusBadRequest, "Personal deduction must be in range between 10,000 - 100,000")
	}

}

func setKreceiptHandler(c echo.Context) error {

	var amount Amount
	err := c.Bind(&amount)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: err.Error()})
	}

	if amount.Amount > 0 && amount.Amount <= 100000 {

		kReceiptRange = amount.Amount

		var receipt KReceipt
		receipt.KReceipt = kReceiptRange

		return c.JSON(http.StatusCreated, receipt)
	} else {
		return c.JSON(http.StatusBadRequest, "k-receipt must be in range between 0 - 100,000")
	}

}
