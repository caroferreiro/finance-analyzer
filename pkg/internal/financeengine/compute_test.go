package financeengine

import (
	"encoding/json"
	"strings"
	"testing"

	demodataset "github.com/Alechan/finance-analyzer/pkg/demo_dataset"
	"github.com/Alechan/finance-analyzer/pkg/internal/pdfcardsummary"
	"github.com/stretchr/testify/require"
)

func TestEngine_Compute_WhenUsingDemoDataset_ThenReturnsAllExpectedTables(t *testing.T) {
	// Given
	rows, err := pdfcardsummary.ParseMovementsWithCardContextCSV([]byte(demodataset.ExtractedCSV))
	require.NoError(t, err)

	var mappings Mappings
	err = json.Unmarshal([]byte(demodataset.MappingsV1JSON), &mappings)
	require.NoError(t, err)

	engine := New()

	// When
	result := engine.Compute(rows, mappings)

	// Then
	require.Equal(t, []string{
		TableIDMetaSummary,
		TableIDOverviewByMonth,
		TableIDOverviewMetricsByMonth,
		TableIDDebtMaturitySchedule,
		TableIDSpendByOwner,
		TableIDSpendByCategory,
		TableIDCategoryBreakdownByMonth,
		TableIDRawExplorerRows,
		TableIDDQIssues,
		TableIDDQSummaryByRule,
	}, tableIDs(result.Tables))

	meta := tableByID(t, result.Tables, TableIDMetaSummary)
	require.Equal(t, [][]string{
		{"row_count", "23"},
		{"statement_month_min", "2025-01-01"},
		{"statement_month_max", "2025-03-01"},
		{"statement_month_count", "3"},
	}, meta.Rows)

	dqSummary := tableByID(t, result.Tables, TableIDDQSummaryByRule)
	require.Equal(t, [][]string{
		{DQRuleMissingCategoryMappingForCardMovement, "1"},
		{DQRuleMissingOwnerMappingForCardMovement, "2"},
	}, dqSummary.Rows)

	debtMaturity := tableByID(t, result.Tables, TableIDDebtMaturitySchedule)
	require.Equal(t, []string{
		"base_statement_month_date",
		"maturity_month_date",
		"month_offset",
		"installment_count",
		"maturity_total_ars",
		"maturity_total_usd",
	}, tableColumnKeys(debtMaturity))
	if len(debtMaturity.Rows) > 0 {
		require.Equal(t, "2025-03-01", debtMaturity.Rows[0][0])
		require.Equal(t, "1", debtMaturity.Rows[0][2])
	}

	rawRows := tableByID(t, result.Tables, TableIDRawExplorerRows)
	require.Len(t, rawRows.Columns, 14)
	require.Len(t, rawRows.Rows, 23)
	require.Equal(t, "card_statement_close_date", rawRows.Columns[0].Key)
	require.Equal(t, "card_statement_due_date", rawRows.Columns[1].Key)
	require.Equal(t, "amount_ars", rawRows.Columns[12].Key)
	require.Equal(t, "amount_usd", rawRows.Columns[13].Key)
	require.True(t, strings.HasPrefix(rawRows.Rows[0][0], "2025-03"))

	columnKeys := tableColumnKeys(rawRows)
	require.NotContains(t, columnKeys, "total_ars")
	require.NotContains(t, columnKeys, "total_usd")
	require.NotContains(t, columnKeys, "card_total_ars")
	require.NotContains(t, columnKeys, "card_total_usd")
}

func tableIDs(tables []Table) []string {
	out := make([]string, 0, len(tables))
	for _, table := range tables {
		out = append(out, table.TableID)
	}
	return out
}

func tableByID(t *testing.T, tables []Table, tableID string) Table {
	t.Helper()
	for _, table := range tables {
		if table.TableID == tableID {
			return table
		}
	}
	t.Fatalf("table %q not found", tableID)
	return Table{}
}

func tableColumnKeys(table Table) []string {
	out := make([]string, 0, len(table.Columns))
	for _, column := range table.Columns {
		out = append(out, column.Key)
	}
	return out
}
