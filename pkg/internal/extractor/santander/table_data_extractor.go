package santander

import (
	"fmt"
	"slices"
	"time"

	"github.com/Alechan/finance-analyzer/pkg/internal/extractor/pdftable"
	"github.com/Alechan/finance-analyzer/pkg/internal/pdfcardsummary"
	"github.com/Alechan/finance-analyzer/pkg/internal/platform/pointersale"
)

type SantanderTableExtractor struct {
	cfg        SantanderExtractorConfig
	rowFactory *pdftable.RowFactory
	// anomalyInfo tracks information about detected glitches (amount-only rows without decimals)
	anomalyInfo AnomalyDetectionInfo
}

// AnomalyDetectionInfo tracks when an amount-only row without decimals is detected
type AnomalyDetectionInfo struct {
	// HasAmountOnlyRowWithoutDecimals indicates if an amount-only row without decimals was detected
	HasAmountOnlyRowWithoutDecimals bool
	// AffectedCardIndex is the index of the card that had the glitched movement
	AffectedCardIndex int
	// AffectedMovementIndex is the index of the movement within the card that was glitched
	AffectedMovementIndex int
}

func NewSantanderTableExtractor(cfg SantanderExtractorConfig) *SantanderTableExtractor {
	return &SantanderTableExtractor{
		cfg:         cfg,
		rowFactory:  pdftable.NewRowFactory(cfg.TableColumnPositions),
		anomalyInfo: AnomalyDetectionInfo{},
	}
}

// TODO: Add unit test for Extract method - test complete table extraction with mock iterator
// Test cases: empty iterator, only SALDO ANTERIOR, complete table with cards and taxes
func (te *SantanderTableExtractor) Extract(rowIterator pdftable.TableIterator) (pdfcardsummary.Table, error) {
	pastPaymentMovements, err := te.extractPastPaymentMovements(rowIterator)
	if err != nil {
		return pdfcardsummary.Table{}, fmt.Errorf("error extracting past payment movements: %w", err)
	}

	cards, taxesMovements, err := te.extractCardsAndTaxesMovs(rowIterator)
	if err != nil {
		return pdfcardsummary.Table{}, fmt.Errorf("error extracting cards: %w", err)
	}

	return pdfcardsummary.Table{
		PastPaymentMovements: pastPaymentMovements,
		Cards:                cards,
		TaxesMovements:       taxesMovements,
	}, nil
}

// TODO: Add unit test for extractPastPaymentMovements - test SALDO ANTERIOR + past payments
// Test cases: only SALDO ANTERIOR, SALDO ANTERIOR + payments, SALDO ANTERIOR + payments + extensions
func (te *SantanderTableExtractor) extractPastPaymentMovements(rowIterator pdftable.TableIterator) ([]pdfcardsummary.Movement, error) {
	prevBalMov, err := te.extractPreviousBalance(rowIterator)
	if err != nil {
		return nil, fmt.Errorf("error extracting previous balance: %w", err)
	}
	restOfPastPaymentsMovs, err := te.extractRestOfPastPaymentsMovs(rowIterator)
	if err != nil {
		return nil, fmt.Errorf("error extracting rest of past payment movements: %w", err)
	}
	pastPaymentMovements := slices.Concat(
		[]pdfcardsummary.Movement{prevBalMov},
		restOfPastPaymentsMovs,
	)
	return pastPaymentMovements, nil
}

// TODO: Add unit test for extractPreviousBalance - test SALDO ANTERIOR extraction
// Test cases: SALDO ANTERIOR found, SALDO ANTERIOR not found, malformed SALDO ANTERIOR
func (te *SantanderTableExtractor) extractPreviousBalance(rowIterator pdftable.TableIterator) (pdfcardsummary.Movement, error) {
	for row, ok := rowIterator.NextUtilRegexIsMatched(te.cfg.SaldoAnteriorRegex); ok; {
		mov, err := ConvertToSaldoAnteriorMovement(row)
		if err != nil {
			return pdfcardsummary.Movement{}, fmt.Errorf("error converting to saldo anterior movement: %w", err)
		}
		return mov, nil
	}
	return pdfcardsummary.Movement{}, fmt.Errorf("no previous balance found in first row of table")
}

