package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func SetPersonaldeductionHandler(c echo.Context) error {

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

func SetKreceiptHandler(c echo.Context) error {

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
