package testdata

import (
	"testing"
	"time"

	"github.com/Alechan/finance-analyzer/pkg/internal/extractor/santander"
	"github.com/Alechan/finance-analyzer/pkg/internal/pdfcardsummary"
	"github.com/Alechan/finance-analyzer/pkg/internal/platform/testsale"
	"github.com/shopspring/decimal"
)

// CardSummaryBuilder helps build test CardSummary objects for testing
type CardSummaryBuilder struct {
	cs pdfcardsummary.CardSummary
	t  *testing.T
}

// NewCardSummaryBuilder creates a new builder with default valid values
func NewCardSummaryBuilder(t *testing.T) *CardSummaryBuilder {
	return &CardSummaryBuilder{
		t: t,
		cs: pdfcardsummary.CardSummary{
			StatementContext: pdfcardsummary.StatementContext{
				CloseDate:      time.Date(2024, time.August, 15, 0, 0, 0, 0, time.UTC),
				ExpirationDate: time.Date(2024, time.August, 23, 0, 0, 0, 0, time.UTC),
				TotalARS:       decimal.Zero,
				TotalUSD:       decimal.Zero,
				Bank:           pdfcardsummary.Bank("?"),
				CardCompany:    pdfcardsummary.CardCompany("?"),
			},
			Table: pdfcardsummary.Table{
				PastPaymentMovements: []pdfcardsummary.Movement{},
				Cards:                []pdfcardsummary.Card{},
				TaxesMovements:       []pdfcardsummary.Movement{},
			},
		},
	}
}

// WithCloseDate sets the close date
func (b *CardSummaryBuilder) WithCloseDate(year int, month time.Month, day int) *CardSummaryBuilder {
	b.cs.StatementContext.CloseDate = time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return b
}

// WithExpirationDate sets the expiration date
func (b *CardSummaryBuilder) WithExpirationDate(year int, month time.Month, day int) *CardSummaryBuilder {
	b.cs.StatementContext.ExpirationDate = time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return b
}

// WithTotalARS sets the total ARS amount
func (b *CardSummaryBuilder) WithTotalARS(amount string) *CardSummaryBuilder {
	b.cs.StatementContext.TotalARS = testsale.AsDecimal(b.t, amount)
	return b
}

// WithTotalUSD sets the total USD amount
func (b *CardSummaryBuilder) WithTotalUSD(amount string) *CardSummaryBuilder {
	b.cs.StatementContext.TotalUSD = testsale.AsDecimal(b.t, amount)
	return b
}

// WithBank sets the bank
func (b *CardSummaryBuilder) WithBank(bank pdfcardsummary.Bank) *CardSummaryBuilder {
	b.cs.StatementContext.Bank = bank
	return b
}

// WithCardCompany sets the card company
func (b *CardSummaryBuilder) WithCardCompany(cardCompany pdfcardsummary.CardCompany) *CardSummaryBuilder {
	b.cs.StatementContext.CardCompany = cardCompany
	return b
}

// WithSaldoAnterior adds a SALDO ANTERIOR movement as the first past payment movement
func (b *CardSummaryBuilder) WithSaldoAnterior(arsAmount, usdAmount string) *CardSummaryBuilder {
	mov := pdfcardsummary.Movement{
		Detail:    "SALDO ANTERIOR",
		AmountARS: testsale.AsDecimal(b.t, arsAmount),
		AmountUSD: testsale.AsDecimal(b.t, usdAmount),
	}
	b.cs.Table.PastPaymentMovements = append([]pdfcardsummary.Movement{mov}, b.cs.Table.PastPaymentMovements...)
	return b
}

// WithPastPaymentMovement adds a past payment movement
func (b *CardSummaryBuilder) WithPastPaymentMovement(date *time.Time, receipt string, detail string, arsAmount, usdAmount string) *CardSummaryBuilder {
	mov := pdfcardsummary.Movement{
		OriginalDate:  date,
		ReceiptNumber: testsale.StrPtr(receipt),
		Detail:        detail,
		AmountARS:     testsale.AsDecimal(b.t, arsAmount),
		AmountUSD:     testsale.AsDecimal(b.t, usdAmount),
	}
	b.cs.Table.PastPaymentMovements = append(b.cs.Table.PastPaymentMovements, mov)
	return b
}

