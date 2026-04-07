package pdfcardsummary

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type CardSummary struct {
	StatementContext StatementContext
	Table            Table
}

// StatementContext is the statement-level context repeated on every extracted CSV row.
type StatementContext struct {
	Bank           Bank
	CardCompany    CardCompany
	CloseDate      time.Time
	ExpirationDate time.Time
	TotalARS       decimal.Decimal
	TotalUSD       decimal.Decimal
}

// ToCSVBytes returns the CSV representation as bytes, ready to write to a file.
func (cs *CardSummary) ToCSVBytes() ([]byte, error) {
	matrix, err := cs.ToCSVMatrix()
	if err != nil {
		return nil, fmt.Errorf("error converting to CSV: %w", err)
	}

	var rowsAsStrings []string
	for _, row := range matrix {
		rowsAsStrings = append(rowsAsStrings, strings.Join(row, ";"))
	}

	finalString := strings.Join(rowsAsStrings, "\n")

	return []byte(finalString), nil
}

// ToCSVMatrix returns the CSV representation as a matrix of strings (rows of columns).
// Useful when you need to manipulate or combine CSV data before converting to bytes.
func (cs *CardSummary) ToCSVMatrix() ([][]string, error) {
	statement := cs.StatementContext

	builder, err := NewCSVBuilder(statement)
	if err != nil {
		return nil, fmt.Errorf("error creating CSV builder: %w", err)
	}

	factRows := NewMovementsWithCardContext(*cs)
	for _, row := range factRows {
		builder.addMovement(row.Movement, row.MovementType, row.CardContext)
	}

	csv, err := builder.BuildWithDefaultColumns()
	if err != nil {
		return nil, fmt.Errorf("error building CSV: %w", err)
	}

	return csv, nil
}
