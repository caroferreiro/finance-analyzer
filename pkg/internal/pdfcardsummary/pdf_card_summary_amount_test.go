package pdfcardsummary

import (
	"fmt"
	"testing"

	"github.com/Alechan/finance-analyzer/pkg/internal/platform/testsale"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestPDFAmountToDecimal(t *testing.T) {
	testCases := []struct {
		name            string
		input           string
		expectedDecimal decimal.Decimal
		expectedError   error
	}{
		{
			name:            "empty string",
			input:           "",
			expectedDecimal: decimal.Zero,
		},
		{
			name:            "zero amount",
			input:           "0,00",
			expectedDecimal: decimal.Zero,
		},
		{
			name:            "positive amount with dots and comma",
			input:           "4.226,53",
			expectedDecimal: decimal.NewFromFloat(4226.53),
		},
		{
			name:            "negative amount with dots and comma",
			input:           "4.226,53-",
			expectedDecimal: decimal.NewFromFloat(-4226.53),
		},
		{
			name:            "positive amount with single digit after comma",
			input:           "4.226,5",
			expectedDecimal: decimal.NewFromFloat(4226.5),
		},
		{
			name:            "positive amount with no decimal places",
			input:           "4.226",
			expectedDecimal: decimal.NewFromFloat(4226),
		},
		{
			name:            "positive amount with multiple thousand separators",
			input:           "123.456.789,00",
			expectedDecimal: decimal.NewFromFloat(123456789.00),
		},
		{
			name:          "invalid amount",
			input:         "invalid",
			expectedError: fmt.Errorf("error converting invalid to decimal: can't convert invalid to decimal"),
		},
		{
			// NOTE: Although multiple dots or commas (e.g. "4.226.53") might seem invalid,
			// the actual data from sources like PDFs can be inconsistent. We rely on the upstream data
			// and do not apply strict validation. Here, all dots are removed, so "4.226.53" becomes "422653".
			name:            "invalid_amount_with_multiple_dots",
			input:           "4.226.53",
			expectedDecimal: decimal.NewFromInt(422653),
		},
		{
			name:          "invalid amount with multiple commas",
			input:         "4,226,53",
			expectedError: fmt.Errorf("error converting 4,226,53 to decimal: can't convert 4.226.53 to decimal: too many .s"),
		},
		{
			name:            "valid amount with multiple thousand separators",
			input:           "1005.583,43",
			expectedDecimal: decimal.NewFromFloat(1005583.43),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// When
			actualDecimal, actualErr := PDFAmountToDecimal(tc.input)

			// Then
			require.Equal(t, tc.expectedError, actualErr)
			testsale.AssertDecimalEqual(t, tc.expectedDecimal, actualDecimal)
		})
	}
}
