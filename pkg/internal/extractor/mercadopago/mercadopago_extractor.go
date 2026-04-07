package mercadopago

import (
	"fmt"
	"time"

	"github.com/Alechan/finance-analyzer/pkg/internal/pdfcardsummary"
	"github.com/Alechan/finance-analyzer/pkg/internal/platform/pdfwrapper"
)

// Extractor implements pdfcardsummaryio.Extractor for MercadoPago credit card statements.
type Extractor struct{}

func NewExtractor() *Extractor {
	return &Extractor{}
}

// ExtractFromBytes is the primary entry point. MercadoPago PDFs require plain-text extraction
// via the pdf library (the pdfwrapper row model does not preserve the layout).
func (e *Extractor) ExtractFromBytes(rawBytes []byte) (pdfcardsummary.CardSummary, error) {
	doc, err := NewDocumentFromBytes(rawBytes)
	if err != nil {
		return pdfcardsummary.CardSummary{}, fmt.Errorf("error parsing MercadoPago PDF: %w", err)
	}
	return doc.ToCardSummary(), nil
}

// ExtractFromDocument extracts from a pre-parsed pdfwrapper.Document.
// This path is less reliable for MercadoPago because the row-based model collapses all text
// into a single row with X=0; prefer ExtractFromBytes when possible.
func (e *Extractor) ExtractFromDocument(pdfDoc pdfwrapper.Document) (pdfcardsummary.CardSummary, error) {
	text := pdfcardsummary.ExtractAllTextFromDocument(pdfDoc)

	closeMonth, err := ExtractCloseMonth(text)
	if err != nil {
		return pdfcardsummary.CardSummary{}, fmt.Errorf("extracting close month: %w", err)
	}
	year := InferYearFromCloseMonth(closeMonth, time.Now())

	closeDate, err := extractCloseDate(text, year)
	if err != nil {
		return pdfcardsummary.CardSummary{}, fmt.Errorf("extracting close date: %w", err)
	}

	expirationDate, err := extractExpirationDate(text, year)
	if err != nil {
		return pdfcardsummary.CardSummary{}, fmt.Errorf("extracting expiration date: %w", err)
	}

	doc := Document{
		Bank:           pdfcardsummary.DetectBankFromText(text),
		CardCompany:    pdfcardsummary.DetectCardCompanyFromText(text),
		CloseDate:      closeDate,
		ExpirationDate: expirationDate,
		CardOwner:      extractCardOwner(text),
		PastPayments:   extractPastPayments(text, closeDate),
		CardMovements:  SplitMovementLines(extractConsumosSection(text), closeDate),
	}
	doc.TotalARS, doc.TotalUSD = extractTotals(text)

	return doc.ToCardSummary(), nil
}
