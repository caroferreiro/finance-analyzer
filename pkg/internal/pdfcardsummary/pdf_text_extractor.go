package pdfcardsummary

import (
	"strings"

	"github.com/Alechan/finance-analyzer/pkg/internal/platform/pdfwrapper"
)

// ExtractAllTextFromDocument extracts all text content from a PDF document
// and returns it as a single string with spaces between words and newlines between rows.
func ExtractAllTextFromDocument(doc pdfwrapper.Document) string {
	var allText []string
	for _, page := range doc.Pages {
		for _, row := range page.Rows {
			// Join all text segments in the row with spaces
			rowText := strings.Join(row.Texts, " ")
			if rowText != "" {
				allText = append(allText, rowText)
			}
		}
	}
	return strings.Join(allText, "\n")
}
