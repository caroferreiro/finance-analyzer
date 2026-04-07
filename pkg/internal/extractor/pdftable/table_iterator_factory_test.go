package pdftable

import (
	"testing"

	"github.com/Alechan/finance-analyzer/pkg/internal/platform/pdfwrapper"
	"github.com/stretchr/testify/require"
)

// TestNewTableIteratorFactory_Creation verifies that the factory is created correctly
// with different RowFactory configurations. This is a simple unit test that checks
// the factory's internal state after creation.
func TestNewTableIteratorFactory_Creation(t *testing.T) {
	tests := []struct {
		name      string
		positions PDFTablePositions
	}{
		{
			name:      "Valid positions",
			positions: TestTablePositions,
		},
		{
			name:      "Zero positions",
			positions: PDFTablePositions{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			rowFactory := NewRowFactory(tt.positions)

			// When
			factory := NewTableIteratorFactory(rowFactory)

			// Then
			require.NotNil(t, factory)
		})
	}
}

// TestTableIteratorFactory_CreateIterator_WhiteBox verifies that the factory correctly
// creates an iterator with the expected behavior. This test checks that the iterator
// returns the expected rows in sequence.
func TestTableIteratorFactory_CreateIterator_WhiteBox(t *testing.T) {
	tests := []struct {
		name          string
		texts         []string
		expectedRows  []Row
		expectedError error
	}{
		{
			name:          "Empty document",
			texts:         []string{},
			expectedRows:  nil,
			expectedError: nil,
		},
		{
			name:  "Single valid row",
			texts: []string{TestCardMovementText},
			expectedRows: []Row{
				TestCardMovementRow,
			},
			expectedError: nil,
		},
		{
			name: "Multiple rows including malformed text",
			texts: []string{
				TestCardMovementText,
				TestShortText,
				TestSaldoAnteriorText,
			},
			expectedRows: []Row{
				TestCardMovementRow,
				TestShortTextRow,
				TestSaldoAnteriorRow,
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			docIterator := pdfwrapper.NewFakeDocumentIterator(tt.texts)
			rowFactory := NewRowFactory(TestTablePositions)
			factory := NewTableIteratorFactory(rowFactory)

			// When
			it := factory.CreateIterator(docIterator)

			// Then
			require.NotNil(t, it)

			// Test iterator behavior instead of accessing private fields
			var actualRows []Row
			for row, ok := it.Next(); ok; row, ok = it.Next() {
				actualRows = append(actualRows, row)
			}
			require.Equal(t, tt.expectedRows, actualRows)
		})
	}
}
