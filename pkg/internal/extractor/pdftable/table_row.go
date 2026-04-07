package pdftable

import (
	"regexp"
	"strconv"
	"strings"
)

func NewRow(rawText, originalDate, receiptNumber, detail, arsAmount, usdAmount string) Row {
	return Row{
		RawText:                        rawText,
		RawOriginalDate:                originalDate,
		RawReceiptNumber:               receiptNumber,
		RawDetailWithMaybeInstallments: detail,
		RawAmountARS:                   arsAmount,
		RawAmountUSD:                   usdAmount,
	}
}

type Row struct {
	// Full text of the row
	RawText string

	// The columns of the row
	RawOriginalDate                string
	RawReceiptNumber               string
	RawDetailWithMaybeInstallments string
	RawAmountARS                   string
	RawAmountUSD                   string
}

var receiptNumberPattern = regexp.MustCompile(`^\d+\s*\*?\s*$`)
var amountPattern = regexp.MustCompile(`^[\d.]+,\d{2}\s*\*?\s*$|^[\d.]+,\d{2}-?\s*$`)

func (m Row) MatchesMovementWithoutYearAndMonth() bool {
	// Check for 2-digit day (existing behavior)
	if len(m.RawOriginalDate) == 2 && canBeConvertedToInt(m.RawOriginalDate) {
		return true
	}

	// Check for empty date with movement data (continuation line)
	if m.RawOriginalDate == "" {
		receipt := strings.TrimSpace(m.RawReceiptNumber)
		arsAmount := strings.TrimSpace(m.RawAmountARS)
		usdAmount := strings.TrimSpace(m.RawAmountUSD)

		// For continuation lines, we require a valid receipt number (digits + optional asterisk)
		// This prevents informational text lines from being matched as movements
		hasValidReceipt := receipt != "" && receiptNumberPattern.MatchString(receipt)
		if !hasValidReceipt {
			return false
		}

		// Must have at least one valid amount (digits, decimal separator, optional asterisk/minus)
		hasValidAmounts := (arsAmount != "" && amountPattern.MatchString(arsAmount)) ||
			(usdAmount != "" && amountPattern.MatchString(usdAmount))

		return hasValidAmounts
	}

	return false
}

func canBeConvertedToInt(rawDate string) bool {
	_, err := strconv.Atoi(rawDate)
	return err == nil
}
