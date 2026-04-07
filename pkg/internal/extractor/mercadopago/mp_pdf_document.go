package mercadopago

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Alechan/finance-analyzer/pkg/internal/pdfcardsummary"
	"github.com/Alechan/pdf"
	"github.com/shopspring/decimal"
)

// Document is the intermediate representation of a parsed MercadoPago credit card statement.
// It holds all extracted fields before they are mapped to the canonical pdfcardsummary.CardSummary.
type Document struct {
	Bank           pdfcardsummary.Bank
	CardCompany    pdfcardsummary.CardCompany
	CloseDate      time.Time
	ExpirationDate time.Time
	TotalARS       decimal.Decimal
	TotalUSD       decimal.Decimal
	CardOwner      string
	PastPayments   []pdfcardsummary.Movement
	CardMovements  []pdfcardsummary.Movement
}

// NewDocumentFromBytes reads a PDF from raw bytes and extracts all statement data.
func NewDocumentFromBytes(rawBytes []byte) (Document, error) {
	text, err := extractPlainText(rawBytes)
	if err != nil {
		return Document{}, fmt.Errorf("extracting plain text: %w", err)
	}
	return NewDocumentFromText(text)
}

// NewDocumentFromText builds a Document by parsing the concatenated plain text of a statement PDF.
func NewDocumentFromText(text string) (Document, error) {
	closeMonth, err := ExtractCloseMonth(text)
	if err != nil {
		return Document{}, fmt.Errorf("extracting close month: %w", err)
	}
	year := InferYearFromCloseMonth(closeMonth, time.Now())

	closeDate, err := extractCloseDate(text, year)
	if err != nil {
		return Document{}, fmt.Errorf("extracting close date: %w", err)
	}

	expirationDate, err := extractExpirationDate(text, year)
	if err != nil {
		return Document{}, fmt.Errorf("extracting expiration date: %w", err)
	}

	totalARS, totalUSD := extractTotals(text)
	bank := pdfcardsummary.DetectBankFromText(text)
	cardCompany := pdfcardsummary.DetectCardCompanyFromText(text)
	cardOwner := extractCardOwner(text)
	pastPayments := extractPastPayments(text, closeDate)
	consumosText := extractConsumosSection(text)
	cardMovements := SplitMovementLines(consumosText, closeDate)

	return Document{
		Bank:           bank,
		CardCompany:    cardCompany,
		CloseDate:      closeDate,
		ExpirationDate: expirationDate,
		TotalARS:       totalARS,
		TotalUSD:       totalUSD,
		CardOwner:      cardOwner,
		PastPayments:   pastPayments,
		CardMovements:  cardMovements,
	}, nil
}

// ToCardSummary maps the intermediate Document to the canonical pdfcardsummary model.
func (d Document) ToCardSummary() pdfcardsummary.CardSummary {
	movementsARS, movementsUSD := sumMovements(d.CardMovements)

	return pdfcardsummary.CardSummary{
		StatementContext: pdfcardsummary.StatementContext{
			Bank:           d.Bank,
			CardCompany:    d.CardCompany,
			CloseDate:      d.CloseDate,
			ExpirationDate: d.ExpirationDate,
			TotalARS:       d.TotalARS,
			TotalUSD:       d.TotalUSD,
		},
		Table: pdfcardsummary.Table{
			PastPaymentMovements: d.PastPayments,
			Cards: []pdfcardsummary.Card{{
				CardContext: pdfcardsummary.CardContext{
					CardOwner:    d.CardOwner,
					CardTotalARS: movementsARS,
					CardTotalUSD: movementsUSD,
				},
				Movements: d.CardMovements,
			}},
		},
	}
}

func extractPlainText(rawBytes []byte) (string, error) {
	reader := strings.NewReader(string(rawBytes))
	r, err := pdf.NewReader(reader, int64(len(rawBytes)))
	if err != nil {
		return "", fmt.Errorf("opening PDF: %w", err)
	}

	var sb strings.Builder
	for p := 1; p <= r.NumPage(); p++ {
		page := r.Page(p)
		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		sb.WriteString(text)
		sb.WriteString("\n")
	}
	return sb.String(), nil
}

func extractCloseDate(text string, year int) (time.Time, error) {
	if t, err := ParseFullSpanishDate(text, "Cierre actual", year); err == nil {
		return t, nil
	}
	return ParseFullSpanishDate(text, "Fecha de cierre", year)
}

