package visaprisma

import (
	"strings"
	"testing"

	"github.com/Alechan/finance-analyzer/pkg/internal/platform/testsale"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestTotalConsumosTarjetaRegex_ParsesCardTotals(t *testing.T) {
	testCases := []struct {
		name               string
		inputText          string
		expectedCardNumber string
		expectedOwner      string
		expectedARSTotal   decimal.Decimal
		expectedUSDTotal   decimal.Decimal
		shouldMatch        bool
	}{
		{
			name:               "card with positive USD amount",
			inputText:          "Total Consumos de TEST USER                                         59.924,78              106,35",
			expectedCardNumber: "",
			expectedOwner:      "TEST USER",
			expectedARSTotal:   testsale.AsDecimal(t, "59924.78"),
			expectedUSDTotal:   testsale.AsDecimal(t, "106.35"),
			shouldMatch:        true,
		},
		{
			name:               "card with negative USD amount (trailing minus)",
			inputText:          "Total Consumos de TEST USER                                        749.036,71            242,99-",
			expectedCardNumber: "",
			expectedOwner:      "TEST USER",
			expectedARSTotal:   testsale.AsDecimal(t, "749036.71"),
			expectedUSDTotal:   testsale.AsDecimal(t, "-242.99"),
			shouldMatch:        true,
		},
		{
			name:               "card with card number and positive USD",
			inputText:          "Tarjeta 1111 Total Consumos de TEST USER                                         59.924,78              0,00",
			expectedCardNumber: "1111",
			expectedOwner:      "TEST USER",
			expectedARSTotal:   testsale.AsDecimal(t, "59924.78"),
			expectedUSDTotal:   decimal.Zero,
			shouldMatch:        true,
		},
		{
			name:               "card with card number and negative USD amount (trailing minus)",
			inputText:          "Tarjeta 2222 Total Consumos de TEST USER                                        749.036,71            242,99-",
			expectedCardNumber: "2222",
			expectedOwner:      "TEST USER",
			expectedARSTotal:   testsale.AsDecimal(t, "749036.71"),
			expectedUSDTotal:   testsale.AsDecimal(t, "-242.99"),
			shouldMatch:        true,
		},
		{
			name:               "card with zero USD amount",
			inputText:          "Total Consumos de TEST USER                                        519.056,00              0,00",
			expectedCardNumber: "",
			expectedOwner:      "TEST USER",
			expectedARSTotal:   testsale.AsDecimal(t, "519056.00"),
			expectedUSDTotal:   decimal.Zero,
			shouldMatch:        true,
		},
		{
			name:        "text that does not match",
			inputText:   "Some random text that doesn't match the pattern",
			shouldMatch: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			// Input text is provided in test case

			// When
			matches := totalConsumosTarjetaRegex.FindStringSubmatch(tc.inputText)

			// Then
			if !tc.shouldMatch {
				require.Empty(t, matches, "regex should not match")
				return
			}

			require.NotEmpty(t, matches, "regex should match")
			require.GreaterOrEqual(t, len(matches), 5, "regex should have at least 5 groups (full match + 4 groups)")

			// Group 1: Card number (optional)
			actualCardNumber := ""
			if matches[1] != "" {
				actualCardNumber = matches[1]
			}
			require.Equal(t, tc.expectedCardNumber, actualCardNumber, "card number should match")

			// Group 2: Owner (trim whitespace as done in production code)
			actualOwner := strings.TrimSpace(matches[2])
			require.Equal(t, tc.expectedOwner, actualOwner, "owner should match")

			// Group 3: ARS total
			actualARSRaw := matches[3]
			actualARS, err := PDFAmountToDecimal(actualARSRaw)
			require.NoError(t, err, "should parse ARS amount")
			testsale.AssertDecimalEqual(t, tc.expectedARSTotal, actualARS)

			// Group 4: USD total - this is the critical test for negative amounts
			actualUSDRaw := matches[4]
			actualUSD, err := PDFAmountToDecimal(actualUSDRaw)
			require.NoError(t, err, "should parse USD amount")
			testsale.AssertDecimalEqual(t, tc.expectedUSDTotal, actualUSD)
		})
	}
}
