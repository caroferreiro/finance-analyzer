package pdfwrapper

import (
	"github.com/Alechan/pdf"
	"github.com/Alechan/finance-analyzer/pkg/internal/platform/sliceale"
)

func newRowFromVendorRow(vendorRow *pdf.Row) Row {
	text := sliceale.ApplyMapFunction(
		vendorRow.Content,
		func(word pdf.Text) string {
			return word.S
		},
	)
	return Row{
		Position: int(vendorRow.Position),
		Texts:    text,
	}
}

// Row is a simplified version of a "PDF Row". It just contains the text of the row, not the position, font, etc.
type Row struct {
	Position int
	Texts    []string
}
