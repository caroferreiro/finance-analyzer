package extractorfuncs

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"

	"github.com/Alechan/finance-analyzer/pkg/internal/platform/pdfwrapper"
)

func TestFirstMatchExtractor_ExtractFirstMatch(t *testing.T) {
	tests := []struct {
		name           string
		regex          *regexp.Regexp
		pages          []pdfwrapper.Page
		deserializerFn func(string) (string, error)
		expected       string
		expectedError  error
	}{
		{
			name:  "successful extraction of first match",
			regex: regexp.MustCompile(`Date: (\d{4}-\d{2}-\d{2})`),
			pages: []pdfwrapper.Page{
				{
					Index: 1,
					Rows: []pdfwrapper.Row{
						{Position: 0, Texts: []string{"Some text", "Date: 2024-03-20", "More text"}},
					},
				},
				{
					Index: 2,
					Rows: []pdfwrapper.Row{
						{Position: 0, Texts: []string{"Date: 2024-03-21", "Other text"}},
					},
				},
			},
			deserializerFn: func(s string) (string, error) { return s, nil },
			expected:       "2024-03-20",
			expectedError:  nil,
		},
		{
			name:  "no match found",
			regex: regexp.MustCompile(`Date: (\d{4}-\d{2}-\d{2})`),
			pages: []pdfwrapper.Page{
				{
					Index: 1,
					Rows: []pdfwrapper.Row{
						{Position: 0, Texts: []string{"Some text without date"}},
					},
				},
				{
					Index: 2,
					Rows: []pdfwrapper.Row{
						{Position: 0, Texts: []string{"More text without date"}},
					},
				},
			},
			deserializerFn: func(s string) (string, error) { return s, nil },
			expected:       "",
			expectedError:  fmt.Errorf("error extracting first match: %w", pdfwrapper.ErrPatternNotFound),
		},
		{
			name:  "deserialization error",
			regex: regexp.MustCompile(`Date: (\d{4}-\d{2}-\d{2})`),
			pages: []pdfwrapper.Page{
				{
					Index: 1,
					Rows: []pdfwrapper.Row{
						{Position: 0, Texts: []string{"Date: 9999-99-99", "Some other text"}},
					},
				},
			},
			deserializerFn: func(s string) (string, error) {
				return "", errors.New("invalid date format")
			},
			expected: "",
			expectedError: fmt.Errorf(
				"error running fold on pages: %w",
				fmt.Errorf(
					"error processing word: %w",
					fmt.Errorf("error extracting from 'Date: 9999-99-99' the value '9999-99-99' and converting to type string: %w",
						errors.New("invalid date format"),
					),
				),
			),
		},
		{
			name:  "multiple matches returns first",
			regex: regexp.MustCompile(`Value: (\d+)`),
			pages: []pdfwrapper.Page{
				{
					Index: 1,
					Rows: []pdfwrapper.Row{
						{Position: 0, Texts: []string{"Value: 100", "Value: 200"}},
					},
				},
				{
					Index: 2,
					Rows: []pdfwrapper.Row{
						{Position: 0, Texts: []string{"Value: 300"}},
					},
				},
			},
			deserializerFn: func(s string) (string, error) { return s, nil },
			expected:       "100",
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			extractor := NewFirstMatchExtractor(tt.regex, tt.deserializerFn)

			// When
			result, err := extractor.ExtractFirstMatch(tt.pages)

			// Then
			require.Equal(t, tt.expected, result)
			require.Equal(t, tt.expectedError, err)
		})
	}
}
