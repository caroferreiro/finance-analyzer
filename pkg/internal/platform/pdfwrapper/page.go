package pdfwrapper

import (
	"fmt"
	"github.com/Alechan/pdf"
)

// newPageFromVendorPage creates a new Page from a pdf.Page. It returns the new Page and a boolean indicating if the page
// valid (it may be empty, for example, in which it should be ignored). If there was an unrecoverable error, it returns
// an error.
func newPageFromVendorPage(index int, vendorPage pdf.Page) (Page, bool, error) {
	if vendorPage.V.IsNull() {
		return Page{}, false, nil
	}

	vendorRows, err := vendorPage.GetTextByRow()
	if err != nil {
		return Page{}, false, fmt.Errorf("error getting text by row: %w", err)
	}

	var rows []Row
	for _, vRow := range vendorRows {
		row := newRowFromVendorRow(vRow)
		rows = append(rows, row)
	}

	return Page{
			Index: index,
			Rows:  rows,
		},
		true,
		nil
}

// Page is a simplified version of a "PDF Page". It contains the minimal information to parse the page.
type Page struct {
	// The Index is the number of the page in the PDF starting on 1 (page 1, page 2, etc.)
	Index int
	// The rows of the page. Each row contains the position of the row in the page and the text (multiple strings) of the row.
	Rows []Row
}
