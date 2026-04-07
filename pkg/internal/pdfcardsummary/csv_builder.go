package pdfcardsummary

import (
	"errors"
	"fmt"
	"github.com/Alechan/finance-analyzer/pkg/internal/platform/decimalale"
	"github.com/Alechan/finance-analyzer/pkg/internal/platform/stringsale"
	"slices"
)

var (
	// The defaultColumns is hardcoded and is coupled to the positions of the records (and vice versa)
	defaultColumns = []string{
		"Bank",
		"CardCompany",
		"CloseDate",
		"ExpirationDate",
		"TotalARS",
		"TotalUSD",
		"CardNumber",
		"CardOwner",
		"CardTotalARS",
		"CardTotalUSD",
		"MovementType",
		"OriginalDate",
		"ReceiptNumber",
		"Detail",
		"CurrentInstallment",
		"TotalInstallments",
		"AmountARS",
		"AmountUSD",
	}
)

type CSVBuilder struct {
	common map[string]string
	rows   []map[string]string
}

func NewCSVBuilder(statement StatementContext) (CSVBuilder, error) {
	if statement.CloseDate.IsZero() {
		return CSVBuilder{}, errors.New("close date cannot be zero")
	}
	if statement.ExpirationDate.IsZero() {
		return CSVBuilder{}, errors.New("expiration date cannot be zero")
	}

	common := map[string]string{
		"Bank":           string(statement.Bank),
		"CardCompany":    string(statement.CardCompany),
		"TotalARS":       decimalale.FormatToArgentineSeparators(statement.TotalARS),
		"TotalUSD":       decimalale.FormatToArgentineSeparators(statement.TotalUSD),
		"CloseDate":      statement.CloseDate.Format("2006-01-02"),
		"ExpirationDate": statement.ExpirationDate.Format("2006-01-02"),
	}

	return CSVBuilder{
		common: common,
		rows:   make([]map[string]string, 0),
	}, nil
}

func (b *CSVBuilder) addMovement(m Movement, movementType MovementType, card *CardContext) {
	row := make(map[string]string)

	// Copy common fields
	for k, v := range b.common {
		row[k] = v
	}

	// Set movement type
	row["MovementType"] = string(movementType)

	// Handle card fields
	if card != nil {
		row["CardNumber"] = stringsale.StringPtrToString(card.CardNumber, "")
		row["CardOwner"] = card.CardOwner
		row["CardTotalARS"] = decimalale.FormatToArgentineSeparators(card.CardTotalARS)
		row["CardTotalUSD"] = decimalale.FormatToArgentineSeparators(card.CardTotalUSD)
	} else {
		row["CardNumber"] = ""
		row["CardOwner"] = ""
		row["CardTotalARS"] = ""
		row["CardTotalUSD"] = ""
	}

	// Movement details
	row["OriginalDate"] = stringsale.TimePtrToYearMonthAndDateString(m.OriginalDate, "")
	row["ReceiptNumber"] = stringsale.StringPtrToString(m.ReceiptNumber, "")
	row["Detail"] = m.Detail
	row["CurrentInstallment"] = stringsale.IntPtrToString(m.CurrentInstallment, "")
	row["TotalInstallments"] = stringsale.IntPtrToString(m.TotalInstallments, "")
	row["AmountARS"] = decimalale.FormatToArgentineSeparators(m.AmountARS)
	row["AmountUSD"] = decimalale.FormatToArgentineSeparators(m.AmountUSD)

	b.rows = append(b.rows, row)
}

func (b *CSVBuilder) Validate(keysToKeep []string) error {
	if len(b.rows) == 0 {
		return ErrNoRowsToBuild
	}

	for _, row := range b.rows {
		for _, col := range keysToKeep {
			if _, exists := row[col]; !exists {
				return fmt.Errorf("missing column %q in CSV row", col)
			}
		}
	}

	return nil
}

func (b *CSVBuilder) BuildWithDefaultColumns() ([][]string, error) {
	keysToKeepInOrder := defaultColumns
	return b.Build(keysToKeepInOrder)
}

func (b *CSVBuilder) Build(keysToKeepInOrder []string) ([][]string, error) {
	if err := b.Validate(keysToKeepInOrder); err != nil {
		return nil, fmt.Errorf("error validating CSVBuilder: %w", err)
	}

	innerRecords, err := stringsale.SliceOfMapsToSliceOfStrings(b.rows, keysToKeepInOrder)
	if err != nil {
		return nil, fmt.Errorf("error building inner records: %w", err)
	}

	allRecords := slices.Concat(
		[][]string{keysToKeepInOrder},
		innerRecords,
	)
	return allRecords, nil
}
