package main

import (
	"errors"
	"reflect"
	"testing"
)

func TestCheckNumber(t *testing.T) {

	testCases := []struct {
		name        string
		totalIncome float64
		wht         float64
		donation    float64
		kReceipt    float64
		want        bool
	}{
		{"Story EXP01", 500000, 0, 0, 0, false},
		{"Story EXP02", 500000, 25000, 0, 0, false},
		{"Story EXP03", 500000, 0, 200000, 0, false},
		{"Story EXP07", 500000, 0, 100000, 200000, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, _ := checkNumber(IncomeData{
				TotalIncome: tc.totalIncome,
				Wht:         tc.wht,
				Allowances: []Allowance{
					{AllowanceType: "donation", Amount: tc.donation},
					{AllowanceType: "k-receipt", Amount: tc.kReceipt},
				},
			})
			if result != tc.want {
				t.Errorf("All input numbers are negative, %t; want %t", result, tc.want)
			}

		})
	}

}

func TestCheckNumberErorr(t *testing.T) {

	testCases := []struct {
		name        string
		totalIncome float64
		wht         float64
		donation    float64
		kReceipt    float64
		want        bool
		err         error
	}{
		{"1-Negative Number", -500000, 0, 100000, 200000, true, errors.New("totalIncome must be positive")},
		{"2-Negative Number", 500000, -3000, 100000, 200000, true, errors.New("wht must be positive")},
		{"3-Negative Number", 500000, 3000, -100000, 200000, true, errors.New("input number must be positive")},
		{"4-Negative Number", 500000, 3000, 100000, -200000, true, errors.New("input number must be positive")},
		{"5-Negative Number", 500000, -3000, -100000, 200000, true, errors.New("wht must be positive")},
		{"1-wht > total income", 5000, 600000, 0, 0, false, errors.New("(wht:with holding tax) must be less than the totalIncome")},
		{"2-wht > total income", 500000, 800000, 100000, 2000000, false, errors.New("(wht:with holding tax) must be less than the totalIncome")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := checkNumber(IncomeData{
				TotalIncome: tc.totalIncome,
				Wht:         tc.wht,
				Allowances: []Allowance{
					{AllowanceType: "donation", Amount: tc.donation},
					{AllowanceType: "k-receipt", Amount: tc.kReceipt},
				},
			})
			if result != tc.want {
				t.Errorf("Result = %t; want = %t", result, tc.want)
			}
			if err.Error() != tc.err.Error() {
				t.Errorf("Result = %v; want = %v", err, tc.err)
			}

		})
	}

}

func TestCheckValueDefault(t *testing.T) {

	testCases := []struct {
		name         string
		totalIncome  float64
		wht          float64
		donation     float64
		kReceipt     float64
		wantDonation float64
		wantKReceipt float64
	}{
		{"Story EXP01", 500000, 0, 0, 0, 0, 0},
		{"Story EXP02", 500000, 25000, 0, 0, 0, 0},
		{"Story EXP03", 500000, 0, 200000, 0, 100000, 0},
		{"Story EXP07", 500000, 0, 100000, 200000, 100000, 50000},
		{"1-Cases", 500000, 0, 200000, 30000, 100000, 30000},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resultDonation, resultKReceipt := checkValue(IncomeData{
				TotalIncome: tc.totalIncome,
				Wht:         tc.wht,
				Allowances: []Allowance{
					{AllowanceType: "donation", Amount: tc.donation},
					{AllowanceType: "k-receipt", Amount: tc.kReceipt},
				},
			})
			if resultDonation != tc.wantDonation && resultKReceipt != tc.wantKReceipt {
				t.Errorf("donation = %v; want %v, k-receipt = %v; want %v", resultDonation, tc.wantDonation, resultKReceipt, tc.wantKReceipt)
			}

		})
	}

}

func TestCheckValueVarykReceiptRange(t *testing.T) {

	testCases := []struct {
		name          string
		totalIncome   float64
		wht           float64
		donation      float64
		kReceipt      float64
		wantDonation  float64
		wantKReceipt  float64
		kReceiptRange float64
	}{
		{"Story EXP01", 500000, 0, 0, 0, 0, 0, 70000},
		{"Story EXP02", 500000, 25000, 0, 0, 0, 0, 70000},
		{"Story EXP03", 500000, 0, 200000, 0, 100000, 0, 70000},
		{"Story EXP07", 500000, 0, 100000, 200000, 100000, 70000, 70000},
		{"1-Cases", 500000, 0, 200000, 30000, 100000, 30000, 80000},
		{"2-Cases", 500000, 0, 100000, 200000, 100000, 100000, 100000},
		{"3-Cases", 500000, 0, 100000, 200000, 100000, 10000, 10000},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resultDonation, resultKReceipt := checkValue(IncomeData{
				TotalIncome: tc.totalIncome,
				Wht:         tc.wht,
				Allowances: []Allowance{
					{AllowanceType: "donation", Amount: tc.donation},
					{AllowanceType: "k-receipt", Amount: tc.kReceipt},
				},
			})
			if resultDonation != tc.wantDonation && resultKReceipt != tc.wantKReceipt {
				t.Errorf("donation = %v; want %v, k-receipt = %v; want %v", resultDonation, tc.wantDonation, resultKReceipt, tc.wantKReceipt)
			}

		})
	}

}

