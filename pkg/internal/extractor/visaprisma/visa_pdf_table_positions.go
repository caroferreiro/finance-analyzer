package visaprisma

// PDFTablePositions are the known positions of the columns of the table of the PDF.
// They start at 0, so the first character is at position 0.
// Left whitespace is not counted, so the first character is the first non-whitespace character.
// All the positions are inclusive but some starts may point to whitespace.
type PDFTablePositions struct {
	// Dates always have the same length, so we can know the start and end of the date
	OriginalDateStart int
	OriginalDateEnd   int
	// Receipts are optional, but even if they are present or not, they are always positioned at the same place
	ReceiptStart int
	ReceiptEnd   int
	// Details are the most variable, so we can only know the start of the detail
	DetailStart int
	// Amounts grow to the left, so we can only know the end of the ARS amount for sure
	ARSAmountStart int
	ARSAmountEnd   int
	// We know the end of the ARS amount, so the next character is the start of the USD amount (probably a whitespace)
	USDAmountStart int
	USDAmountEnd   int
}
