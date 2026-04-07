package validation

import (
	"fmt"
	"time"

	"github.com/Alechan/finance-analyzer/pkg/internal/pdfcardsummary"
	"github.com/Alechan/finance-analyzer/pkg/internal/platform/timeale"
)

const (
	// minValidYear is the minimum valid year for dates in card summaries
	// Years before 2000 are considered invalid for credit card statements
	minValidYear = 2000

	// maxValidYear is the maximum valid year for dates in card summaries
	// Years more than 10 years in the future are considered invalid
	maxValidYearOffset = 10
)

// ValidateCloseDateBeforeExpirationDate ensures the close date is before or equal to the expiration date
func ValidateCloseDateBeforeExpirationDate(cs pdfcardsummary.CardSummary) error {
	if cs.StatementContext.CloseDate.After(cs.StatementContext.ExpirationDate) {
		daysDiff := timeale.DaysBetweenDates(cs.StatementContext.ExpirationDate, cs.StatementContext.CloseDate)
		return fmt.Errorf("close date %s is after expiration date %s (difference: %d days)",
			cs.StatementContext.CloseDate.Format("2006-01-02"), cs.StatementContext.ExpirationDate.Format("2006-01-02"), daysDiff)
	}
	return nil
}

// ValidateMovementDatesWithinStatementPeriod ensures all movement dates are within a reasonable range
// of the close date (typically within 30 days before or after)
func ValidateMovementDatesWithinStatementPeriod(cs pdfcardsummary.CardSummary) error {
	maxDaysAfterClose := 30 * 24 * time.Hour

	// Check card movements
	for cardIdx, card := range cs.Table.Cards {
		for movIdx, mov := range card.Movements {
			if mov.OriginalDate == nil {
				continue // nil dates are valid for SALDO ANTERIOR
			}

			if mov.OriginalDate.After(cs.StatementContext.CloseDate.Add(maxDaysAfterClose)) {
				return fmt.Errorf("card %d (%s) movement %d (%s) date %s is more than 30 days after close date %s",
					cardIdx, card.IdentifiableInfo(), movIdx, mov.IdentifiableInfo(),
					mov.OriginalDate.Format("2006-01-02"), cs.StatementContext.CloseDate.Format("2006-01-02"))
			}
		}
	}

	// Check past payment movements
	for movIdx, mov := range cs.Table.PastPaymentMovements {
		if mov.OriginalDate == nil {
			continue // nil dates are valid for SALDO ANTERIOR
		}

		if mov.OriginalDate.After(cs.StatementContext.CloseDate.Add(maxDaysAfterClose)) {
			return fmt.Errorf("past payment movement %d (%s) date %s is more than 30 days after close date %s",
				movIdx, mov.IdentifiableInfo(),
				mov.OriginalDate.Format("2006-01-02"), cs.StatementContext.CloseDate.Format("2006-01-02"))
		}
	}

	// Check tax movements
	for movIdx, mov := range cs.Table.TaxesMovements {
		if mov.OriginalDate == nil {
			continue
		}

		if mov.OriginalDate.After(cs.StatementContext.CloseDate.Add(maxDaysAfterClose)) {
			return fmt.Errorf("tax movement %d (%s) date %s is more than 30 days after close date %s",
				movIdx, mov.IdentifiableInfo(),
				mov.OriginalDate.Format("2006-01-02"), cs.StatementContext.CloseDate.Format("2006-01-02"))
		}
	}

	return nil
}