// TODO: Add unit test for extractRestOfPastPaymentsMovs - test past payment movements
// Test cases: regular movements, movement extensions, movements with only day, extension without previous movement
func (te *SantanderTableExtractor) extractRestOfPastPaymentsMovs(rowIterator pdftable.TableIterator) ([]pdfcardsummary.Movement, error) {
	var movs []pdfcardsummary.Movement
	for row, ok := rowIterator.Next(); ok; row, ok = rowIterator.Next() {
		if te.cfg.EndOfPastPaymentsRegex.MatchString(row.RawText) {
			break
		}
		mov, err := ConvertRawWithMonthToMovement(row, te.cfg.CuotasSubDetailRegex)
		if err == nil {
			movs = append(movs, mov)
			continue
		}
		movExt, err := convertToMovementExtension(row)
		if err == nil {
			atLeastOnePrevMov := len(movs) > 0
			if !atLeastOnePrevMov {
				return nil, fmt.Errorf("no movement found for extension %v", movExt)
			}
			err = ExtendPreviousMovementDetail(&movs[len(movs)-1], movExt)
			if err != nil {
				return nil, fmt.Errorf("error extending previous movement: %w", err)
			}
			continue
		}
		movWithOnlyDay, err := convertToMovementWithOnlyDay(row)
		if err == nil {
			atLeastOnePrevMov := len(movs) > 0
			if !atLeastOnePrevMov {
				return nil, fmt.Errorf("no movement found to get full date %v", movWithOnlyDay)
			}
			prevMov := movs[len(movs)-1]
			mov, err := ConvertToFullDate(movWithOnlyDay, prevMov)
			if err != nil {
				return nil, fmt.Errorf("error converting to full date: %w", err)
			}
			movs = append(movs, mov)
			continue
		}
	}
	return movs, nil
}

// TODO: Add unit test for extractCardsAndTaxesMovs - test card and tax extraction
// Test cases: single card, multiple cards, cards with movements, cards without movements, missing card numbers
func (te *SantanderTableExtractor) extractCardsAndTaxesMovs(rowIterator pdftable.TableIterator) ([]pdfcardsummary.Card, []pdfcardsummary.Movement, error) {
	cardBuilder := pdfcardsummary.NewCardBuilder()
	var cards []pdfcardsummary.Card
	var pendingMovement *pdfcardsummary.Movement
	var currentCardMovementCount int // Track movement count for current card
	for row, ok := rowIterator.Next(); ok; row, ok = rowIterator.Next() {
		// First, check if there's a pending movement and if the current row is an amount-only row
		if pendingMovement != nil {
			if looksLikeAmountOnlyRow(row, te.cfg) {
				hasNoDecimals, err := mergeAmountsIntoMovement(pendingMovement, row, te.cfg)
				if err != nil {
					return nil, nil, fmt.Errorf("error merging amounts into pending movement: %w", err)
				}
				// Track anomaly if amount had no decimals
				te.trackAnomalyIfNeeded(hasNoDecimals, len(cards), currentCardMovementCount)
				cardBuilder.AddMovement(*pendingMovement)
				currentCardMovementCount++
				pendingMovement = nil
				continue
			}
			// If not an amount-only row, add the pending movement as-is (it has zero amounts)
			// This shouldn't happen in valid PDFs, but we handle it gracefully
			cardBuilder.AddMovement(*pendingMovement)
			currentCardMovementCount++
			pendingMovement = nil
		}

		mov, err := ConvertRawWithMonthToMovement(row, te.cfg.CuotasSubDetailRegex)
		if err == nil {
			// Check if movement has zero amounts (broken line case)
			if mov.AmountARS.IsZero() && mov.AmountUSD.IsZero() {
				pendingMovement = &mov
				continue
			}
			cardBuilder.AddMovement(mov)
			currentCardMovementCount++
			continue
		}

		// Check for card total line BEFORE checking for movements without month
		// This prevents card total lines from being incorrectly matched as movements
		matchesTotalConsumosTarjeta := te.cfg.TotalConsumosTarjetaRegex.FindStringSubmatch(row.RawText)
		if len(matchesTotalConsumosTarjeta) > 1 {
			// Before finishing the card, add any pending movement
			if pendingMovement != nil {
				cardBuilder.AddMovement(*pendingMovement)
				currentCardMovementCount++
				pendingMovement = nil
			}
			card, err := finishBuildingCard(cardBuilder, matchesTotalConsumosTarjeta)
			if err != nil {
				return nil, nil, fmt.Errorf("error finishing building card: %w", err)
			}
			cards = append(cards, card)
			cardBuilder = pdfcardsummary.NewCardBuilder()
			currentCardMovementCount = 0 // Reset for next card
			continue
		}

		matchesMovWithoutMonth := row.MatchesMovementWithoutYearAndMonth()
		if matchesMovWithoutMonth {
			lastMovementDate, err := cardBuilder.GetLastMovementDate()
			if err != nil {
				return nil, nil, fmt.Errorf("error getting last movement's date: %w", err)
			}
			mov, err := te.parseCardMovementWithoutMonth(row.RawText, lastMovementDate)
			if err != nil {
				return nil, nil, fmt.Errorf("error creating card movement without month: %w", err)
			}
			// Check if movement has zero amounts (broken line case)
			if mov.AmountARS.IsZero() && mov.AmountUSD.IsZero() {
				pendingMovement = &mov
				continue
			}
			cardBuilder.AddMovement(mov)
			currentCardMovementCount++
			continue
		}
	}
	// Add any remaining pending movement before returning
	if pendingMovement != nil {
		cardBuilder.AddMovement(*pendingMovement)
	}
	movsWithoutCards := cardBuilder.MovementsAccumulated()
	return cards, movsWithoutCards, nil
}

