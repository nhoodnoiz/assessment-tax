package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-playground/validator"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
)

type Allowance struct {
	AllowanceType string  `json:"allowanceType"`
	Amount        float64 `json:"amount"`
}

type IncomeData struct {
	TotalIncome float64     `json:"totalIncome" validate:"required"`
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

// Response body CSV - EXP06
// type Taxes struct {
// 	TotalIncome float64 `json:"totalIncome"`
// 	TaxDetails  []Tax
// 	TaxRefunds  []TaxRefund
// }

type Taxes struct {
	Taxes []TaxesAll `json:"taxes"`
}

type TaxesAll struct {
	TotalIncome float64 `json:"totalIncome"`
	Tax         float64 `json:"tax"`
	TaxRefund   float64 `json:"taxRefund"`
}

// type TaxesAll struct {
// 	TotalIncome float64     `json:"totalIncome"`
// 	Tax         []Tax       `json:"tax"`
// 	TaxRefund   []TaxRefund `json:"taxRefund"`
// }

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

// Validate JSON request
type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
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
	e.Validator = &CustomValidator{validator: validator.New()}
	e.GET("/health", healthHandler)
	e.POST("/tax/calculations", getTaxHandler)
	e.POST("/tax/calculations/upload-csv", uploadCsvHandler)

	g := e.Group("/admin")

	adminUsername := os.Getenv("ADMIN_USERNAME")
	adminPassword := os.Getenv("ADMIN_PASSWORD")

	g.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if username == adminUsername && password == adminPassword {
			return true, nil
		}
		return false, nil
	}))

	// e.GET("/health", healthHandler)
	// e.POST("/tax/calculations", getTaxHandler)
	g.POST("/deductions/personal", setPersonaldeductionHandler)
	g.POST("/deductions/k-receipt", setKreceiptHandler)
	// e.POST("/tax/calculations/upload-csv", uploadCsvHandler)

	// Gracefully shutdown

	// Start http server
	port := os.Getenv("PORT")
	// e.Logger.Fatal(e.Start(":" + port))

	go func() {
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	<-shutdown
	fmt.Println("shutting down the server")

	// Wait for interrupt signal or kill signal to gracefully shutdown the server with a timeout of 10 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
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
	// Validate JSON
	if err = c.Validate(&incomeData); err != nil {
		return err
	}

	value, err := checkNumber(incomeData)
	if value {
		return c.JSON(http.StatusBadRequest, Err{Message: "The input numbers must be all positive"})
	}
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: err.Error()})
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
func checkNumber(data IncomeData) (bool, error) {
	// check negative number
	if data.TotalIncome < 0 {
		return true, nil
	}
	if data.Wht < 0 {
		return true, nil
	}
	for _, value := range data.Allowances {
		if value.Amount < 0 {
			return true, nil
		}
	}
	if data.Wht > data.TotalIncome {
		return false, errors.New("(wht:with holding tax) must be less than the totalIncome")
	}
	return false, nil
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

	var taxDecimal decimal.Decimal
	var tax float64

	// tax = data.TotalIncome - personalDeduction - donation - kReceipt
	// totalIncomeDecimal := decimal.NewFromFloat(data.TotalIncome)
	// personalDeductionDecimal := decimal.NewFromFloat(personalDeduction)
	// donationDecimal := decimal.NewFromFloat(donation)
	// kReceiptDecimal := decimal.NewFromFloat(kReceipt)

	// taxDecimal = totalIncomeDecimal.Sub(personalDeductionDecimal).Sub(donationDecimal).Sub(kReceiptDecimal)
	// taxDecimal = taxDecimal.Sub(decimal.NewFromInt(2000000)).Mul(decimal.NewFromFloat(0.35)).Add(decimal.NewFromInt(200000 + 75000 + 35000)).Sub(decimal.NewFromFloat(data.Wht))

	// fmt.Println("taxDecimal =", taxDecimal)

	// tax = taxDecimal.InexactFloat64()

	tax = data.TotalIncome - personalDeduction - donation - kReceipt

	fmt.Printf("totalIncome = %v, personalDeduction = %v, donation = %v, kReceipt = %v\n", data.TotalIncome, personalDeduction, donation, kReceipt)

	txLevel := 0
	var txLv1 float64
	var txLv2 float64
	var txLv3 float64
	var txLv4 float64
	if tax > 2000000 {
		// tax = tax - 2000000
		// tax = 0.35 * tax
		// txLv4 = tax
		// tax = tax + (1000000 * 0.2) + (500000 * 0.15) + (350000 * 0.1)
		// txLevel = 4
		// txLv1 = 35000
		// txLv2 = 75000
		// txLv3 = 200000

		// convert to decimal.NewFromFloat
		// var taxDecimal decimal.Decimal
		taxDecimal = decimal.NewFromFloat(tax).Sub(decimal.NewFromInt(2000000))
		taxDecimal = taxDecimal.Mul(decimal.NewFromFloat(0.35))
		txLv4 = taxDecimal.InexactFloat64()
		taxDecimal = taxDecimal.Add(decimal.NewFromInt(200000 + 75000 + 35000))
		txLevel = 4
		txLv1 = 35000
		txLv2 = 75000
		txLv3 = 200000
		// tax = taxDecimal.InexactFloat64()
		// fmt.Println("From taxDecimal to tax float64 =", tax)

	} else if tax > 1000000 && tax <= 2000000 {
		// tax = tax - 1000000
		// tax = 0.2 * tax
		// txLv3 = tax
		// tax = tax + (500000 * 0.15) + (350000 * 0.1)
		// txLevel = 3
		// txLv1 = 35000
		// txLv2 = 75000

		// convert to decimal.NewFromFloat
		// var taxDecimal decimal.Decimal
		taxDecimal = decimal.NewFromFloat(tax).Sub(decimal.NewFromInt(1000000))
		taxDecimal = taxDecimal.Mul(decimal.NewFromFloat(0.2))
		txLv4 = taxDecimal.InexactFloat64()
		taxDecimal = taxDecimal.Add(decimal.NewFromInt(75000 + 35000))
		txLevel = 3
		txLv1 = 35000
		txLv2 = 75000

	} else if tax > 500000 && tax <= 1000000 {
		// tax = tax - 500000
		// tax = 0.15 * tax
		// txLv2 = tax
		// tax = tax + (350000 * 0.1)
		// txLevel = 2
		// txLv1 = 35000

		// convert to decimal.NewFromFloat
		// var taxDecimal decimal.Decimal
		taxDecimal = decimal.NewFromFloat(tax).Sub(decimal.NewFromInt(500000))
		taxDecimal = taxDecimal.Mul(decimal.NewFromFloat(0.15))
		txLv4 = taxDecimal.InexactFloat64()
		taxDecimal = taxDecimal.Add(decimal.NewFromInt(35000))
		txLevel = 2
		txLv1 = 35000

	} else if tax > 150000 && tax <= 500000 {
		// tax = tax - 150000
		// tax = 0.1 * tax
		// txLv1 = tax
		// txLevel = 1

		// convert to decimal.NewFromFloat
		// var taxDecimal decimal.Decimal
		taxDecimal = decimal.NewFromFloat(tax).Sub(decimal.NewFromInt(150000))
		taxDecimal = taxDecimal.Mul(decimal.NewFromFloat(0.1))
		txLv4 = taxDecimal.InexactFloat64()
		txLevel = 1

	} else {
		tax = 0
		txLevel = 0
	}
	// tax = tax - data.Wht
	taxDecimal = taxDecimal.Sub(decimal.NewFromFloat(data.Wht))
	tax = taxDecimal.InexactFloat64()

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
					Tax:   txLv1,
				},
				{
					Level: "500,001-1,000,000",
					Tax:   txLv2,
				},
				{
					Level: "1,000,001-2,000,000",
					Tax:   txLv3,
				},
				{
					Level: "2,000,001 ขึ้นไป",
					Tax:   txLv4,
				},
			},
		}

		// response.TaxLevel[txLevel].Tax = tax
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

