package pdfcardsummary

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type CardBuilder struct {
	Number    *string
	Owner     *string
	Movements []Movement
	TotalARS  *decimal.Decimal
	TotalUSD  *decimal.Decimal
}

func NewCardBuilder() *CardBuilder {
	return &CardBuilder{}
}

func (cb *CardBuilder) SetNumber(number *string) {
	if number != nil && *number == "" {
		// Unify "" to nil
		number = nil
	}

	cb.Number = number
}

func (cb *CardBuilder) SetOwner(owner string) {
	withoutSpaces := strings.TrimSpace(owner)
	cb.Owner = &withoutSpaces
}

func (cb *CardBuilder) AddMovement(m Movement) {
	cb.Movements = append(cb.Movements, m)
}

func (cb *CardBuilder) SetTotalARS(rawTotalARS string) error {
	totalARS, err := PDFAmountToDecimal(rawTotalARS)
	if err != nil {
		return err
	}
	cb.TotalARS = &totalARS
	return nil
}

func (cb *CardBuilder) SetTotalUSD(rawTotalUSD string) error {
	totalUSD, err := PDFAmountToDecimal(rawTotalUSD)
	if err != nil {
		return err
	}
	cb.TotalUSD = &totalUSD
	return nil
}

func (cb *CardBuilder) Build() (Card, error) {
	if cb.Owner == nil {
		return Card{}, fmt.Errorf("card owner is required")
	}
	if len(cb.Movements) == 0 {
		return Card{}, fmt.Errorf("card movements are required")
	}
	if cb.TotalARS == nil {
		return Card{}, fmt.Errorf("card total ARS is required")
	}
	if cb.TotalUSD == nil {
		return Card{}, fmt.Errorf("card total USD is required")
	}

	return Card{
		CardContext: CardContext{
			CardNumber:   cb.Number,
			CardOwner:    *cb.Owner,
			CardTotalARS: *cb.TotalARS,
			CardTotalUSD: *cb.TotalUSD,
		},
		Movements: cb.Movements,
	}, nil
}

func (cb *CardBuilder) GetLastMovementDate() (time.Time, error) {

	if len(cb.Movements) == 0 {
		return time.Time{}, fmt.Errorf("can't take the last movement date from an empty list of movements")
	}

	lastMovDatePtr := cb.Movements[len(cb.Movements)-1].OriginalDate
	if lastMovDatePtr == nil {
		return time.Time{}, fmt.Errorf("last movement date is nil")
	}

	return *lastMovDatePtr, nil
}

// MovementsAccumulated returns a list of movements that are currently accumulated in the builder.
// It should be used when we've finished reading all the cards but we still have movements to process.
// These movements are automatically assumed to be taxes movements.
func (cb *CardBuilder) MovementsAccumulated() []Movement {
	return cb.Movements
}