func extractExpirationDate(text string, year int) (time.Time, error) {
	if t, err := ParseFullSpanishDate(text, "Vencimiento actual", year); err == nil {
		return t, nil
	}
	return ParseFullSpanishDate(text, "Fecha de vencimiento", year)
}

// extractTotals locates "Total a pagar" and reads the ARS and optional USD amounts after it.
// The first occurrence on page 1 often uses a display-only format without proper separators,
// so we use the last occurrence which is inside the movement detail pages.
func extractTotals(text string) (decimal.Decimal, decimal.Decimal) {
	totalRe := regexp.MustCompile(`(?i)Total a pagar`)
	allLocs := totalRe.FindAllStringIndex(text, -1)
	if len(allLocs) == 0 {
		return decimal.Zero, decimal.Zero
	}

	loc := allLocs[len(allLocs)-1]
	after := text[loc[1]:]
	if len(after) > 200 {
		after = after[:200]
	}

	totalARS := decimal.Zero
	if m := arsAmountRe.FindString(after); m != "" {
		totalARS = ParseARSAmount(m)
	}
	totalUSD := decimal.Zero
	if m := usdAmountRe.FindString(after); m != "" {
		totalUSD = ParseARSAmount(strings.TrimPrefix(m, "US"))
	}

	return totalARS, totalUSD
}

func extractCardOwner(text string) string {
	re := regexp.MustCompile(`(?i)pertenece a\s+([^,]+),\s*CUIL`)
	m := re.FindStringSubmatch(text)
	if m != nil {
		return strings.TrimSpace(m[1])
	}
	return ""
}

func extractPastPayments(text string, closeDate time.Time) []pdfcardsummary.Movement {
	var movements []pdfcardsummary.Movement

	prevRe := regexp.MustCompile(`(\d{1,2}/[a-zA-Záéíóú]{3})\s*Resumen de \w+\s*(\$\s*[\d.]+,\d{2})`)
	if m := prevRe.FindStringSubmatch(text); m != nil {
		date := ParseShortDate(m[1], closeDate)
		movements = append(movements, pdfcardsummary.Movement{
			OriginalDate: date,
			Detail:       "SALDO ANTERIOR",
			AmountARS:    ParseARSAmount(m[2]),
		})
	}

	payRe := regexp.MustCompile(`(\d{1,2}/[a-zA-Záéíóú]{3})\s*Pago de tarjeta\s*(-\$\s*[\d.]+,\d{2})`)
	for _, m := range payRe.FindAllStringSubmatch(text, -1) {
		date := ParseShortDate(m[1], closeDate)
		movements = append(movements, pdfcardsummary.Movement{
			OriginalDate: date,
			Detail:       "PAGO DE TARJETA",
			AmountARS:    ParseARSAmount(m[2]),
		})
	}

	return movements
}

// extractConsumosSection returns the raw text of all card movements from the "Consumos" section.
// Handles multi-page statements where continuation pages repeat "DETALLE DE MOVIMIENTOS" headers.
func extractConsumosSection(text string) string {
	sectionStartRe := regexp.MustCompile(`(?i)Con tarjeta virtual`)
	firstDateRe := regexp.MustCompile(`\d{1,2}/[a-záéíóú]{3}`)

	loc := sectionStartRe.FindStringIndex(text)
	if loc == nil {
		return ""
	}

	after := text[loc[1]:]
	dateLoc := firstDateRe.FindStringIndex(after)
	if dateLoc == nil {
		return ""
	}
	sectionBody := after[dateLoc[0]:]

	subtotalRe := regexp.MustCompile(`(?i)Subtotal`)
	allSubtotals := subtotalRe.FindAllStringIndex(sectionBody, -1)
	if len(allSubtotals) > 0 {
		sectionBody = sectionBody[:allSubtotals[0][0]]
	}

	continuationRe := regexp.MustCompile(`(?i)DETALLE DE MOVIMIENTOS[A-Za-záéíóúñÁÉÍÓÚÑ\s]*(?:Pesos|D[oó]lares)`)
	sectionBody = continuationRe.ReplaceAllString(sectionBody, "")

	return sectionBody
}

func sumMovements(movements []pdfcardsummary.Movement) (decimal.Decimal, decimal.Decimal) {
	ars, usd := decimal.Zero, decimal.Zero
	for _, m := range movements {
		ars = ars.Add(m.AmountARS)
		usd = usd.Add(m.AmountUSD)
	}
	return ars, usd
}
