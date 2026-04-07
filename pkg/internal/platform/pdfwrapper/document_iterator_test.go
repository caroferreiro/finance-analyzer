package pdfwrapper

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestContinueProcessingPages(t *testing.T) {
	// Test "ContinueProcessingPages" function of the iterator.
	// For each test we keep a slice of strings that should be processed by the processFunc, which is called for each word
	tests := []struct {
		name                string
		pages               []Page
		startingRowPosition int
		processFunc         func(word string, processed *[]string) (ProcessResult, error)
		expectedWords       []string
		expectError         error
	}{
		{
			name:                "Empty pages",
			pages:               []Page{},
			startingRowPosition: 10,
			processFunc: func(word string, processed *[]string) (ProcessResult, error) {
				*processed = append(*processed, word)
				return Continue, nil
			},
			expectedWords: nil,
			expectError:   nil,
		},
		{
			name: "Row with empty text",
			pages: []Page{
				{
					Index: 0,
					Rows: []Row{
						{Position: 8, Texts: []string{}},
						{Position: 5, Texts: []string{"a"}},
					},
				},
			},
			startingRowPosition: 10,
			processFunc: func(word string, processed *[]string) (ProcessResult, error) {
				*processed = append(*processed, word)
				return Continue, nil
			},
			expectedWords: []string{"a"},
			expectError:   nil,
		},
		{
			name: "Full processing",
			pages: []Page{
				{
					Index: 0,
					Rows: []Row{
						{Position: 8, Texts: []string{"a", "b"}},
						{Position: 5, Texts: []string{"c", "d"}},
					},
				},
			},
			startingRowPosition: 10,
			processFunc: func(word string, processed *[]string) (ProcessResult, error) {
				*processed = append(*processed, word)
				return Continue, nil
			},
			expectedWords: []string{"a", "b", "c", "d"},
			expectError:   nil,
		},
		{
			name: "Skip rows based on position",
			pages: []Page{
				{
					Index: 0,
					Rows: []Row{
						// This row is skipped because 12 > startingRowPosition (10)
						{Position: 12, Texts: []string{"a", "b"}},
						// This row is processed because 8 <= 10
						{Position: 8, Texts: []string{"c", "d"}},
					},
				},
			},
			startingRowPosition: 10,
			processFunc: func(word string, processed *[]string) (ProcessResult, error) {
				*processed = append(*processed, word)
				return Continue, nil
			},
			expectedWords: []string{"c", "d"},
			expectError:   nil,
		},
		{
			name: "Stop iteration",
			pages: []Page{
				{
					Index: 0,
					Rows: []Row{
						// When "b" is encountered, break from the whole document.
						{Position: 8, Texts: []string{"a", "b", "c"}},
					},
				},
				{
					Index: 1,
					Rows: []Row{
						{Position: 7, Texts: []string{"d", "e"}},
					},
				},
			},
			startingRowPosition: 10,
			processFunc: func(word string, processed *[]string) (ProcessResult, error) {
				*processed = append(*processed, word)
				if word == "b" {
					// break from the whole document
					return StopIteration, nil
				}
				return Continue, nil
			},
			// Expect processing stops immediately after "b".
			expectedWords: []string{"a", "b"},
			expectError:   nil,
		},
		{
			name: "Error from processWord",
			pages: []Page{
				{
					Index: 0,
					Rows: []Row{
						{Position: 8, Texts: []string{"a", "b", "c"}},
					},
				},
			},
			startingRowPosition: 10,
			processFunc: func(word string, processed *[]string) (ProcessResult, error) {
				*processed = append(*processed, word)
				if word == "b" {
					return StopIteration, fmt.Errorf("test error")
				}
				return Continue, nil
			},
			// Expect that processing stops on error (after processing "b").
			expectedWords: []string{"a", "b"},
			expectError:   fmt.Errorf("error processing word: %w", errors.New("test error")),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var processed []string
			// Wrap the table's processFunc into the signature required by ContinueProcessingPages.
			processWord := func(word string) (ProcessResult, error) {
				return tc.processFunc(word, &processed)
			}
			// Given
			doc := NewDocument(tc.pages)
			iter := NewRealDocumentIterator(doc, tc.startingRowPosition)

			// When
			err := iter.ContinueProcessingPages(processWord)

			// Then
			require.Equal(t, tc.expectError, err)
			require.Equal(t, tc.expectedWords, processed)
		})
	}
}

