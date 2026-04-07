package visaprisma

import (
	"github.com/Alechan/finance-analyzer/pkg/internal/platform/pointersale"
	"github.com/shopspring/decimal"
)

func NewPDFCard(number, owner string, movements []PDFMovement, totalARS, totalUSD decimal.Decimal) (PDFCard, error) {
	var numberPtr *string
	if number != "" {
		numberPtr = pointersale.ToPointer(number)
	}

	return PDFCard{
		Number:    numberPtr,
		Owner:     owner,
		Movements: movements,
		TotalARS:  totalARS,
		TotalUSD:  totalUSD,
	}, nil
}

type PDFCard struct {
	Number    *string
	Owner     string
	Movements []PDFMovement
	TotalARS  decimal.Decimal
	TotalUSD  decimal.Decimal
}

func addAllCardsAmounts(cards []PDFCard) (decimal.Decimal, decimal.Decimal) {
	arsSum := decimal.Zero
	usdSum := decimal.Zero
	for _, c := range cards {
		arsSum = arsSum.Add(c.TotalARS)
		usdSum = usdSum.Add(c.TotalUSD)
	}

	return arsSum, usdSum
}