// trackAnomalyIfNeeded tracks anomaly information if a glitch is detected
func (te *SantanderTableExtractor) trackAnomalyIfNeeded(hasNoDecimals bool, cardIndex, movementIndex int) {
	if hasNoDecimals {
		te.anomalyInfo.HasAmountOnlyRowWithoutDecimals = true
		te.anomalyInfo.AffectedCardIndex = cardIndex
		te.anomalyInfo.AffectedMovementIndex = movementIndex
	}
}

// GetAnomalyInfo returns the anomaly detection information
func (te *SantanderTableExtractor) GetAnomalyInfo() AnomalyDetectionInfo {
	return te.anomalyInfo
}

// TODO: Add unit test for parseCardMovementWithMonth - test card movement parsing with month
// Test cases: valid movement, invalid movement, movement with installments
func (te *SantanderTableExtractor) parseCardMovementWithMonth(text string) (pdfcardsummary.Movement, error) {
	row := te.rowFactory.CreateRow(text)
	mov, err := ConvertRawWithMonthToMovement(row, te.cfg.CuotasSubDetailRegex)
	if err != nil {
		return pdfcardsummary.Movement{}, fmt.Errorf("error creating movement from text with month: %w", err)
	}
	return mov, nil
}

// TODO: Add unit test for parseCardMovementWithoutMonth - test card movement parsing without month
// Test cases: valid movement, invalid movement, movement with previous date
func (te *SantanderTableExtractor) parseCardMovementWithoutMonth(text string, lastMovementDate time.Time) (pdfcardsummary.Movement, error) {
	row := te.rowFactory.CreateRow(text)
	mov, err := ConvertRawWithoutMonthToMovement(row, te.cfg.CuotasSubDetailRegex, lastMovementDate)
	if err != nil {
		return pdfcardsummary.Movement{}, fmt.Errorf("error creating movement from text without month: %w", err)
	}
	return mov, nil
}

// TODO: Add unit test for finishBuildingCard - test card building from regex matches
// Test cases: valid matches, invalid ARS amount, invalid USD amount, missing card number
func finishBuildingCard(cardBuilder *pdfcardsummary.CardBuilder, matchesTotalConsumosTarjeta []string) (pdfcardsummary.Card, error) {
	cardBuilder.SetNumber(pointersale.ToPointer(matchesTotalConsumosTarjeta[1]))
	cardBuilder.SetOwner(matchesTotalConsumosTarjeta[2])

	err := cardBuilder.SetTotalARS(matchesTotalConsumosTarjeta[3])
	if err != nil {
		return pdfcardsummary.Card{}, fmt.Errorf("error setting total ARS: %w", err)
	}

	err = cardBuilder.SetTotalUSD(matchesTotalConsumosTarjeta[4])
	if err != nil {
		return pdfcardsummary.Card{}, fmt.Errorf("error setting total USD: %w", err)
	}

	card, err := cardBuilder.Build()
	if err != nil {
		return pdfcardsummary.Card{}, fmt.Errorf("error building card: %w", err)
	}
	return card, nil
}
