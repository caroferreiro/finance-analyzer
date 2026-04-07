package santander

import (
	"regexp"

	"github.com/Alechan/finance-analyzer/pkg/internal/extractor/pdftable"
)

func DefaultConfig() SantanderExtractorConfig {
	return SantanderExtractorConfig{
		ClosingDateRegex:             regexp.MustCompile(`^CIERRE\s+(\d{2}\s+\w{3}\s+\d{2})$`),
		ExpirationDateRegex:          regexp.MustCompile(`^VENCIMIENTO\s+(\d{2}\s+\w{3}\s+\d{2})$`),
		SaldoAnteriorRegex:           regexp.MustCompile(`^\s*SALDO\s+ANTERIOR\s+([\d.,]+)\s+([\d.,]+-?)\s*$`),
		CuotasSubDetailRegex:         regexp.MustCompile(`\s*C\.(\d+)/(\d+)$`),
		EndOfPastPaymentsRegex:       regexp.MustCompile("^[\\-_]{3,}\\s*$"),
		TotalConsumosTarjetaRegex:    regexp.MustCompile(`^\s*(?:Tarjeta (\d{4}) )?Total Consumos de (.*?)\s+([\d.]+,\d{2}) \* +([\d.]+,\d{2}) \*.*$`),
		TotalAmountsSanityCheckRegex: regexp.MustCompile(`^\s*(?:TNA|TEM)\s+\d+,\d{3}(?:\s+(?:TNA|TEM)\s+\d+,\d{3})*\s*$`),
		AmountOnlyRowRegex:           regexp.MustCompile(`^\s{0,75}((?:\d{1,3}\.)+\d+(?:,\d{2})?|\d+,\d{2})-?\s*$|^\s{0,92}((?:\d{1,3}\.)+\d+(?:,\d{2})?|\d+,\d{2})-?\s*$`),
		TableFirstRowPositionInPage:  457,

		// All the PDFs should use the same length for each column, so we can use the same positions for all of them
		TableColumnPositions: pdftable.PDFTablePositions{
			OriginalDateStart: 0,
			OriginalDateEnd:   12,

			ReceiptStart: 14,
			ReceiptEnd:   22,

			DetailStart: 24,
			DetailEnd:   74,

			ARSAmountStart: 76,
			ARSAmountEnd:   91,

			USDAmountStart: 93,
			USDAmountEnd:   110,
		},
		// The total amounts should be in the last page, in row 212
		TotalAmountsPage: -1,
		TotalAmountsRow:  212,
	}
}

type SantanderExtractorConfig struct {
	ClosingDateRegex             *regexp.Regexp
	ExpirationDateRegex          *regexp.Regexp
	SaldoAnteriorRegex           *regexp.Regexp
	EndOfPastPaymentsRegex       *regexp.Regexp
	TotalConsumosTarjetaRegex    *regexp.Regexp
	CuotasSubDetailRegex         *regexp.Regexp
	AmountOnlyRowRegex           *regexp.Regexp
	TableColumnPositions         pdftable.PDFTablePositions
	TableFirstRowPositionInPage  int
	TotalAmountsPage             int
	TotalAmountsRow              int
	TotalAmountsSanityCheckRegex *regexp.Regexp
}
