package pdfcardsummary

import (
	"fmt"
	"github.com/Alechan/finance-analyzer/pkg/internal/platform/sliceale"
	"github.com/shopspring/decimal"
	"time"
)

func SumOfMovements(movements []Movement) map[string]decimal.Decimal {
	sum := sliceale.ApplyReduceFunction(
		movements,
		func(mov Movement, accum map[string]decimal.Decimal) map[string]decimal.Decimal {
			accum["ARS"] = accum["ARS"].Add(mov.AmountARS)
			accum["USD"] = accum["USD"].Add(mov.AmountUSD)

			return accum
		},
		map[string]decimal.Decimal{
			"ARS": decimal.Zero,
			"USD": decimal.Zero,
		},
	)

	return sum
}

type Movement struct {
	OriginalDate       *time.Time
	ReceiptNumber      *string
	Detail             string
	CurrentInstallment *int
	TotalInstallments  *int
	AmountARS          decimal.Decimal
	AmountUSD          decimal.Decimal
}

// IdentifiableInfo returns a string containing key identifying information about the movement.
// This is useful for logging, debugging, or displaying the movement in contexts where
// a brief identifier is needed. It does not include all movement fields.
func (m Movement) IdentifiableInfo() string {
	dateStr := "<nil>"
	if m.OriginalDate != nil {
		dateStr = m.OriginalDate.Format("2006-01-02")
	}
	receiptStr := "<nil>"
	if m.ReceiptNumber != nil {
		receiptStr = *m.ReceiptNumber
	}
	return fmt.Sprintf("date: %s, detail: %q, receipt: %s", dateStr, m.Detail, receiptStr)
}

// MovementExtension is a special kind of movement "extends" another movement. That is, it doesn't make sense on its own.
type MovementExtension struct {
	OriginalDateDay int
	Detail          string
}

// MovementWithOnlyDay is a special kind of movement that needs a "previous" movement to be complete because it only has the day
// and needs the year and month
type MovementWithOnlyDay struct {
	OriginalDateDay    int
	ReceiptNumber      *string
	Detail             string
	CurrentInstallment *int
	TotalInstallments  *int
	AmountARS          decimal.Decimal
	AmountUSD          decimal.Decimal
}
