package pdfwrapper

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFakeDocumentIterator_NextText(t *testing.T) {
	tests := []struct {
		name     string
		texts    []string
		expected []string
	}{
		{
			name:     "Empty texts",
			texts:    []string{},
			expected: nil,
		},
		{
			name:     "Single text",
			texts:    []string{"hello"},
			expected: []string{"hello"},
		},
		{
			name:     "Multiple texts",
			texts:    []string{"hello", "world", "!"},
			expected: []string{"hello", "world", "!"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			it := NewFakeDocumentIterator(tt.texts)

			// When
			var actual []string
			for text, ok := it.NextText(); ok; text, ok = it.NextText() {
				actual = append(actual, text)
			}

			// Then
			require.Equal(t, tt.expected, actual)

			// Verify that NextText returns false after exhausting all texts
			text, ok := it.NextText()
			require.Empty(t, text)
			require.False(t, ok)
		})
	}
}
