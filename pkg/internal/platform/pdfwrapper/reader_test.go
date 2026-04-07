package pdfwrapper

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/Alechan/finance-analyzer/pkg/internal/platform/sliceale"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	invalidPDF = toBytes([]string{
		"this is not a PDF",
	})

	//go:embed test_files/small_valid.pdf
	smallValidPDF []byte
)

func TestReaderWrapper_ReadFromBytes(t *testing.T) {
	tests := []struct {
		name          string
		rawBytes      []byte
		expectedDoc   Document
		expectedError error
	}{
		{
			name:          "nil bytes",
			rawBytes:      nil,
			expectedDoc:   Document{},
			expectedError: ErrNilOrEmptyRawBytes,
		},
		{
			name:          "empty bytes",
			rawBytes:      []byte{},
			expectedDoc:   Document{},
			expectedError: ErrNilOrEmptyRawBytes,
		},
		{
			name:        "invalid PDF",
			rawBytes:    invalidPDF,
			expectedDoc: Document{},
			expectedError: fmt.Errorf(
				"%w: error finding/triming lines until header: :not a valid PDF file: missing header",
				ErrCreatingVendorReader,
			),
		},
		{
			name:     "small valid PDF",
			rawBytes: smallValidPDF,
			expectedDoc: Document{
				Pages: expectedPagesForSmallValid(),
			},
			expectedError: nil,
		},
		{
			name:     "technically invalid but readable PDF with [BEGIN] prefix",
			rawBytes: smallValidPDF,
			expectedDoc: Document{
				// The [BEGIN] prefix should be ignored
				Pages: expectedPagesForSmallValid(),
			},
			expectedError: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			r := &ReaderWrapper{}

			// When
			actualDoc, actualErr := r.ReadFromBytes(tt.rawBytes)

			// Then
			require.Equal(t, tt.expectedError, actualErr)
			require.Equal(t, tt.expectedDoc, actualDoc)
		})
	}
}

func expectedPagesForSmallValid() []Page {
	return []Page{
		{
			Index: 1,
			Rows: []Row{
				{
					Texts: []string{
						"",
						"My PDF Heading",
						"",
						"This is a sample paragraph in the PDF document. It provides some introductory content for the",
						"",
						"reader.",
						"",
						"    This paragraph has leading and trailing spaces.    ",
					}},
			},
		},
		{
			Index: 2,
			Rows: []Row{
				{
					Texts: []string{
						"",
						"Second Page Heading",
						"",
						"This is the content on the second page. Here you can add additional information, details, or any",
						"",
						"extra text you'd like to display.",
					},
				},
			},
		}}
}

func toBytes(stringsSlice []string) []byte {
	eachAsBytes := sliceale.ApplyMapFunction(
		stringsSlice,
		func(s string) []byte {
			return []byte(s)
		},
	)
	return bytes.Join(eachAsBytes, []byte{})
}
