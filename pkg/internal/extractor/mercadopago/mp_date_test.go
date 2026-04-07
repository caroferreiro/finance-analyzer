package mercadopago

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseShortDate(t *testing.T) {
	closeDate := time.Date(2025, time.October, 12, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		input    string
		expected time.Time
	}{
		{name: "same month", input: "9/oct", expected: time.Date(2025, time.October, 9, 0, 0, 0, 0, time.UTC)},
		{name: "earlier month", input: "23/ago", expected: time.Date(2025, time.August, 23, 0, 0, 0, 0, time.UTC)},
		{name: "later month wraps to prev year", input: "5/dic", expected: time.Date(2024, time.December, 5, 0, 0, 0, 0, time.UTC)},
		{name: "single digit day", input: "2/oct", expected: time.Date(2025, time.October, 2, 0, 0, 0, 0, time.UTC)},
		{name: "alternative sept abbreviation", input: "15/set", expected: time.Date(2025, time.September, 15, 0, 0, 0, 0, time.UTC)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseShortDate(tt.input, closeDate)
			require.NotNil(t, got)
			assert.Equal(t, tt.expected, *got)
		})
	}
}

func TestParseShortDate_Invalid(t *testing.T) {
	closeDate := time.Date(2025, time.October, 12, 0, 0, 0, 0, time.UTC)

	assert.Nil(t, ParseShortDate("5/xyz", closeDate))
	assert.Nil(t, ParseShortDate("not-a-date", closeDate))
	assert.Nil(t, ParseShortDate("", closeDate))
}

func TestParseFullSpanishDate(t *testing.T) {
	text := "Cierre actual12 de eneroVencimiento actual19 de enero"
	date, err := ParseFullSpanishDate(text, "Cierre actual", 2026)
	require.NoError(t, err)
	assert.Equal(t, time.January, date.Month())
	assert.Equal(t, 12, date.Day())
	assert.Equal(t, 2026, date.Year())
}

func TestParseFullSpanishDate_NotFound(t *testing.T) {
	_, err := ParseFullSpanishDate("no date here", "Cierre actual", 2026)
	assert.Error(t, err)
}

func TestInferYearFromCloseMonth(t *testing.T) {
	now := time.Date(2026, time.April, 7, 0, 0, 0, 0, time.UTC)

	assert.Equal(t, 2026, InferYearFromCloseMonth(time.January, now))
	assert.Equal(t, 2026, InferYearFromCloseMonth(time.March, now))
	assert.Equal(t, 2026, InferYearFromCloseMonth(time.April, now))
	assert.Equal(t, 2025, InferYearFromCloseMonth(time.October, now))
	assert.Equal(t, 2025, InferYearFromCloseMonth(time.December, now))
}

func TestExtractCloseMonth(t *testing.T) {
	text := "blahCierre actual12 de octubreVencimiento"
	month, err := ExtractCloseMonth(text)
	require.NoError(t, err)
	assert.Equal(t, time.October, month)
}

func TestExtractCloseMonth_FallbackLabel(t *testing.T) {
	text := "Fecha de cierre12 de eneroFecha de vencimiento"
	month, err := ExtractCloseMonth(text)
	require.NoError(t, err)
	assert.Equal(t, time.January, month)
}
