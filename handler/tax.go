package handler

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/shopspring/decimal"
)

func checkNumber(data IncomeData) (bool, error) {
	// check negative number
	if data.TotalIncome < 0 {
		return true, errors.New("totalIncome must be positive")
	}
	if data.Wht < 0 {
		return true, errors.New("wht must be positive")
	}
	for _, value := range data.Allowances {
		if value.Amount < 0 {
			return true, errors.New("input number must be positive")
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
		txLv3 = taxDecimal.InexactFloat64()
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
		txLv2 = taxDecimal.InexactFloat64()
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
		txLv1 = taxDecimal.InexactFloat64()
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
	}

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

func getTaxesCsv(file io.Reader) (Taxes, error) {

	// Parse the file
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
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

		if responseRefund.TaxRefund != 0 {
			taxes = TaxesAll{TotalIncome: data.TotalIncome, Tax: tx.Tax, TaxRefund: responseRefund.TaxRefund}

			taxesSlice.Taxes = append(taxesSlice.Taxes, taxes)
		} else {
			taxes = TaxesAll{TotalIncome: data.TotalIncome, Tax: tx.Tax}

			taxesSlice.Taxes = append(taxesSlice.Taxes, taxes)

		}

	}
	return taxesSlice, nil

}
