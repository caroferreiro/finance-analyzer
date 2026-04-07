package santander

import (
	"testing"

	"github.com/Alechan/finance-analyzer/pkg/internal/extractor/pdftable"
	"github.com/stretchr/testify/require"
)

// TestRowFactoryWithSantanderPositions verifies that RowFactory.CreateRow correctly parses rows using Santander's
// standard table column positions. These positions are specific to Santander's PDF format and
// are defined in santander.DefaultConfig().TableColumnPositions.
func TestRowFactoryWithSantanderPositions(t *testing.T) {
	// Define standard table positions used in most tests - these are Santander-specific positions
	standardPositions := DefaultConfig().TableColumnPositions
	factory := pdftable.NewRowFactory(standardPositions)

	tests := []struct {
		name          string
		rawText       string
		positions     pdftable.PDFTablePositions
		expectedRow   pdftable.Row
		expectedError error // Using error field for explicit error expectations
	}{
		{
			name:      "full date + receipt number + detail + ARS + WITH trailing whitespace",
			rawText:   "24 Julio   01 123456 *  MERCHANT XYZ            C.07/09                       12.345,67                    ",
			positions: standardPositions,
			expectedRow: pdftable.Row{
				RawText:                        "24 Julio   01 123456 *  MERCHANT XYZ            C.07/09                       12.345,67                    ",
				RawOriginalDate:                "24 Julio   01",
				RawReceiptNumber:               "123456 *",
				RawDetailWithMaybeInstallments: "MERCHANT XYZ            C.07/09",
				RawAmountARS:                   "12.345,67",
				RawAmountUSD:                   "",
			},
		},
		{
			name:      "full date + receipt number + detail + ARS + WITHOUT trailing whitespace",
			rawText:   "24 Julio   01 123456 *  MERCHANT XYZ            C.07/09                       12.345,67",
			positions: standardPositions,
			expectedRow: pdftable.Row{
				RawText:                        "24 Julio   01 123456 *  MERCHANT XYZ            C.07/09                       12.345,67",
				RawOriginalDate:                "24 Julio   01",
				RawReceiptNumber:               "123456 *",
				RawDetailWithMaybeInstallments: "MERCHANT XYZ            C.07/09",
				RawAmountARS:                   "12.345,67",
				RawAmountUSD:                   "",
			},
		},
		{
			name:      "saldo anterior (only ARS and USD)",
			rawText:   "                        SALDO ANTERIOR                                           123.456,78              98,76",
			positions: standardPositions,
			expectedRow: pdftable.Row{
				RawText:                        "                        SALDO ANTERIOR                                           123.456,78              98,76",
				RawOriginalDate:                "",
				RawReceiptNumber:               "",
				RawDetailWithMaybeInstallments: "SALDO ANTERIOR",
				RawAmountARS:                   "123.456,78",
				RawAmountUSD:                   "98,76",
			},
		},
		{
			name:      "no year/month + detail + ARS",
			rawText:   "           31 771488 *  SERVICE PROVIDER ABC                                     45.678,90",
			positions: standardPositions,
			expectedRow: pdftable.Row{
				RawText:                        "           31 771488 *  SERVICE PROVIDER ABC                                     45.678,90",
				RawOriginalDate:                "31",
				RawReceiptNumber:               "771488 *",
				RawDetailWithMaybeInstallments: "SERVICE PROVIDER ABC",
				RawAmountARS:                   "45.678,90",
				RawAmountUSD:                   "",
			},
		},
		{
			name:      "no year/month + detail + USD",
			rawText:   "           20           USD PAYMENT                                                                    98,76-",
			positions: standardPositions,
			expectedRow: pdftable.Row{
				RawText:                        "           20           USD PAYMENT                                                                    98,76-",
				RawOriginalDate:                "20",
				RawReceiptNumber:               "",
				RawDetailWithMaybeInstallments: "USD PAYMENT",
				RawAmountARS:                   "",
				RawAmountUSD:                   "98,76-",
			},
		},
		{
			name:      "row with negative ARS amount",
			rawText:   "           20           ARS PAYMENT                                             234.567,89-",
			positions: standardPositions,
			expectedRow: pdftable.Row{
				RawText:                        "           20           ARS PAYMENT                                             234.567,89-",
				RawOriginalDate:                "20",
				RawReceiptNumber:               "",
				RawDetailWithMaybeInstallments: "ARS PAYMENT",
				RawAmountARS:                   "234.567,89-",
				RawAmountUSD:                   "",
			},
		},
		{
			name:      "row with long detail and embedded numbers",
			rawText:   "24 Diciem. 25 222222 *  MERCHANT*PAYMENT         99999999999                   1.234,56",
			positions: standardPositions,
			expectedRow: pdftable.Row{
				RawText:                        "24 Diciem. 25 222222 *  MERCHANT*PAYMENT         99999999999                   1.234,56",
				RawOriginalDate:                "24 Diciem. 25",
				RawReceiptNumber:               "222222 *",
				RawDetailWithMaybeInstallments: "MERCHANT*PAYMENT         99999999999",
				RawAmountARS:                   "1.234,56",
				RawAmountUSD:                   "",
			},
		},
		{
			name:      "row with special characters in detail",
			rawText:   "24 Diciem. 16           MERCHANT*2PRODUCTS                                        23.456,78-",
			positions: standardPositions,
			expectedRow: pdftable.Row{
				RawText:                        "24 Diciem. 16           MERCHANT*2PRODUCTS                                        23.456,78-",
				RawOriginalDate:                "24 Diciem. 16",
				RawReceiptNumber:               "",
				RawDetailWithMaybeInstallments: "MERCHANT*2PRODUCTS",
				RawAmountARS:                   "23.456,78-",
				RawAmountUSD:                   "",
			},
		},
		{
			name:      "date cut short should still include the date",
			rawText:   "25-08",
			positions: standardPositions,
			expectedRow: pdftable.Row{
				RawText:                        "25-08",
				RawOriginalDate:                "25-08",
				RawReceiptNumber:               "",
				RawDetailWithMaybeInstallments: "",
				RawAmountARS:                   "",
				RawAmountUSD:                   "",
			},
		},
		{
			name:      "just enough for first column",
			rawText:   "1234567890123", // 13 chars, may cover OriginalDateStart/End only
			positions: standardPositions,
			expectedRow: pdftable.Row{
				RawText:                        "1234567890123",
				RawOriginalDate:                "1234567890123",
				RawReceiptNumber:               "",
				RawDetailWithMaybeInstallments: "",
				RawAmountARS:                   "",
				RawAmountUSD:                   "",
			},
		},
		{
			name:      "just enough for first two columns",
			rawText:   "1234567890123 123456789",
			positions: standardPositions,
			expectedRow: pdftable.Row{
				RawText:                        "1234567890123 123456789",
				RawOriginalDate:                "1234567890123",
				RawReceiptNumber:               "123456789",
				RawDetailWithMaybeInstallments: "",
				RawAmountARS:                   "",
				RawAmountUSD:                   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			// Test data is already set up in the test case

			// When
			actualRow := factory.CreateRow(tt.rawText)

			// Then
			require.Equal(t, tt.expectedRow, actualRow)
		})
	}
}
