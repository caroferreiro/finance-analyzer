package validation_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/Alechan/finance-analyzer/pkg/internal/pdfcardsummary"
	"github.com/Alechan/finance-analyzer/pkg/internal/platform/testsale"
	"github.com/Alechan/finance-analyzer/pkg/internal/validation"
	"github.com/Alechan/finance-analyzer/pkg/internal/validation/testdata"
	"github.com/stretchr/testify/require"
)

func TestValidateAtLeastOneCardRequired(t *testing.T) {
	testCases := []struct {
		name          string
		cardSummary   pdfcardsummary.CardSummary
		expectedError error
	}{
		{
			name: "valid - has cards",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithCard("1234", "OWNER", "1000.00", "0.00")
			}),
			expectedError: nil,
		},
		{
			name: "valid - has multiple cards",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithCard("1234", "OWNER1", "1000.00", "0.00")
				b.WithCard("5678", "OWNER2", "2000.00", "0.00")
			}),
			expectedError: nil,
		},
		{
			name: "invalid - no cards",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				// No cards
			}),
			expectedError: fmt.Errorf("no cards found in statement"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			// CardSummary is set up in test case

			// When
			err := validation.ValidateAtLeastOneCardRequired(tc.cardSummary)

			// Then
			require.Equal(t, tc.expectedError, err)
		})
	}
}

func TestValidateSaldoAnteriorPresence(t *testing.T) {
	testCases := []struct {
		name          string
		cardSummary   pdfcardsummary.CardSummary
		expectedError error
	}{
		{
			name: "valid - SALDO ANTERIOR is first",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithSaldoAnterior("1000.00", "0.00")
			}),
			expectedError: nil,
		},
		{
			name: "valid - SALDO ANTERIOR is first with other movements",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithSaldoAnterior("1000.00", "0.00")
				b.WithPastPaymentMovement(testsale.DatePtr(2024, time.August, 10), "123456*", "PAYMENT", "-500.00", "0.00")
			}),
			expectedError: nil,
		},
		{
			name: "invalid - no past payment movements",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				// No past payment movements
			}),
			expectedError: fmt.Errorf("no past payment movements found"),
		},
		{
			name: "invalid - first movement is not SALDO ANTERIOR",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithPastPaymentMovement(testsale.DatePtr(2024, time.August, 10), "123456*", "OTHER MOVEMENT", "1000.00", "0.00")
			}),
			expectedError: fmt.Errorf("first past payment movement is not SALDO ANTERIOR, got: OTHER MOVEMENT"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			// CardSummary is set up in test case

			// When
			err := validation.ValidateSaldoAnteriorPresence(tc.cardSummary)

			// Then
			require.Equal(t, tc.expectedError, err)
		})
	}
}

func TestValidateCardOwnerNonEmpty(t *testing.T) {
	testCases := []struct {
		name          string
		cardSummary   pdfcardsummary.CardSummary
		expectedError error
	}{
		{
			name: "valid - all cards have owners",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithCard("1234", "OWNER1", "1000.00", "0.00")
				b.WithCard("5678", "OWNER2", "2000.00", "0.00")
			}),
			expectedError: nil,
		},
		{
			name: "invalid - card has empty owner",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithCard("1234", "", "1000.00", "0.00")
			}),
			expectedError: fmt.Errorf("card 0 (number: 1234) has empty owner"),
		},
		{
			name: "invalid - card has whitespace-only owner",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithCard("1234", "   ", "1000.00", "0.00")
			}),
			expectedError: fmt.Errorf("card 0 (number: 1234) has empty owner"),
		},
		{
			name: "invalid - second card has empty owner",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithCard("1234", "OWNER1", "1000.00", "0.00")
				b.WithCard("5678", "", "2000.00", "0.00")
			}),
			expectedError: fmt.Errorf("card 1 (number: 5678) has empty owner"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			// CardSummary is set up in test case

			// When
			err := validation.ValidateCardOwnerNonEmpty(tc.cardSummary)

			// Then
			require.Equal(t, tc.expectedError, err)
		})
	}
}

