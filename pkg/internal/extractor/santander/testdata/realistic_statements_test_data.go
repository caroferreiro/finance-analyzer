package testdata

import (
	_ "embed"
	"testing"
	"time"

	"github.com/Alechan/finance-analyzer/pkg/internal/pdfcardsummary"
	"github.com/Alechan/finance-analyzer/pkg/internal/platform/testsale"
	"github.com/shopspring/decimal"

	"github.com/Alechan/finance-analyzer/pkg/internal/extractor/pdftable"
)

//go:embed realistic_statement.txt
var RawRealisticStatementData []byte

// RealisticStatementRows represents the expected rows when parsing realistic_statement.txt
var RealisticStatementRows = []pdftable.Row{
	// 1: "// Previous balance"
	{RawText: "// Previous balance", RawOriginalDate: "// Previous b", RawReceiptNumber: "lance", RawDetailWithMaybeInstallments: "", RawAmountARS: "", RawAmountUSD: ""},
	// 2: "                        SALDO ANTERIOR                                           866.169,54              54,68"
	{RawText: "                        SALDO ANTERIOR                                           866.169,54              54,68", RawOriginalDate: "", RawReceiptNumber: "", RawDetailWithMaybeInstallments: "SALDO ANTERIOR", RawAmountARS: "866.169,54", RawAmountUSD: "54,68"},
	// 3: "// Past payment movements"
	{RawText: "// Past payment movements", RawOriginalDate: "// Past payme", RawReceiptNumber: "t movemen", RawDetailWithMaybeInstallments: "s", RawAmountARS: "", RawAmountUSD: ""},
	// 4: "24 Noviem. 16           ZZZZZZZZZZZZZZZZZZ                                        23.202,00-"
	{RawText: "24 Noviem. 16           ZZZZZZZZZZZZZZZZZZ                                        23.202,00-", RawOriginalDate: "24 Noviem. 16", RawReceiptNumber: "", RawDetailWithMaybeInstallments: "ZZZZZZZZZZZZZZZZZZ", RawAmountARS: "23.202,00-", RawAmountUSD: ""},
	// 5: "           20           SU PAGO EN PESOS                                         832.054,80-"
	{RawText: "           20           SU PAGO EN PESOS                                         832.054,80-", RawOriginalDate: "20", RawReceiptNumber: "", RawDetailWithMaybeInstallments: "SU PAGO EN PESOS", RawAmountARS: "832.054,80-", RawAmountUSD: ""},
	// 6: "           20           SU PAGO EN USD                                                                   54,68-"
	{RawText: "           20           SU PAGO EN USD                                                                   54,68-", RawOriginalDate: "20", RawReceiptNumber: "", RawDetailWithMaybeInstallments: "SU PAGO EN USD", RawAmountARS: "", RawAmountUSD: "54,68-"},
	// 7: "           20           CR.IMPUESTO PAIS 30%                                      16.691,07-"
	{RawText: "           20           CR.IMPUESTO PAIS 30%                                      16.691,07-", RawOriginalDate: "20", RawReceiptNumber: "", RawDetailWithMaybeInstallments: "CR.IMPUESTO PAIS 30%", RawAmountARS: "16.691,07-", RawAmountUSD: ""},
	// 8: "           20           CR.RG 5617 30% M                                          16.691,07-"
	{RawText: "           20           CR.RG 5617 30% M                                          16.691,07-", RawOriginalDate: "20", RawReceiptNumber: "", RawDetailWithMaybeInstallments: "CR.RG 5617 30% M", RawAmountARS: "16.691,07-", RawAmountUSD: ""},
	// 9: "// Separator"
	{RawText: "// Separator", RawOriginalDate: "// Separator", RawReceiptNumber: "", RawDetailWithMaybeInstallments: "", RawAmountARS: "", RawAmountUSD: ""},
	// 10: "________________________________________________________________________________________________________________"
	{RawText: "________________________________________________________________________________________________________________", RawOriginalDate: "_____________", RawReceiptNumber: "_________", RawDetailWithMaybeInstallments: "___________________________________________________", RawAmountARS: "________________", RawAmountUSD: "__________________"},
	// 11: "// Card movements (first card only, a few rows)"
	{RawText: "// Card movements (first card only, a few rows)", RawOriginalDate: "// Card movem", RawReceiptNumber: "nts (firs", RawDetailWithMaybeInstallments: "card only, a few rows)", RawAmountARS: "", RawAmountUSD: ""},
	// 12: "24 Junio   06 123456 *  AAAAA                       C.07/09                       45.776,88"
	{RawText: "24 Junio   06 123456 *  AAAAA                       C.07/09                       45.776,88", RawOriginalDate: "24 Junio   06", RawReceiptNumber: "123456 *", RawDetailWithMaybeInstallments: "AAAAA                       C.07/09", RawAmountARS: "45.776,88", RawAmountUSD: ""},
	// 13: "24 Noviem. 25 789012 *  BBBBBBBBBBB         12345678901                            4.226,53"
	{RawText: "24 Noviem. 25 789012 *  BBBBBBBBBBB         12345678901                            4.226,53", RawOriginalDate: "24 Noviem. 25", RawReceiptNumber: "789012 *", RawDetailWithMaybeInstallments: "BBBBBBBBBBB         12345678901", RawAmountARS: "4.226,53", RawAmountUSD: ""},
	// 14: "           27 345678 *  SEGURO DE VIDA      9876543210                             9.373,07"
	{RawText: "           27 345678 *  SEGURO DE VIDA      9876543210                             9.373,07", RawOriginalDate: "27", RawReceiptNumber: "345678 *", RawDetailWithMaybeInstallments: "SEGURO DE VIDA      9876543210", RawAmountARS: "9.373,07", RawAmountUSD: ""},
	// 15: "// Card total"
	{RawText: "// Card total", RawOriginalDate: "// Card total", RawReceiptNumber: "", RawDetailWithMaybeInstallments: "", RawAmountARS: "", RawAmountUSD: ""},
	// 16: "Tarjeta 1234 Total Consumos de JOOOOOHHN SMITH                                    59.376,48 *            0,00 *"
	{RawText: "Tarjeta 1234 Total Consumos de JOOOOOHHN SMITH                                    59.376,48 *            0,00 *", RawOriginalDate: "Tarjeta 1234", RawReceiptNumber: "otal Cons", RawDetailWithMaybeInstallments: "mos de JOOOOOHHN SMITH", RawAmountARS: "59.376,48", RawAmountUSD: "0,00 *"},
	// 17: "" (blank line)
	{RawText: "", RawOriginalDate: "", RawReceiptNumber: "", RawDetailWithMaybeInstallments: "", RawAmountARS: "", RawAmountUSD: ""},
}

