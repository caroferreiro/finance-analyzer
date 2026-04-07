package validation

import (
	"fmt"
	"slices"

	"github.com/Alechan/finance-analyzer/pkg/internal/pdfcardsummary"
)

// ValidateDocumentTotalsMatchComponents ensures document totals match the sum of all components
func ValidateDocumentTotalsMatchComponents(cs pdfcardsummary.CardSummary) error {
	allMovements := slices.Concat(cs.Table.PastPaymentMovements, cs.Table.TaxesMovements)
	movementsSumMap := pdfcardsummary.SumOfMovements(allMovements)
	movementsARS, movementsUSD := movementsSumMap["ARS"], movementsSumMap["USD"]

	cardsARS, cardsUSD := pdfcardsummary.AddCardsAmounts(cs.Table.Cards)

	expectedARS := movementsARS.Add(cardsARS)
	expectedUSD := movementsUSD.Add(cardsUSD)

	if !cs.StatementContext.TotalARS.Equal(expectedARS) {
		difference := cs.StatementContext.TotalARS.Sub(expectedARS)
		return fmt.Errorf("document total ARS %s does not match sum %s (cards: %s + movements: %s) (difference: %s)",
			cs.StatementContext.TotalARS, expectedARS, cardsARS, movementsARS, difference)
	}

	if !cs.StatementContext.TotalUSD.Equal(expectedUSD) {
		difference := cs.StatementContext.TotalUSD.Sub(expectedUSD)
		return fmt.Errorf("document total USD %s does not match sum %s (cards: %s + movements: %s) (difference: %s)",
			cs.StatementContext.TotalUSD, expectedUSD, cardsUSD, movementsUSD, difference)
	}

	return nil
}

// ValidateCardTotalsMatchMovements ensures each card's totals match the sum of its movements
func ValidateCardTotalsMatchMovements(cs pdfcardsummary.CardSummary) error {
	for i, card := range cs.Table.Cards {
		movementsSumMap := pdfcardsummary.SumOfMovements(card.Movements)
		movementsARS := movementsSumMap["ARS"]
		movementsUSD := movementsSumMap["USD"]

		if !card.CardContext.CardTotalARS.Equal(movementsARS) {
			difference := card.CardContext.CardTotalARS.Sub(movementsARS)
			return fmt.Errorf("card %d (%s) total ARS %s does not match sum of movements %s (difference: %s)",
				i, card.IdentifiableInfo(), card.CardContext.CardTotalARS, movementsARS, difference)
		}

		if !card.CardContext.CardTotalUSD.Equal(movementsUSD) {
			difference := card.CardContext.CardTotalUSD.Sub(movementsUSD)
			return fmt.Errorf("card %d (%s) total USD %s does not match sum of movements %s (difference: %s)",
				i, card.IdentifiableInfo(), card.CardContext.CardTotalUSD, movementsUSD, difference)
		}
	}

	return nil
}

// ValidateMovementsHaveAmounts ensures all movements have at least one non-zero amount.
// SALDO ANTERIOR is an exception and can have zero amounts.
func ValidateMovementsHaveAmounts(cs pdfcardsummary.CardSummary) error {
	// Check card movements
	for cardIdx, card := range cs.Table.Cards {
		for movIdx, mov := range card.Movements {
			if mov.AmountARS.IsZero() && mov.AmountUSD.IsZero() {
				return fmt.Errorf("card %d (%s) movement %d (%s) has both ARS and USD amounts zero",
					cardIdx, card.IdentifiableInfo(), movIdx, mov.IdentifiableInfo())
			}
		}
	}

	// Check past payment movements (except SALDO ANTERIOR)
	for movIdx, mov := range cs.Table.PastPaymentMovements {
		if mov.Detail == "SALDO ANTERIOR" {
			continue // SALDO ANTERIOR can have zero amounts
		}
		if mov.AmountARS.IsZero() && mov.AmountUSD.IsZero() {
			return fmt.Errorf("past payment movement %d (%s) has both ARS and USD amounts zero",
				movIdx, mov.IdentifiableInfo())
		}
	}

	// Check tax movements
	for movIdx, mov := range cs.Table.TaxesMovements {
		if mov.AmountARS.IsZero() && mov.AmountUSD.IsZero() {
			return fmt.Errorf("tax movement %d (%s) has both ARS and USD amounts zero",
				movIdx, mov.IdentifiableInfo())
		}
	}

	return nil
}
