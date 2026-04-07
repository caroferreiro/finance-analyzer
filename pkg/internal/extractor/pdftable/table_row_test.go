package pdftable_test

import (
	"testing"

	"github.com/Alechan/finance-analyzer/pkg/internal/extractor/pdftable"
	"github.com/stretchr/testify/require"
)

func TestRow_MatchesMovementWithoutYearAndMonth(t *testing.T) {
	tests := []struct {
		name           string
		row            pdftable.Row
		expectedResult bool
	}{
		{
			name: "Valid two-digit date",
			row: pdftable.Row{
				RawOriginalDate: "23",
			},
			expectedResult: true,
		},
		{
			name: "Invalid date - too long",
			row: pdftable.Row{
				RawOriginalDate: "123",
			},
			expectedResult: false,
		},
		{
			name: "Invalid date - non-numeric",
			row: pdftable.Row{
				RawOriginalDate: "ab",
			},
			expectedResult: false,
		},
		{
			name: "Invalid date - empty",
			row: pdftable.Row{
				RawOriginalDate: "",
			},
			expectedResult: false,
		},
		{
			name: "Continuation line - empty date with receipt and detail and ARS amount",
			row: pdftable.Row{
				RawOriginalDate:                "",
				RawReceiptNumber:               "158328 *",
				RawDetailWithMaybeInstallments: "SEGURO DE VIDA 0576315018",
				RawAmountARS:                   "10.310,71",
				RawAmountUSD:                   "",
			},
			expectedResult: true,
		},
		{
			name: "Continuation line - empty date with receipt and detail and ARS amount (MARKOVA)",
			row: pdftable.Row{
				RawOriginalDate:                "",
				RawReceiptNumber:               "903064 *",
				RawDetailWithMaybeInstallments: "MERPAGO*MARKOVA             C.09/12",
				RawAmountARS:                   "32.500,00",
				RawAmountUSD:                   "",
			},
			expectedResult: true,
		},
		{
			name: "Continuation line - empty date with receipt and detail and ARS amount (LASPEPAS)",
			row: pdftable.Row{
				RawOriginalDate:                "",
				RawReceiptNumber:               "129323 *",
				RawDetailWithMaybeInstallments: "MERPAGO*LASPEPAS            C.07/09",
				RawAmountARS:                   "55.990,00",
				RawAmountUSD:                   "",
			},
			expectedResult: true,
		},
		{
			name: "Continuation line - empty date with receipt and detail and ARS amount (LAS MARGARITAS)",
			row: pdftable.Row{
				RawOriginalDate:                "",
				RawReceiptNumber:               "157300 *",
				RawDetailWithMaybeInstallments: "LAS MARGARITAS              C.07/12",
				RawAmountARS:                   "12.695,83",
				RawAmountUSD:                   "",
			},
			expectedResult: true,
		},
		{
			name: "Continuation line - empty date with receipt and detail and ARS amount (SBSLIBRERIAS)",
			row: pdftable.Row{
				RawOriginalDate:                "",
				RawReceiptNumber:               "098301 *",
				RawDetailWithMaybeInstallments: "MERPAGO*SBSLIBRERIAS        C.05/06",
				RawAmountARS:                   "2.548,50",
				RawAmountUSD:                   "",
			},
			expectedResult: true,
		},
		{
			name: "Continuation line - empty date with receipt and detail and ARS amount (FARMACITY)",
			row: pdftable.Row{
				RawOriginalDate:                "",
				RawReceiptNumber:               "515488 *",
				RawDetailWithMaybeInstallments: "FARMACITY                   C.05/06",
				RawAmountARS:                   "4.077,86",
				RawAmountUSD:                   "",
			},
			expectedResult: true,
		},
		{
			name: "Continuation line - empty date with receipt and detail and ARS amount (LAS MARGARITAS CABIL)",
			row: pdftable.Row{
				RawOriginalDate:                "",
				RawReceiptNumber:               "421484 *",
				RawDetailWithMaybeInstallments: "LAS MARGARITAS CABIL        C.05/12",
				RawAmountARS:                   "10.114,16",
				RawAmountUSD:                   "",
			},
			expectedResult: true,
		},
		{
			name: "Continuation line - empty date with receipt and detail and ARS amount (=COMPLOT)",
			row: pdftable.Row{
				RawOriginalDate:                "",
				RawReceiptNumber:               "507303 *",
				RawDetailWithMaybeInstallments: "=COMPLOT                    C.05/06",
				RawAmountARS:                   "19.975,00",
				RawAmountUSD:                   "",
			},
			expectedResult: true,
		},
		{
			name: "Continuation line - empty date with receipt and detail and ARS amount (2PRODUCTOS)",
			row: pdftable.Row{
				RawOriginalDate:                "",
				RawReceiptNumber:               "615302 *",
				RawDetailWithMaybeInstallments: "MERPAGO*2PRODUCTOS          C.05/06",
				RawAmountARS:                   "8.496,83",
				RawAmountUSD:                   "",
			},
			expectedResult: true,
		},
		{
			name: "Empty date but no movement data - should not match",
			row: pdftable.Row{
				RawOriginalDate:                "",
				RawReceiptNumber:               "",
				RawDetailWithMaybeInstallments: "",
				RawAmountARS:                   "",
				RawAmountUSD:                   "",
			},
			expectedResult: false,
		},
		{
			name: "Empty date with only detail but no amounts - should not match",
			row: pdftable.Row{
				RawOriginalDate:                "",
				RawReceiptNumber:               "",
				RawDetailWithMaybeInstallments: "SOME DETAIL",
				RawAmountARS:                   "",
				RawAmountUSD:                   "",
			},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			// Test data is already set up in the test case

			// When
			actualResult := tt.row.MatchesMovementWithoutYearAndMonth()

			// Then
			require.Equal(t, tt.expectedResult, actualResult)
		})
	}
}