func TestCalculateTaxTx(t *testing.T) {
	testCases := []struct {
		name        string
		totalIncome float64
		wht         float64
		donation    float64
		kReceipt    float64
		wantTax     Tax
	}{
		{"Story EXP01", 500000, 0, 0, 0, Tax{Tax: 29000}},
		{"Story EXP02", 500000, 25000, 0, 0, Tax{Tax: 4000}},
		{"Story EXP03", 500000, 0, 200000, 0, Tax{Tax: 19000}},
		{"Story EXP07", 500000, 0, 100000, 200000, Tax{Tax: 14000}},
		{"1-Cases", 2160001, 200000.35, 200000, 0, Tax{Tax: 110000}},
		{"2-Cases", 660000, 0, 100000, 200000, Tax{Tax: 30000}},
		{"3-Cases", 600000, 36000, 20000, 0, Tax{Tax: 2000}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resultDonation, resultKReceipt := checkValue(IncomeData{
				TotalIncome: tc.totalIncome,
				Wht:         tc.wht,
				Allowances: []Allowance{
					{AllowanceType: "donation", Amount: tc.donation},
					{AllowanceType: "k-receipt", Amount: tc.kReceipt},
				},
			})
			resultTx, _, _ := calculateTax(IncomeData{
				TotalIncome: tc.totalIncome,
				Wht:         tc.wht,
				Allowances: []Allowance{
					{AllowanceType: "donation", Amount: tc.donation},
					{AllowanceType: "k-receipt", Amount: tc.kReceipt},
				},
			}, resultDonation, resultKReceipt)

			if resultTx != tc.wantTax {
				t.Errorf("resultTx = %v; want = %v", resultTx, tc.wantTax)
			}

		})
	}

}

func TestCalculateTaxResponse(t *testing.T) {
	testCases := []struct {
		name        string
		totalIncome float64
		wht         float64
		donation    float64
		kReceipt    float64
		wantTax     TaxData
	}{
		{"Story EXP01", 500000, 0, 0, 0, TaxData{Tax: 29000, TaxLevel: []TaxLevel{

			{
				Level: "0-150,000",
				Tax:   0.0,
			},
			{
				Level: "150,001-500,000",
				Tax:   29000,
			},
			{
				Level: "500,001-1,000,000",
				Tax:   0,
			},
			{
				Level: "1,000,001-2,000,000",
				Tax:   0,
			},
			{
				Level: "2,000,001 ขึ้นไป",
				Tax:   0,
			},
		}}},
		{"Story EXP02", 500000, 25000, 0, 0, TaxData{Tax: 4000, TaxLevel: []TaxLevel{

			{
				Level: "0-150,000",
				Tax:   0.0,
			},
			{
				Level: "150,001-500,000",
				Tax:   29000,
			},
			{
				Level: "500,001-1,000,000",
				Tax:   0,
			},
			{
				Level: "1,000,001-2,000,000",
				Tax:   0,
			},
			{
				Level: "2,000,001 ขึ้นไป",
				Tax:   0,
			},
		}}},
		{"Story EXP03", 500000, 0, 200000, 0, TaxData{Tax: 19000, TaxLevel: []TaxLevel{

			{
				Level: "0-150,000",
				Tax:   0.0,
			},
			{
				Level: "150,001-500,000",
				Tax:   19000,
			},
			{
				Level: "500,001-1,000,000",
				Tax:   0,
			},
			{
				Level: "1,000,001-2,000,000",
				Tax:   0,
			},
			{
				Level: "2,000,001 ขึ้นไป",
				Tax:   0,
			},
		}}},
		{"Story EXP07", 500000, 0, 100000, 200000, TaxData{Tax: 14000, TaxLevel: []TaxLevel{

			{
				Level: "0-150,000",
				Tax:   0.0,
			},
			{
				Level: "150,001-500,000",
				Tax:   14000,
			},
			{
				Level: "500,001-1,000,000",
				Tax:   0,
			},
			{
				Level: "1,000,001-2,000,000",
				Tax:   0,
			},
			{
				Level: "2,000,001 ขึ้นไป",
				Tax:   0,
			},
		}}},
		{"1-Cases", 2160001, 200000.35, 200000, 0, TaxData{Tax: 110000, TaxLevel: []TaxLevel{

			{
				Level: "0-150,000",
				Tax:   0.0,
			},
			{
				Level: "150,001-500,000",
				Tax:   35000,
			},
			{
				Level: "500,001-1,000,000",
				Tax:   75000,
			},
			{
				Level: "1,000,001-2,000,000",
				Tax:   200000,
			},
			{
				Level: "2,000,001 ขึ้นไป",
				Tax:   0.35,
			},
		}}},
		{"2-Cases", 660000, 0, 100000, 200000, TaxData{Tax: 30000, TaxLevel: []TaxLevel{

			{
				Level: "0-150,000",
				Tax:   0.0,
			},
			{
				Level: "150,001-500,000",
				Tax:   30000,
			},
			{
				Level: "500,001-1,000,000",
				Tax:   0,
			},
			{
				Level: "1,000,001-2,000,000",
				Tax:   0,
			},
			{
				Level: "2,000,001 ขึ้นไป",
				Tax:   0,
			},
		}}},
		{"3-Cases", 600000, 36000, 20000, 0, TaxData{Tax: 2000, TaxLevel: []TaxLevel{

			{
				Level: "0-150,000",
				Tax:   0.0,
			},
			{
				Level: "150,001-500,000",
				Tax:   35000,
			},
			{
				Level: "500,001-1,000,000",
				Tax:   3000,
			},
			{
				Level: "1,000,001-2,000,000",
				Tax:   0,
			},
			{
				Level: "2,000,001 ขึ้นไป",
				Tax:   0,
			},
		}}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resultDonation, resultKReceipt := checkValue(IncomeData{
				TotalIncome: tc.totalIncome,
				Wht:         tc.wht,
				Allowances: []Allowance{
					{AllowanceType: "donation", Amount: tc.donation},
					{AllowanceType: "k-receipt", Amount: tc.kReceipt},
				},
			})
			_, resultResponse, _ := calculateTax(IncomeData{
				TotalIncome: tc.totalIncome,
				Wht:         tc.wht,
				Allowances: []Allowance{
					{AllowanceType: "donation", Amount: tc.donation},
					{AllowanceType: "k-receipt", Amount: tc.kReceipt},
				},
			}, resultDonation, resultKReceipt)

			if !reflect.DeepEqual(resultResponse, tc.wantTax) {
				t.Errorf("resultResponse = %v; want = %v", resultResponse, tc.wantTax)
			}

		})
	}

}

