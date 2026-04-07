package timeale

import (
	"fmt"
	"testing"
	"time"
)

func TestCardSummarySpanishMonthDateToTime(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      time.Time
		expectedError error
	}{
		{
			name:          "valid date with Ene",
			input:         "15 Ene 24",
			expected:      time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			expectedError: nil,
		},
		{
			name:          "valid date with Feb",
			input:         "29 Feb 24",
			expected:      time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
			expectedError: nil,
		},
		{
			name:          "valid date with Sep (both Set and Sep)",
			input:         "30 Sep 24",
			expected:      time.Date(2024, 9, 30, 0, 0, 0, 0, time.UTC),
			expectedError: nil,
		},
		{
			name:          "valid date with Set (both Set and Sep)",
			input:         "30 Set 24",
			expected:      time.Date(2024, 9, 30, 0, 0, 0, 0, time.UTC),
			expectedError: nil,
		},
		{
			name:          "valid date with Dic",
			input:         "31 Dic 24",
			expected:      time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
			expectedError: nil,
		},
		{
			name:          "invalid date format - missing parts",
			input:         "15 Ene",
			expectedError: fmt.Errorf("invalid date format: 15 Ene"),
		},
		{
			name:          "invalid date format - extra parts",
			input:         "15 Ene 24 extra",
			expectedError: fmt.Errorf("invalid date format: 15 Ene 24 extra"),
		},
		{
			name:          "invalid month abbreviation",
			input:         "15 Invalid 24",
			expectedError: fmt.Errorf("invalid month abbreviation: Invalid"),
		},
		{
			name:          "invalid day",
			input:         "32 Ene 24",
			expectedError: fmt.Errorf("error parsing date: parsing time \"32 01 24\": day out of range"),
		},
		{
			name:          "invalid year format",
			input:         "15 Ene 2024",
			expectedError: fmt.Errorf("error parsing date: parsing time \"15 01 2024\": extra text: \"24\""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CardSummarySpanishMonthDateToTime(tt.input)

			// Check error cases
			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("CardSummarySpanishMonthDateToTime() expected error %v but got none", tt.expectedError)
					return
				}
				if err.Error() != tt.expectedError.Error() {
					t.Errorf("CardSummarySpanishMonthDateToTime() error = %v, want %v", err, tt.expectedError)
				}
				return
			}

			// Check non-error cases
			if err != nil {
				t.Errorf("CardSummarySpanishMonthDateToTime() unexpected error: %v", err)
				return
			}

			// Compare the parsed time with expected time
			if !got.Equal(tt.expected) {
				t.Errorf("CardSummarySpanishMonthDateToTime() = %v, want %v", got, tt.expected)
			}
		})
	}
}
