package pdfcardsummary

import (
	"fmt"

	"github.com/shopspring/decimal"
)

func AddCardsAmounts(cards []Card) (decimal.Decimal, decimal.Decimal) {
	var cardsARS, cardsUSD decimal.Decimal
	for _, card := range cards {
		cardsARS = cardsARS.Add(card.CardContext.CardTotalARS)
		cardsUSD = cardsUSD.Add(card.CardContext.CardTotalUSD)
	}
	return cardsARS, cardsUSD
}

type Card struct {
	CardContext CardContext
	Movements   []Movement
}

// CardContext is the card-level context repeated on every extracted CSV row when MovementType is CardMovement.
type CardContext struct {
	CardNumber   *string
	CardOwner    string
	CardTotalARS decimal.Decimal
	CardTotalUSD decimal.Decimal
}

// IdentifiableInfo returns a string containing key identifying information about the card.
// This is useful for logging, debugging, or displaying the card in contexts where
// a brief identifier is needed. It does not include all card fields.
func (c Card) IdentifiableInfo() string {
	numberStr := "<nil>"
	if c.CardContext.CardNumber != nil {
		numberStr = *c.CardContext.CardNumber
	}
	return fmt.Sprintf("owner: %q, number: %s, movements: %d", c.CardContext.CardOwner, numberStr, len(c.Movements))
}
