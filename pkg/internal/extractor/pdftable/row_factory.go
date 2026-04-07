package pdftable

import (
	"strings"
)

// RowFactory is responsible for creating Row instances with specific table position configurations.
type RowFactory struct {
	positions PDFTablePositions
}

// NewRowFactory creates a new RowFactory with the given table positions.
func NewRowFactory(positions PDFTablePositions) *RowFactory {
	return &RowFactory{
		positions: positions,
	}
}

// CreateRow creates a new Row instance using the factory's stored positions and the provided text.
func (f *RowFactory) CreateRow(rawText string) Row {
	paddedText := getPaddedRawText(rawText, f.positions)

	originalDate := extractField(rawText, paddedText, f.positions.OriginalDateStart, f.positions.OriginalDateEnd)
	receiptNumber := extractField(rawText, paddedText, f.positions.ReceiptStart, f.positions.ReceiptEnd)
	detail := extractField(rawText, paddedText, f.positions.DetailStart, f.positions.DetailEnd)
	arsAmount := extractField(rawText, paddedText, f.positions.ARSAmountStart, f.positions.ARSAmountEnd)
	usdAmount := extractField(rawText, paddedText, f.positions.USDAmountStart, f.positions.USDAmountEnd)

	return NewRow(rawText, originalDate, receiptNumber, detail, arsAmount, usdAmount)
}

// getPaddedRawText returns a padded version of the input text that is long enough to accommodate all column positions.
// The padding is done by adding spaces at the end of the text if necessary.
func getPaddedRawText(rawText string, positions PDFTablePositions) string {
	// Determine the maximum end index needed
	maxEnd := positions.OriginalDateEnd
	if positions.ReceiptEnd > maxEnd {
		maxEnd = positions.ReceiptEnd
	}
	if positions.DetailEnd > maxEnd {
		maxEnd = positions.DetailEnd
	}
	if positions.ARSAmountEnd > maxEnd {
		maxEnd = positions.ARSAmountEnd
	}
	if positions.USDAmountEnd > maxEnd {
		maxEnd = positions.USDAmountEnd
	}

	// Pad rawText if necessary
	if len(rawText) < maxEnd+1 {
		return rawText + strings.Repeat(" ", maxEnd+1-len(rawText))
	}
	return rawText
}

// extractField extracts a field from paddedRawText, but only if the field's start is within the original (unpadded) text.
func extractField(originalText, paddedRawText string, startInclusive, endInclusive int) string {
	if startInclusive >= len(originalText) {
		return ""
	}
	endExclusive := endInclusive + 1
	if endExclusive > len(paddedRawText) {
		endExclusive = len(paddedRawText)
	}
	field := strings.TrimSpace(paddedRawText[startInclusive:endExclusive])
	if field == "" {
		return ""
	}
	return field
}
