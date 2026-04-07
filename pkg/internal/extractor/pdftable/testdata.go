package pdftable

// TestTablePositions defines the standard table positions used in tests.
// These positions match the Santander PDF format used in most test cases.
var TestTablePositions = PDFTablePositions{
	OriginalDateStart: 0,
	OriginalDateEnd:   12,
	ReceiptStart:      14,
	ReceiptEnd:        22,
	DetailStart:       24,
	DetailEnd:         74,
	ARSAmountStart:    76,
	ARSAmountEnd:      91,
	USDAmountStart:    93,
	USDAmountEnd:      110,
}

// TestCardMovementText is the raw text that produces TestCardMovementRow when parsed.
var TestCardMovementText = "23 Diciem. 30 111111 *  AN ESTABLISHMENT A          C.07/12                       11.111,11                    "

// TestCardMovementRow represents a typical credit card movement row with all fields populated.
var TestCardMovementRow = Row{
	RawText:                        TestCardMovementText,
	RawOriginalDate:                "23 Diciem. 30",
	RawReceiptNumber:               "111111 *",
	RawDetailWithMaybeInstallments: "AN ESTABLISHMENT A          C.07/12",
	RawAmountARS:                   "11.111,11",
	RawAmountUSD:                   "",
}

// TestSaldoAnteriorText is the raw text that produces TestSaldoAnteriorRow when parsed.
var TestSaldoAnteriorText = "                        SALDO ANTERIOR                                           222.111,66             110,00 "

// TestSaldoAnteriorRow represents a "SALDO ANTERIOR" row with ARS and USD amounts.
var TestSaldoAnteriorRow = Row{
	RawText:                        TestSaldoAnteriorText,
	RawOriginalDate:                "",
	RawReceiptNumber:               "",
	RawDetailWithMaybeInstallments: "SALDO ANTERIOR",
	RawAmountARS:                   "222.111,66",
	RawAmountUSD:                   "110,00",
}

// TestShortText is the raw text that produces TestShortTextRow when parsed.
var TestShortText = "short text"

// TestShortTextRow represents a row with text too short to contain all fields.
var TestShortTextRow = Row{
	RawText:                        TestShortText,
	RawOriginalDate:                "short text",
	RawReceiptNumber:               "",
	RawDetailWithMaybeInstallments: "",
	RawAmountARS:                   "",
	RawAmountUSD:                   "",
}

// TestBoundaryText is the raw text that produces TestBoundaryRow when parsed.
// It's exactly at column ends to test boundary conditions.
var TestBoundaryText = "abcdeFGHIJKLMN"

// TestBoundaryPositions defines positions that match exactly with TestBoundaryText.
var TestBoundaryPositions = PDFTablePositions{
	OriginalDateStart: 0, OriginalDateEnd: 2,
	ReceiptStart: 3, ReceiptEnd: 5,
	DetailStart: 6, DetailEnd: 8,
	ARSAmountStart: 9, ARSAmountEnd: 11,
	USDAmountStart: 12, USDAmountEnd: 13,
}

// TestBoundaryRow represents a row with text exactly matching column boundaries.
var TestBoundaryRow = Row{
	RawText:                        TestBoundaryText,
	RawOriginalDate:                "abc",
	RawReceiptNumber:               "deF",
	RawDetailWithMaybeInstallments: "GHI",
	RawAmountARS:                   "JKL",
	RawAmountUSD:                   "MN",
}

// TestWhitespaceText is the raw text that produces TestWhitespaceRow when parsed.
var TestWhitespaceText = "   01 Enero  01  123456 *  Detalle con espacios   y\tcaracteres!@#   1.234,56     78,90   "

// TestWhitespaceRow represents a row with various whitespace and special characters.
var TestWhitespaceRow = Row{
	RawText:                        TestWhitespaceText,
	RawOriginalDate:                "01 Enero",
	RawReceiptNumber:               "1  123456",
	RawDetailWithMaybeInstallments: "*  Detalle con espacios   y\tcaracteres!@#   1.234,5",
	RawAmountARS:                   "78,90",
	RawAmountUSD:                   "",
}
