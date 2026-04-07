package decimalale

import (
	"github.com/shopspring/decimal"
	"strings"
)

func FormatToFinancialString(dec decimal.Decimal) string {
	// Convert Decimal to string
	decStr := dec.StringFixed(2)

	// Split into integer and fractional parts
	parts := strings.Split(decStr, ".")
	intPart := parts[0]
	var fracPart string
	if len(parts) > 1 {
		fracPart = "." + parts[1]
	}

	// Insert thousands separators in the integer part
	var formattedIntPart string
	for i := len(intPart) - 1; i >= 0; i-- {
		formattedIntPart = string(intPart[i]) + formattedIntPart
		if (len(intPart)-i)%3 == 0 && i > 0 {
			formattedIntPart = " " + formattedIntPart
		}
	}
	if formattedIntPart[0] == '-' && formattedIntPart[1] == ' ' {
		formattedIntPart = "-" + formattedIntPart[2:]
	}

	// Combine formatted integer and fractional parts
	formattedDecimal := formattedIntPart + fracPart
	return formattedDecimal
}

func FormatToArgentineSeparators(dec decimal.Decimal) string {
	// Convert Decimal to string
	decStr := dec.StringFixed(2)

	// Split into integer and fractional parts
	parts := strings.Split(decStr, ".")
	intPart := parts[0]
	var fracPart string
	if len(parts) > 1 {
		fracPart = "," + parts[1]
	}

	// Insert thousands separators in the integer part
	var formattedIntPart string
	for i := len(intPart) - 1; i >= 0; i-- {
		formattedIntPart = string(intPart[i]) + formattedIntPart
		if (len(intPart)-i)%3 == 0 && i > 0 {
			formattedIntPart = "." + formattedIntPart
		}
	}
	if formattedIntPart[0] == '-' && formattedIntPart[1] == '.' {
		formattedIntPart = "-" + formattedIntPart[2:]
	}

	// Combine formatted integer and fractional parts
	formattedDecimal := formattedIntPart + fracPart
	return formattedDecimal
}
