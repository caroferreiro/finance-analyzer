package santander

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Alechan/finance-analyzer/pkg/internal/extractor/pdftable"
	"github.com/Alechan/finance-analyzer/pkg/internal/extractor/santander/testdata"
	"github.com/Alechan/finance-analyzer/pkg/internal/pdfcardsummary"
	"github.com/Alechan/finance-analyzer/pkg/internal/platform/testsale"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestConvertRawWithMonthToMovement(t *testing.T) {
	cfg := DefaultConfig()

	// Get expected movements from realistic statement data
	expectedTable := testdata.ExpectedSantanderTableData(t)

	testCases := []struct {
		name          string
		row           pdftable.Row
		expected      pdfcardsummary.Movement
		expectedError error
	}{
		{
			name:     "valid movement with installments",
			row:      testdata.RealisticStatementRows[11], // Card movement with installments (C.07/09)
			expected: expectedTable.Cards[0].Movements[0], // First card, first movement
		},
		{
			name:     "valid movement without installments",
			row:      testdata.RealisticStatementRows[12], // Card movement without installments
			expected: expectedTable.Cards[0].Movements[1], // First card, second movement
		},
		{
			name: "valid movement with USD amount",
			row:  pdftable.NewRow("24 Diciem. 20 123456 *  PAGO EN USD                                                                   54,68-", "24 Diciem. 20", "123456 *", "PAGO EN USD", "", "54,68-"),
			expected: pdfcardsummary.Movement{
				OriginalDate:       testsale.DatePtr(2024, time.December, 20),
				ReceiptNumber:      testsale.StrPtr("123456 *"),
				Detail:             "PAGO EN USD",
				CurrentInstallment: nil,
				TotalInstallments:  nil,
				AmountARS:          decimal.Zero,
				AmountUSD:          testsale.AsDecimal(t, "-54.68"),
			},
		},
		{
			name:          "invalid date format",
			row:           pdftable.NewRow("24 Invalid 25 123456 *  SOME DETAIL                                                            1.234,56", "24 Invalid 25", "123456 *", "SOME DETAIL", "1.234,56", ""),
			expectedError: errors.New("error extracting date from text with month: error converting month Invalid to time.Month: month Invalid not found in mapper"),
		},
		{
			name:          "invalid ARS amount",
			row:           pdftable.NewRow("24 Julio   25 123456 *  SOME DETAIL                                                            invalid", "24 Julio   25", "123456 *", "SOME DETAIL", "", "invalid"),
			expectedError: errors.New("error extracting USD amount: error converting invalid to decimal: can't convert invalid to decimal"),
		},
		{
			name:          "invalid USD amount",
			row:           pdftable.NewRow("24 Julio   25 123456 *  SOME DETAIL                                                            1.234,56 invalid", "24 Julio   25", "123456 *", "SOME DETAIL", "1.234,56", "invalid"),
			expectedError: errors.New("error extracting USD amount: error converting invalid to decimal: can't convert invalid to decimal"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			// Row is already set up in the test case

			// When
			actualMovement, actualError := ConvertRawWithMonthToMovement(tc.row, cfg.CuotasSubDetailRegex)

			// Then
			require.Equal(t, tc.expectedError, actualError)
			require.Equal(t, tc.expected, actualMovement)
		})
	}
}

