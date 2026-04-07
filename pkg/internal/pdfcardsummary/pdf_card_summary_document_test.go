package pdfcardsummary

import (
	"fmt"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestToCSVMatrix(t *testing.T) {
	nov1 := time.Date(2023, 11, 1, 0, 0, 0, 0, time.UTC)
	nov15 := time.Date(2023, 11, 15, 0, 0, 0, 0, time.UTC)
	oct25 := time.Date(2023, 10, 25, 0, 0, 0, 0, time.UTC)

	cardNumber := "4111-1111-1111-1111"
	receipt123 := "RC-123"
	installment1 := 1
	installment3 := 3

	tests := []struct {
		name          string
		summary       CardSummary
		expectedRows  [][]string
		expectedError error
	}{
		{
			name: "full_data",
			summary: CardSummary{
				StatementContext: StatementContext{
					Bank:           Bank("Santander"),
					CardCompany:    CardCompany("VISA"),
					TotalARS:       decimal.NewFromFloat(1000.00),
					TotalUSD:       decimal.NewFromFloat(50.00),
					CloseDate:      nov1,
					ExpirationDate: nov15,
				},
				Table: Table{
					PastPaymentMovements: []Movement{
						{
							OriginalDate:       &oct25,
							ReceiptNumber:      &receipt123,
							Detail:             "Past Payment",
							CurrentInstallment: &installment1,
							TotalInstallments:  &installment3,
							AmountARS:          decimal.NewFromFloat(200.00),
							AmountUSD:          decimal.NewFromFloat(10.00),
						},
					},
					TaxesMovements: []Movement{
						{
							Detail:    "Tax Charge",
							AmountARS: decimal.NewFromFloat(100.00),
							AmountUSD: decimal.NewFromFloat(5.00),
						},
					},
					Cards: []Card{
						{
							CardContext: CardContext{
								CardNumber:   &cardNumber,
								CardOwner:    "John Doe",
								CardTotalARS: decimal.NewFromFloat(500.00),
								CardTotalUSD: decimal.NewFromFloat(25.00),
							},
							Movements: []Movement{
								{
									OriginalDate: &nov1,
									Detail:       "Purchase",
									AmountARS:    decimal.NewFromFloat(500.00),
									AmountUSD:    decimal.NewFromFloat(25.00),
								},
							},
						},
					},
				},
			},
			expectedRows: [][]string{
				{"Bank", "CardCompany", "CloseDate", "ExpirationDate", "TotalARS", "TotalUSD", "CardNumber", "CardOwner", "CardTotalARS", "CardTotalUSD", "MovementType", "OriginalDate", "ReceiptNumber", "Detail", "CurrentInstallment", "TotalInstallments", "AmountARS", "AmountUSD"},
				// Past Payment
				{"Santander", "VISA", "2023-11-01", "2023-11-15", "1.000,00", "50,00", "", "", "", "", "PastPayment", "2023-10-25", "RC-123", "Past Payment", "1", "3", "200,00", "10,00"},
				// Tax
				{"Santander", "VISA", "2023-11-01", "2023-11-15", "1.000,00", "50,00", "", "", "", "", "Tax", "", "", "Tax Charge", "", "", "100,00", "5,00"},
				// Card Movement
				{"Santander", "VISA", "2023-11-01", "2023-11-15", "1.000,00", "50,00", "4111-1111-1111-1111", "John Doe", "500,00", "25,00", "CardMovement", "2023-11-01", "", "Purchase", "", "", "500,00", "25,00"},
			},
			expectedError: nil,
		},
		{
			name: "empty_movements",
			summary: CardSummary{
				StatementContext: StatementContext{
					Bank:           Bank("Santander"),
					CardCompany:    CardCompany("VISA"),
					CloseDate:      nov1,
					ExpirationDate: nov15,
				},
			},
			expectedRows: nil,
			expectedError: fmt.Errorf(
				"error building CSV: %w",
				fmt.Errorf(
					"error validating CSVBuilder: %w",
					ErrNoRowsToBuild,
				),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			actualRows, actualError := tt.summary.ToCSVMatrix()

			// Then
			require.Equal(t, tt.expectedRows, actualRows)
			require.Equal(t, tt.expectedError, actualError)

		})
	}
}
