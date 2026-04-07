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

func TestValidateDocumentTotalsMatchComponents(t *testing.T) {
	testCases := []struct {
		name          string
		cardSummary   pdfcardsummary.CardSummary
		expectedError error
	}{
		{
			name: "valid - totals match components",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithTotalARS("3000.00")
				b.WithTotalUSD("0.00")
				b.WithSaldoAnterior("1000.00", "0.00")
				b.WithCard("1234", "OWNER1", "1000.00", "0.00")
				b.WithCardMovement(0, testsale.DatePtr(2024, time.August, 10), "123456*", "DETAIL1", "1000.00", "0.00")
				b.WithCard("5678", "OWNER2", "1000.00", "0.00")
				b.WithCardMovement(1, testsale.DatePtr(2024, time.August, 11), "123457*", "DETAIL2", "1000.00", "0.00")
				b.WithTaxMovement(testsale.DatePtr(2024, time.August, 12), "TAX", "0.00", "0.00")
			}),
			expectedError: nil,
		},
		{
			name: "valid - with taxes movements",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithTotalARS("3100.00")
				b.WithTotalUSD("0.00")
				b.WithSaldoAnterior("1000.00", "0.00")
				b.WithCard("1234", "OWNER", "2000.00", "0.00")
				b.WithCardMovement(0, testsale.DatePtr(2024, time.August, 10), "123456*", "DETAIL", "2000.00", "0.00")
				b.WithTaxMovement(testsale.DatePtr(2024, time.August, 12), "TAX", "100.00", "0.00")
			}),
			expectedError: nil,
		},
		{
			name: "invalid - ARS total does not match",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithTotalARS("9999.00")
				b.WithTotalUSD("0.00")
				b.WithSaldoAnterior("1000.00", "0.00")
				b.WithCard("1234", "OWNER", "1000.00", "0.00")
				b.WithCardMovement(0, testsale.DatePtr(2024, time.August, 10), "123456*", "DETAIL", "1000.00", "0.00")
			}),
			expectedError: fmt.Errorf("document total ARS %s does not match sum %s (cards: %s + movements: %s) (difference: %s)",
				testsale.AsDecimal(t, "9999.00"), testsale.AsDecimal(t, "2000.00"), testsale.AsDecimal(t, "1000.00"), testsale.AsDecimal(t, "1000.00"),
				testsale.AsDecimal(t, "9999.00").Sub(testsale.AsDecimal(t, "2000.00"))),
		},
		{
			name: "invalid - USD total does not match",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithTotalARS("0.00")
				b.WithTotalUSD("9999.00")
				b.WithSaldoAnterior("0.00", "1000.00")
				b.WithCard("1234", "OWNER", "0.00", "1000.00")
				b.WithCardMovement(0, testsale.DatePtr(2024, time.August, 10), "123456*", "DETAIL", "0.00", "1000.00")
			}),
			expectedError: fmt.Errorf("document total USD %s does not match sum %s (cards: %s + movements: %s) (difference: %s)",
				testsale.AsDecimal(t, "9999.00"), testsale.AsDecimal(t, "2000.00"), testsale.AsDecimal(t, "1000.00"), testsale.AsDecimal(t, "1000.00"),
				testsale.AsDecimal(t, "9999.00").Sub(testsale.AsDecimal(t, "2000.00"))),
		},
		{
			name: "valid - zero totals",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithTotalARS("0.00")
				b.WithTotalUSD("0.00")
				b.WithSaldoAnterior("0.00", "0.00")
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
			err := validation.ValidateDocumentTotalsMatchComponents(tc.cardSummary)

			// Then
			require.Equal(t, tc.expectedError, err)
		})
	}
}