// TestNewRow verifies that NewRow correctly assigns individual field values without any parsing logic.
// This is a "dumb" constructor that simply assigns the provided values to the Row struct.
func TestNewRow(t *testing.T) {
	tests := []struct {
		name          string
		rawText       string
		originalDate  string
		receiptNumber string
		detail        string
		arsAmount     string
		usdAmount     string
		expectedRow   pdftable.Row
	}{
		{
			name:          "all fields populated",
			rawText:       "complete row text",
			originalDate:  "25",
			receiptNumber: "123456",
			detail:        "PURCHASE AT STORE",
			arsAmount:     "1500.50",
			usdAmount:     "15.00",
			expectedRow: pdftable.Row{
				RawText:                        "complete row text",
				RawOriginalDate:                "25",
				RawReceiptNumber:               "123456",
				RawDetailWithMaybeInstallments: "PURCHASE AT STORE",
				RawAmountARS:                   "1500.50",
				RawAmountUSD:                   "15.00",
			},
		},
		{
			name:          "empty fields",
			rawText:       "incomplete row",
			originalDate:  "",
			receiptNumber: "",
			detail:        "",
			arsAmount:     "",
			usdAmount:     "",
			expectedRow: pdftable.Row{
				RawText:                        "incomplete row",
				RawOriginalDate:                "",
				RawReceiptNumber:               "",
				RawDetailWithMaybeInstallments: "",
				RawAmountARS:                   "",
				RawAmountUSD:                   "",
			},
		},
		{
			name:          "mixed populated and empty fields",
			rawText:       "mixed row",
			originalDate:  "15",
			receiptNumber: "",
			detail:        "SOME DETAIL",
			arsAmount:     "500.00",
			usdAmount:     "",
			expectedRow: pdftable.Row{
				RawText:                        "mixed row",
				RawOriginalDate:                "15",
				RawReceiptNumber:               "",
				RawDetailWithMaybeInstallments: "SOME DETAIL",
				RawAmountARS:                   "500.00",
				RawAmountUSD:                   "",
			},
		},
		{
			name:          "whitespace in fields",
			rawText:       "whitespace test",
			originalDate:  " 20 ",
			receiptNumber: " 789 ",
			detail:        " DETAIL WITH SPACES ",
			arsAmount:     " 1000.00 ",
			usdAmount:     " 10.00 ",
			expectedRow: pdftable.Row{
				RawText:                        "whitespace test",
				RawOriginalDate:                " 20 ",
				RawReceiptNumber:               " 789 ",
				RawDetailWithMaybeInstallments: " DETAIL WITH SPACES ",
				RawAmountARS:                   " 1000.00 ",
				RawAmountUSD:                   " 10.00 ",
			},
		},
		{
			name:          "special characters in fields",
			rawText:       "special chars: !@#$%^&*()",
			originalDate:  "31",
			receiptNumber: "ABC-123",
			detail:        "PURCHASE @ STORE #1",
			arsAmount:     "$1,234.56",
			usdAmount:     "€10.50",
			expectedRow: pdftable.Row{
				RawText:                        "special chars: !@#$%^&*()",
				RawOriginalDate:                "31",
				RawReceiptNumber:               "ABC-123",
				RawDetailWithMaybeInstallments: "PURCHASE @ STORE #1",
				RawAmountARS:                   "$1,234.56",
				RawAmountUSD:                   "€10.50",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			actualRow := pdftable.NewRow(tt.rawText, tt.originalDate, tt.receiptNumber, tt.detail, tt.arsAmount, tt.usdAmount)

			// Then
			require.Equal(t, tt.expectedRow, actualRow)
		})
	}
}
