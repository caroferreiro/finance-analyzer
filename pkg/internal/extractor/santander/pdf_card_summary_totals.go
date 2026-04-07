package santander

import (
	"fmt"
	"strings"

	"github.com/Alechan/finance-analyzer/pkg/internal/pdfcardsummary"
	"github.com/Alechan/finance-analyzer/pkg/internal/platform/pdfwrapper"
	"github.com/shopspring/decimal"
)

func (r *SantanderExtractor) extractDocumentTotals(pages []pdfwrapper.Page) (decimal.Decimal, decimal.Decimal, error) {
	page, err := r.getTotalAmountsPage(pages)
	if err != nil {
		return decimal.Decimal{}, decimal.Decimal{}, fmt.Errorf("error getting total amounts page: %w", err)
	}

	row, err := r.getTotalAmountsRowContents(page)
	if err != nil {
		return decimal.Decimal{}, decimal.Decimal{}, fmt.Errorf("error getting total amounts row: %w", err)
	}

	if len(row) < 3 {
		return decimal.Decimal{}, decimal.Decimal{}, fmt.Errorf("expected at least 3 columns, got %d", len(row))
	}

	// The first text should pass the sanity check
	if !r.cfg.TotalAmountsSanityCheckRegex.MatchString(row[0]) {
		return decimal.Decimal{}, decimal.Decimal{}, fmt.Errorf("first column doesn't pass the sanity check: %s", row[0])
	}

	// The second text should be the amount in ARS
	arsAmount, err := pdfcardsummary.PDFAmountToDecimal(strings.TrimSpace(row[1]))
	if err != nil {
		return decimal.Decimal{}, decimal.Decimal{}, fmt.Errorf("error converting ARS amount '%s' to decimal: %w", row[1], err)
	}

	// The third text should be the amount in USD
	usdAmount, err := pdfcardsummary.PDFAmountToDecimal(strings.TrimSpace(row[2]))
	if err != nil {
		return decimal.Decimal{}, decimal.Decimal{}, fmt.Errorf("error converting USD amount '%s' to decimal: %w", row[2], err)
	}

	return arsAmount, usdAmount, nil
}

func (r *SantanderExtractor) getTotalAmountsRowContents(page pdfwrapper.Page) ([]string, error) {
	pageRows := page.Rows
	var result []string
	for _, row := range pageRows {
		// We're looking to the row with the specified position
		if row.Position <= r.cfg.TotalAmountsRow {
			for iContent := 0; iContent < len(row.Texts); iContent++ {
				text := row.Texts[iContent]
				result = append(result, text)
			}
			return result, nil
		}
	}

	return nil, fmt.Errorf("couldn't find row with position %d on page %d", r.cfg.TotalAmountsRow, r.cfg.TotalAmountsPage)
}

func (r *SantanderExtractor) getTotalAmountsPage(pages []pdfwrapper.Page) (pdfwrapper.Page, error) {
	if len(pages) <= r.cfg.TotalAmountsPage {
		return pdfwrapper.Page{}, fmt.Errorf(
			"expected at least %d pages, got %d",
			r.cfg.TotalAmountsPage,
			len(pages),
		)
	}

	pageIndex := r.cfg.TotalAmountsPage
	if r.cfg.TotalAmountsPage < 0 {
		pageIndex = len(pages) + r.cfg.TotalAmountsPage
	}

	page := pages[pageIndex]
	return page, nil
}
