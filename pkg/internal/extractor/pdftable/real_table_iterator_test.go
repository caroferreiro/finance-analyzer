package pdftable

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRealTableIterator_Next covers all scenarios for the Next method.
func TestRealTableIterator_Next(t *testing.T) {
	tests := []struct {
		name         string
		rows         []Row
		expectedRows []Row
	}{
		{
			name:         "Empty",
			rows:         nil,
			expectedRows: nil,
		},
		{
			name:         "Single row",
			rows:         []Row{TestCardMovementRow},
			expectedRows: []Row{TestCardMovementRow},
		},
		{
			name:         "Multiple rows",
			rows:         []Row{TestCardMovementRow, TestShortTextRow, TestSaldoAnteriorRow},
			expectedRows: []Row{TestCardMovementRow, TestShortTextRow, TestSaldoAnteriorRow},
		},
		{
			name:         "Single saldo anterior",
			rows:         []Row{TestSaldoAnteriorRow},
			expectedRows: []Row{TestSaldoAnteriorRow},
		},
		{
			name:         "Single valid card movement",
			rows:         []Row{TestCardMovementRow},
			expectedRows: []Row{TestCardMovementRow},
		},
		{
			name:         "Multiple rows: saldo anterior and card movement",
			rows:         []Row{TestSaldoAnteriorRow, TestCardMovementRow},
			expectedRows: []Row{TestSaldoAnteriorRow, TestCardMovementRow},
		},
		{
			name:         "Malformed/short line: too short for all columns",
			rows:         []Row{TestShortTextRow},
			expectedRows: []Row{TestShortTextRow},
		},
		{
			name:         "Boundary: line exactly at column ends",
			rows:         []Row{TestBoundaryRow},
			expectedRows: []Row{TestBoundaryRow},
		},
		{
			name:         "Whitespace and special characters",
			rows:         []Row{TestWhitespaceRow},
			expectedRows: []Row{TestWhitespaceRow},
		},
		{
			name:         "Mixed row types",
			rows:         []Row{TestSaldoAnteriorRow, TestCardMovementRow, TestShortTextRow, TestBoundaryRow, TestWhitespaceRow},
			expectedRows: []Row{TestSaldoAnteriorRow, TestCardMovementRow, TestShortTextRow, TestBoundaryRow, TestWhitespaceRow},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			it := NewRealTableIterator(tt.rows)

			// When
			var actualRows []Row
			for row, ok := it.Next(); ok; row, ok = it.Next() {
				actualRows = append(actualRows, row)
			}

			// Then
			require.Equal(t, tt.expectedRows, actualRows)

			// Exhaustion: Next returns false after all rows
			row, ok := it.Next()
			require.Empty(t, row)
			require.False(t, ok)
		})
	}
}

// TestRealTableIterator_NextUtilRegexIsMatched covers all scenarios for regex-based iteration.
func TestRealTableIterator_NextUtilRegexIsMatched(t *testing.T) {
	tests := []struct {
		name          string
		rows          []Row
		regex         *regexp.Regexp
		expectedRow   Row
		expectedFound bool
	}{
		{
			name:          "Empty",
			rows:          nil,
			regex:         regexp.MustCompile(`SALDO`),
			expectedRow:   Row{},
			expectedFound: false,
		},
		{
			name:          "Match found in first row",
			rows:          []Row{TestSaldoAnteriorRow, TestCardMovementRow},
			regex:         regexp.MustCompile(`SALDO`),
			expectedRow:   TestSaldoAnteriorRow,
			expectedFound: true,
		},
		{
			name:          "Match found in second row",
			rows:          []Row{TestCardMovementRow, TestSaldoAnteriorRow},
			regex:         regexp.MustCompile(`SALDO`),
			expectedRow:   TestSaldoAnteriorRow,
			expectedFound: true,
		},
		{
			name:          "No match found",
			rows:          []Row{TestCardMovementRow, TestShortTextRow},
			regex:         regexp.MustCompile(`SALDO`),
			expectedRow:   Row{},
			expectedFound: false,
		},
		{
			name:          "Match with whitespace and special characters",
			rows:          []Row{TestCardMovementRow, TestWhitespaceRow, TestSaldoAnteriorRow},
			regex:         regexp.MustCompile(`Enero`),
			expectedRow:   TestWhitespaceRow,
			expectedFound: true,
		},
		{
			name:          "Match with boundary row",
			rows:          []Row{TestCardMovementRow, TestBoundaryRow, TestSaldoAnteriorRow},
			regex:         regexp.MustCompile(`abc`),
			expectedRow:   TestBoundaryRow,
			expectedFound: true,
		},
		{
			name:          "Match with short text row",
			rows:          []Row{TestCardMovementRow, TestShortTextRow, TestSaldoAnteriorRow},
			regex:         regexp.MustCompile(`short`),
			expectedRow:   TestShortTextRow,
			expectedFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			it := NewRealTableIterator(tt.rows)

			// When
			row, found := it.NextUtilRegexIsMatched(tt.regex)

			// Then
			require.Equal(t, tt.expectedFound, found)
			require.Equal(t, tt.expectedRow, row)

			// Exhaustion: NextUtilRegexIsMatched returns false after all rows
			row, found = it.NextUtilRegexIsMatched(tt.regex)
			require.Empty(t, row)
			require.False(t, found)
		})
	}
}

// TestRealTableIterator_NextUtilRegexIsMatched_ComplexPatterns covers regex edge cases and complex patterns.
func TestRealTableIterator_NextUtilRegexIsMatched_ComplexPatterns(t *testing.T) {
	tests := []struct {
		name          string
		rows          []Row
		regex         *regexp.Regexp
		expectedRow   Row
		expectedFound bool
	}{
		{
			name:          "Match with case insensitive pattern",
			rows:          []Row{TestCardMovementRow, TestSaldoAnteriorRow},
			regex:         regexp.MustCompile(`(?i)saldo`),
			expectedRow:   TestSaldoAnteriorRow,
			expectedFound: true,
		},
		{
			name:          "Match with word boundary",
			rows:          []Row{TestCardMovementRow, TestWhitespaceRow, TestSaldoAnteriorRow},
			regex:         regexp.MustCompile(`\bEnero\b`),
			expectedRow:   TestWhitespaceRow,
			expectedFound: true,
		},
		{
			name:          "Match with digit pattern",
			rows:          []Row{TestCardMovementRow, TestSaldoAnteriorRow},
			regex:         regexp.MustCompile(`222\.111,66`),
			expectedRow:   TestSaldoAnteriorRow,
			expectedFound: true,
		},
		{
			name:          "Match with special characters",
			rows:          []Row{TestCardMovementRow, TestWhitespaceRow, TestSaldoAnteriorRow},
			regex:         regexp.MustCompile(`!@#`),
			expectedRow:   TestWhitespaceRow,
			expectedFound: true,
		},
		{
			name:          "No match with complex pattern",
			rows:          []Row{TestCardMovementRow, TestShortTextRow},
			regex:         regexp.MustCompile(`\d{4}-\d{2}-\d{2}`),
			expectedRow:   Row{},
			expectedFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			it := NewRealTableIterator(tt.rows)

			// When
			row, found := it.NextUtilRegexIsMatched(tt.regex)

			// Then
			require.Equal(t, tt.expectedFound, found)
			require.Equal(t, tt.expectedRow, row)

			// Exhaustion: NextUtilRegexIsMatched returns false after all rows
			row, found = it.NextUtilRegexIsMatched(tt.regex)
			require.Empty(t, row)
			require.False(t, found)
		})
	}
}
