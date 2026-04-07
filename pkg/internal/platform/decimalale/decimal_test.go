package decimalale_test

import (
	"github.com/Alechan/finance-analyzer/pkg/internal/platform/decimalale"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFormatToFinancialString(t *testing.T) {
	testCases := []struct {
		input    decimal.Decimal
		expected string
	}{
		{decimal.NewFromFloat(1234567.89), "1 234 567.89"},
		{decimal.NewFromFloat(1234.56789), "1 234.57"},
		{decimal.NewFromFloat(123.456789), "123.46"},
		{decimal.NewFromFloat(123456789.0), "123 456 789.00"},
		{decimal.NewFromFloat(123456789.01234567), "123 456 789.01"},
		{decimal.NewFromFloat(-123456789.01234567), "-123 456 789.01"},
		{decimal.NewFromFloat(0.0123456789), "0.01"},
		{decimal.NewFromFloat(0.0), "0.00"},
		{decimal.NewFromFloat(-0.0), "0.00"},
		{decimal.NewFromFloat(-0.0123456789), "-0.01"},
		{decimal.NewFromFloat(-123456789.0), "-123 456 789.00"},
		{decimal.NewFromFloat(-123.456789), "-123.46"},
	}

	for _, tc := range testCases {
		t.Run(tc.input.String(), func(t *testing.T) {
			formatted := decimalale.FormatToFinancialString(tc.input)
			if formatted != tc.expected {
				t.Errorf("Expected %s, but got %s", tc.expected, formatted)
			}
		})
	}
}

func TestFormatToArgentineSeparators(t *testing.T) {
	tests := []struct {
		name           string
		input          decimal.Decimal
		expectedString string
	}{
		{
			name:           "1234567.89",
			input:          decimal.NewFromFloat(1234567.89),
			expectedString: "1.234.567,89",
		},
		{
			name:           "1234.56789",
			input:          decimal.NewFromFloat(1234.56789),
			expectedString: "1.234,57",
		},
		{
			name:           "0",
			input:          decimal.NewFromFloat(0),
			expectedString: "0,00",
		},
		{
			name:           "-0.01",
			input:          decimal.NewFromFloat(-0.01),
			expectedString: "-0,01",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			actualString := decimalale.FormatToArgentineSeparators(tt.input)

			// Then
			require.Equal(t, tt.expectedString, actualString)
		})
	}
}
