package testsale

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func AsDecimal(t *testing.T, s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	require.NoError(t, err)
	return d
}

func DatePtr(year int, month time.Month, day int) *time.Time {
	date := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return &date
}

func StrPtr(s string) *string {
	return &s
}

func IntPtr(s int) *int {
	return &s
}

// AssertDecimalEqual compares two decimal values using decimal.Equal and provides
// detailed error output with string representations if they don't match.
func AssertDecimalEqual(t TestingTB, expected, actual decimal.Decimal) {
	t.Helper()
	if !expected.Equal(actual) {
		t.Fatalf("decimal values do not match:\nexpected: %s\nactual:   %s", expected.String(), actual.String())
	}
}
