package validation

import (
	"fmt"

	"github.com/Alechan/finance-analyzer/pkg/internal/pdfcardsummary"
)

// Validator validates CardSummary objects
type Validator struct{}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// Validate runs all validations on the CardSummary and returns the first error encountered
func (v *Validator) Validate(cs pdfcardsummary.CardSummary) error {
	if err := ValidateCloseDateBeforeExpirationDate(cs); err != nil {
		return fmt.Errorf("date validation failed: %w", err)
	}

	if err := ValidateMovementDatesWithinStatementPeriod(cs); err != nil {
		return fmt.Errorf("movement date validation failed: %w", err)
	}

	if err := ValidateDateRangeReasonableness(cs); err != nil {
		return fmt.Errorf("date range validation failed: %w", err)
	}

	if err := ValidateAtLeastOneCardRequired(cs); err != nil {
		return fmt.Errorf("structure validation failed: %w", err)
	}

	if err := ValidateSaldoAnteriorPresence(cs); err != nil {
		return fmt.Errorf("structure validation failed: %w", err)
	}

	if err := ValidateCardOwnerNonEmpty(cs); err != nil {
		return fmt.Errorf("structure validation failed: %w", err)
	}

	if err := ValidateCardMovementsHaveReceiptNumbers(cs); err != nil {
		return fmt.Errorf("structure validation failed: %w", err)
	}

	if err := ValidateTaxMovementsHaveNoReceiptNumbers(cs); err != nil {
		return fmt.Errorf("structure validation failed: %w", err)
	}

	if err := ValidateDocumentTotalsMatchComponents(cs); err != nil {
		return fmt.Errorf("amount validation failed: %w", err)
	}

	if err := ValidateCardTotalsMatchMovements(cs); err != nil {
		return fmt.Errorf("amount validation failed: %w", err)
	}

	if err := ValidateMovementsHaveAmounts(cs); err != nil {
		return fmt.Errorf("amount validation failed: %w", err)
	}

	return nil
}
