package santander

import (
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/Alechan/finance-analyzer/pkg/internal/pdfcardsummary"
	"github.com/Alechan/finance-analyzer/pkg/internal/platform/timeale"
	"github.com/shopspring/decimal"

	"github.com/Alechan/finance-analyzer/pkg/internal/extractor/extractorfuncs"
	"github.com/Alechan/finance-analyzer/pkg/internal/extractor/pdftable"
	"github.com/Alechan/finance-analyzer/pkg/internal/platform/pdfwrapper"
)

const (
	// AnomalyAdjustmentDetail is the detail text for anomaly adjustment movements
	AnomalyAdjustmentDetail = "WARNING: AJUSTE POR FILA DESALINEADA - DECIMALES PERDIDOS"
	// maxAnomalyAdjustmentARS is the maximum ARS amount difference allowed for anomaly adjustment
	maxAnomalyAdjustmentARS = 1
	// maxAnomalyAdjustmentUSD is the maximum USD amount difference allowed for anomaly adjustment
	maxAnomalyAdjustmentUSD = 0.01
)

// Currency represents a currency type
type Currency string

const (
	CurrencyARS Currency = "ARS"
	CurrencyUSD Currency = "USD"
)

// AnomalyAdjustment represents a calculated anomaly adjustment
type AnomalyAdjustment struct {
	Amount      decimal.Decimal
	Currency    Currency
	ShouldApply bool
}

func NewSantanderExtractorFromDefaultCfg() *SantanderExtractor {
	cfg := DefaultConfig()
	reader := NewSantanderExtractor(cfg)
	return reader
}

func NewSantanderExtractor(cfg SantanderExtractorConfig) *SantanderExtractor {
	// TODO: use IoC and receive the rowFactory as a parameter
	rowFactory := pdftable.NewRowFactory(cfg.TableColumnPositions)
	return &SantanderExtractor{
		cfg:                     cfg,
		closingDateExtractor:    extractorfuncs.NewFirstMatchExtractor(cfg.ClosingDateRegex, timeale.CardSummarySpanishMonthDateToTime),
		expirationDateExtractor: extractorfuncs.NewFirstMatchExtractor(cfg.ExpirationDateRegex, timeale.CardSummarySpanishMonthDateToTime),
		tableExtractor:          NewSantanderTableExtractor(cfg),
		tableIteratorFactory:    pdftable.NewTableIteratorFactory(rowFactory),
	}
}

// SantanderExtractor extracts the Card Summary from Santander PDFs
// The inspection is ignored because having the name in the structs helps to disambiguate the files between different extractors
//
//goland:noinspection ALL
type SantanderExtractor struct {
	cfg                     SantanderExtractorConfig
	closingDateExtractor    *extractorfuncs.FirstMatchExtractor[time.Time]
	expirationDateExtractor *extractorfuncs.FirstMatchExtractor[time.Time]
	tableExtractor          *SantanderTableExtractor
	tableIteratorFactory    *pdftable.TableIteratorFactory
}

func (r *SantanderExtractor) ExtractFromBytes(rawBytes []byte) (pdfcardsummary.CardSummary, error) {
	pdfReader := pdfwrapper.NewReaderWrapper()
	pdfDoc, err := pdfReader.ReadFromBytes(rawBytes)
	if err != nil {
		return pdfcardsummary.CardSummary{}, fmt.Errorf("error reading pdf: %w", err)
	}
	return r.ExtractFromDocument(pdfDoc)
}

