package pdfcardsummary

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// ParseMovementsWithCardContextCSV parses extracted CSV bytes into fact rows.
// It is filename-agnostic by design; upper layers should attach filename context to returned errors.
func ParseMovementsWithCardContextCSV(csvBytes []byte) ([]MovementWithCardContext, error) {
	reader := csv.NewReader(bytes.NewReader(csvBytes))
	reader.Comma = ';'
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV records: %v", err)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("CSV has no rows")
	}

	header := records[0]
	if !slices.Equal(header, defaultColumns) {
		return nil, fmt.Errorf("invalid CSV header: expected %v, got %v", defaultColumns, header)
	}

	rows := make([]MovementWithCardContext, 0, len(records)-1)
	for i, record := range records[1:] {
		rowIndex1 := i + 1
		row, err := parseMovementWithCardContextRecord(rowIndex1, record)
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}

	return rows, nil
}

func parseMovementWithCardContextRecord(rowIndex1 int, record []string) (MovementWithCardContext, error) {
	if len(record) != len(defaultColumns) {
		return MovementWithCardContext{}, fmt.Errorf("row %d: expected %d columns, got %d", rowIndex1, len(defaultColumns), len(record))
	}

	valueByColumn := make(map[string]string, len(defaultColumns))
	for i, column := range defaultColumns {
		valueByColumn[column] = record[i]
	}

	statement, err := parseStatementContext(rowIndex1, valueByColumn)
	if err != nil {
		return MovementWithCardContext{}, err
	}

	movementType, err := parseMovementType(rowIndex1, valueByColumn["MovementType"])
	if err != nil {
		return MovementWithCardContext{}, err
	}

	cardContext, err := parseCardContext(rowIndex1, movementType, valueByColumn)
	if err != nil {
		return MovementWithCardContext{}, err
	}

	movement, err := parseMovement(rowIndex1, valueByColumn)
	if err != nil {
		return MovementWithCardContext{}, err
	}

	return MovementWithCardContext{
		StatementContext: statement,
		CardContext:      cardContext,
		MovementType:     movementType,
		Movement:         movement,
	}, nil
}

func parseStatementContext(rowIndex1 int, m map[string]string) (StatementContext, error) {
	closeDate, err := parseRequiredDate(rowIndex1, "CloseDate", m["CloseDate"])
	if err != nil {
		return StatementContext{}, err
	}
	expirationDate, err := parseRequiredDate(rowIndex1, "ExpirationDate", m["ExpirationDate"])
	if err != nil {
		return StatementContext{}, err
	}

	totalARS, err := parseRequiredDecimalArgentine(rowIndex1, "TotalARS", m["TotalARS"])
	if err != nil {
		return StatementContext{}, err
	}
	totalUSD, err := parseRequiredDecimalArgentine(rowIndex1, "TotalUSD", m["TotalUSD"])
	if err != nil {
		return StatementContext{}, err
	}

	return StatementContext{
		Bank:           Bank(m["Bank"]),
		CardCompany:    CardCompany(m["CardCompany"]),
		CloseDate:      closeDate,
		ExpirationDate: expirationDate,
		TotalARS:       totalARS,
		TotalUSD:       totalUSD,
	}, nil
}

