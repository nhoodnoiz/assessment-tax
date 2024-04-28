package handler

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
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

type Taxes struct {
	Taxes []TaxesAll `json:"taxes"`
}

type TaxesAll struct {
	TotalIncome float64 `json:"totalIncome"`
	Tax         float64 `json:"tax"`
	TaxRefund   float64 `json:"taxRefund"`
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

// var db *sql.DB

type Err struct {
	Message string `json:"message"`
}

// Validate JSON request
// type CustomValidator struct {
// 	validator *validator.Validate
// }

// func (cv *CustomValidator) Validate(i interface{}) error {
// 	if err := cv.validator.Struct(i); err != nil {
// 		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	return nil
// }

func HealthHandler(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, Go Bootcamp!")
}

func GetTaxHandler(c echo.Context) error {

	var incomeData IncomeData
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
		return c.JSON(http.StatusBadRequest, Err{Message: err.Error()})
	}
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: err.Error()})
	}

	donation, kReceipt = checkValue(incomeData)

	// implement story: EXP04 and EXP07; provide the tax level detail
	_, response, responseRefund := calculateTax(incomeData, donation, kReceipt)

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

func UploadCsvHandler(c echo.Context) error {
	file, err := c.FormFile("taxFile")
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