func (r *SantanderExtractor) ExtractFromDocument(pdfDoc pdfwrapper.Document) (pdfcardsummary.CardSummary, error) {
	pages := pdfDoc.Pages
	// We're going to read the PDF rows in order and extract the information piece by piece
	closingDate, err := r.closingDateExtractor.ExtractFirstMatch(pages)
	if err != nil {
		return pdfcardsummary.CardSummary{}, fmt.Errorf("error finding the closing date: %w", err)
	}

	expirationDate, err := r.expirationDateExtractor.ExtractFirstMatch(pages)
	if err != nil {
		return pdfcardsummary.CardSummary{}, fmt.Errorf("error finding the expiration date: %w", err)
	}

	// 2. Extract table data
	pi := pdfwrapper.NewRealDocumentIterator(pdfDoc, r.cfg.TableFirstRowPositionInPage)
	rowIterator := r.tableIteratorFactory.CreateIterator(pi)
	table, err := r.tableExtractor.Extract(rowIterator)
	if err != nil {
		return pdfcardsummary.CardSummary{}, fmt.Errorf("error extracting table data: %w", err)
	}

	arsTotal, usdTotal, err := r.extractDocumentTotals(pages)
	if err != nil {
		return pdfcardsummary.CardSummary{}, fmt.Errorf("error extracting document totals: %w", err)
	}

	// Check for anomalies and add adjustment if needed
	anomalyInfo := r.tableExtractor.GetAnomalyInfo()
	if err := r.applyAnomalyAdjustmentIfNeeded(&table, anomalyInfo); err != nil {
		return pdfcardsummary.CardSummary{}, fmt.Errorf("error applying anomaly adjustment: %w", err)
	}

	// Extract bank and card company from PDF text
	allText := pdfcardsummary.ExtractAllTextFromDocument(pdfDoc)
	bank := pdfcardsummary.DetectBankFromText(allText)
	cardCompany := pdfcardsummary.DetectCardCompanyFromText(allText)

	return pdfcardsummary.CardSummary{
		StatementContext: pdfcardsummary.StatementContext{
			TotalARS:       arsTotal,
			TotalUSD:       usdTotal,
			CloseDate:      closingDate,
			ExpirationDate: expirationDate,
			Bank:           bank,
			CardCompany:    cardCompany,
		},
		Table: table,
	}, nil
}

// applyAnomalyAdjustmentIfNeeded checks for anomalies and applies adjustment if needed
func (r *SantanderExtractor) applyAnomalyAdjustmentIfNeeded(table *pdfcardsummary.Table, anomalyInfo AnomalyDetectionInfo) error {
	if !anomalyInfo.HasAmountOnlyRowWithoutDecimals || anomalyInfo.AffectedCardIndex >= len(table.Cards) || anomalyInfo.AffectedMovementIndex < 0 {
		return nil
	}

	affectedCard := &table.Cards[anomalyInfo.AffectedCardIndex]
	adjustment := calculateAnomalyAdjustment(affectedCard)
	if !adjustment.ShouldApply {
		return nil
	}

	anomalyMov := createAnomalyAdjustmentMovement(adjustment)
	insertAnomalyMovement(affectedCard, anomalyMov, anomalyInfo.AffectedMovementIndex)

	log.Printf("WARNING: Added anomaly adjustment movement for card %d (difference: %s %s)",
		anomalyInfo.AffectedCardIndex, adjustment.Amount.String(), adjustment.Currency)

	// Verify that the adjustment actually fixed the mismatch
	if err := verifyAnomalyAdjustment(affectedCard); err != nil {
		log.Printf("WARNING: Anomaly adjustment applied but verification failed: %v", err)
		// Continue anyway - the adjustment was applied, verification failure is logged
	}

	return nil
}

// isWithinAnomalyThreshold checks if an amount difference is within the allowed threshold for anomaly adjustment.
// Returns true if the absolute value of the amount is less than or equal to the maximum allowed
// for the given currency (maxAnomalyAdjustmentARS for ARS, maxAnomalyAdjustmentUSD for USD).
func isWithinAnomalyThreshold(amount decimal.Decimal, currency Currency) bool {
	if currency == CurrencyARS {
		return amount.Abs().LessThanOrEqual(decimal.NewFromInt(maxAnomalyAdjustmentARS))
	}
	return amount.Abs().LessThanOrEqual(decimal.NewFromFloat(maxAnomalyAdjustmentUSD))
}

