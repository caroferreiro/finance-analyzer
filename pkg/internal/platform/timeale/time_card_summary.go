package timeale

import (
	"fmt"
	"strings"
	"time"
)

// Map of Spanish month abbreviations to their numbers
var spanishMonths = map[string]string{
	"Ene": "01",
	"Feb": "02",
	"Mar": "03",
	"Abr": "04",
	"May": "05",
	"Jun": "06",
	"Jul": "07",
	"Ago": "08",
	"Set": "09",
	"Sep": "09",
	"Oct": "10",
	"Nov": "11",
	"Dic": "12",
}

func CardSummarySpanishMonthDateToTime(dateStr string) (time.Time, error) {
	// Split the date into its components
	parts := strings.Split(dateStr, " ")

	if len(parts) != 3 {
		return time.Time{}, fmt.Errorf("invalid date format: %s", dateStr)
	}

	// Replace the Spanish month abbreviation with its numeric equivalent
	month, ok := spanishMonths[parts[1]]
	if !ok {
		return time.Time{}, fmt.Errorf("invalid month abbreviation: %s", parts[1])
	}

	// Reconstruct the date string in the format "02 01 06"
	formattedDate := fmt.Sprintf("%s %s %s", parts[0], month, parts[2])

	// Define the layout
	layout := "02 01 06"

	// Parse the formatted date string
	date, err := time.Parse(layout, formattedDate)
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing date: %v", err)
	}
	return date, err
}