// WithCard adds a new card to the statement
func (b *CardSummaryBuilder) WithCard(cardNumber string, owner string, totalARS, totalUSD string) *CardSummaryBuilder {
	card := pdfcardsummary.Card{
		CardContext: pdfcardsummary.CardContext{
			CardNumber:   testsale.StrPtr(cardNumber),
			CardOwner:    owner,
			CardTotalARS: testsale.AsDecimal(b.t, totalARS),
			CardTotalUSD: testsale.AsDecimal(b.t, totalUSD),
		},
		Movements: []pdfcardsummary.Movement{},
	}
	b.cs.Table.Cards = append(b.cs.Table.Cards, card)
	return b
}

// WithCardMovement adds a movement to the specified card
func (b *CardSummaryBuilder) WithCardMovement(cardIndex int, date *time.Time, receipt string, detail string, arsAmount, usdAmount string) *CardSummaryBuilder {
	return b.WithCardMovementWithInstallments(cardIndex, date, receipt, detail, nil, nil, arsAmount, usdAmount)
}

// WithCardMovementWithInstallments adds a movement with installments to the specified card
func (b *CardSummaryBuilder) WithCardMovementWithInstallments(cardIndex int, date *time.Time, receipt string, detail string, currentInstallment, totalInstallments *int, arsAmount, usdAmount string) *CardSummaryBuilder {
	mov := pdfcardsummary.Movement{
		OriginalDate:       date,
		ReceiptNumber:      testsale.StrPtr(receipt),
		Detail:             detail,
		CurrentInstallment: currentInstallment,
		TotalInstallments:  totalInstallments,
		AmountARS:          testsale.AsDecimal(b.t, arsAmount),
		AmountUSD:          testsale.AsDecimal(b.t, usdAmount),
	}
	b.cs.Table.Cards[cardIndex].Movements = append(b.cs.Table.Cards[cardIndex].Movements, mov)
	return b
}

// WithTaxMovement adds a tax movement
func (b *CardSummaryBuilder) WithTaxMovement(date *time.Time, detail string, arsAmount, usdAmount string) *CardSummaryBuilder {
	mov := pdfcardsummary.Movement{
		OriginalDate:  date,
		ReceiptNumber: nil, // Taxes typically don't have receipt numbers
		Detail:        detail,
		AmountARS:     testsale.AsDecimal(b.t, arsAmount),
		AmountUSD:     testsale.AsDecimal(b.t, usdAmount),
	}
	b.cs.Table.TaxesMovements = append(b.cs.Table.TaxesMovements, mov)
	return b
}

// WithAnomalyAdjustmentMovement adds an anomaly adjustment movement to the specified card
// This is used to represent adjustments for glitched PDFs where decimal values are missing
func (b *CardSummaryBuilder) WithAnomalyAdjustmentMovement(cardIndex int, arsAmount, usdAmount string) *CardSummaryBuilder {
	mov := pdfcardsummary.Movement{
		OriginalDate:       nil,
		ReceiptNumber:      nil,
		Detail:             santander.AnomalyAdjustmentDetail,
		CurrentInstallment: nil,
		TotalInstallments:  nil,
		AmountARS:          testsale.AsDecimal(b.t, arsAmount),
		AmountUSD:          testsale.AsDecimal(b.t, usdAmount),
	}
	b.cs.Table.Cards[cardIndex].Movements = append(b.cs.Table.Cards[cardIndex].Movements, mov)
	return b
}

// Build returns the constructed CardSummary
func (b *CardSummaryBuilder) Build() pdfcardsummary.CardSummary {
	return b.cs
}

// BuildCardSummary is a convenience function for building CardSummary with a function
func BuildCardSummary(t *testing.T, fn func(*CardSummaryBuilder)) pdfcardsummary.CardSummary {
	builder := NewCardSummaryBuilder(t)
	fn(builder)
	return builder.Build()
}
