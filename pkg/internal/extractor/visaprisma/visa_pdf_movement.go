package visaprisma

import (
	"github.com/shopspring/decimal"
	"time"
)

type PDFMovement struct {
	OriginalDate       *time.Time
	ReceiptNumber      *string
	Detail             string
	CurrentInstallment *int
	TotalInstallments  *int
	AmountARS          decimal.Decimal
	AmountUSD          decimal.Decimal
}

func addAllMovementsAmounts(movements []PDFMovement) (decimal.Decimal, decimal.Decimal) {
	arsSum := decimal.Zero
	usdSum := decimal.Zero
	for _, m := range movements {
		arsSum = arsSum.Add(m.AmountARS)
		usdSum = usdSum.Add(m.AmountUSD)
	}

	return arsSum, usdSum
}
