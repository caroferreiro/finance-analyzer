package pdfwrapper

import (
	"fmt"
	"github.com/Alechan/pdf"
	"strings"
)

func NewReaderWrapper() *ReaderWrapper {
	return &ReaderWrapper{}
}

// ReaderWrapper wraps the vendor pdf.Reader to provide a more convenient interface and encapsulate behavior that
// shouldn't be exposed to the client.
type ReaderWrapper struct {
}

// ReadFromBytes reads the PDF from a byte slice and returns the rows of the pages. It returns an error if it encounters
// any problem.
// You should first read the file into a byte slice and then call this function.
func (r *ReaderWrapper) ReadFromBytes(rawBytes []byte) (Document, error) {
	if rawBytes == nil || len(rawBytes) == 0 {
		return Document{}, fmt.Errorf("nil or empty raw bytes")
	}
	readerAt := strings.NewReader(string(rawBytes))
	size := int64(len(rawBytes))

	vendorReader, err := pdf.NewReader(readerAt, size)
	if err != nil {
		return Document{}, fmt.Errorf(
			"%w: %v",
			ErrCreatingVendorReader,
			err,
		)
	}

	allRows, err := extractAllPages(vendorReader)
	if err != nil {
		return Document{}, fmt.Errorf("error extracting all rows: %w", err)
	}
	return Document{
			Pages: allRows,
		},
		nil
}

func extractAllPages(r *pdf.Reader) ([]Page, error) {
	nPages := r.NumPage()

	var allPages []Page
	for pageIndex := 1; pageIndex <= nPages; pageIndex++ {
		p := r.Page(pageIndex)

		page, nonEmpty, err := newPageFromVendorPage(pageIndex, p)
		if err != nil {
			return nil, fmt.Errorf("error creating new page from vendor page: %w", err)
		}

		if nonEmpty {
			allPages = append(allPages, page)
		}
	}
	return allPages, nil
}
