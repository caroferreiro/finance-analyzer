package mercadopago

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Alechan/finance-analyzer/pkg/internal/pdfcardsummary"
	"github.com/shopspring/decimal"
)

// operationNumberDigits is the fixed length of MercadoPago operation IDs.
const operationNumberDigits = 6

// ParseMovementChunk parses a single movement from a text chunk that starts with a short date.
// Returns nil if the chunk is a section header (e.g., "Subtotal", "Resumen de", "Pago de tarjeta").
func ParseMovementChunk(chunk string, closeDate time.Time) *pdfcardsummary.Movement {
	dateRe := regexp.MustCompile(`^(\d{1,2}/[a-zA-Záéíóú]{3})`)
	dateMatch := dateRe.FindString(chunk)
	if dateMatch == "" {
		return nil
	}
	date := ParseShortDate(dateMatch, closeDate)
	rest := strings.TrimSpace(chunk[len(dateMatch):])

	if rest == "" {
		return nil
	}
	if isSectionHeader(rest) {
		return nil
	}

	amountARS, amountUSD, rest := extractAmounts(rest)

	rest = strings.TrimSpace(rest)
	if rest == "" {
		return nil
	}

	currentInst, totalInst, receipt, rest := extractInstallmentAndOperation(rest)

	detail := strings.TrimSpace(rest)
	if detail == "" {
		return nil
	}

	return &pdfcardsummary.Movement{
		OriginalDate:       date,
		ReceiptNumber:      &receipt,
		Detail:             detail,
		CurrentInstallment: currentInst,
		TotalInstallments:  totalInst,
		AmountARS:          amountARS,
		AmountUSD:          amountUSD,
	}
}

// SplitMovementLines splits concatenated movement text into individual chunks by date boundaries.
// MercadoPago PDFs concatenate lines without whitespace, so an amount like "$5.619,90" runs
// directly into the next date "4/ene". We split by detecting the boundary between an amount
// ending (",\d{2}") and a date start ("\d{1,2}/mmm").
func SplitMovementLines(text string, closeDate time.Time) []pdfcardsummary.Movement {
	var movements []pdfcardsummary.Movement

	text = strings.ReplaceAll(text, "\n", "|")

	splitRe := regexp.MustCompile(`(,\d{2})(\d{1,2}/[a-zA-Záéíóú]{3})`)
	separated := splitRe.ReplaceAllString(text, "$1|$2")

	dateStartRe := regexp.MustCompile(`^\d{1,2}/[a-zA-Záéíóú]{3}`)
	chunks := strings.Split(separated, "|")

	for _, chunk := range chunks {
		chunk = strings.TrimSpace(chunk)
		if chunk == "" || !dateStartRe.MatchString(chunk) {
			continue
		}
		if mov := ParseMovementChunk(chunk, closeDate); mov != nil {
			movements = append(movements, *mov)
		}
	}

	return movements
}

func isSectionHeader(text string) bool {
	lower := strings.ToLower(text)
	return strings.HasPrefix(lower, "resumen de") ||
		strings.HasPrefix(lower, "pago de tarjeta") ||
		strings.HasPrefix(lower, "subtotal")
}

func extractAmounts(rest string) (decimal.Decimal, decimal.Decimal, string) {
	amountUSD := decimal.Zero
	usdMatch := usdAmountRe.FindString(rest)
	if usdMatch != "" {
		amountUSD = ParseARSAmount(strings.TrimPrefix(usdMatch, "US"))
		rest = strings.Replace(rest, usdMatch, "", 1)
	}

	amountARS := decimal.Zero
	arsMatch := arsAmountRe.FindString(rest)
	if arsMatch != "" {
		amountARS = ParseARSAmount(arsMatch)
		idx := strings.Index(rest, arsMatch)
		rest = rest[:idx]
	}

	return amountARS, amountUSD, rest
}

// extractInstallmentAndOperation parses the tail of a movement detail for installments
// and/or an operation number.
//
// MercadoPago formats vary across statements:
//   - "3 de 3954796" → installment 3/3, operation 954796 (glued together)
//   - "1 de 3"       → installment 1/3, no operation
//   - "461159"       → no installment, standalone operation
//   - (nothing)      → neither
func extractInstallmentAndOperation(rest string) (*int, *int, string, string) {
	var currentInst, totalInst *int
	var receipt string

	instOpRe := regexp.MustCompile(`(\d{1,2})\s+de\s+(\d{1,2})(\d{` +
		strconv.Itoa(operationNumberDigits) + `})?\s*$`)
	if m := instOpRe.FindStringSubmatch(rest); m != nil {
		c, _ := strconv.Atoi(m[1])
		t, _ := strconv.Atoi(m[2])
		currentInst = &c
		totalInst = &t
		if m[3] != "" {
			receipt = m[3]
		}
		rest = strings.TrimSpace(rest[:len(rest)-len(m[0])])
		return currentInst, totalInst, receipt, rest
	}

	opRe := regexp.MustCompile(`(\d{` + strconv.Itoa(operationNumberDigits) + `})\s*$`)
	if m := opRe.FindStringSubmatch(rest); m != nil {
		receipt = m[1]
		rest = strings.TrimSpace(rest[:len(rest)-len(m[0])])
	}

	return currentInst, totalInst, receipt, rest
}