// ExpectedSantanderTableData should be derived from RealisticStatementRows
func ExpectedSantanderTableData(t *testing.T) pdfcardsummary.Table {
	return pdfcardsummary.Table{
		PastPaymentMovements: []pdfcardsummary.Movement{
			{
				OriginalDate:  nil,
				ReceiptNumber: nil,
				Detail:        "SALDO ANTERIOR",
				AmountARS:     testsale.AsDecimal(t, "866169.54"),
				AmountUSD:     testsale.AsDecimal(t, "54.68"),
			},
			{
				OriginalDate:  testsale.DatePtr(2024, time.November, 16),
				ReceiptNumber: nil,
				Detail:        "ZZZZZZZZZZZZZZZZZZ",
				AmountARS:     testsale.AsDecimal(t, "-23202.00"),
				AmountUSD:     decimal.Zero,
			},
			{
				OriginalDate:  testsale.DatePtr(2024, time.November, 20),
				ReceiptNumber: nil,
				Detail:        "SU PAGO EN PESOS",
				AmountARS:     testsale.AsDecimal(t, "-832054.80"),
				AmountUSD:     decimal.Zero,
			},
			{
				OriginalDate:  testsale.DatePtr(2024, time.November, 20),
				ReceiptNumber: nil,
				Detail:        "SU PAGO EN USD",
				AmountARS:     decimal.Zero,
				AmountUSD:     testsale.AsDecimal(t, "-54.68"),
			},
			{
				OriginalDate:  testsale.DatePtr(2024, time.November, 20),
				ReceiptNumber: nil,
				Detail:        "CR.IMPUESTO PAIS 30%",
				AmountARS:     testsale.AsDecimal(t, "-16691.07"),
				AmountUSD:     decimal.Zero,
			},
			{
				OriginalDate:  testsale.DatePtr(2024, time.November, 20),
				ReceiptNumber: nil,
				Detail:        "CR.RG 5617 30% M",
				AmountARS:     testsale.AsDecimal(t, "-16691.07"),
				AmountUSD:     decimal.Zero,
			},
		},
		Cards: []pdfcardsummary.Card{
			{
				CardContext: pdfcardsummary.CardContext{
					CardNumber:   testsale.StrPtr("1234"),
					CardOwner:    "JOOOOOHHN SMITH",
					CardTotalARS: testsale.AsDecimal(t, "59376.48"),
					CardTotalUSD: decimal.Zero,
				},
				Movements: []pdfcardsummary.Movement{
					{testsale.DatePtr(2024, time.June, 6), testsale.StrPtr("123456 *"), "AAAAA", testsale.IntPtr(7), testsale.IntPtr(9), testsale.AsDecimal(t, "45776.88"), decimal.Zero},
					{testsale.DatePtr(2024, time.November, 25), testsale.StrPtr("789012 *"), "BBBBBBBBBBB 12345678901", nil, nil, testsale.AsDecimal(t, "4226.53"), decimal.Zero},
					{testsale.DatePtr(2024, time.November, 27), testsale.StrPtr("345678 *"), "SEGURO DE VIDA 9876543210", nil, nil, testsale.AsDecimal(t, "9373.07"), decimal.Zero},
				},
			},
		},
		TaxesMovements: nil,
	}
}