// calculateCardMovementsSum calculates the sum of all movements for a card, returning ARS and USD totals separately.
func calculateCardMovementsSum(card *pdfcardsummary.Card) (ars, usd decimal.Decimal) {
	movementsSumMap := pdfcardsummary.SumOfMovements(card.Movements)
	return movementsSumMap["ARS"], movementsSumMap["USD"]
}

// calculateAnomalyAdjustment calculates the anomaly adjustment needed for a card
func calculateAnomalyAdjustment(card *pdfcardsummary.Card) AnomalyAdjustment {
	movementsARS, movementsUSD := calculateCardMovementsSum(card)

	diffARS := card.CardContext.CardTotalARS.Sub(movementsARS)
	diffUSD := card.CardContext.CardTotalUSD.Sub(movementsUSD)

	// Only add anomaly adjustment if:
	// 1. The difference is small (<= maxAnomalyAdjustmentARS ARS or <= maxAnomalyAdjustmentUSD USD)
	// 2. Exactly one currency has a difference
	// 3. The difference matches what we'd expect from missing decimals
	if !diffARS.IsZero() && isWithinAnomalyThreshold(diffARS, CurrencyARS) && diffUSD.IsZero() {
		return AnomalyAdjustment{
			Amount:      diffARS,
			Currency:    CurrencyARS,
			ShouldApply: true,
		}
	}
	if !diffUSD.IsZero() && isWithinAnomalyThreshold(diffUSD, CurrencyUSD) && diffARS.IsZero() {
		return AnomalyAdjustment{
			Amount:      diffUSD,
			Currency:    CurrencyUSD,
			ShouldApply: true,
		}
	}

	return AnomalyAdjustment{ShouldApply: false}
}

// createAnomalyAdjustmentMovement creates a movement for anomaly adjustment
func createAnomalyAdjustmentMovement(adjustment AnomalyAdjustment) pdfcardsummary.Movement {
	mov := pdfcardsummary.Movement{
		Detail:    AnomalyAdjustmentDetail,
		AmountARS: decimal.Zero,
		AmountUSD: decimal.Zero,
	}
	if adjustment.Currency == CurrencyARS {
		mov.AmountARS = adjustment.Amount
	} else {
		mov.AmountUSD = adjustment.Amount
	}
	return mov
}

// verifyAnomalyAdjustment verifies that the anomaly adjustment actually fixed the mismatch.
// Returns an error if card totals still don't match the sum of movements after adjustment.
func verifyAnomalyAdjustment(card *pdfcardsummary.Card) error {
	movementsARS, movementsUSD := calculateCardMovementsSum(card)

	if !card.CardContext.CardTotalARS.Equal(movementsARS) || !card.CardContext.CardTotalUSD.Equal(movementsUSD) {
		return fmt.Errorf("card totals still don't match after adjustment (ARS: expected %s, got %s; USD: expected %s, got %s)",
			card.CardContext.CardTotalARS, movementsARS, card.CardContext.CardTotalUSD, movementsUSD)
	}
	return nil
}

// insertAnomalyMovement inserts an anomaly adjustment movement at the specified index.
// insertIndex is calculated as affectedMovementIndex + 1, which should be within bounds
// (0 <= insertIndex <= len(card.Movements)). The slices.Insert function allows inserting
// at index equal to length (which appends), so the condition insertIndex <= len is correct.
// We include a fallback to append if the index is out of range, which could happen
// if movements were added/removed after anomaly detection but before adjustment.
func insertAnomalyMovement(card *pdfcardsummary.Card, anomalyMov pdfcardsummary.Movement, affectedMovementIndex int) {
	insertIndex := affectedMovementIndex + 1
	if insertIndex <= len(card.Movements) {
		card.Movements = slices.Insert(card.Movements, insertIndex, anomalyMov)
	} else {
		// Fallback: append if index is out of range (should not happen in normal flow,
		// but protects against edge cases where movement count changed after detection)
		card.Movements = append(card.Movements, anomalyMov)
	}
}