func TestValidateCardTotalsMatchMovements(t *testing.T) {
	testCases := []struct {
		name          string
		cardSummary   pdfcardsummary.CardSummary
		expectedError error
	}{
		{
			name: "valid - card totals match movements",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithSaldoAnterior("0", "0")
				b.WithCard("1234", "OWNER", "2000.00", "0.00")
				b.WithCardMovement(0, testsale.DatePtr(2024, time.August, 10), "123456*", "DETAIL1", "1000.00", "0.00")
				b.WithCardMovement(0, testsale.DatePtr(2024, time.August, 11), "123457*", "DETAIL2", "1000.00", "0.00")
			}),
			expectedError: nil,
		},
		{
			name: "valid - multiple cards with matching totals",
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
			name: "valid - card with USD movements",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithSaldoAnterior("0", "0")
				b.WithCard("1234", "OWNER", "0.00", "1000.00")
				b.WithCardMovement(0, testsale.DatePtr(2024, time.August, 10), "123456*", "DETAIL", "0.00", "1000.00")
			}),
			expectedError: nil,
		},
		{
			name: "valid - card with mixed ARS and USD",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithSaldoAnterior("0", "0")
				b.WithCard("1234", "OWNER", "1000.00", "500.00")
				b.WithCardMovement(0, testsale.DatePtr(2024, time.August, 10), "123456*", "DETAIL1", "1000.00", "0.00")
				b.WithCardMovement(0, testsale.DatePtr(2024, time.August, 11), "123457*", "DETAIL2", "0.00", "500.00")
			}),
			expectedError: nil,
		},
		{
			name: "valid - card with no movements",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithSaldoAnterior("0", "0")
				b.WithCard("1234", "OWNER", "0.00", "0.00")
			}),
			expectedError: nil,
		},
		{
			name: "invalid - card ARS total does not match",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithSaldoAnterior("0", "0")
				b.WithCard("1234", "OWNER", "9999.00", "0.00")
				b.WithCardMovement(0, testsale.DatePtr(2024, time.August, 10), "123456*", "DETAIL", "1000.00", "0.00")
			}),
			expectedError: fmt.Errorf("card 0 (owner: %q, number: %s, movements: %d) total ARS %s does not match sum of movements %s (difference: %s)",
				"OWNER", "1234", 1,
				testsale.AsDecimal(t, "9999.00"), testsale.AsDecimal(t, "1000.00"),
				testsale.AsDecimal(t, "9999.00").Sub(testsale.AsDecimal(t, "1000.00"))),
		},
		{
			name: "invalid - card USD total does not match",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithSaldoAnterior("0", "0")
				b.WithCard("1234", "OWNER", "0.00", "9999.00")
				b.WithCardMovement(0, testsale.DatePtr(2024, time.August, 10), "123456*", "DETAIL", "0.00", "1000.00")
			}),
			expectedError: fmt.Errorf("card 0 (owner: %q, number: %s, movements: %d) total USD %s does not match sum of movements %s (difference: %s)",
				"OWNER", "1234", 1,
				testsale.AsDecimal(t, "9999.00"), testsale.AsDecimal(t, "1000.00"),
				testsale.AsDecimal(t, "9999.00").Sub(testsale.AsDecimal(t, "1000.00"))),
		},
		{
			name: "invalid - second card total does not match",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithSaldoAnterior("0", "0")
				b.WithCard("1234", "OWNER1", "1000.00", "0.00")
				b.WithCardMovement(0, testsale.DatePtr(2024, time.August, 10), "123456*", "DETAIL1", "1000.00", "0.00")
				b.WithCard("5678", "OWNER2", "9999.00", "0.00")
				b.WithCardMovement(1, testsale.DatePtr(2024, time.August, 11), "123457*", "DETAIL2", "2000.00", "0.00")
			}),
			expectedError: fmt.Errorf("card 1 (owner: %q, number: %s, movements: %d) total ARS %s does not match sum of movements %s (difference: %s)",
				"OWNER2", "5678", 1,
				testsale.AsDecimal(t, "9999.00"), testsale.AsDecimal(t, "2000.00"),
				testsale.AsDecimal(t, "9999.00").Sub(testsale.AsDecimal(t, "2000.00"))),
		},
		{
			name: "invalid - small difference without anomaly adjustment",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithSaldoAnterior("0", "0")
				b.WithCard("1234", "OWNER", "1000.68", "0.00")
				b.WithCardMovement(0, testsale.DatePtr(2024, time.August, 10), "123456*", "DETAIL", "1000.00", "0.00")
			}),
			expectedError: fmt.Errorf("card 0 (owner: %q, number: %s, movements: %d) total ARS %s does not match sum of movements %s (difference: %s)",
				"OWNER", "1234", 1,
				testsale.AsDecimal(t, "1000.68"), testsale.AsDecimal(t, "1000.00"),
				testsale.AsDecimal(t, "1000.68").Sub(testsale.AsDecimal(t, "1000.00"))),
		},
		{
			name: "valid - small difference WITH anomaly adjustment",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithSaldoAnterior("0", "0")
				b.WithCard("1234", "OWNER", "1000.68", "0.00")
				b.WithCardMovement(0, testsale.DatePtr(2024, time.August, 10), "123456*", "DETAIL", "1000.00", "0.00")
				b.WithAnomalyAdjustmentMovement(0, "0.68", "0.00")
			}),
			expectedError: nil,
		},
		{
			name: "valid - USD difference WITH anomaly adjustment",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithSaldoAnterior("0", "0")
				b.WithCard("1234", "OWNER", "0.00", "100.01")
				b.WithCardMovement(0, testsale.DatePtr(2024, time.August, 10), "123456*", "DETAIL", "0.00", "100.00")
				b.WithAnomalyAdjustmentMovement(0, "0.00", "0.01")
			}),
			expectedError: nil,
		},
		{
			name: "invalid - multiple cards, only one has mismatch without anomaly",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithSaldoAnterior("0", "0")
				b.WithCard("1234", "OWNER1", "1000.00", "0.00")
				b.WithCardMovement(0, testsale.DatePtr(2024, time.August, 10), "123456*", "DETAIL1", "1000.00", "0.00")
				b.WithCard("5678", "OWNER2", "1000.68", "0.00")
				b.WithCardMovement(1, testsale.DatePtr(2024, time.August, 11), "123457*", "DETAIL2", "1000.00", "0.00")
			}),
			expectedError: fmt.Errorf("card 1 (owner: %q, number: %s, movements: %d) total ARS %s does not match sum of movements %s (difference: %s)",
				"OWNER2", "5678", 1,
				testsale.AsDecimal(t, "1000.68"), testsale.AsDecimal(t, "1000.00"),
				testsale.AsDecimal(t, "1000.68").Sub(testsale.AsDecimal(t, "1000.00"))),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			// CardSummary is set up in test case

			// When
			err := validation.ValidateCardTotalsMatchMovements(tc.cardSummary)

			// Then
			require.Equal(t, tc.expectedError, err)
		})
	}
}

