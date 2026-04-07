package validation

import (
	"fmt"
	"strings"

	"github.com/Alechan/finance-analyzer/pkg/internal/extractor/santander"
	"github.com/Alechan/finance-analyzer/pkg/internal/pdfcardsummary"
)

// ValidateAtLeastOneCardRequired ensures the statement has at least one card
func ValidateAtLeastOneCardRequired(cs pdfcardsummary.CardSummary) error {
	if len(cs.Table.Cards) == 0 {
		return fmt.Errorf("no cards found in statement")
	}
	return nil
}

// ValidateSaldoAnteriorPresence ensures SALDO ANTERIOR is present and is the first past payment movement
func ValidateSaldoAnteriorPresence(cs pdfcardsummary.CardSummary) error {
	if len(cs.Table.PastPaymentMovements) == 0 {
		return fmt.Errorf("no past payment movements found")
	}

	firstMov := cs.Table.PastPaymentMovements[0]
	if firstMov.Detail != "SALDO ANTERIOR" {
		return fmt.Errorf("first past payment movement is not SALDO ANTERIOR, got: %s", firstMov.Detail)
	}

	return nil
}

// ValidateCardOwnerNonEmpty ensures all cards have a non-empty owner
func ValidateCardOwnerNonEmpty(cs pdfcardsummary.CardSummary) error {
	for i, card := range cs.Table.Cards {
		if strings.TrimSpace(card.CardContext.CardOwner) == "" {
			numberStr := "<nil>"
			if card.CardContext.CardNumber != nil {
				numberStr = *card.CardContext.CardNumber
			}
			return fmt.Errorf("card %d (number: %s) has empty owner", i, numberStr)
		}
	}
	return nil
}

// ValidateCardMovementsHaveReceiptNumbers ensures all card movements have receipt numbers.
// Anomaly adjustment movements (with detail exactly matching AnomalyAdjustmentDetail) are exempt from this requirement.
func ValidateCardMovementsHaveReceiptNumbers(cs pdfcardsummary.CardSummary) error {
	for cardIdx, card := range cs.Table.Cards {
		for movIdx, mov := range card.Movements {
			// Skip anomaly adjustment movements - they don't have receipt numbers
			if mov.Detail == santander.AnomalyAdjustmentDetail {
				continue
			}
			if mov.ReceiptNumber == nil {
				return fmt.Errorf("card %d (%s) movement %d (%s) does not have a receipt number",
					cardIdx, card.IdentifiableInfo(), movIdx, mov.IdentifiableInfo())
			}
		}
	}
	return nil
}

// ValidateTaxMovementsHaveNoReceiptNumbers ensures tax movements do not have receipt numbers
func ValidateTaxMovementsHaveNoReceiptNumbers(cs pdfcardsummary.CardSummary) error {
	for movIdx, mov := range cs.Table.TaxesMovements {
		if mov.ReceiptNumber != nil {
			return fmt.Errorf("tax movement %d (%s) should not have a receipt number, got: %s",
				movIdx, mov.IdentifiableInfo(), *mov.ReceiptNumber)
		}
	}
	return nil
}