func uploadCsvHandler(c echo.Context) error {
	file, err := c.FormFile("taxesFile")
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: err.Error()})
	}
	if !strings.HasSuffix(file.Filename, ".csv") {
		return c.JSON(http.StatusBadRequest, "invalid file type, must be .csv")
	}
	csvFile, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: err.Error()})
	}
	defer csvFile.Close()

	taxesSlice, err := getTaxesCsv(csvFile)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: err.Error()})
	}

	return c.JSON(http.StatusCreated, taxesSlice)
}

func getTaxesCsv(file io.Reader) (Taxes, error) {
	// file, err := os.Open("taxes.csv")
	// if err != nil {
	// 	fmt.Println("Error opening file:", err)
	// 	// return
	// }
	// defer file.Close()

	// Parse the file
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		// fmt.Println("Error reading CSV:", err)
		return Taxes{}, err
	}

	// skip the header row
	header := records[:1]
	records = records[1:]

	// Slice to hold parsed records
	var incomeDataSlice []IncomeData

	// Iterate over each record
	for _, row := range records {
		// Convert string fields to float64
		totalIncome, err := strconv.ParseFloat(row[0], 64)
		if err != nil {
			fmt.Println("Error converting totalIncome:", err)
		}
		wht, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			fmt.Println("Error converting wht:", err)
		}
		donation, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			fmt.Println("Error converting donation:", err)
		}

		// create a new record and append to parsedRecords
		// taxes := TaxesCSV{
		// 	TotalIncome: totalIncome,
		// 	Wht:         wht,
		// 	Donation:    donation,
		// }
		income := IncomeData{
			TotalIncome: totalIncome,
			Wht:         wht,
		}
		income.Allowances = append(income.Allowances, Allowance{
			AllowanceType: header[0][2],
			Amount:        donation,
		})

		// Append to the slice
		incomeDataSlice = append(incomeDataSlice, income)
	}

	// var taxes Taxes
	var taxesSlice Taxes
	for _, data := range incomeDataSlice {

		var taxes TaxesAll

		value, err := checkNumber(data)
		if value {
			return Taxes{}, errors.New("the input numbers in .csv file must be all positive")
		}
		if err != nil {
			return Taxes{}, err
		}

		donation, kReceipt = checkValue(data)

		// implement story: EXP04 and EXP07; provide the tax level detail
		tx, _, responseRefund := calculateTax(data, donation, kReceipt)

		// if responseRefund.TaxRefund != 0 {
		// 	taxes.TotalIncome = data.TotalIncome
		// 	taxes.TaxRefunds = append(taxes.TaxRefunds, responseRefund)
		// 	taxesSlice = append(taxesSlice, taxes)

		// } else {
		// 	taxes.TotalIncome = data.TotalIncome
		// 	taxes.TaxDetails = append(taxes.TaxDetails, tx)
		// 	taxesSlice = append(taxesSlice, taxes)
		// }

		if responseRefund.TaxRefund != 0 {
			taxes = TaxesAll{TotalIncome: data.TotalIncome, Tax: tx.Tax, TaxRefund: responseRefund.TaxRefund}

			taxesSlice.Taxes = append(taxesSlice.Taxes, taxes)
		} else {
			taxes = TaxesAll{TotalIncome: data.TotalIncome, Tax: tx.Tax}

			taxesSlice.Taxes = append(taxesSlice.Taxes, taxes)

		}

		// if responseRefund.TaxRefund != 0 {
		// 	taxes = TaxesAll{TotalIncome: data.TotalIncome, TaxRefund: []TaxRefund{responseRefund}}

		// 	taxesSlice.Taxes = append(taxesSlice.Taxes, taxes)
		// } else {
		// 	taxes = TaxesAll{TotalIncome: data.TotalIncome, Tax: []Tax{tx}}

		// 	taxesSlice.Taxes = append(taxesSlice.Taxes, taxes)

		// }

	}
	return taxesSlice, nil

}
