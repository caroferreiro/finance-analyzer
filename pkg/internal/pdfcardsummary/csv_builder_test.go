package pdfcardsummary

import (
	"errors"
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestNewCSVBuilder(t *testing.T) {
	type args struct {
		statement StatementContext
	}
	tests := []struct {
		name            string
		args            args
		expectedBuilder CSVBuilder
		expectedError   error
	}{
		{
			name: "all required fields are non zero",
			args: args{
				statement: StatementContext{
					Bank:           Bank("Santander"),
					CardCompany:    CardCompany("VISA"),
					CloseDate:      time.Date(2023, 11, 1, 0, 0, 0, 0, time.UTC),
					ExpirationDate: time.Date(2023, 11, 15, 0, 0, 0, 0, time.UTC),
					TotalARS:       decimal.NewFromFloat(1000.00),
					TotalUSD:       decimal.NewFromFloat(50.00),
				},
			},
			expectedBuilder: CSVBuilder{
				common: map[string]string{
					"Bank":           "Santander",
					"CardCompany":    "VISA",
					"TotalARS":       "1.000,00",
					"TotalUSD":       "50,00",
					"CloseDate":      "2023-11-01",
					"ExpirationDate": "2023-11-15",
				},
				rows: []map[string]string{},
			},
		},
		{
			name: "all zero amounts but non-zero dates",
			args: args{
				statement: StatementContext{
					Bank:           Bank("Galicia"),
					CardCompany:    CardCompany("AMEX"),
					CloseDate:      time.Date(2023, 11, 1, 0, 0, 0, 0, time.UTC),
					ExpirationDate: time.Date(2023, 11, 15, 0, 0, 0, 0, time.UTC),
					TotalARS:       decimal.Decimal{},
					TotalUSD:       decimal.Decimal{},
				},
			},
			expectedBuilder: CSVBuilder{
				common: map[string]string{
					"Bank":           "Galicia",
					"CardCompany":    "AMEX",
					"TotalARS":       "0,00",
					"TotalUSD":       "0,00",
					"CloseDate":      "2023-11-01",
					"ExpirationDate": "2023-11-15",
				},
				rows: []map[string]string{},
			},
		},
		{
			name: "missing close date",
			args: args{
				statement: StatementContext{
					Bank:           Bank("Santander"),
					CardCompany:    CardCompany("VISA"),
					CloseDate:      time.Time{},
					ExpirationDate: time.Date(2023, 11, 15, 0, 0, 0, 0, time.UTC),
					TotalARS:       decimal.Decimal{},
					TotalUSD:       decimal.Decimal{},
				},
			},
			expectedBuilder: CSVBuilder{},
			expectedError:   errors.New("close date cannot be zero"),
		},
		{
			name: "missing expiration date",
			args: args{
				statement: StatementContext{
					Bank:           Bank("Santander"),
					CardCompany:    CardCompany("VISA"),
					CloseDate:      time.Date(2023, 11, 1, 0, 0, 0, 0, time.UTC),
					ExpirationDate: time.Time{},
					TotalARS:       decimal.Decimal{},
					TotalUSD:       decimal.Decimal{},
				},
			},
			expectedBuilder: CSVBuilder{},
			expectedError:   errors.New("expiration date cannot be zero"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			actualBuilder, actualError := NewCSVBuilder(tt.args.statement)

			// Then
			require.Equal(t, tt.expectedBuilder, actualBuilder)
			require.Equal(t, tt.expectedError, actualError)
		})
	}
}