func parseCardContext(rowIndex1 int, movementType MovementType, m map[string]string) (*CardContext, error) {
	cardNumberRaw := m["CardNumber"]
	cardOwnerRaw := m["CardOwner"]
	cardTotalARSRaw := m["CardTotalARS"]
	cardTotalUSDRaw := m["CardTotalUSD"]

	if movementType != MovementTypeCard {
		if cardNumberRaw != "" || cardOwnerRaw != "" || cardTotalARSRaw != "" || cardTotalUSDRaw != "" {
			return nil, fmt.Errorf(
				"row %d: non-card movement type %q must have empty card columns, got CardNumber=%q CardOwner=%q CardTotalARS=%q CardTotalUSD=%q",
				rowIndex1, movementType, cardNumberRaw, cardOwnerRaw, cardTotalARSRaw, cardTotalUSDRaw,
			)
		}
		return nil, nil
	}

	cardTotalARS, err := parseRequiredDecimalArgentine(rowIndex1, "CardTotalARS", cardTotalARSRaw)
	if err != nil {
		return nil, err
	}
	cardTotalUSD, err := parseRequiredDecimalArgentine(rowIndex1, "CardTotalUSD", cardTotalUSDRaw)
	if err != nil {
		return nil, err
	}

	var cardNumber *string
	if cardNumberRaw != "" {
		s := cardNumberRaw
		cardNumber = &s
	}

	return &CardContext{
		CardNumber:   cardNumber,
		CardOwner:    cardOwnerRaw,
		CardTotalARS: cardTotalARS,
		CardTotalUSD: cardTotalUSD,
	}, nil
}

func parseMovement(rowIndex1 int, m map[string]string) (Movement, error) {
	originalDate, err := parseOptionalDate(rowIndex1, "OriginalDate", m["OriginalDate"])
	if err != nil {
		return Movement{}, err
	}

	currentInstallment, err := parseOptionalInt(rowIndex1, "CurrentInstallment", m["CurrentInstallment"])
	if err != nil {
		return Movement{}, err
	}
	totalInstallments, err := parseOptionalInt(rowIndex1, "TotalInstallments", m["TotalInstallments"])
	if err != nil {
		return Movement{}, err
	}

	amountARS, err := parseRequiredDecimalArgentine(rowIndex1, "AmountARS", m["AmountARS"])
	if err != nil {
		return Movement{}, err
	}
	amountUSD, err := parseRequiredDecimalArgentine(rowIndex1, "AmountUSD", m["AmountUSD"])
	if err != nil {
		return Movement{}, err
	}

	var receiptNumber *string
	if m["ReceiptNumber"] != "" {
		s := m["ReceiptNumber"]
		receiptNumber = &s
	}

	return Movement{
		OriginalDate:       originalDate,
		ReceiptNumber:      receiptNumber,
		Detail:             m["Detail"],
		CurrentInstallment: currentInstallment,
		TotalInstallments:  totalInstallments,
		AmountARS:          amountARS,
		AmountUSD:          amountUSD,
	}, nil
}

func parseMovementType(rowIndex1 int, raw string) (MovementType, error) {
	switch MovementType(raw) {
	case MovementTypePastPayment, MovementTypeTax, MovementTypeCard:
		return MovementType(raw), nil
	default:
		return "", fmt.Errorf("row %d col %q: invalid movement type %q", rowIndex1, "MovementType", raw)
	}
}

func parseRequiredDate(rowIndex1 int, column string, raw string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("row %d col %q: invalid date %q: %v", rowIndex1, column, raw, err)
	}
	return t, nil
}

func parseOptionalDate(rowIndex1 int, column string, raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, fmt.Errorf("row %d col %q: invalid date %q: %v", rowIndex1, column, raw, err)
	}
	return &t, nil
}

func parseOptionalInt(rowIndex1 int, column string, raw string) (*int, error) {
	if raw == "" {
		return nil, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return nil, fmt.Errorf("row %d col %q: invalid integer %q: %v", rowIndex1, column, raw, err)
	}
	return &v, nil
}

func parseRequiredDecimalArgentine(rowIndex1 int, column string, raw string) (decimal.Decimal, error) {
	if raw == "" {
		return decimal.Decimal{}, fmt.Errorf("row %d col %q: empty decimal value", rowIndex1, column)
	}
	normalized := strings.ReplaceAll(raw, ".", "")
	normalized = strings.ReplaceAll(normalized, ",", ".")
	d, err := decimal.NewFromString(normalized)
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("row %d col %q: invalid decimal %q: %v", rowIndex1, column, raw, err)
	}
	return d, nil
}
