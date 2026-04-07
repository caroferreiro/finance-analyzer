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

func TestValidateCloseDateBeforeExpirationDate(t *testing.T) {
	testCases := []struct {
		name          string
		cardSummary   pdfcardsummary.CardSummary
		expectedError error
	}{
		{
			name: "valid - close date before expiration date",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithCloseDate(2024, time.August, 15)
				b.WithExpirationDate(2024, time.August, 23)
			}),
			expectedError: nil,
		},
		{
			name: "valid - same dates (edge case)",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithCloseDate(2024, time.August, 15)
				b.WithExpirationDate(2024, time.August, 15)
			}),
			expectedError: nil,
		},
		{
			name: "invalid - close date after expiration date",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithCloseDate(2024, time.August, 23)
				b.WithExpirationDate(2024, time.August, 15)
			}),
			expectedError: fmt.Errorf("close date 2024-08-23 is after expiration date 2024-08-15 (difference: 8 days)"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			// CardSummary is set up in test case

			// When
			err := validation.ValidateCloseDateBeforeExpirationDate(tc.cardSummary)

			// Then
			require.Equal(t, tc.expectedError, err)
		})
	}
}

func TestValidateMovementDatesWithinStatementPeriod(t *testing.T) {
	testCases := []struct {
		name          string
		cardSummary   pdfcardsummary.CardSummary
		expectedError error
	}{
		{
			name: "valid - all dates within reasonable range",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithCloseDate(2024, time.August, 15)
				b.WithCard("1234", "OWNER", "1000.00", "0.00")
				b.WithCardMovement(0, testsale.DatePtr(2024, time.August, 10), "123456*", "DETAIL", "1000.00", "0.00")
			}),
			expectedError: nil,
		},
		{
			name: "valid - movement date exactly 30 days after close date",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithCloseDate(2024, time.August, 15)
				b.WithCard("1234", "OWNER", "1000.00", "0.00")
				b.WithCardMovement(0, testsale.DatePtr(2024, time.September, 14), "123456*", "DETAIL", "1000.00", "0.00")
			}),
			expectedError: nil,
		},
		{
			name: "invalid - movement date too far in future",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithCloseDate(2024, time.August, 15)
				b.WithCard("1234", "OWNER", "1000.00", "0.00")
				// Movement date is 31 days after close date
				futureDate := testsale.DatePtr(2024, time.September, 15)
				b.WithCardMovement(0, futureDate, "123456*", "DETAIL", "1000.00", "0.00")
			}),
			expectedError: fmt.Errorf(`card 0 (owner: "OWNER", number: 1234, movements: 1) movement 0 (date: 2024-09-15, detail: "DETAIL", receipt: 123456*) date 2024-09-15 is more than 30 days after close date 2024-08-15`),
		},
		{
			name: "valid - movement date is nil (SALDO ANTERIOR)",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithCloseDate(2024, time.August, 15)
				b.WithSaldoAnterior("1000.00", "0.00") // nil date is valid
			}),
			expectedError: nil,
		},
		{
			name: "valid - movement date before close date",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithCloseDate(2024, time.August, 15)
				b.WithCard("1234", "OWNER", "1000.00", "0.00")
				b.WithCardMovement(0, testsale.DatePtr(2024, time.July, 10), "123456*", "DETAIL", "1000.00", "0.00")
			}),
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			// CardSummary is set up in test case

			// When
			err := validation.ValidateMovementDatesWithinStatementPeriod(tc.cardSummary)

			// Then
			require.Equal(t, tc.expectedError, err)
		})
	}
}

func TestValidateDateRangeReasonableness(t *testing.T) {
	now := time.Now()
	currentYear := now.Year()

	testCases := []struct {
		name          string
		cardSummary   pdfcardsummary.CardSummary
		expectedError error
	}{
		{
			name: "valid - current date",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithCloseDate(currentYear, time.August, 15)
			}),
			expectedError: nil,
		},
		{
			name: "invalid - close date too old (before 2000)",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithCloseDate(1999, time.January, 1)
			}),
			expectedError: fmt.Errorf("close date 1999-01-01 is before minimum valid year 2000"),
		},
		{
			name: "invalid - close date too far in future",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithCloseDate(currentYear+11, time.January, 1)
			}),
			expectedError: fmt.Errorf("close date %d-01-01 is more than 10 years in the future", currentYear+11),
		},
		{
			name: "invalid - expiration date too old",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithCloseDate(2024, time.August, 15)
				b.WithExpirationDate(1999, time.January, 1)
			}),
			expectedError: fmt.Errorf("expiration date 1999-01-01 is before minimum valid year 2000"),
		},
		{
			name: "invalid - movement date year too old (catches year parsing errors like '24' -> 2024)",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithCloseDate(2024, time.August, 15)
				b.WithCard("1234", "OWNER", "1000.00", "0.00")
				// This simulates the bug where a 2-digit year "24" was incorrectly parsed as year 24 (instead of 2024)
				// Year 24 is way before 2000, so it should be caught
				b.WithCardMovement(0, testsale.DatePtr(24, time.August, 15), "123456*", "DETAIL", "1000.00", "0.00")
			}),
			expectedError: fmt.Errorf(`card 0 (owner: "OWNER", number: 1234, movements: 1) movement 0 (date: 0024-08-15, detail: "DETAIL", receipt: 123456*) date 0024-08-15 has year 24 which is before minimum valid year 2000`),
		},
		{
			name: "invalid - movement date year too old (another edge case)",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithCloseDate(2024, time.August, 15)
				b.WithCard("1234", "OWNER", "1000.00", "0.00")
				b.WithCardMovement(0, testsale.DatePtr(1999, time.August, 15), "123456*", "DETAIL", "1000.00", "0.00")
			}),
			expectedError: fmt.Errorf(`card 0 (owner: "OWNER", number: 1234, movements: 1) movement 0 (date: 1999-08-15, detail: "DETAIL", receipt: 123456*) date 1999-08-15 has year 1999 which is before minimum valid year 2000`),
		},
		{
			name: "invalid - movement date year too far in future",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithCloseDate(2024, time.August, 15)
				b.WithCard("1234", "OWNER", "1000.00", "0.00")
				b.WithCardMovement(0, testsale.DatePtr(currentYear+11, time.August, 15), "123456*", "DETAIL", "1000.00", "0.00")
			}),
			expectedError: fmt.Errorf(`card 0 (owner: "OWNER", number: 1234, movements: 1) movement 0 (date: %d-08-15, detail: "DETAIL", receipt: 123456*) date %d-08-15 has year %d which is more than 10 years in the future`, currentYear+11, currentYear+11, currentYear+11),
		},
		{
			name: "valid - movement date year at boundary (2000)",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithCloseDate(2024, time.August, 15)
				b.WithCard("1234", "OWNER", "1000.00", "0.00")
				b.WithCardMovement(0, testsale.DatePtr(2000, time.August, 15), "123456*", "DETAIL", "1000.00", "0.00")
			}),
			expectedError: nil,
		},
		{
			name: "valid - movement date year at future boundary",
			cardSummary: testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
				b.WithCloseDate(2024, time.August, 15)
				b.WithCard("1234", "OWNER", "1000.00", "0.00")
				b.WithCardMovement(0, testsale.DatePtr(currentYear+10, time.August, 15), "123456*", "DETAIL", "1000.00", "0.00")
			}),
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			// CardSummary is set up in test case

			// When
			err := validation.ValidateDateRangeReasonableness(tc.cardSummary)

			// Then
			require.Equal(t, tc.expectedError, err)
		})
	}
}
