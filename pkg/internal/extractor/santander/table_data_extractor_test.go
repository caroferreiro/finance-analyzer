package santander

import (
	_ "embed"
	"testing"
	"time"

	"github.com/Alechan/finance-analyzer/pkg/internal/extractor/pdftable"
	"github.com/Alechan/finance-analyzer/pkg/internal/extractor/santander/testdata"
	"github.com/Alechan/finance-analyzer/pkg/internal/pdfcardsummary"
	"github.com/Alechan/finance-analyzer/pkg/internal/platform/testsale"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestSantanderTableExtractor_WithRealisticStatement(t *testing.T) {
	cfg := DefaultConfig()
	extractor := NewSantanderTableExtractor(cfg)

	// Given: RealisticStatementRows (from I/O tests)
	iterator := pdftable.NewFakeTableIterator(testdata.RealisticStatementRows)

	// When: Extract business data
	got, err := extractor.Extract(iterator)

	// Then: Compare with expected SantanderTableData
	require.NoError(t, err)

	require.Equal(t, testdata.ExpectedSantanderTableData(t), got)
}

func TestExtractCardsAndTaxesMovs_BrokenLineCase(t *testing.T) {
	cfg := DefaultConfig()
	extractor := NewSantanderTableExtractor(cfg)

	testCases := []struct {
		name          string
		rows          []pdftable.Row
		expectedCards []pdfcardsummary.Card
		expectedTaxes []pdfcardsummary.Movement
		expectedError error
	}{
		{
			name: "movement with zero amounts followed by amount-only row - merges correctly",
			rows: []pdftable.Row{
				// Movement with zero amounts (broken line)
				// Date format: "YY Month DD" where "22 Diciem. 01" = 2022-12-01
				pdftable.NewRow("22 Diciem. 01 111111 *  TEST MERCHANT                                                                           ", "22 Diciem. 01", "111111 *", "TEST MERCHANT", "", ""),
				// Amount-only row
				pdftable.NewRow("                                                                    1.316,68", "", "", "", "1.316,68", ""),
				// Card total - regex captures "1.316,68" and "0,00" (without asterisks)
				pdftable.NewRow("Tarjeta 1234 Total Consumos de TEST USER                                                                    1.316,68 *            0,00 *", "Tarjeta 1234", "otal Cons", "mos de TEST USER", "1.316,68", "0,00"),
			},
			expectedCards: []pdfcardsummary.Card{
				{
					CardContext: pdfcardsummary.CardContext{
						CardNumber:   testsale.StrPtr("1234"),
						CardOwner:    "TEST USER",
						CardTotalARS: testsale.AsDecimal(t, "1316.68"),
						CardTotalUSD: decimal.Zero,
					},
					Movements: []pdfcardsummary.Movement{
						{
							OriginalDate:       testsale.DatePtr(2022, time.December, 1),
							ReceiptNumber:      testsale.StrPtr("111111 *"),
							Detail:             "TEST MERCHANT",
							CurrentInstallment: nil,
							TotalInstallments:  nil,
							AmountARS:          testsale.AsDecimal(t, "1316.68"),
							AmountUSD:          decimal.Zero,
						},
					},
				},
			},
			expectedTaxes: nil,
			expectedError: nil,
		},
		{
			name: "movement with zero amounts followed by regular movement - adds zero-amount movement",
			rows: []pdftable.Row{
				// Movement with zero amounts (broken line, but no amount-only row follows)
				pdftable.NewRow("22 Diciem. 01 111111 *  TEST MERCHANT                                                                           ", "22 Diciem. 01", "111111 *", "TEST MERCHANT", "", ""),
				// Regular movement
				pdftable.NewRow("22 Diciem. 02 123456 *  OTHER MOVEMENT                                                                    500,00", "22 Diciem. 02", "123456 *", "OTHER MOVEMENT", "500,00", ""),
				// Card total
				pdftable.NewRow("Tarjeta 1234 Total Consumos de TEST USER                                                                      500,00 *            0,00 *", "Tarjeta 1234", "otal Cons", "mos de TEST USER", "500,00", "0,00"),
			},
			expectedCards: []pdfcardsummary.Card{
				{
					CardContext: pdfcardsummary.CardContext{
						CardNumber:   testsale.StrPtr("1234"),
						CardOwner:    "TEST USER",
						CardTotalARS: testsale.AsDecimal(t, "500.00"),
						CardTotalUSD: decimal.Zero,
					},
					Movements: []pdfcardsummary.Movement{
						{
							OriginalDate:       testsale.DatePtr(2022, time.December, 1),
							ReceiptNumber:      testsale.StrPtr("111111 *"),
							Detail:             "TEST MERCHANT",
							CurrentInstallment: nil,
							TotalInstallments:  nil,
							AmountARS:          decimal.Zero,
							AmountUSD:          decimal.Zero,
						},
						{
							OriginalDate:       testsale.DatePtr(2022, time.December, 2),
							ReceiptNumber:      testsale.StrPtr("123456 *"),
							Detail:             "OTHER MOVEMENT",
							CurrentInstallment: nil,
							TotalInstallments:  nil,
							AmountARS:          testsale.AsDecimal(t, "500.00"),
							AmountUSD:          decimal.Zero,
						},
					},
				},
			},
			expectedTaxes: nil,
			expectedError: nil,
		},
		{
			name: "multiple movements with broken line in the middle",
			rows: []pdftable.Row{
				// Regular movement
				pdftable.NewRow("22 Diciem. 01 123456 *  FIRST MOVEMENT                                                                    100,00", "22 Diciem. 01", "123456 *", "FIRST MOVEMENT", "100,00", ""),
				// Movement with zero amounts (broken line)
				pdftable.NewRow("22 Diciem. 02 111111 *  TEST MERCHANT                                                                           ", "22 Diciem. 02", "111111 *", "TEST MERCHANT", "", ""),
				// Amount-only row
				pdftable.NewRow("                                                                    1.316,68", "", "", "", "1.316,68", ""),
				// Regular movement
				pdftable.NewRow("22 Diciem. 03 789012 *  THIRD MOVEMENT                                                                    200,00", "22 Diciem. 03", "789012 *", "THIRD MOVEMENT", "200,00", ""),
				// Card total
				pdftable.NewRow("Tarjeta 1234 Total Consumos de TEST USER                                                                  1.616,68 *            0,00 *", "Tarjeta 1234", "otal Cons", "mos de TEST USER", "1.616,68", "0,00"),
			},
			expectedCards: []pdfcardsummary.Card{
				{
					CardContext: pdfcardsummary.CardContext{
						CardNumber:   testsale.StrPtr("1234"),
						CardOwner:    "TEST USER",
						CardTotalARS: testsale.AsDecimal(t, "1616.68"),
						CardTotalUSD: decimal.Zero,
					},
					Movements: []pdfcardsummary.Movement{
						{
							OriginalDate:       testsale.DatePtr(2022, time.December, 1),
							ReceiptNumber:      testsale.StrPtr("123456 *"),
							Detail:             "FIRST MOVEMENT",
							CurrentInstallment: nil,
							TotalInstallments:  nil,
							AmountARS:          testsale.AsDecimal(t, "100.00"),
							AmountUSD:          decimal.Zero,
						},
						{
							OriginalDate:       testsale.DatePtr(2022, time.December, 2),
							ReceiptNumber:      testsale.StrPtr("111111 *"),
							Detail:             "TEST MERCHANT",
							CurrentInstallment: nil,
							TotalInstallments:  nil,
							AmountARS:          testsale.AsDecimal(t, "1316.68"),
							AmountUSD:          decimal.Zero,
						},
						{
							OriginalDate:       testsale.DatePtr(2022, time.December, 3),
							ReceiptNumber:      testsale.StrPtr("789012 *"),
							Detail:             "THIRD MOVEMENT",
							CurrentInstallment: nil,
							TotalInstallments:  nil,
							AmountARS:          testsale.AsDecimal(t, "200.00"),
							AmountUSD:          decimal.Zero,
						},
					},
				},
			},
			expectedTaxes: nil,
			expectedError: nil,
		},
		{
			name: "movement with zero amounts at end of card - adds before card total",
			rows: []pdftable.Row{
				// Regular movement
				pdftable.NewRow("22 Diciem. 01 123456 *  FIRST MOVEMENT                                                                    100,00", "22 Diciem. 01", "123456 *", "FIRST MOVEMENT", "100,00", ""),
				// Movement with zero amounts (broken line)
				pdftable.NewRow("22 Diciem. 02 111111 *  TEST MERCHANT                                                                           ", "22 Diciem. 02", "111111 *", "TEST MERCHANT", "", ""),
				// Card total (no amount-only row, so zero-amount movement is added before card total)
				pdftable.NewRow("Tarjeta 1234 Total Consumos de TEST USER                                                                    100,00 *            0,00 *", "Tarjeta 1234", "otal Cons", "mos de TEST USER", "100,00", "0,00"),
			},
			expectedCards: []pdfcardsummary.Card{
				{
					CardContext: pdfcardsummary.CardContext{
						CardNumber:   testsale.StrPtr("1234"),
						CardOwner:    "TEST USER",
						CardTotalARS: testsale.AsDecimal(t, "100.00"),
						CardTotalUSD: decimal.Zero,
					},
					Movements: []pdfcardsummary.Movement{
						{
							OriginalDate:       testsale.DatePtr(2022, time.December, 1),
							ReceiptNumber:      testsale.StrPtr("123456 *"),
							Detail:             "FIRST MOVEMENT",
							CurrentInstallment: nil,
							TotalInstallments:  nil,
							AmountARS:          testsale.AsDecimal(t, "100.00"),
							AmountUSD:          decimal.Zero,
						},
						{
							OriginalDate:       testsale.DatePtr(2022, time.December, 2),
							ReceiptNumber:      testsale.StrPtr("111111 *"),
							Detail:             "TEST MERCHANT",
							CurrentInstallment: nil,
							TotalInstallments:  nil,
							AmountARS:          decimal.Zero,
							AmountUSD:          decimal.Zero,
						},
					},
				},
			},
			expectedTaxes: nil,
			expectedError: nil,
		},
		{
			name: "normal movements without broken lines - works as before",
			rows: []pdftable.Row{
				// Regular movement
				pdftable.NewRow("22 Diciem. 01 123456 *  FIRST MOVEMENT                                                                    100,00", "22 Diciem. 01", "123456 *", "FIRST MOVEMENT", "100,00", ""),
				// Regular movement
				pdftable.NewRow("22 Diciem. 02 789012 *  SECOND MOVEMENT                                                                  200,00", "22 Diciem. 02", "789012 *", "SECOND MOVEMENT", "200,00", ""),
				// Card total
				pdftable.NewRow("Tarjeta 1234 Total Consumos de TEST USER                                                                    300,00 *            0,00 *", "Tarjeta 1234", "otal Cons", "mos de TEST USER", "300,00", "0,00"),
			},
			expectedCards: []pdfcardsummary.Card{
				{
					CardContext: pdfcardsummary.CardContext{
						CardNumber:   testsale.StrPtr("1234"),
						CardOwner:    "TEST USER",
						CardTotalARS: testsale.AsDecimal(t, "300.00"),
						CardTotalUSD: decimal.Zero,
					},
					Movements: []pdfcardsummary.Movement{
						{
							OriginalDate:       testsale.DatePtr(2022, time.December, 1),
							ReceiptNumber:      testsale.StrPtr("123456 *"),
							Detail:             "FIRST MOVEMENT",
							CurrentInstallment: nil,
							TotalInstallments:  nil,
							AmountARS:          testsale.AsDecimal(t, "100.00"),
							AmountUSD:          decimal.Zero,
						},
						{
							OriginalDate:       testsale.DatePtr(2022, time.December, 2),
							ReceiptNumber:      testsale.StrPtr("789012 *"),
							Detail:             "SECOND MOVEMENT",
							CurrentInstallment: nil,
							TotalInstallments:  nil,
							AmountARS:          testsale.AsDecimal(t, "200.00"),
							AmountUSD:          decimal.Zero,
						},
					},
				},
			},
			expectedTaxes: nil,
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			iterator := pdftable.NewFakeTableIterator(tc.rows)

			// When
			actualCards, actualTaxes, actualError := extractor.extractCardsAndTaxesMovs(iterator)

			// Then
			require.Equal(t, tc.expectedError, actualError)
			require.Equal(t, tc.expectedCards, actualCards)
			require.Equal(t, tc.expectedTaxes, actualTaxes)
		})
	}
}
