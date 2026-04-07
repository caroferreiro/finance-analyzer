package santander

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Alechan/finance-analyzer/pkg/internal/extractor/pdftable"

	"github.com/Alechan/finance-analyzer/pkg/internal/pdfcardsummary"
	"github.com/Alechan/finance-analyzer/pkg/internal/platform/stringsale"
)

func ConvertRawWithMonthToMovement(row pdftable.Row, cuotasSubDetailRegex *regexp.Regexp) (pdfcardsummary.Movement, error) {
	date, err := parseRawDateWithMonth(row.RawOriginalDate)
	if err != nil {
		return pdfcardsummary.Movement{}, fmt.Errorf("error extracting date from text with month: %v", err)
	}

	receiptNumber := parseRawReceipt(row.RawReceiptNumber)

	detail, currInstallment, totalInstallments, err := parseRawDetail(row.RawDetailWithMaybeInstallments, cuotasSubDetailRegex)
	if err != nil {
		return pdfcardsummary.Movement{}, fmt.Errorf("error extracting detail: %w", err)
	}

	arsAmount, err := pdfcardsummary.PDFAmountToDecimal(row.RawAmountARS)
	if err != nil {
		return pdfcardsummary.Movement{}, fmt.Errorf("error extracting ARS amount: %w", err)
	}

	amountUSD, err := pdfcardsummary.PDFAmountToDecimal(row.RawAmountUSD)
	if err != nil {
		return pdfcardsummary.Movement{}, fmt.Errorf("error extracting USD amount: %v", err)
	}

	return pdfcardsummary.Movement{
		OriginalDate:       &date,
		ReceiptNumber:      receiptNumber,
		Detail:             detail,
		CurrentInstallment: currInstallment,
		TotalInstallments:  totalInstallments,
		AmountARS:          arsAmount,
		AmountUSD:          amountUSD,
	}, nil
}

func ConvertRawWithoutMonthToMovement(row pdftable.Row, cuotasSubDetailRegex *regexp.Regexp, lastMovementDate time.Time) (pdfcardsummary.Movement, error) {
	date, err := parseRawDateWithoutMonth(row.RawOriginalDate, lastMovementDate)
	if err != nil {
		return pdfcardsummary.Movement{}, fmt.Errorf("error extracting date from text without month: %w", err)
	}

	receiptNumber := parseRawReceipt(row.RawReceiptNumber)

	detail, currInstallment, totalInstallments, err := parseRawDetail(row.RawDetailWithMaybeInstallments, cuotasSubDetailRegex)
	if err != nil {
		return pdfcardsummary.Movement{}, fmt.Errorf("error extracting detail: %w", err)
	}

	arsAmount, err := pdfcardsummary.PDFAmountToDecimal(row.RawAmountARS)
	if err != nil {
		return pdfcardsummary.Movement{}, fmt.Errorf("error extracting ARS amount: %w", err)
	}

	amountUSD, err := pdfcardsummary.PDFAmountToDecimal(row.RawAmountUSD)
	if err != nil {
		return pdfcardsummary.Movement{}, fmt.Errorf("error extracting USD amount: %w", err)
	}

	return pdfcardsummary.Movement{
		OriginalDate:       &date,
		ReceiptNumber:      receiptNumber,
		Detail:             detail,
		CurrentInstallment: currInstallment,
		TotalInstallments:  totalInstallments,
		AmountARS:          arsAmount,
		AmountUSD:          amountUSD,
	}, nil
}

func ConvertToSaldoAnteriorMovement(rawMovement pdftable.Row) (pdfcardsummary.Movement, error) {
	arsAmount, err := pdfcardsummary.PDFAmountToDecimal(rawMovement.RawAmountARS)
	if err != nil {
		return pdfcardsummary.Movement{}, fmt.Errorf("error converting ARS amount %s to decimal: %w", rawMovement.RawAmountARS, err)
	}

	usdAmount, err := pdfcardsummary.PDFAmountToDecimal(rawMovement.RawAmountUSD)
	if err != nil {
		return pdfcardsummary.Movement{}, fmt.Errorf("error converting USD amount %s to decimal: %w", rawMovement.RawAmountUSD, err)
	}

	return pdfcardsummary.Movement{
		Detail:    "SALDO ANTERIOR",
		AmountARS: arsAmount,
		AmountUSD: usdAmount,
	}, nil

}

func ExtendPreviousMovementDetail(prevMov *pdfcardsummary.Movement, movExtension pdfcardsummary.MovementExtension) error {
	// For it to be an extension, it should have the same day as the previous movement
	if prevMov.OriginalDate == nil || prevMov.OriginalDate.Day() != movExtension.OriginalDateDay {
		return fmt.Errorf("previous movement date %v is different from extension day %d", prevMov.OriginalDate, movExtension.OriginalDateDay)
	}

	prevMov.Detail += " " + movExtension.Detail
	return nil
}