func TestCalculateTaxResponseRefund(t *testing.T) {
	testCases := []struct {
		name          string
		totalIncome   float64
		wht           float64
		donation      float64
		kReceipt      float64
		wantTaxRefund TaxRefund
	}{
		{"Story EXP01", 500000, 30000, 0, 0, TaxRefund{Tax: 0, TaxRefund: 1000}},
		{"Story EXP02", 500000, 31000, 0, 0, TaxRefund{Tax: 0, TaxRefund: 2000}},
		{"Story EXP03", 500000, 22000, 200000, 0, TaxRefund{Tax: 0, TaxRefund: 3000}},
		{"Story EXP07", 500000, 19000, 100000, 200000, TaxRefund{Tax: 0, TaxRefund: 5000}},
		{"1-Cases", 2160001, 310001.35, 200000, 0, TaxRefund{Tax: 0, TaxRefund: 1}},
		{"2-Cases", 660000, 30500, 100000, 200000, TaxRefund{Tax: 0, TaxRefund: 500}},
		{"3-Cases", 600000, 40000, 20000, 0, TaxRefund{Tax: 0, TaxRefund: 2000}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resultDonation, resultKReceipt := checkValue(IncomeData{
				TotalIncome: tc.totalIncome,
				Wht:         tc.wht,
				Allowances: []Allowance{
					{AllowanceType: "donation", Amount: tc.donation},
					{AllowanceType: "k-receipt", Amount: tc.kReceipt},
				},
			})
			_, _, resultResponseRefund := calculateTax(IncomeData{
				TotalIncome: tc.totalIncome,
				Wht:         tc.wht,
				Allowances: []Allowance{
					{AllowanceType: "donation", Amount: tc.donation},
					{AllowanceType: "k-receipt", Amount: tc.kReceipt},
				},
			}, resultDonation, resultKReceipt)

			if !reflect.DeepEqual(resultResponseRefund, tc.wantTaxRefund) {
				t.Errorf("resultResponse = %v; want = %v", resultResponseRefund, tc.wantTaxRefund)
			}

		})
	}

}

// func TestSetPersonaldeduction(t *testing.T) {

// 	t.Run("Change personalDeduction", func(t *testing.T) {

// 		userJSON := `{"amount": 700000}`

// 		e := echo.New()

// 		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(userJSON))
// 		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 		rec := httptest.NewRecorder()

// 		c := e.NewContext(req, rec)
// 		h := &Amount{Amount: 70000}

// 		err := setPersonaldeductionHandler(c)

// 		if err == nil {
// 			t.Errorf("Result %v; want %v", err, nil)
// 		}
// 	})

// }
