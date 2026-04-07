package financeengine

import (
	"strconv"

	"github.com/shopspring/decimal"
)

func formatDecimal(d decimal.Decimal) string {
	return d.StringFixed(2)
}

func intToString(v int) string {
	return strconv.Itoa(v)
}