func TestMergeAmountsIntoMovement(t *testing.T) {
	testCases := []struct {
		name                  string
		movement              pdfcardsummary.Movement
		amountOnlyRow         pdftable.Row
		expected              pdfcardsummary.Movement
		expectedHasNoDecimals bool
		expectedError         error
	}{
		{
			name: "merge ARS amount only",
			movement: pdfcardsummary.Movement{
				OriginalDate:       testsale.DatePtr(2022, time.December, 1),
				ReceiptNumber:      testsale.StrPtr("111111 *"),
				Detail:             "TEST MERCHANT",
				CurrentInstallment: nil,
				TotalInstallments:  nil,
				AmountARS:          decimal.Zero,
				AmountUSD:          decimal.Zero,
			},
			amountOnlyRow: pdftable.NewRow("                                                                    1.316,68", "", "", "", "1.316,68", ""),
			expected: pdfcardsummary.Movement{
				OriginalDate:       testsale.DatePtr(2022, time.December, 1),
				ReceiptNumber:      testsale.StrPtr("111111 *"),
				Detail:             "TEST MERCHANT",
				CurrentInstallment: nil,
				TotalInstallments:  nil,
				AmountARS:          testsale.AsDecimal(t, "1316.68"),
				AmountUSD:          decimal.Zero,
			},
			expectedHasNoDecimals: false,
			expectedError:         nil,
		},
		{
			name: "merge USD amount only",
			movement: pdfcardsummary.Movement{
				OriginalDate:       testsale.DatePtr(2024, time.August, 10),
				ReceiptNumber:      testsale.StrPtr("123456*"),
				Detail:             "USD PAYMENT",
				CurrentInstallment: nil,
				TotalInstallments:  nil,
				AmountARS:          decimal.Zero,
				AmountUSD:          decimal.Zero,
			},
			amountOnlyRow: pdftable.NewRow("                                                                                    54,68", "", "", "", "", "54,68"),
			expected: pdfcardsummary.Movement{
				OriginalDate:       testsale.DatePtr(2024, time.August, 10),
				ReceiptNumber:      testsale.StrPtr("123456*"),
				Detail:             "USD PAYMENT",
				CurrentInstallment: nil,
				TotalInstallments:  nil,
				AmountARS:          decimal.Zero,
				AmountUSD:          testsale.AsDecimal(t, "54.68"),
			},
			expectedHasNoDecimals: false,
			expectedError:         nil,
		},
		{
			name: "merge both ARS and USD amounts",
			movement: pdfcardsummary.Movement{
				OriginalDate:       testsale.DatePtr(2024, time.August, 10),
				ReceiptNumber:      testsale.StrPtr("123456*"),
				Detail:             "MIXED PAYMENT",
				CurrentInstallment: nil,
				TotalInstallments:  nil,
				AmountARS:          decimal.Zero,
				AmountUSD:          decimal.Zero,
			},
			amountOnlyRow: pdftable.NewRow("                                                                    1.000,00                50,00", "", "", "", "1.000,00", "50,00"),
			expected: pdfcardsummary.Movement{
				OriginalDate:       testsale.DatePtr(2024, time.August, 10),
				ReceiptNumber:      testsale.StrPtr("123456*"),
				Detail:             "MIXED PAYMENT",
				CurrentInstallment: nil,
				TotalInstallments:  nil,
				AmountARS:          testsale.AsDecimal(t, "1000.00"),
				AmountUSD:          testsale.AsDecimal(t, "50.00"),
			},
			expectedHasNoDecimals: false,
			expectedError:         nil,
		},
		{
			name: "merge negative ARS amount",
			movement: pdfcardsummary.Movement{
				OriginalDate:       testsale.DatePtr(2024, time.August, 10),
				ReceiptNumber:      testsale.StrPtr("123456*"),
				Detail:             "REFUND",
				CurrentInstallment: nil,
				TotalInstallments:  nil,
				AmountARS:          decimal.Zero,
				AmountUSD:          decimal.Zero,
			},
			amountOnlyRow: pdftable.NewRow("                                                                    1.316,68-", "", "", "", "1.316,68-", ""),
			expected: pdfcardsummary.Movement{
				OriginalDate:       testsale.DatePtr(2024, time.August, 10),
				ReceiptNumber:      testsale.StrPtr("123456*"),
				Detail:             "REFUND",
				CurrentInstallment: nil,
				TotalInstallments:  nil,
				AmountARS:          testsale.AsDecimal(t, "-1316.68"),
				AmountUSD:          decimal.Zero,
			},
			expectedHasNoDecimals: false,
			expectedError:         nil,
		},
		{
			name: "merge negative USD amount",
			movement: pdfcardsummary.Movement{
				OriginalDate:       testsale.DatePtr(2024, time.August, 10),
				ReceiptNumber:      testsale.StrPtr("123456*"),
				Detail:             "USD REFUND",
				CurrentInstallment: nil,
				TotalInstallments:  nil,
				AmountARS:          decimal.Zero,
				AmountUSD:          decimal.Zero,
			},
			amountOnlyRow: pdftable.NewRow("                                                                                    54,68-", "", "", "", "", "54,68-"),
			expected: pdfcardsummary.Movement{
				OriginalDate:       testsale.DatePtr(2024, time.August, 10),
				ReceiptNumber:      testsale.StrPtr("123456*"),
				Detail:             "USD REFUND",
				CurrentInstallment: nil,
				TotalInstallments:  nil,
				AmountARS:          decimal.Zero,
				AmountUSD:          testsale.AsDecimal(t, "-54.68"),
			},
			expectedHasNoDecimals: false,
			expectedError:         nil,
		},
		{
			name: "error - invalid ARS amount format",
			movement: pdfcardsummary.Movement{
				OriginalDate:       testsale.DatePtr(2024, time.August, 10),
				ReceiptNumber:      testsale.StrPtr("123456*"),
				Detail:             "INVALID",
				CurrentInstallment: nil,
				TotalInstallments:  nil,
				AmountARS:          decimal.Zero,
				AmountUSD:          decimal.Zero,
			},
			amountOnlyRow: pdftable.NewRow("                                                                    invalid", "", "", "", "invalid", ""),
			expected: pdfcardsummary.Movement{
				OriginalDate:       testsale.DatePtr(2024, time.August, 10),
				ReceiptNumber:      testsale.StrPtr("123456*"),
				Detail:             "INVALID",
				CurrentInstallment: nil,
				TotalInstallments:  nil,
				AmountARS:          decimal.Zero,
				AmountUSD:          decimal.Zero,
			},
			expectedHasNoDecimals: false,
			expectedError:         fmt.Errorf("error parsing ARS amount from amount-only row: %w", errors.New("error converting invalid to decimal: can't convert invalid to decimal")),
		},
		{
			name: "error - invalid USD amount format",
			movement: pdfcardsummary.Movement{
				OriginalDate:       testsale.DatePtr(2024, time.August, 10),
				ReceiptNumber:      testsale.StrPtr("123456*"),
				Detail:             "INVALID",
				CurrentInstallment: nil,
				TotalInstallments:  nil,
				AmountARS:          decimal.Zero,
				AmountUSD:          decimal.Zero,
			},
			amountOnlyRow: pdftable.NewRow("                                                                                    invalid", "", "", "", "", "invalid"),
			expected: pdfcardsummary.Movement{
				OriginalDate:       testsale.DatePtr(2024, time.August, 10),
				ReceiptNumber:      testsale.StrPtr("123456*"),
				Detail:             "INVALID",
				CurrentInstallment: nil,
				TotalInstallments:  nil,
				AmountARS:          decimal.Zero,
				AmountUSD:          decimal.Zero,
			},
			expectedHasNoDecimals: false,
			expectedError:         fmt.Errorf("error parsing USD amount from amount-only row: %w", errors.New("error converting invalid to decimal: can't convert invalid to decimal")),
		},
		{
			name: "merge ARS amount from regex extraction (empty RawAmountARS) - broken line case",
			movement: pdfcardsummary.Movement{
				OriginalDate:       testsale.DatePtr(2022, time.December, 1),
				ReceiptNumber:      testsale.StrPtr("111111 *"),
				Detail:             "TEST MERCHANT",
				CurrentInstallment: nil,
				TotalInstallments:  nil,
				AmountARS:          decimal.Zero,
				AmountUSD:          decimal.Zero,
			},
			// RawAmountARS is empty, but RawText contains the amount (broken line case)
			amountOnlyRow: pdftable.NewRow("            1.316", "", "", "", "", ""),
			expected: pdfcardsummary.Movement{
				OriginalDate:       testsale.DatePtr(2022, time.December, 1),
				ReceiptNumber:      testsale.StrPtr("111111 *"),
				Detail:             "TEST MERCHANT",
				CurrentInstallment: nil,
				TotalInstallments:  nil,
				AmountARS:          testsale.AsDecimal(t, "1316.00"),
				AmountUSD:          decimal.Zero,
			},
			expectedHasNoDecimals: true, // Amount "1.316" has no decimals
			expectedError:         nil,
		},
		{
			name: "merge ARS amount from regex extraction with decimal part (empty RawAmountARS)",
			movement: pdfcardsummary.Movement{
				OriginalDate:       testsale.DatePtr(2022, time.December, 1),
				ReceiptNumber:      testsale.StrPtr("111111 *"),
				Detail:             "TEST MOVEMENT",
				CurrentInstallment: nil,
				TotalInstallments:  nil,
				AmountARS:          decimal.Zero,
				AmountUSD:          decimal.Zero,
			},
			// RawAmountARS is empty, but RawText contains the amount with decimal part
			amountOnlyRow: pdftable.NewRow("                                                                    1.316,68", "", "", "", "", ""),
			expected: pdfcardsummary.Movement{
				OriginalDate:       testsale.DatePtr(2022, time.December, 1),
				ReceiptNumber:      testsale.StrPtr("111111 *"),
				Detail:             "TEST MOVEMENT",
				CurrentInstallment: nil,
				TotalInstallments:  nil,
				AmountARS:          testsale.AsDecimal(t, "1316.68"),
				AmountUSD:          decimal.Zero,
			},
			expectedHasNoDecimals: false,
			expectedError:         nil,
		},
		{
			name: "merge USD amount from regex extraction (empty RawAmountUSD)",
			movement: pdfcardsummary.Movement{
				OriginalDate:       testsale.DatePtr(2024, time.August, 10),
				ReceiptNumber:      testsale.StrPtr("123456*"),
				Detail:             "USD PAYMENT",
				CurrentInstallment: nil,
				TotalInstallments:  nil,
				AmountARS:          decimal.Zero,
				AmountUSD:          decimal.Zero,
			},
			// RawAmountUSD is empty, but RawText contains the amount at USD position
			amountOnlyRow: pdftable.NewRow("                                                                                    54,68", "", "", "", "", ""),
			expected: pdfcardsummary.Movement{
				OriginalDate:       testsale.DatePtr(2024, time.August, 10),
				ReceiptNumber:      testsale.StrPtr("123456*"),
				Detail:             "USD PAYMENT",
				CurrentInstallment: nil,
				TotalInstallments:  nil,
				AmountARS:          decimal.Zero,
				AmountUSD:          testsale.AsDecimal(t, "54.68"),
			},
			expectedHasNoDecimals: false,
			expectedError:         nil,
		},
		{
			name: "merge negative ARS amount from regex extraction (empty RawAmountARS)",
			movement: pdfcardsummary.Movement{
				OriginalDate:       testsale.DatePtr(2024, time.August, 10),
				ReceiptNumber:      testsale.StrPtr("123456*"),
				Detail:             "REFUND",
				CurrentInstallment: nil,
				TotalInstallments:  nil,
				AmountARS:          decimal.Zero,
				AmountUSD:          decimal.Zero,
			},
			// RawAmountARS is empty, but RawText contains negative amount
			amountOnlyRow: pdftable.NewRow("                                                                    1.316,68-", "", "", "", "", ""),
			expected: pdfcardsummary.Movement{
				OriginalDate:       testsale.DatePtr(2024, time.August, 10),
				ReceiptNumber:      testsale.StrPtr("123456*"),
				Detail:             "REFUND",
				CurrentInstallment: nil,
				TotalInstallments:  nil,
				AmountARS:          testsale.AsDecimal(t, "-1316.68"),
				AmountUSD:          decimal.Zero,
			},
			expectedHasNoDecimals: false,
			expectedError:         nil,
		},
		{
			name: "merge negative ARS amount without decimal part from regex extraction (empty RawAmountARS)",
			movement: pdfcardsummary.Movement{
				OriginalDate:       testsale.DatePtr(2024, time.August, 10),
				ReceiptNumber:      testsale.StrPtr("123456*"),
				Detail:             "REFUND",
				CurrentInstallment: nil,
				TotalInstallments:  nil,
				AmountARS:          decimal.Zero,
				AmountUSD:          decimal.Zero,
			},
			// RawAmountARS is empty, but RawText contains negative amount without decimal part
			amountOnlyRow: pdftable.NewRow("            1.316-", "", "", "", "", ""),
			expected: pdfcardsummary.Movement{
				OriginalDate:       testsale.DatePtr(2024, time.August, 10),
				ReceiptNumber:      testsale.StrPtr("123456*"),
				Detail:             "REFUND",
				CurrentInstallment: nil,
				TotalInstallments:  nil,
				AmountARS:          testsale.AsDecimal(t, "-1316.00"),
				AmountUSD:          decimal.Zero,
			},
			expectedHasNoDecimals: true, // Amount "1.316-" has no decimals
			expectedError:         nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			mov := tc.movement

			// When
			hasNoDecimals, err := mergeAmountsIntoMovement(&mov, tc.amountOnlyRow, DefaultConfig())

			// Then
			require.Equal(t, tc.expectedError, err)
			require.Equal(t, tc.expected, mov)
			require.Equal(t, tc.expectedHasNoDecimals, hasNoDecimals)
		})
	}
}