func TestValidateCardMovementsHaveReceiptNumbers(t *testing.T) {
	testCases := []struct {
		name          string
		cardSummary   pdfcardsummary.CardSummary
		expectedError error
	}{
		{
			name: "valid - all card movements have receipt numbers",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithSaldoAnterior("0", "0")
				b.WithCard("1234", "OWNER", "1000.00", "0.00")
				b.WithCardMovement(0, testsale.DatePtr(2024, time.August, 10), "123456*", "DETAIL", "1000.00", "0.00")
			}),
			expectedError: nil,
		},
		{
			name: "valid - multiple cards with movements",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithSaldoAnterior("0", "0")
				b.WithCard("1234", "OWNER1", "1000.00", "0.00")
				b.WithCardMovement(0, testsale.DatePtr(2024, time.August, 10), "123456*", "DETAIL1", "1000.00", "0.00")
				b.WithCard("5678", "OWNER2", "2000.00", "0.00")
				b.WithCardMovement(1, testsale.DatePtr(2024, time.August, 11), "123457*", "DETAIL2", "2000.00", "0.00")
			}),
			expectedError: nil,
		},
		{
			name: "invalid - card movement missing receipt number",
			cardSummary: func() pdfcardsummary.CardSummary {
				builder := testdata.NewCardSummaryBuilder(t)
				builder.WithSaldoAnterior("0", "0")
				builder.WithCard("1234", "OWNER", "1000.00", "0.00")
				mov := pdfcardsummary.Movement{
					OriginalDate:  testsale.DatePtr(2024, time.August, 10),
					ReceiptNumber: nil,
					Detail:        "DETAIL",
					AmountARS:     testsale.AsDecimal(t, "1000.00"),
					AmountUSD:     testsale.AsDecimal(t, "0.00"),
				}
				cs := builder.Build()
				cs.Table.Cards[0].Movements = append(cs.Table.Cards[0].Movements, mov)
				return cs
			}(),
			expectedError: fmt.Errorf(`card 0 (owner: "OWNER", number: 1234, movements: 1) movement 0 (date: 2024-08-10, detail: "DETAIL", receipt: <nil>) does not have a receipt number`),
		},
		{
			name: "invalid - second movement missing receipt number",
			cardSummary: func() pdfcardsummary.CardSummary {
				builder := testdata.NewCardSummaryBuilder(t)
				builder.WithSaldoAnterior("0", "0")
				builder.WithCard("1234", "OWNER", "2000.00", "0.00")
				builder.WithCardMovement(0, testsale.DatePtr(2024, time.August, 10), "123456*", "DETAIL1", "1000.00", "0.00")
				mov := pdfcardsummary.Movement{
					OriginalDate:  testsale.DatePtr(2024, time.August, 11),
					ReceiptNumber: nil,
					Detail:        "DETAIL2",
					AmountARS:     testsale.AsDecimal(t, "1000.00"),
					AmountUSD:     testsale.AsDecimal(t, "0.00"),
				}
				cs := builder.Build()
				cs.Table.Cards[0].Movements = append(cs.Table.Cards[0].Movements, mov)
				return cs
			}(),
			expectedError: fmt.Errorf(`card 0 (owner: "OWNER", number: 1234, movements: 2) movement 1 (date: 2024-08-11, detail: "DETAIL2", receipt: <nil>) does not have a receipt number`),
		},
		{
			name: "valid - card with no movements",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithSaldoAnterior("0", "0")
				b.WithCard("1234", "OWNER", "0.00", "0.00")
			}),
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			// CardSummary is set up in test case

			// When
			err := validation.ValidateCardMovementsHaveReceiptNumbers(tc.cardSummary)

			// Then
			require.Equal(t, tc.expectedError, err)
		})
	}
}

func TestValidateTaxMovementsHaveNoReceiptNumbers(t *testing.T) {
	testCases := []struct {
		name          string
		cardSummary   pdfcardsummary.CardSummary
		expectedError error
	}{
		{
			name: "valid - tax movements have no receipt numbers",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithSaldoAnterior("0", "0")
				b.WithCard("1234", "OWNER", "1000.00", "0.00")
				b.WithTaxMovement(testsale.DatePtr(2024, time.August, 10), "TAX DETAIL", "100.00", "0.00")
			}),
			expectedError: nil,
		},
		{
			name: "valid - multiple tax movements without receipt numbers",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithSaldoAnterior("0", "0")
				b.WithCard("1234", "OWNER", "1000.00", "0.00")
				b.WithTaxMovement(testsale.DatePtr(2024, time.August, 10), "TAX1", "100.00", "0.00")
				b.WithTaxMovement(testsale.DatePtr(2024, time.August, 11), "TAX2", "200.00", "0.00")
			}),
			expectedError: nil,
		},
		{
			name: "valid - no tax movements",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithSaldoAnterior("0", "0")
				b.WithCard("1234", "OWNER", "1000.00", "0.00")
			}),
			expectedError: nil,
		},
		{
			name: "invalid - tax movement has receipt number",
			cardSummary: func() pdfcardsummary.CardSummary {
				builder := testdata.NewCardSummaryBuilder(t)
				builder.WithSaldoAnterior("0", "0")
				builder.WithCard("1234", "OWNER", "1000.00", "0.00")
				mov := pdfcardsummary.Movement{
					OriginalDate:  testsale.DatePtr(2024, time.August, 10),
					ReceiptNumber: testsale.StrPtr("123456*"),
					Detail:        "TAX DETAIL",
					AmountARS:     testsale.AsDecimal(t, "100.00"),
					AmountUSD:     testsale.AsDecimal(t, "0.00"),
				}
				cs := builder.Build()
				cs.Table.TaxesMovements = append(cs.Table.TaxesMovements, mov)
				return cs
			}(),
			expectedError: fmt.Errorf(`tax movement 0 (date: 2024-08-10, detail: "TAX DETAIL", receipt: 123456*) should not have a receipt number, got: 123456*`),
		},
		{
			name: "invalid - second tax movement has receipt number",
			cardSummary: func() pdfcardsummary.CardSummary {
				builder := testdata.NewCardSummaryBuilder(t)
				builder.WithSaldoAnterior("0", "0")
				builder.WithCard("1234", "OWNER", "1000.00", "0.00")
				builder.WithTaxMovement(testsale.DatePtr(2024, time.August, 10), "TAX1", "100.00", "0.00")
				mov := pdfcardsummary.Movement{
					OriginalDate:  testsale.DatePtr(2024, time.August, 11),
					ReceiptNumber: testsale.StrPtr("123457*"),
					Detail:        "TAX2",
					AmountARS:     testsale.AsDecimal(t, "200.00"),
					AmountUSD:     testsale.AsDecimal(t, "0.00"),
				}
				cs := builder.Build()
				cs.Table.TaxesMovements = append(cs.Table.TaxesMovements, mov)
				return cs
			}(),
			expectedError: fmt.Errorf(`tax movement 1 (date: 2024-08-11, detail: "TAX2", receipt: 123457*) should not have a receipt number, got: 123457*`),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			// CardSummary is set up in test case

			// When
			err := validation.ValidateTaxMovementsHaveNoReceiptNumbers(tc.cardSummary)

			// Then
			require.Equal(t, tc.expectedError, err)
		})
	}
}
