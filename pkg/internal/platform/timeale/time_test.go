package timeale

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestDaysBetweenDates(t *testing.T) {
	t.Run("date1 is before date2 by 1 day", func(t *testing.T) {
		date1 := time.Date(2023, time.September, 1, 0, 0, 0, 0, time.UTC)
		date2 := time.Date(2023, time.September, 2, 0, 0, 0, 0, time.UTC)
		expected := 1
		result := DaysBetweenDates(date1, date2)
		require.Equal(t, expected, result)
	})

	t.Run("date1 is before date2 by 2 days", func(t *testing.T) {
		date1 := time.Date(2023, time.September, 1, 0, 0, 0, 0, time.UTC)
		date2 := time.Date(2023, time.September, 3, 0, 0, 0, 0, time.UTC)
		expected := 2
		result := DaysBetweenDates(date1, date2)
		require.Equal(t, expected, result)
	})

	t.Run("date1 is the same as date2", func(t *testing.T) {
		date1 := time.Date(2023, time.September, 1, 0, 0, 0, 0, time.UTC)
		date2 := time.Date(2023, time.September, 1, 0, 0, 0, 0, time.UTC)
		expected := 0
		result := DaysBetweenDates(date1, date2)
		require.Equal(t, expected, result)
	})

	t.Run("date1 is after date2 by 1 day", func(t *testing.T) {
		date1 := time.Date(2023, time.September, 2, 0, 0, 0, 0, time.UTC)
		date2 := time.Date(2023, time.September, 1, 0, 0, 0, 0, time.UTC)
		expected := -1
		result := DaysBetweenDates(date1, date2)
		require.Equal(t, expected, result)
	})

	t.Run("date1 is after date2 by 2 days", func(t *testing.T) {
		date1 := time.Date(2023, time.September, 3, 0, 0, 0, 0, time.UTC)
		date2 := time.Date(2023, time.September, 1, 0, 0, 0, 0, time.UTC)
		expected := -2
		result := DaysBetweenDates(date1, date2)
		require.Equal(t, expected, result)
	})

	t.Run("date1 is after date2 by 1 year", func(t *testing.T) {
		date1 := time.Date(2024, time.September, 1, 0, 0, 0, 0, time.UTC)
		date2 := time.Date(2023, time.September, 1, 0, 0, 0, 0, time.UTC)
		expected := -366
		result := DaysBetweenDates(date1, date2)
		require.Equal(t, expected, result)
	})

	t.Run("date1 is after date2 by 1 hour on the same day", func(t *testing.T) {
		date1 := time.Date(2023, time.September, 1, 1, 0, 0, 0, time.UTC)
		date2 := time.Date(2023, time.September, 1, 0, 0, 0, 0, time.UTC)
		expected := 0
		result := DaysBetweenDates(date1, date2)
		require.Equal(t, expected, result)
	})

	t.Run("date1 is before date2 by 1 hour on the same day", func(t *testing.T) {
		date1 := time.Date(2023, time.September, 1, 0, 0, 0, 0, time.UTC)
		date2 := time.Date(2023, time.September, 1, 1, 0, 0, 0, time.UTC)
		expected := 0
		result := DaysBetweenDates(date1, date2)
		require.Equal(t, expected, result)
	})

	t.Run("date1 is before date2 by 1 hour on different days", func(t *testing.T) {
		date1 := time.Date(2023, time.September, 1, 0, 0, 0, 0, time.UTC)
		date2 := time.Date(2023, time.September, 2, 1, 0, 0, 0, time.UTC)
		expected := 1
		result := DaysBetweenDates(date1, date2)
		require.Equal(t, expected, result)
	})

	t.Run("date1 is after date2 by 1 hour on different days", func(t *testing.T) {
		date1 := time.Date(2023, time.September, 2, 1, 0, 0, 0, time.UTC)
		date2 := time.Date(2023, time.September, 1, 0, 0, 0, 0, time.UTC)
		expected := -1
		result := DaysBetweenDates(date1, date2)
		require.Equal(t, expected, result)
	})
}

func TestMiddleOfNextMonth(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "first of the month",
			input:    time.Date(2023, time.September, 1, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2023, time.October, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "last day of the month",
			input:    time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2023, time.October, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "middle of the month",
			input:    time.Date(2023, time.September, 15, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2023, time.October, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "last day of the month in 31 days month",
			input:    time.Date(2023, time.January, 31, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2023, time.February, 15, 0, 0, 0, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := MiddleOfNextMonth(tt.input)
			require.Equal(t, tt.expected, actual)
		})
	}
}
