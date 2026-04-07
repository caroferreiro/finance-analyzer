package pdfcardsummary

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

func PDFAmountToDecimal(rawAmount string) (decimal.Decimal, error) {
	// NOTE: We must allow empty strings and treat them as decimal.Zero due to upstream data quirks (e.g., missing values in PDFs).
	if rawAmount == "" {
		return decimal.Zero, nil
	}

	// Move sign from the end to the beginning
	if strings.HasSuffix(rawAmount, "-") {
		rawAmount = "-" + strings.TrimSuffix(rawAmount, "-")
	}

	// Remove all dots "."
	res := strings.ReplaceAll(rawAmount, ".", "")
	// Replace comma "," with dot "."
	res = strings.ReplaceAll(res, ",", ".")
	asDec, err := decimal.NewFromString(res)
	if err != nil {
		return decimal.Zero, fmt.Errorf("error converting %s to decimal: %v", rawAmount, err)
	}

	// Converge all zeros to decimal.Zero (the "abstract 0" can be represented in many ways and it fails our tests)
	if asDec.IsZero() {
		return decimal.Zero, nil
	}

	return asDec, err
}