// ValidateDateRangeReasonableness ensures dates are within reasonable bounds
func ValidateDateRangeReasonableness(cs pdfcardsummary.CardSummary) error {
	now := time.Now()
	minValidDate := time.Date(minValidYear, 1, 1, 0, 0, 0, 0, time.UTC)
	maxValidDate := now.AddDate(maxValidYearOffset, 0, 0)

	if cs.StatementContext.CloseDate.Before(minValidDate) {
		return fmt.Errorf("close date %s is before minimum valid year %d", cs.StatementContext.CloseDate.Format("2006-01-02"), minValidYear)
	}

	if cs.StatementContext.CloseDate.After(maxValidDate) {
		return fmt.Errorf("close date %s is more than %d years in the future", cs.StatementContext.CloseDate.Format("2006-01-02"), maxValidYearOffset)
	}

	if cs.StatementContext.ExpirationDate.Before(minValidDate) {
		return fmt.Errorf("expiration date %s is before minimum valid year %d", cs.StatementContext.ExpirationDate.Format("2006-01-02"), minValidYear)
	}

	if cs.StatementContext.ExpirationDate.After(maxValidDate) {
		return fmt.Errorf("expiration date %s is more than %d years in the future", cs.StatementContext.ExpirationDate.Format("2006-01-02"), maxValidYearOffset)
	}

	// Validate card movement dates
	for cardIdx, card := range cs.Table.Cards {
		for movIdx, mov := range card.Movements {
			if mov.OriginalDate == nil {
				continue
			}

			if mov.OriginalDate.Year() < minValidYear {
				return fmt.Errorf("card %d (%s) movement %d (%s) date %s has year %d which is before minimum valid year %d",
					cardIdx, card.IdentifiableInfo(), movIdx, mov.IdentifiableInfo(),
					mov.OriginalDate.Format("2006-01-02"), mov.OriginalDate.Year(), minValidYear)
			}

			if mov.OriginalDate.Year() > now.Year()+maxValidYearOffset {
				return fmt.Errorf("card %d (%s) movement %d (%s) date %s has year %d which is more than %d years in the future",
					cardIdx, card.IdentifiableInfo(), movIdx, mov.IdentifiableInfo(),
					mov.OriginalDate.Format("2006-01-02"), mov.OriginalDate.Year(), maxValidYearOffset)
			}
		}
	}

	// Validate past payment movement dates
	for movIdx, mov := range cs.Table.PastPaymentMovements {
		if mov.OriginalDate == nil {
			continue
		}

		if mov.OriginalDate.Year() < minValidYear {
			return fmt.Errorf("past payment movement %d (%s) date %s has year %d which is before minimum valid year %d",
				movIdx, mov.IdentifiableInfo(),
				mov.OriginalDate.Format("2006-01-02"), mov.OriginalDate.Year(), minValidYear)
		}

		if mov.OriginalDate.Year() > now.Year()+maxValidYearOffset {
			return fmt.Errorf("past payment movement %d (%s) date %s has year %d which is more than %d years in the future",
				movIdx, mov.IdentifiableInfo(),
				mov.OriginalDate.Format("2006-01-02"), mov.OriginalDate.Year(), maxValidYearOffset)
		}
	}

	// Validate tax movement dates
	for movIdx, mov := range cs.Table.TaxesMovements {
		if mov.OriginalDate == nil {
			continue
		}

		if mov.OriginalDate.Year() < minValidYear {
			return fmt.Errorf("tax movement %d (%s) date %s has year %d which is before minimum valid year %d",
				movIdx, mov.IdentifiableInfo(),
				mov.OriginalDate.Format("2006-01-02"), mov.OriginalDate.Year(), minValidYear)
		}

		if mov.OriginalDate.Year() > now.Year()+maxValidYearOffset {
			return fmt.Errorf("tax movement %d (%s) date %s has year %d which is more than %d years in the future",
				movIdx, mov.IdentifiableInfo(),
				mov.OriginalDate.Format("2006-01-02"), mov.OriginalDate.Year(), maxValidYearOffset)
		}
	}

	return nil
}

// getAllMovements returns all movements from past payments, cards, and taxes
func getAllMovements(cs pdfcardsummary.CardSummary) []pdfcardsummary.Movement {
	var allMovements []pdfcardsummary.Movement

	allMovements = append(allMovements, cs.Table.PastPaymentMovements...)
	allMovements = append(allMovements, cs.Table.TaxesMovements...)

	for _, card := range cs.Table.Cards {
		allMovements = append(allMovements, card.Movements...)
	}

	return allMovements
}
