package pdftable

// PDFTablePositions are the known positions of the columns of the table of the PDF.
// They start at 0, so the first character is at position 0.
// All the positions are inclusive
type PDFTablePositions struct {
	OriginalDateStart int
	OriginalDateEnd   int

	ReceiptStart int
	ReceiptEnd   int

	DetailStart int
	DetailEnd   int

	ARSAmountStart int
	ARSAmountEnd   int

	USDAmountStart int
	USDAmountEnd   int
}
