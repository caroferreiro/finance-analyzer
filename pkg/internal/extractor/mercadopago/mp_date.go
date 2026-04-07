package mercadopago

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var spanishFullMonths = map[string]time.Month{
	"enero":      time.January,
	"febrero":    time.February,
	"marzo":      time.March,
	"abril":      time.April,
	"mayo":       time.May,
	"junio":      time.June,
	"julio":      time.July,
	"agosto":     time.August,
	"septiembre": time.September,
	"octubre":    time.October,
	"noviembre":  time.November,
	"diciembre":  time.December,
}

var spanishShortMonths = map[string]time.Month{
	"ene": time.January,
	"feb": time.February,
	"mar": time.March,
	"abr": time.April,
	"may": time.May,
	"jun": time.June,
	"jul": time.July,
	"ago": time.August,
	"sep": time.September,
	"set": time.September,
	"oct": time.October,
	"nov": time.November,
	"dic": time.December,
}

const monthAlternation = "enero|febrero|marzo|abril|mayo|junio|julio|agosto|septiembre|octubre|noviembre|diciembre"

// ParseFullSpanishDate extracts a date from text matching "label" followed by "DD de month".
// MercadoPago concatenates text without spaces, so the regex matches known month names
// explicitly to avoid capturing trailing text (e.g., "eneroFecha").
func ParseFullSpanishDate(text string, label string, year int) (time.Time, error) {
	pattern := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(label) + `\s*(\d{1,2})\s+de\s+(` + monthAlternation + `)`)
	m := pattern.FindStringSubmatch(text)
	if m == nil {
		return time.Time{}, fmt.Errorf("label %q not found", label)
	}

	day, _ := strconv.Atoi(m[1])
	monthName := strings.ToLower(m[2])
	month, ok := spanishFullMonths[monthName]
	if !ok {
		return time.Time{}, fmt.Errorf("unknown month %q after label %q", monthName, label)
	}

	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC), nil
}

// ParseShortDate parses abbreviated dates like "23/ago" or "1/nov" (d/mmm or dd/mmm).
// The year is inferred relative to closeDate: if the movement month is after the
// close date month, it belongs to the previous year.
func ParseShortDate(s string, closeDate time.Time) *time.Time {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		return nil
	}
	day, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil
	}
	monthAbbr := strings.ToLower(parts[1])
	month, ok := spanishShortMonths[monthAbbr]
	if !ok {
		return nil
	}

	year := closeDate.Year()
	if month > closeDate.Month() {
		year--
	}

	t := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return &t
}

// InferYearFromCloseMonth determines the statement year given the close date's month.
// MercadoPago PDFs omit explicit years, so we assume the most recent occurrence
// of the given month that is not in the future.
func InferYearFromCloseMonth(closeMonth time.Month, now time.Time) int {
	year := now.Year()
	if closeMonth > now.Month() {
		year--
	}
	return year
}

// ExtractCloseMonth searches the text for the close date month label and returns it.
func ExtractCloseMonth(text string) (time.Month, error) {
	closeDateRe := regexp.MustCompile(`(?i)(?:Cierre actual|Fecha de cierre)\s*\d{1,2}\s+de\s+(` + monthAlternation + `)`)
	m := closeDateRe.FindStringSubmatch(text)
	if m == nil {
		return 0, fmt.Errorf("could not extract close month from text")
	}

	monthName := strings.ToLower(m[1])
	month, ok := spanishFullMonths[monthName]
	if !ok {
		return 0, fmt.Errorf("unknown close month %q", monthName)
	}
	return month, nil
}
