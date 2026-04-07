package mercadopago

import (
	"regexp"
	"strings"

	"github.com/shopspring/decimal"
)

var arsAmountRe = regexp.MustCompile(`-?\$\s*[\d.]+,\d{2}`)
var usdAmountRe = regexp.MustCompile(`US\$\s*[\d.]+,\d{2}`)

// ParseARSAmount converts an Argentine peso string like "$ 354.174,29" or "-$ 159.387,70"
// into a decimal.Decimal. Periods are thousands separators, comma is the decimal separator.
func ParseARSAmount(s string) decimal.Decimal {
	negative := strings.Contains(s, "-")
	s = strings.ReplaceAll(s, "$", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", ".")

	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Zero
	}
	if negative {
		d = d.Neg()
	}
	return d
}
