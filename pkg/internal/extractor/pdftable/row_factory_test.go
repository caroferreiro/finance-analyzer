package pdftable

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRowFactory_CreateRow(t *testing.T) {
	tests := []struct {
		name        string
		rawText     string
		positions   PDFTablePositions
		expectedRow Row
	}{
		{
			name:        "typical card movement with all fields",
			rawText:     TestCardMovementText,
			positions:   TestTablePositions,
			expectedRow: TestCardMovementRow,
		},
		{
			name:        "saldo anterior row with ARS and USD amounts",
			rawText:     TestSaldoAnteriorText,
			positions:   TestTablePositions,
			expectedRow: TestSaldoAnteriorRow,
		},
		{
			name:        "short text that only covers first column",
			rawText:     TestShortText,
			positions:   TestTablePositions,
			expectedRow: TestShortTextRow,
		},
		{
			name:        "empty text produces empty row",
			rawText:     "",
			positions:   TestTablePositions,
			expectedRow: Row{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			factory := NewRowFactory(tt.positions)

			// When
			actualRow := factory.CreateRow(tt.rawText)

			// Then
			require.Equal(t, tt.expectedRow, actualRow)
		})
	}
}

// TestRowFactory_CreateRow_Generic verifies that RowFactory.CreateRow correctly parses rows using generic table positions.
// This test focuses on core functionality and edge cases that apply to any PDF table format.
func TestRowFactory_CreateRow_Generic(t *testing.T) {
	tests := []struct {
		name          string
		rawText       string
		positions     PDFTablePositions
		expectedRow   Row
		expectedError error
	}{
		{
			name:        "typical card movement with all fields",
			rawText:     TestCardMovementText,
			positions:   TestTablePositions,
			expectedRow: TestCardMovementRow,
		},
		{
			name:        "saldo anterior row with ARS and USD amounts",
			rawText:     TestSaldoAnteriorText,
			positions:   TestTablePositions,
			expectedRow: TestSaldoAnteriorRow,
		},
		{
			name:        "short text that only covers first column",
			rawText:     TestShortText,
			positions:   TestTablePositions,
			expectedRow: TestShortTextRow,
		},
		{
			name:        "boundary conditions - text exactly at column ends",
			rawText:     TestBoundaryText,
			positions:   TestBoundaryPositions,
			expectedRow: TestBoundaryRow,
		},
		{
			name:        "whitespace and special characters",
			rawText:     TestWhitespaceText,
			positions:   TestTablePositions,
			expectedRow: TestWhitespaceRow,
		},
		{
			name:      "empty text",
			rawText:   "",
			positions: TestTablePositions,
			expectedRow: Row{
				RawText:                        "",
				RawOriginalDate:                "",
				RawReceiptNumber:               "",
				RawDetailWithMaybeInstallments: "",
				RawAmountARS:                   "",
				RawAmountUSD:                   "",
			},
		},
		{
			name:    "text shorter than any column start",
			rawText: "abc",
			positions: PDFTablePositions{
				OriginalDateStart: 10,
				OriginalDateEnd:   15,
				ReceiptStart:      20,
				ReceiptEnd:        25,
				DetailStart:       30,
				DetailEnd:         35,
				ARSAmountStart:    40,
				ARSAmountEnd:      45,
				USDAmountStart:    50,
				USDAmountEnd:      55,
			},
			expectedRow: Row{
				RawText:                        "abc",
				RawOriginalDate:                "",
				RawReceiptNumber:               "",
				RawDetailWithMaybeInstallments: "",
				RawAmountARS:                   "",
				RawAmountUSD:                   "",
			},
		},
		{
			name:    "text exactly at column start boundary",
			rawText: "1234567890123456789012345678901234567890",
			positions: PDFTablePositions{
				OriginalDateStart: 0,
				OriginalDateEnd:   9,
				ReceiptStart:      10,
				ReceiptEnd:        19,
				DetailStart:       20,
				DetailEnd:         29,
				ARSAmountStart:    30,
				ARSAmountEnd:      39,
				USDAmountStart:    40,
				USDAmountEnd:      49,
			},
			expectedRow: Row{
				RawText:                        "1234567890123456789012345678901234567890",
				RawOriginalDate:                "1234567890",
				RawReceiptNumber:               "1234567890",
				RawDetailWithMaybeInstallments: "1234567890",
				RawAmountARS:                   "1234567890",
				RawAmountUSD:                   "",
			},
		},
		{
			name:    "overlapping column positions",
			rawText: "abcdefghijklmnop",
			positions: PDFTablePositions{
				OriginalDateStart: 0,
				OriginalDateEnd:   5,
				ReceiptStart:      3,
				ReceiptEnd:        8,
				DetailStart:       6,
				DetailEnd:         11,
				ARSAmountStart:    9,
				ARSAmountEnd:      14,
				USDAmountStart:    12,
				USDAmountEnd:      15,
			},
			expectedRow: Row{
				RawText:                        "abcdefghijklmnop",
				RawOriginalDate:                "abcdef",
				RawReceiptNumber:               "defghi",
				RawDetailWithMaybeInstallments: "ghijkl",
				RawAmountARS:                   "jklmno",
				RawAmountUSD:                   "mnop",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			factory := NewRowFactory(tt.positions)

			// When
			actualRow := factory.CreateRow(tt.rawText)

			// Then
			require.Equal(t, tt.expectedRow, actualRow)
		})
	}
}

func TestNewRowFactory(t *testing.T) {
	// Given
	positions := TestTablePositions

	// When
	factory := NewRowFactory(positions)

	// Then
	require.NotNil(t, factory)
	require.Equal(t, positions, factory.positions)
}