func TestValidateMovementsHaveAmounts(t *testing.T) {
	testCases := []struct {
		name          string
		cardSummary   pdfcardsummary.CardSummary
		expectedError error
	}{
		{
			name: "valid - card movement with ARS amount only",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithSaldoAnterior("0", "0")
				b.WithCard("1234", "OWNER", "1000.00", "0.00")
				b.WithCardMovement(0, testsale.DatePtr(2024, time.August, 10), "123456*", "DETAIL", "1000.00", "0.00")
			}),
			expectedError: nil,
		},
		{
			name: "valid - card movement with USD amount only",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithSaldoAnterior("0", "0")
				b.WithCard("1234", "OWNER", "0.00", "1000.00")
				b.WithCardMovement(0, testsale.DatePtr(2024, time.August, 10), "123456*", "DETAIL", "0.00", "1000.00")
			}),
			expectedError: nil,
		},
		{
			name: "valid - card movement with both ARS and USD amounts",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithSaldoAnterior("0", "0")
				b.WithCard("1234", "OWNER", "1000.00", "500.00")
				b.WithCardMovement(0, testsale.DatePtr(2024, time.August, 10), "123456*", "DETAIL", "1000.00", "500.00")
			}),
			expectedError: nil,
		},
		{
			name: "valid - SALDO ANTERIOR with zero amounts (exception)",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithSaldoAnterior("0.00", "0.00")
				b.WithCard("1234", "OWNER", "0.00", "0.00")
			}),
			expectedError: nil,
		},
		{
			name: "invalid - card movement with both amounts zero",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithSaldoAnterior("0", "0")
				b.WithCard("1234", "OWNER", "0.00", "0.00")
				b.WithCardMovement(0, testsale.DatePtr(2024, time.August, 10), "123456*", "DETAIL", "0.00", "0.00")
			}),
			expectedError: fmt.Errorf(`card 0 (owner: "OWNER", number: 1234, movements: 1) movement 0 (date: 2024-08-10, detail: "DETAIL", receipt: 123456*) has both ARS and USD amounts zero`),
		},
		{
			name: "invalid - past payment movement (non-SALDO) with both amounts zero",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithSaldoAnterior("1000.00", "0.00")
				b.WithPastPaymentMovement(testsale.DatePtr(2024, time.August, 5), "PAYMENT", "SU PAGO", "0.00", "0.00")
				b.WithCard("1234", "OWNER", "0.00", "0.00")
			}),
			expectedError: fmt.Errorf(`past payment movement 1 (date: 2024-08-05, detail: "SU PAGO", receipt: PAYMENT) has both ARS and USD amounts zero`),
		},
		{
			name: "invalid - tax movement with both amounts zero",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithSaldoAnterior("0", "0")
				b.WithCard("1234", "OWNER", "0.00", "0.00")
				b.WithTaxMovement(testsale.DatePtr(2024, time.August, 12), "TAX", "0.00", "0.00")
			}),
			expectedError: fmt.Errorf(`tax movement 0 (date: 2024-08-12, detail: "TAX", receipt: <nil>) has both ARS and USD amounts zero`),
		},
		{
			name: "invalid - second card movement with both amounts zero",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithSaldoAnterior("0", "0")
				b.WithCard("1234", "OWNER1", "1000.00", "0.00")
				b.WithCardMovement(0, testsale.DatePtr(2024, time.August, 10), "123456*", "DETAIL1", "1000.00", "0.00")
				b.WithCard("5678", "OWNER2", "0.00", "0.00")
				b.WithCardMovement(1, testsale.DatePtr(2024, time.August, 11), "123457*", "DETAIL2", "0.00", "0.00")
			}),
			expectedError: fmt.Errorf(`card 1 (owner: "OWNER2", number: 5678, movements: 1) movement 0 (date: 2024-08-11, detail: "DETAIL2", receipt: 123457*) has both ARS and USD amounts zero`),
		},
		{
			name: "valid - multiple movements with amounts",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithSaldoAnterior("1000.00", "0.00")
				b.WithCard("1234", "OWNER", "2000.00", "0.00")
				b.WithCardMovement(0, testsale.DatePtr(2024, time.August, 10), "123456*", "DETAIL1", "1000.00", "0.00")
				b.WithCardMovement(0, testsale.DatePtr(2024, time.August, 11), "123457*", "DETAIL2", "1000.00", "0.00")
				b.WithTaxMovement(testsale.DatePtr(2024, time.August, 12), "TAX", "100.00", "0.00")
			}),
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			// CardSummary is set up in test case

			// When
			err := validation.ValidateMovementsHaveAmounts(tc.cardSummary)

			// Then
			require.Equal(t, tc.expectedError, err)
		})
	}
}
