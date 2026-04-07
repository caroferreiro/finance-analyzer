package pdfcardsummary

import (
	"strings"
)

// CardCompany represents a card company (VISA, AMEX, Mastercard, etc.)
type CardCompany string

// DetectCardCompanyFromText extracts the card company from PDF text content.
// It searches for card company keywords (VISA, AMEX/American Express, Mastercard)
// and returns the normalized card company name (e.g., "VISA", "AMEX", "Mastercard"), or "?" if no match is found.
// Priority: VISA > AMEX > Mastercard (first match found)
func DetectCardCompanyFromText(text string) CardCompany {
	textUpper := strings.ToUpper(text)

	// Check for VISA (most common, check first)
	if strings.Contains(textUpper, "VISA") {
		return CardCompany("VISA")
	}

	// Check for AMEX / American Express
	if strings.Contains(textUpper, "AMEX") || strings.Contains(textUpper, "AMERICAN EXPRESS") {
		return CardCompany("AMEX")
	}

	// Check for Mastercard
	if strings.Contains(textUpper, "MASTERCARD") {
		return CardCompany("Mastercard")
	}

	return CardCompany("?")
}
