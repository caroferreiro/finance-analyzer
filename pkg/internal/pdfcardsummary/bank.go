package pdfcardsummary

import (
	"strings"
)

// Bank represents a bank institution
type Bank string

var (
	// wordToBank maps individual words and phrases to canonical bank names.
	// All keys are lowercase for case-insensitive matching.
	// Multiple words can map to the same bank (e.g., "santander" matches "Santander").
	// Multi-word banks like "Mercado Pago" can be matched by either word.
	wordToBank = map[string]string{
		// Santander variations
		"santander":       "Santander",
		"banco santander": "Santander",

		// Galicia variations
		"galicia":       "Galicia",
		"banco galicia": "Galicia",

		// Acronyms
		"bbva":       "BBVA",
		"banco bbva": "BBVA",
		"hsbc":       "HSBC",
		"icbc":       "ICBC",
		"bind":       "BIND",

		// Other banks
		"macro":                                 "Macro",
		"banco macro":                           "Macro",
		"brubank":                               "Brubank",
		"naranja":                               "Naranja X",
		"cencosud":                              "Cencosud",
		"uala":                                  "Ualá",
		"ualá":                                  "Ualá", // Handle accent
		"itau":                                  "Itaú",
		"itaú":                                  "Itaú",
		"banco itaú":                            "Itaú",
		"banco itau":                            "Itaú",
		"patagonia":                             "Patagonia",
		"banco patagonia":                       "Patagonia",
		"supervielle":                           "Supervielle",
		"banco supervielle":                     "Supervielle",
		"comafi":                                "Comafi",
		"banco comafi":                          "Comafi",
		"hipotecario":                           "Hipotecario",
		"banco hipotecario":                     "Hipotecario",
		"credicoop":                             "Credicoop",
		"banco credicoop":                       "Credicoop",
		"columbia":                              "Columbia",
		"banco columbia":                        "Columbia",
		"piano":                                 "Piano",
		"banco piano":                           "Piano",
		"sol":                                   "Sol",
		"banco del sol":                         "Sol",
		"nacion":                                "Nación",
		"banco de la nacion argentina":          "Nación",
		"provincia":                             "Provincia",
		"banco de la provincia de buenos aires": "Provincia",
		"banco ciudad":                          "Ciudad",

		// Multi-word phrases
		"mercado pago": "Mercado Pago",
		"naranja x":    "Naranja X",
	}
)

// extractWords extracts words from text, handling punctuation and whitespace.
// It normalizes text to lowercase and handles accents by keeping them as-is.
func extractWords(text string) []string {
	// Replace punctuation with spaces, then split by whitespace
	text = strings.ToLower(text)
	// Remove common punctuation but keep word boundaries
	// For acronyms with dots (e.g., "B.B.V.A."), we want to remove dots but keep letters together
	// So we replace dots with empty string first for potential acronyms, then replace other punctuation with spaces
	text = strings.ReplaceAll(text, ".", "")
	text = strings.ReplaceAll(text, ",", " ")
	text = strings.ReplaceAll(text, ":", " ")
	text = strings.ReplaceAll(text, ";", " ")
	text = strings.ReplaceAll(text, "!", " ")
	text = strings.ReplaceAll(text, "?", " ")
	text = strings.ReplaceAll(text, "(", " ")
	text = strings.ReplaceAll(text, ")", " ")
	text = strings.ReplaceAll(text, "[", " ")
	text = strings.ReplaceAll(text, "]", " ")
	text = strings.ReplaceAll(text, "-", " ")
	text = strings.ReplaceAll(text, "_", " ")

	// Split by whitespace and filter empty strings
	words := strings.Fields(text)
	return words
}

// multiWordBanks lists multi-word bank names that should also be matched as
// substrings. PDF extractors sometimes concatenate words (e.g., "deMercado Pago"
// or "Mercado Pago.2"), which breaks word-boundary matching. A case-insensitive
// substring check catches these cases.
var multiWordBanks = []struct {
	substring string
	canonical string
}{
	{"mercado pago", "Mercado Pago"},
	{"naranja x", "Naranja X"},
}

// findBankInText searches for bank names by matching words in the text.
// Returns the canonical bank name if found, or empty string if not found.
// Multiple words can match the same bank - this is acceptable and expected.
func findBankInText(text string) string {
	textLower := strings.ToLower(text)

	// Split text into words (handle punctuation, spaces, newlines)
	words := extractWords(textLower)

	// Check 2-word phrases first (e.g., "mercado pago", "naranja x")
	for i := 0; i < len(words)-1; i++ {
		phrase := words[i] + " " + words[i+1]
		if bank, ok := wordToBank[phrase]; ok {
			return bank
		}
	}

	// Check each word against the word-to-bank map
	for _, word := range words {
		if bank, ok := wordToBank[word]; ok {
			return bank
		}
	}

	// Fallback: substring search for multi-word banks whose words may be
	// concatenated with adjacent text in the PDF (e.g., "deMercado Pago.2").
	for _, mw := range multiWordBanks {
		if strings.Contains(textLower, mw.substring) {
			return mw.canonical
		}
	}

	return ""
}

// DetectBankFromText extracts the bank name from PDF text content.
// It uses simple word-based matching and returns the canonical bank name,
// or "?" if no match is found.
func DetectBankFromText(text string) Bank {
	matchedBank := findBankInText(text)
	if matchedBank == "" {
		return Bank("?")
	}
	return Bank(matchedBank)
}