func ConvertToFullDate(movWithOnlyDay pdfcardsummary.MovementWithOnlyDay, prevMov pdfcardsummary.Movement) (pdfcardsummary.Movement, error) {
	// If the previous movement doesn't have a date, it's an error
	if prevMov.OriginalDate == nil {
		return pdfcardsummary.Movement{}, fmt.Errorf("previous movement doesn't have a date to get full date %v", movWithOnlyDay)
	}

	date := time.Date(
		prevMov.OriginalDate.Year(),
		prevMov.OriginalDate.Month(),
		movWithOnlyDay.OriginalDateDay,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	return pdfcardsummary.Movement{
		OriginalDate:       &date,
		ReceiptNumber:      movWithOnlyDay.ReceiptNumber,
		Detail:             movWithOnlyDay.Detail,
		CurrentInstallment: movWithOnlyDay.CurrentInstallment,
		TotalInstallments:  movWithOnlyDay.TotalInstallments,
		AmountARS:          movWithOnlyDay.AmountARS,
		AmountUSD:          movWithOnlyDay.AmountUSD,
	}, nil
}

func convertToMovementExtension(row pdftable.Row) (pdfcardsummary.MovementExtension, error) {
	day, err := parseDayFromRawDate(row.RawOriginalDate)
	if err != nil {
		return pdfcardsummary.MovementExtension{}, fmt.Errorf("error extracting day from raw date: %w", err)
	}

	validDay := day != 0
	noARSAmount := strings.TrimSpace(row.RawAmountARS) == ""
	noUSDAmount := strings.TrimSpace(row.RawAmountUSD) == ""
	nonEmptyDetail := row.RawDetailWithMaybeInstallments != ""

	isValidExtension := validDay && noARSAmount && noUSDAmount && nonEmptyDetail
	if !isValidExtension {
		return pdfcardsummary.MovementExtension{}, fmt.Errorf("movement is not a valid extension: %v", row)
	}

	return pdfcardsummary.MovementExtension{
		OriginalDateDay: day,
		Detail:          row.RawDetailWithMaybeInstallments,
	}, nil
}

func convertToMovementWithOnlyDay(row pdftable.Row) (pdfcardsummary.MovementWithOnlyDay, error) {
	day, err := parseDayFromRawDate(row.RawOriginalDate)
	if err != nil {
		return pdfcardsummary.MovementWithOnlyDay{}, fmt.Errorf("error extracting day from raw date: %w", err)
	}

	detail, currInstallment, totalInstallments, err := parseRawDetail(row.RawDetailWithMaybeInstallments, nil)
	if err != nil {
		return pdfcardsummary.MovementWithOnlyDay{}, fmt.Errorf("error extracting detail: %w", err)
	}

	arsAmount, err := pdfcardsummary.PDFAmountToDecimal(row.RawAmountARS)
	if err != nil {
		return pdfcardsummary.MovementWithOnlyDay{}, fmt.Errorf("error extracting ARS amount: %w", err)
	}

	amountUSD, err := pdfcardsummary.PDFAmountToDecimal(row.RawAmountUSD)
	if err != nil {
		return pdfcardsummary.MovementWithOnlyDay{}, fmt.Errorf("error extracting USD amount: %w", err)
	}

	return pdfcardsummary.MovementWithOnlyDay{
		OriginalDateDay:    day,
		ReceiptNumber:      nil,
		Detail:             detail,
		CurrentInstallment: currInstallment,
		TotalInstallments:  totalInstallments,
		AmountARS:          arsAmount,
		AmountUSD:          amountUSD,
	}, nil
}

func parseRawDetail(rawDetailWithMaybeInstallments string, cuotasSubDetailRegex *regexp.Regexp) (string, *int, *int, error) {
	// See if the last part of the detail is the installments
	trimmed := strings.TrimSpace(rawDetailWithMaybeInstallments)
	if cuotasSubDetailRegex == nil {
		return stringsale.RemoveDuplicateSpaces(trimmed), nil, nil, nil
	}

	matches := cuotasSubDetailRegex.FindStringSubmatch(trimmed)
	if len(matches) > 1 {
		currInstallment, err := strconv.Atoi(matches[1])
		if err != nil {
			return "", nil, nil, fmt.Errorf("error converting current installment %s to int: %w", matches[1], err)
		}

		totalInstallments, err := strconv.Atoi(matches[2])
		if err != nil {
			return "", nil, nil, fmt.Errorf("error converting total installments %s to int: %w", matches[2], err)
		}

		// The detail should be the part before the installments
		firstMatchedCuotaRegex := len(trimmed) - len(matches[0])
		detailWithoutInstallments := trimmed[:firstMatchedCuotaRegex]
		return stringsale.RemoveDuplicateSpaces(detailWithoutInstallments), &currInstallment, &totalInstallments, nil
	}

	return stringsale.RemoveDuplicateSpaces(trimmed), nil, nil, nil
}

func parseRawDateWithMonth(rawDate string) (time.Time, error) {
	//"24 Agosto 12"
	if len(rawDate) != 13 {
		return time.Time{}, fmt.Errorf("expected date with 13 characters, got '%s'", rawDate)
	}

	// The first 2 characters should correspond to the year (2024 -> 24)
	rawYear := "20" + rawDate[:2]
	year, err := strconv.Atoi(rawYear)
	if err != nil {
		return time.Time{}, fmt.Errorf("error converting year %s to int: %w", rawYear, err)
	}

	// The month comes in Spanish and may be abbreviated
	rawMonth := strings.TrimSpace(rawDate[3:10])
	month, err := convertSpanishMaybeAbbreviatedMonthToTime(rawMonth)
	if err != nil {
		return time.Time{}, fmt.Errorf("error converting month %s to time.Month: %v", rawMonth, err)
	}

	// The day should be the last 2 characters
	rawDay := strings.TrimSpace(rawDate[11:])
	day, err := strconv.Atoi(rawDay)
	if err != nil {
		return time.Time{}, fmt.Errorf("error converting day %s to int: %w", rawDay, err)
	}

	date := time.Date(
		year,
		month,
		day,
		0,
		0,
		0,
		0,
		time.UTC,
	)
	return date, nil
}

func parseRawDateWithoutMonth(rawDate string, lastMovementDate time.Time) (time.Time, error) {
	// Handle empty date (continuation line) - use previous date directly
	if rawDate == "" {
		return lastMovementDate, nil
	}

	// The length should be exactly 2 because only the day is included
	if len(rawDate) != 2 {
		return time.Time{}, fmt.Errorf("expected date with exactly 2 characters, got '%s'", rawDate)
	}

	// The day should be both characters
	day, err := strconv.Atoi(rawDate)
	if err != nil {
		return time.Time{}, fmt.Errorf("error converting day %s to int: %w", rawDate, err)
	}

	// The year and month are taken from the last movement date
	year := lastMovementDate.Year()
	month := lastMovementDate.Month()

	date := time.Date(
		year,
		month,
		day,
		0,
		0,
		0,
		0,
		time.UTC,
	)
	return date, nil
}

func convertSpanishMaybeAbbreviatedMonthToTime(rawMonth string) (time.Month, error) {
	mapper := map[string]time.Month{
		"Enero":   time.January,
		"Febrero": time.February,
		"Marzo":   time.March,
		"Abril":   time.April,
		"Mayo":    time.May,
		"Junio":   time.June,
		"Julio":   time.July,
		"Agosto":  time.August,
		"Setiem.": time.September,
		"Octubre": time.October,
		"Octubr.": time.October,
		"Noviem.": time.November,
		"Diciem.": time.December,
	}

	month, ok := mapper[rawMonth]
	if !ok {
		return time.Month(0), fmt.Errorf("month %s not found in mapper", rawMonth)
	}

	return month, nil
}

func parseDayFromRawDate(rawOriginalDate string) (int, error) {
	if len(rawOriginalDate) != 2 {
		return 0, fmt.Errorf("expected date with at least 2 characters, got %s", rawOriginalDate)
	}

	// The 2 characters should be the day
	day, err := strconv.Atoi(rawOriginalDate)
	if err != nil {
		return 0, fmt.Errorf("error converting day %s to int: %w", rawOriginalDate, err)
	}

	return day, nil
}

func parseRawReceipt(rawReceipt string) *string {
	receipt := strings.TrimSpace(rawReceipt)
	if receipt == "" {
		return nil
	}

	receiptNumber := &receipt
	return receiptNumber
}

// looksLikeAmountOnlyRow checks if a row matches the pattern of an amount-only row (broken line).
// An amount-only row has amounts but no date, receipt, or detail fields.
// This handles the edge case where a movement's description appears on one row and the amount appears on the next row.
func looksLikeAmountOnlyRow(row pdftable.Row, cfg SantanderExtractorConfig) bool {
	return cfg.AmountOnlyRowRegex.MatchString(row.RawText)
}

// parseAmountFromRegexMatch parses an amount from a regex match, handling negative signs and decimal formatting.
// The amountStr is the captured group, and fullMatch is the full regex match (used to detect trailing minus sign).
func parseAmountFromRegexMatch(amountStr string, fullMatch string) string {
	amountStr = strings.TrimSpace(amountStr)
	// Check the full match for trailing minus sign (for negative amounts)
	// The minus sign is outside the capturing group, so we check the full match
	isNegative := strings.HasSuffix(fullMatch, "-")
	// If the amount doesn't have a comma (no decimal part), append ",00" to ensure 2 decimal places
	// This handles cases like "1.316" which should be parsed as "1.316,00" = 1316.00
	if !strings.Contains(amountStr, ",") {
		amountStr = amountStr + ",00"
	}
	if isNegative {
		amountStr = "-" + amountStr
	}
	return amountStr
}

// hasNoDecimals checks if an amount string has no decimal separator (comma).
// Returns true if the string does not contain a comma, indicating missing decimal values.
func hasNoDecimals(amountStr string) bool {
	return !strings.Contains(amountStr, ",")
}

// mergeAmountsIntoMovement merges the amounts from an amount-only row into a movement.
// The movement should have zero amounts, and the row should be an amount-only row.
// Only non-empty amounts from the row are merged (ARS and/or USD).
// Returns true if the amount had no decimals (indicating a glitch).
func mergeAmountsIntoMovement(mov *pdfcardsummary.Movement, amountOnlyRow pdftable.Row, cfg SantanderExtractorConfig) (bool, error) {
	// Track if any amount had no decimals (glitch indicator)
	var hasNoDecimalsFlag bool

	// Check RawAmountARS and RawAmountUSD first, as they're more reliable if extracted correctly
	// These take precedence over regex extraction if they exist
	if strings.TrimSpace(amountOnlyRow.RawAmountARS) != "" {
		if hasNoDecimals(amountOnlyRow.RawAmountARS) {
			hasNoDecimalsFlag = true
		}
		arsAmount, err := pdfcardsummary.PDFAmountToDecimal(amountOnlyRow.RawAmountARS)
		if err != nil {
			return false, fmt.Errorf("error parsing ARS amount from amount-only row: %w", err)
		}
		mov.AmountARS = arsAmount
	}

	if strings.TrimSpace(amountOnlyRow.RawAmountUSD) != "" {
		if hasNoDecimals(amountOnlyRow.RawAmountUSD) {
			hasNoDecimalsFlag = true
		}
		usdAmount, err := pdfcardsummary.PDFAmountToDecimal(amountOnlyRow.RawAmountUSD)
		if err != nil {
			return false, fmt.Errorf("error parsing USD amount from amount-only row: %w", err)
		}
		mov.AmountUSD = usdAmount
	}

	// Extract amount from RawText using the regex for any currency not already set
	// This handles cases where RowFactory didn't extract amounts correctly
	// We check regex for each currency independently, as one might be extracted while the other isn't
	if mov.AmountARS.IsZero() || mov.AmountUSD.IsZero() {
		matches := cfg.AmountOnlyRowRegex.FindStringSubmatch(amountOnlyRow.RawText)
		if len(matches) > 1 {
			// The regex has two alternatives, so we need to find which capturing group has the amount
			// Group 1 is for ARS amounts (0-75 spaces), Group 2 is for USD amounts (0-92 spaces)
			if matches[1] != "" && mov.AmountARS.IsZero() {
				// Extract ARS from regex if not already set
				amountStr := matches[1]
				if hasNoDecimals(amountStr) {
					hasNoDecimalsFlag = true
				}
				formattedAmount := parseAmountFromRegexMatch(amountStr, matches[0])
				parsedAmount, err := pdfcardsummary.PDFAmountToDecimal(formattedAmount)
				if err != nil {
					return false, fmt.Errorf("error parsing ARS amount from amount-only row: %w", err)
				}
				mov.AmountARS = parsedAmount
			}
			if len(matches) > 2 && matches[2] != "" && mov.AmountUSD.IsZero() {
				// Extract USD from regex if not already set
				amountStr := matches[2]
				if hasNoDecimals(amountStr) {
					hasNoDecimalsFlag = true
				}
				formattedAmount := parseAmountFromRegexMatch(amountStr, matches[0])
				parsedAmount, err := pdfcardsummary.PDFAmountToDecimal(formattedAmount)
				if err != nil {
					return false, fmt.Errorf("error parsing USD amount from amount-only row: %w", err)
				}
				mov.AmountUSD = parsedAmount
			}
		}
	}

	return hasNoDecimalsFlag, nil
}
