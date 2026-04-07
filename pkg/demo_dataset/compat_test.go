package demodataset

import (
	"testing"
	"time"

	"github.com/Alechan/finance-analyzer/pkg/internal/pdfcardsummary"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestDemoDataset_CanBeParsedIntoRepoCompatibleStructs(t *testing.T) {
	// Given
	csvBytes := []byte(ExtractedCSV)
	expectedRows := expectedMovementWithCardContextRows()

	// When
	actualRows, err := pdfcardsummary.ParseMovementsWithCardContextCSV(csvBytes)

	// Then
	require.NoError(t, err)
	require.Equal(t, expectedRows, actualRows)
}

func expectedMovementWithCardContextRows() []pdfcardsummary.MovementWithCardContext {
	statementJan := pdfcardsummary.StatementContext{
		Bank:           pdfcardsummary.Bank("DEMO_BANK"),
		CardCompany:    pdfcardsummary.CardCompany("VISA"),
		CloseDate:      date(2025, time.January, 25),
		ExpirationDate: date(2025, time.February, 10),
		TotalARS:       dec("0"),
		TotalUSD:       dec("0"),
	}
	statementFeb := pdfcardsummary.StatementContext{
		Bank:           pdfcardsummary.Bank("DEMO_BANK"),
		CardCompany:    pdfcardsummary.CardCompany("VISA"),
		CloseDate:      date(2025, time.February, 25),
		ExpirationDate: date(2025, time.March, 10),
		TotalARS:       dec("0"),
		TotalUSD:       dec("0"),
	}
	statementMarVisa := pdfcardsummary.StatementContext{
		Bank:           pdfcardsummary.Bank("DEMO_BANK"),
		CardCompany:    pdfcardsummary.CardCompany("VISA"),
		CloseDate:      date(2025, time.March, 25),
		ExpirationDate: date(2025, time.April, 10),
		TotalARS:       dec("0"),
		TotalUSD:       dec("0"),
	}
	statementMarAmex := pdfcardsummary.StatementContext{
		Bank:           pdfcardsummary.Bank("DEMO_BANK"),
		CardCompany:    pdfcardsummary.CardCompany("AMEX"),
		CloseDate:      date(2025, time.March, 25),
		ExpirationDate: date(2025, time.April, 10),
		TotalARS:       dec("0"),
		TotalUSD:       dec("0"),
	}

	card0000OwnerA := &pdfcardsummary.CardContext{
		CardNumber:   strPtr("0000"),
		CardOwner:    "OWNER A",
		CardTotalARS: dec("0"),
		CardTotalUSD: dec("0"),
	}
	card1111OwnerB := &pdfcardsummary.CardContext{
		CardNumber:   strPtr("1111"),
		CardOwner:    "OWNER B",
		CardTotalARS: dec("0"),
		CardTotalUSD: dec("0"),
	}
	card2222OwnerA := &pdfcardsummary.CardContext{
		CardNumber:   strPtr("2222"),
		CardOwner:    "OWNER A",
		CardTotalARS: dec("0"),
		CardTotalUSD: dec("0"),
	}

	return []pdfcardsummary.MovementWithCardContext{
		{StatementContext: statementJan, CardContext: nil, MovementType: pdfcardsummary.MovementTypePastPayment, Movement: pdfcardsummary.Movement{Detail: "SALDO ANTERIOR", AmountARS: dec("-10000"), AmountUSD: dec("0")}},
		{StatementContext: statementJan, CardContext: nil, MovementType: pdfcardsummary.MovementTypeTax, Movement: pdfcardsummary.Movement{OriginalDate: datePtr(2025, time.January, 24), Detail: "IVA DEMO", AmountARS: dec("500"), AmountUSD: dec("0")}},
		{StatementContext: statementJan, CardContext: card0000OwnerA, MovementType: pdfcardsummary.MovementTypeCard, Movement: pdfcardsummary.Movement{OriginalDate: datePtr(2025, time.January, 10), ReceiptNumber: strPtr("1001"), Detail: "SUPERMARKET DEMO", AmountARS: dec("12345.67"), AmountUSD: dec("0")}},
		{StatementContext: statementJan, CardContext: card0000OwnerA, MovementType: pdfcardsummary.MovementTypeCard, Movement: pdfcardsummary.Movement{OriginalDate: datePtr(2025, time.January, 12), ReceiptNumber: strPtr("1002"), Detail: "STREAMING DEMO", AmountARS: dec("2000"), AmountUSD: dec("0")}},
		{StatementContext: statementJan, CardContext: card1111OwnerB, MovementType: pdfcardsummary.MovementTypeCard, Movement: pdfcardsummary.Movement{OriginalDate: datePtr(2025, time.January, 15), ReceiptNumber: strPtr("1003"), Detail: "UNMAPPED MERCHANT", AmountARS: dec("1111"), AmountUSD: dec("0")}},
		{StatementContext: statementJan, CardContext: card0000OwnerA, MovementType: pdfcardsummary.MovementTypeCard, Movement: pdfcardsummary.Movement{OriginalDate: datePtr(2025, time.January, 20), ReceiptNumber: strPtr("1010"), Detail: "REFUNDABLE DEMO", AmountARS: dec("1234"), AmountUSD: dec("0")}},
		{StatementContext: statementJan, CardContext: card0000OwnerA, MovementType: pdfcardsummary.MovementTypeCard, Movement: pdfcardsummary.Movement{OriginalDate: datePtr(2025, time.January, 22), ReceiptNumber: strPtr("1011"), Detail: "PHONE DEMO", CurrentInstallment: intPtr(1), TotalInstallments: intPtr(3), AmountARS: dec("5000"), AmountUSD: dec("0")}},

		{StatementContext: statementFeb, CardContext: nil, MovementType: pdfcardsummary.MovementTypePastPayment, Movement: pdfcardsummary.Movement{Detail: "SALDO ANTERIOR", AmountARS: dec("-15000"), AmountUSD: dec("0")}},
		{StatementContext: statementFeb, CardContext: nil, MovementType: pdfcardsummary.MovementTypeTax, Movement: pdfcardsummary.Movement{OriginalDate: datePtr(2025, time.February, 24), Detail: "IMPUESTO DEMO", AmountARS: dec("800"), AmountUSD: dec("0")}},
		{StatementContext: statementFeb, CardContext: card0000OwnerA, MovementType: pdfcardsummary.MovementTypeCard, Movement: pdfcardsummary.Movement{OriginalDate: datePtr(2025, time.February, 3), ReceiptNumber: strPtr("2001"), Detail: "TRANSPORT DEMO", AmountARS: dec("3500"), AmountUSD: dec("0")}},
		{StatementContext: statementFeb, CardContext: card0000OwnerA, MovementType: pdfcardsummary.MovementTypeCard, Movement: pdfcardsummary.Movement{OriginalDate: datePtr(2025, time.February, 7), ReceiptNumber: strPtr("2002"), Detail: "SUPERMARKET DEMO", AmountARS: dec("8000"), AmountUSD: dec("0")}},
		{StatementContext: statementFeb, CardContext: card0000OwnerA, MovementType: pdfcardsummary.MovementTypeCard, Movement: pdfcardsummary.Movement{OriginalDate: datePtr(2025, time.February, 9), ReceiptNumber: strPtr("2003"), Detail: "PHONE DEMO", CurrentInstallment: intPtr(2), TotalInstallments: intPtr(3), AmountARS: dec("5000"), AmountUSD: dec("0")}},
		{StatementContext: statementFeb, CardContext: card0000OwnerA, MovementType: pdfcardsummary.MovementTypeCard, Movement: pdfcardsummary.Movement{OriginalDate: datePtr(2025, time.February, 10), ReceiptNumber: strPtr("2010"), Detail: "REFUNDABLE DEMO", AmountARS: dec("-1234"), AmountUSD: dec("0")}},
		{StatementContext: statementFeb, CardContext: card0000OwnerA, MovementType: pdfcardsummary.MovementTypeCard, Movement: pdfcardsummary.Movement{OriginalDate: datePtr(2025, time.February, 12), ReceiptNumber: strPtr("2011"), Detail: "SAME MONTH PAIR DEMO", AmountARS: dec("777"), AmountUSD: dec("0")}},
		{StatementContext: statementFeb, CardContext: card0000OwnerA, MovementType: pdfcardsummary.MovementTypeCard, Movement: pdfcardsummary.Movement{OriginalDate: datePtr(2025, time.February, 13), ReceiptNumber: strPtr("2012"), Detail: "SAME MONTH PAIR DEMO", AmountARS: dec("-777"), AmountUSD: dec("0")}},
		{StatementContext: statementFeb, CardContext: card1111OwnerB, MovementType: pdfcardsummary.MovementTypeCard, Movement: pdfcardsummary.Movement{OriginalDate: datePtr(2025, time.February, 11), ReceiptNumber: strPtr("2004"), Detail: "SUPERMARKET DEMO", AmountARS: dec("4444"), AmountUSD: dec("0")}},

		{StatementContext: statementMarVisa, CardContext: nil, MovementType: pdfcardsummary.MovementTypePastPayment, Movement: pdfcardsummary.Movement{Detail: "SALDO ANTERIOR", AmountARS: dec("-9000"), AmountUSD: dec("0")}},
		{StatementContext: statementMarVisa, CardContext: nil, MovementType: pdfcardsummary.MovementTypeTax, Movement: pdfcardsummary.Movement{OriginalDate: datePtr(2025, time.March, 24), Detail: "IVA DEMO", AmountARS: dec("600"), AmountUSD: dec("0")}},
		{StatementContext: statementMarVisa, CardContext: card0000OwnerA, MovementType: pdfcardsummary.MovementTypeCard, Movement: pdfcardsummary.Movement{OriginalDate: datePtr(2025, time.March, 1), ReceiptNumber: strPtr("3000"), Detail: "SUPERMARKET DEMO", AmountARS: dec("7777"), AmountUSD: dec("0")}},
		{StatementContext: statementMarVisa, CardContext: card0000OwnerA, MovementType: pdfcardsummary.MovementTypeCard, Movement: pdfcardsummary.Movement{OriginalDate: datePtr(2025, time.March, 2), ReceiptNumber: strPtr("3003"), Detail: "SUPERMARKET  DEMO", AmountARS: dec("1000"), AmountUSD: dec("0")}},
		{StatementContext: statementMarVisa, CardContext: card0000OwnerA, MovementType: pdfcardsummary.MovementTypeCard, Movement: pdfcardsummary.Movement{OriginalDate: datePtr(2025, time.March, 6), ReceiptNumber: strPtr("3002"), Detail: "SOFTWARE DEMO USD", AmountARS: dec("0"), AmountUSD: dec("50")}},
		{StatementContext: statementMarVisa, CardContext: card0000OwnerA, MovementType: pdfcardsummary.MovementTypeCard, Movement: pdfcardsummary.Movement{OriginalDate: datePtr(2025, time.March, 9), ReceiptNumber: strPtr("3004"), Detail: "PHONE DEMO", CurrentInstallment: intPtr(3), TotalInstallments: intPtr(3), AmountARS: dec("5000"), AmountUSD: dec("0")}},
		{StatementContext: statementMarAmex, CardContext: card2222OwnerA, MovementType: pdfcardsummary.MovementTypeCard, Movement: pdfcardsummary.Movement{OriginalDate: datePtr(2025, time.March, 5), ReceiptNumber: strPtr("3001"), Detail: "AMEX MERCHANT DEMO", AmountARS: dec("6500"), AmountUSD: dec("0")}},
	}
}

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func datePtr(year int, month time.Month, day int) *time.Time {
	t := date(year, month, day)
	return &t
}

func dec(s string) decimal.Decimal {
	d := decimal.RequireFromString(s)
	return decimal.RequireFromString(d.StringFixed(2))
}

func strPtr(s string) *string {
	return &s
}

func intPtr(v int) *int {
	return &v
}
