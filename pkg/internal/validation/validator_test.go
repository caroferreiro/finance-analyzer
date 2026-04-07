package validation_test

import (
	"testing"
	"time"

	"github.com/Alechan/finance-analyzer/pkg/internal/pdfcardsummary"
	"github.com/Alechan/finance-analyzer/pkg/internal/platform/testsale"
	"github.com/Alechan/finance-analyzer/pkg/internal/validation"
	"github.com/Alechan/finance-analyzer/pkg/internal/validation/testdata"
	"github.com/stretchr/testify/require"
)

func TestValidator_Validate_HappyPath(t *testing.T) {
	// Given
	validCardSummary := testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
		b.WithCloseDate(2024, time.August, 15)
		b.WithExpirationDate(2024, time.August, 23)
		b.WithTotalARS("2000.00")
		b.WithSaldoAnterior("1000.00", "0.00")
		b.WithCard("1234", "OWNER", "1000.00", "0.00")
		b.WithCardMovement(0, testsale.DatePtr(2024, time.August, 10), "123456*", "DETAIL", "1000.00", "0.00")
	})

	validator := validation.NewValidator()

	// When
	err := validator.Validate(validCardSummary)

	// Then
	require.NoError(t, err)
}

func TestValidator_Validate_MultipleErrors(t *testing.T) {
	testCases := []struct {
		name          string
		cardSummary   pdfcardsummary.CardSummary
		expectedError string // Partial error message to check
	}{
		{
			name: "multiple errors - returns first error",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithCloseDate(2024, time.August, 23)      // After expiration
				b.WithExpirationDate(2024, time.August, 15) // Before close
				b.WithTotalARS("0.00")                      // Zero but cards exist
				// No cards
				// No SALDO ANTERIOR
			}),
			expectedError: "date validation failed",
		},
		{
			name: "no cards error",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithCloseDate(2024, time.August, 15)
				b.WithExpirationDate(2024, time.August, 23)
				// No cards
			}),
			expectedError: "structure validation failed",
		},
		{
			name: "missing SALDO ANTERIOR",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithCloseDate(2024, time.August, 15)
				b.WithExpirationDate(2024, time.August, 23)
				b.WithCard("1234", "OWNER", "1000.00", "0.00")
				// No SALDO ANTERIOR
			}),
			expectedError: "structure validation failed",
		},
		{
			name: "year validation catches parsing error",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithCloseDate(2024, time.August, 15)
				b.WithExpirationDate(2024, time.August, 23)
				b.WithSaldoAnterior("1000.00", "0.00")
				b.WithCard("1234", "OWNER", "1000.00", "0.00")
				// This simulates the bug where year "24" was parsed as year 24 instead of 2024
				b.WithCardMovement(0, testsale.DatePtr(24, time.August, 15), "123456*", "DETAIL", "1000.00", "0.00")
			}),
			expectedError: "date range validation failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			validator := validation.NewValidator()

			// When
			err := validator.Validate(tc.cardSummary)

			// Then
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

func TestValidator_Validate_YearParsingBug(t *testing.T) {
	// Given
	// This test specifically validates that we catch the bug where year "24" was parsed as year 24
	cardSummary := testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
		b.WithCloseDate(2024, time.August, 15)
		b.WithExpirationDate(2024, time.August, 23)
		b.WithSaldoAnterior("1000.00", "0.00")
		b.WithCard("1234", "OWNER", "1000.00", "0.00")
		// Simulate the bug: year "24" was incorrectly parsed as year 24 (instead of 2024)
		b.WithCardMovement(0, testsale.DatePtr(24, time.August, 15), "123456*", "DETAIL", "1000.00", "0.00")
	})

	validator := validation.NewValidator()

	// When
	err := validator.Validate(cardSummary)

	// Then
	require.Error(t, err)
	require.Contains(t, err.Error(), "date 0024-08-15 has year 24 which is before minimum valid year 2000")
	require.Contains(t, err.Error(), "date range validation failed")
}
