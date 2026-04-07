package visaprisma

import (
	"fmt"
	"github.com/shopspring/decimal"
	"strings"
)

func PDFAmountToDecimal(rawAmount string) (decimal.Decimal, error) {
	if rawAmount == "" {
		return decimal.Zero, nil
	}

	if strings.HasSuffix(rawAmount, "-") {
		rawAmount = "-" + strings.TrimSuffix(rawAmount, "-")
	}

	// Remove all dots "."
	res := strings.ReplaceAll(rawAmount, ".", "")
	// Replace comma "," with dot "."
	res = strings.ReplaceAll(res, ",", ".")
	asDec, err := decimal.NewFromString(res)
	if err != nil {
		return decimal.Zero, fmt.Errorf("error converting %s to decimal: %w", rawAmount, err)
	}

	// Converge all zeros to decimal.Zero
	if asDec.IsZero() {
		return decimal.Zero, nil
	}

	return asDec, err
}
