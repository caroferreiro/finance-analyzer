package visaprisma

import (
	"fmt"

	"github.com/Alechan/finance-analyzer/pkg/internal/pdfcardsummary"
	"github.com/Alechan/finance-analyzer/pkg/internal/platform/pdfwrapper"
)

// VisaprismaExtractor implements the pdfcardsummaryio.Extractor interface
type VisaprismaExtractor struct{}

// NewVisaprismaExtractor creates a new VISA PRISMA extractor
func NewVisaprismaExtractor() *VisaprismaExtractor {
	return &VisaprismaExtractor{}
}

// ExtractFromBytes extracts a CardSummary from raw PDF bytes
func (e *VisaprismaExtractor) ExtractFromBytes(rawBytes []byte) (pdfcardsummary.CardSummary, error) {
	document, err := NewDocumentFromBytes(rawBytes)
	if err != nil {
		return pdfcardsummary.CardSummary{}, err
	}

	// Extract bank and card company from PDF text
	pdfReader := pdfwrapper.NewReaderWrapper()
	pdfDoc, err := pdfReader.ReadFromBytes(rawBytes)
	if err != nil {
		return pdfcardsummary.CardSummary{}, fmt.Errorf("error reading pdf for bank/card company detection: %w", err)
	}
	allText := pdfcardsummary.ExtractAllTextFromDocument(pdfDoc)
	bank := pdfcardsummary.DetectBankFromText(allText)
	cardCompany := pdfcardsummary.DetectCardCompanyFromText(allText)

	cardSummary := e.convertDocumentToCardSummary(document)
	cardSummary.StatementContext.Bank = bank
	cardSummary.StatementContext.CardCompany = cardCompany

	return cardSummary, nil
}

// ExtractFromDocument extracts a CardSummary from a pdfwrapper.Document
//
// NOTE: This method is currently NOT IMPLEMENTED and returns an error.
//
// The VISA PRISMA extractor's parsing logic (NewDocumentFromPDFRows) requires []*pdf.Row
// from github.com/Alechan/pdf directly, not pdfwrapper.Row. The pdfwrapper abstraction
// only provides simplified Row objects with Position and Texts []string, which loses
// the detailed formatting/positioning information needed by VISA PRISMA's parsing logic.
//
// TODO: Once VISA PRISMA extractor is migrated to use pdfwrapper abstraction (Phase 2),
// ExtractFromBytes() should call this method, similar to how SantanderExtractor.ExtractFromBytes()
// calls SantanderExtractor.ExtractFromDocument(). See pkg/internal/extractor/santander/santander_extractor.go
// for reference implementation.
//
// For now, use ExtractFromBytes() which works correctly with raw PDF bytes.
func (e *VisaprismaExtractor) ExtractFromDocument(pdfDoc pdfwrapper.Document) (pdfcardsummary.CardSummary, error) {
	return pdfcardsummary.CardSummary{}, fmt.Errorf("ExtractFromDocument not implemented for VISA PRISMA extractor. Use ExtractFromBytes() instead. See method comment for details")
}

// convertDocumentToCardSummary converts a VISA PRISMA Document to a CardSummary
func (e *VisaprismaExtractor) convertDocumentToCardSummary(document Document) pdfcardsummary.CardSummary {
	// Convert PDFCards to Cards
	var cards []pdfcardsummary.Card
	for _, pdfCard := range document.Cards {
		card := pdfcardsummary.Card{
			CardContext: pdfcardsummary.CardContext{
				CardNumber:   pdfCard.Number,
				CardOwner:    pdfCard.Owner,
				CardTotalARS: pdfCard.TotalARS,
				CardTotalUSD: pdfCard.TotalUSD,
			},
			Movements: e.convertPDFMovementsToMovements(pdfCard.Movements),
		}
		cards = append(cards, card)
	}

	// Convert PDFMovements to Movements
	pastPaymentMovements := e.convertPDFMovementsToMovements(document.PastPaymentMovements)
	taxesMovements := e.convertPDFMovementsToMovements(document.TaxesMovements)

	return pdfcardsummary.CardSummary{
		StatementContext: pdfcardsummary.StatementContext{
			TotalARS:       document.TotalARS,
			TotalUSD:       document.TotalUSD,
			CloseDate:      document.CloseDate,
			ExpirationDate: document.ExpirationDate,
		},
		Table: pdfcardsummary.Table{
			PastPaymentMovements: pastPaymentMovements,
			Cards:                cards,
			TaxesMovements:       taxesMovements,
		},
	}
}

// convertPDFMovementsToMovements converts PDFMovements to Movements
func (e *VisaprismaExtractor) convertPDFMovementsToMovements(pdfMovements []PDFMovement) []pdfcardsummary.Movement {
	var movements []pdfcardsummary.Movement
	for _, pdfMov := range pdfMovements {
		movement := pdfcardsummary.Movement{
			OriginalDate:       pdfMov.OriginalDate,
			ReceiptNumber:      pdfMov.ReceiptNumber,
			Detail:             pdfMov.Detail,
			CurrentInstallment: pdfMov.CurrentInstallment,
			TotalInstallments:  pdfMov.TotalInstallments,
			AmountARS:          pdfMov.AmountARS,
			AmountUSD:          pdfMov.AmountUSD,
		}
		movements = append(movements, movement)
	}
	return movements
}
