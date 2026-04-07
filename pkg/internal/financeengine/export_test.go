package financeengine

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExportTableCSV_WhenTableIsValid_ThenReturnsExpectedCSV(t *testing.T) {
	// Given
	table := Table{
		TableID: "test_table",
		Title:   "Test Table",
		Columns: []TableColumn{
			{Key: "month", Label: "Month", Type: ColumnTypeDate},
			{Key: "amount_ars", Label: "Amount ARS", Type: ColumnTypeMoneyARS},
		},
		Rows: [][]string{
			{"2025-01-01", "100.00"},
			{"2025-02-01", "200.00"},
		},
	}

	// When
	actual, err := ExportTableCSV(table)

	// Then
	require.NoError(t, err)
	require.Equal(t, "month;amount_ars\n2025-01-01;100.00\n2025-02-01;200.00\n", actual)
}

func TestExportTableCSV_WhenRowWidthMismatch_ThenReturnsError(t *testing.T) {
	// Given
	table := Table{
		TableID: "bad_table",
		Title:   "Bad Table",
		Columns: []TableColumn{
			{Key: "col_a", Label: "A", Type: ColumnTypeString},
			{Key: "col_b", Label: "B", Type: ColumnTypeString},
		},
		Rows: [][]string{
			{"only_one_column"},
		},
	}

	// When
	csvText, err := ExportTableCSV(table)

	// Then
	require.Equal(t, "", csvText)
	require.EqualError(t, err, "row 1 has 1 columns, expected 2")
}