func Test_DocumentIterator_Next(t *testing.T) {
	tests := []struct {
		name                string
		doc                 Document
		startingRowPosition int
		// expected holds the sequence of texts that should be returned by successive calls to NextText.
		expected []string
	}{
		{
			name:                "Empty Document",
			doc:                 Document{Pages: []Page{}},
			startingRowPosition: 0,
			expected:            nil,
		},
		{
			name: "Single page, single row, one text",
			doc: Document{
				Pages: []Page{
					{
						Index: 0,
						Rows: []Row{
							{Position: 0, Texts: []string{"Hello"}},
						},
					},
				},
			},
			// With startingRowPosition set to 1, row with Position 0 qualifies (0 <= 1)
			startingRowPosition: 1,
			expected:            []string{"Hello"},
		},
		{
			name: "Single page, single row, multiple texts",
			doc: Document{
				Pages: []Page{
					{
						Index: 0,
						Rows: []Row{
							{Position: 0, Texts: []string{"Hello", "World"}},
						},
					},
				},
			},
			startingRowPosition: 1,
			// Expected behavior: first call returns "Hello", second call returns "World"
			expected: []string{"Hello", "World"},
		},
		{
			name: "Single page, multiple rows",
			doc: Document{
				Pages: []Page{
					{
						Index: 0,
						Rows: []Row{
							// Row qualifies if row.Position <= startingRowPosition.
							{Position: 2, Texts: []string{"A"}},
							{Position: 1, Texts: []string{"B", "C"}},
						},
					},
				},
			},
			// Setting startingRowPosition to 2 means both rows qualify (2 <= 2 and 1 <= 2).
			startingRowPosition: 2,
			// Expected: "A" from the first row then "B" and "C" from the second.
			expected: []string{"A", "B", "C"},
		},
		{
			name: "Multiple pages, startingRowPosition = 1",
			doc: Document{
				Pages: []Page{
					{
						Index: 0,
						Rows: []Row{
							{Position: 1, Texts: []string{"Page1Row1"}},
							{Position: 0, Texts: []string{"Page1Row2"}},
						},
					},
					{
						Index: 1,
						Rows: []Row{
							// This row is skipped because 2 > startingRowPosition (1)
							{Position: 2, Texts: []string{"Page2Row1", "Extra"}},
							{Position: 0, Texts: []string{"Page2Row2"}},
						},
					},
				},
			},
			startingRowPosition: 1,
			// Expected: only rows where Position <= 1 are processed.
			expected: []string{"Page1Row1", "Page1Row2", "Page2Row2"},
		},
		{
			name: "Multiple pages, startingRowPosition = 2",
			doc: Document{
				Pages: []Page{
					{
						Index: 0,
						Rows: []Row{
							{Position: 1, Texts: []string{"Page1Row1"}},
							{Position: 0, Texts: []string{"Page1Row2"}},
						},
					},
					{
						Index: 1,
						Rows: []Row{
							// Now this row qualifies because 2 <= 2.
							{Position: 2, Texts: []string{"Page2Row1", "Extra"}},
							{Position: 0, Texts: []string{"Page2Row2"}},
						},
					},
				},
			},
			startingRowPosition: 2,
			// Expected: all rows on page 0 and both texts in the qualifying row on page 1.
			expected: []string{"Page1Row1", "Page1Row2", "Page2Row1", "Extra", "Page2Row2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			it := NewRealDocumentIterator(tt.doc, tt.startingRowPosition)

			// When
			actualTexts := fromIteratorToSlice(it)

			// Then
			require.Equal(t, tt.expected, actualTexts)
		})
	}
}

func fromIteratorToSlice(it *RealDocumentIterator) []string {
	var texts []string
	for text, ok := it.NextText(); ok; text, ok = it.NextText() {
		texts = append(texts, text)
	}
	return texts
}
