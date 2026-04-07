package financeengine

import (
	"encoding/json"
	"testing"
	"time"

	demodataset "github.com/Alechan/finance-analyzer/pkg/demo_dataset"
	"github.com/Alechan/finance-analyzer/pkg/internal/pdfcardsummary"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestEngine_OverviewByStatementMonth_WhenUsingDemoDataset_ThenReturnsExpectedRows(t *testing.T) {
	// Given
	parsedRows, err := pdfcardsummary.ParseMovementsWithCardContextCSV([]byte(demodataset.ExtractedCSV))
	require.NoError(t, err)

	engine := New()
	expectedRows := []OverviewByStatementMonthRow{
		{
			StatementMonthDate:   time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
			CardMovementTotalARS: decimal2("21690.67"),
			CardMovementTotalUSD: decimal2("0"),
		},
		{
			StatementMonthDate:   time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC),
			CardMovementTotalARS: decimal2("19710"),
			CardMovementTotalUSD: decimal2("0"),
		},
		{
			StatementMonthDate:   time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC),
			CardMovementTotalARS: decimal2("20277"),
			CardMovementTotalUSD: decimal2("50"),
		},
	}

	// When
	actualRows := engine.OverviewByStatementMonth(parsedRows)

	// Then
	require.Equal(t, expectedRows, normalizeOverviewRows(actualRows))
}

func TestEngine_SpendByOwner_WhenUsingDemoDataset_ThenReturnsExpectedRows(t *testing.T) {
	// Given
	parsedRows, err := pdfcardsummary.ParseMovementsWithCardContextCSV([]byte(demodataset.ExtractedCSV))
	require.NoError(t, err)

	engine := New()
	expectedRows := []SpendByOwnerRow{
		{
			Owner:                "OWNER A",
			Month:                time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC),
			CardMovementTotalARS: decimal2("20277"),
			CardMovementTotalUSD: decimal2("50"),
		},
		{
			Owner:                "OWNER A",
			Month:                time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC),
			CardMovementTotalARS: decimal2("15266"),
			CardMovementTotalUSD: decimal2("0"),
		},
		{
			Owner:                "OWNER B",
			Month:                time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC),
			CardMovementTotalARS: decimal2("4444"),
			CardMovementTotalUSD: decimal2("0"),
		},
		{
			Owner:                "OWNER A",
			Month:                time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
			CardMovementTotalARS: decimal2("20579.67"),
			CardMovementTotalUSD: decimal2("0"),
		},
		{
			Owner:                "OWNER B",
			Month:                time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
			CardMovementTotalARS: decimal2("1111"),
			CardMovementTotalUSD: decimal2("0"),
		},
	}

	// When
	actualRows := engine.SpendByOwner(parsedRows)

	// Then
	require.Equal(t, expectedRows, normalizeSpendByOwnerRows(actualRows))
}

func TestEngine_SpendByCategory_WhenUsingDemoDataset_ThenReturnsExpectedPivotTable(t *testing.T) {
	// Given
	parsedRows, err := pdfcardsummary.ParseMovementsWithCardContextCSV([]byte(demodataset.ExtractedCSV))
	require.NoError(t, err)

	var mappings Mappings
	require.NoError(t, json.Unmarshal([]byte(demodataset.MappingsV1JSON), &mappings))

	engine := New()

	// When
	result := engine.Compute(parsedRows, mappings)
	tbl := tableByID(t, result.Tables, TableIDSpendByCategory)

	// Then: Month first column, months as rows (desc), categories as columns (ARS + USD each)
	require.Equal(t, "month", tbl.Columns[0].Key)
	require.GreaterOrEqual(t, len(tbl.Columns), 3) // Month + at least one category (ARS + USD)
	require.Equal(t, 3, len(tbl.Rows))             // 3 statement months

	// Months sorted newest first
	require.Equal(t, "2025-03-01", tbl.Rows[0][0])
	require.Equal(t, "2025-02-01", tbl.Rows[1][0])
	require.Equal(t, "2025-01-01", tbl.Rows[2][0])
}

func TestEngine_Compute_WhenUsingDemoDataset_ThenIncludesCategoryBreakdownByMonthTable(t *testing.T) {
	// Given
	parsedRows, err := pdfcardsummary.ParseMovementsWithCardContextCSV([]byte(demodataset.ExtractedCSV))
	require.NoError(t, err)

	var mappings Mappings
	require.NoError(t, json.Unmarshal([]byte(demodataset.MappingsV1JSON), &mappings))

	engine := New()

	// When
	result := engine.Compute(parsedRows, mappings)

	// Then
	tableByID(t, result.Tables, "category_breakdown_by_month")
}

func TestEngine_Compute_WhenUsingDemoDataset_ThenCategoryBreakdownByMonthHasExpectedRows(t *testing.T) {
	// Given
	parsedRows, err := pdfcardsummary.ParseMovementsWithCardContextCSV([]byte(demodataset.ExtractedCSV))
	require.NoError(t, err)

	var mappings Mappings
	require.NoError(t, json.Unmarshal([]byte(demodataset.MappingsV1JSON), &mappings))

	engine := New()

	// When
	result := engine.Compute(parsedRows, mappings)
	tbl := tableByID(t, result.Tables, "category_breakdown_by_month")

	// Then: mapped categories are used; "SUPERMARKET DEMO" and "SUPERMARKET  DEMO" both map to "Groceries",
	// collapsing 2 previously separate rows into 1, so 15 raw-detail rows → 13 mapped-category rows.
	require.Equal(t, 13, len(tbl.Rows), "one row per month+category aggregate")
	require.Equal(t, []string{"2025-03-01", "Groceries", "8777.00", "0.00", "43.29", "0.00"}, tbl.Rows[0])
	require.Contains(t, tbl.Rows, []string{"2025-02-01", "Groceries", "12444.00", "0.00", "63.14", "0.00"})
}

func TestEngine_OverviewMetricsByStatementMonth_WhenRowsHaveMixedMovementTypes_ThenReturnsExpectedRows(t *testing.T) {
	// Given
	rows := []pdfcardsummary.MovementWithCardContext{
		makeFactRow(time.Date(2025, time.January, 15, 0, 0, 0, 0, time.UTC), pdfcardsummary.MovementTypeCard, "100", "10", nil, nil),
		makeFactRow(time.Date(2025, time.January, 16, 0, 0, 0, 0, time.UTC), pdfcardsummary.MovementTypeCard, "200", "20", intPtr(1), intPtr(3)),
		makeFactRow(time.Date(2025, time.January, 17, 0, 0, 0, 0, time.UTC), pdfcardsummary.MovementTypeCard, "300", "30", intPtr(2), intPtr(3)),
		makeFactRow(time.Date(2025, time.January, 18, 0, 0, 0, 0, time.UTC), pdfcardsummary.MovementTypeCard, "50", "5", intPtr(4), intPtr(3)), // invalid installment (total < current)
		makeFactRow(time.Date(2025, time.January, 19, 0, 0, 0, 0, time.UTC), pdfcardsummary.MovementTypeTax, "40", "4", nil, nil),
		makeFactRow(time.Date(2025, time.January, 20, 0, 0, 0, 0, time.UTC), pdfcardsummary.MovementTypePastPayment, "-60", "-6", nil, nil),

		makeFactRow(time.Date(2025, time.February, 11, 0, 0, 0, 0, time.UTC), pdfcardsummary.MovementTypeTax, "70", "7", nil, nil),
		makeFactRow(time.Date(2025, time.February, 12, 0, 0, 0, 0, time.UTC), pdfcardsummary.MovementTypeCard, "500", "50", intPtr(3), intPtr(3)),
		makeFactRow(time.Date(2025, time.February, 13, 0, 0, 0, 0, time.UTC), pdfcardsummary.MovementTypeCard, "80", "8", intPtr(0), intPtr(3)), // invalid installment (non-positive current)

		makeFactRow(time.Date(2025, time.March, 10, 0, 0, 0, 0, time.UTC), pdfcardsummary.MovementTypePastPayment, "-20", "-2", nil, nil),
	}

	engine := New()
	expected := []OverviewMetricsByStatementMonthRow{
		{
			StatementMonthDate: time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
			NetStatementARS:    decimal2("630"),
			NetStatementUSD:    decimal2("63"),
			CardMovementsARS:   decimal2("650"),
			CardMovementsUSD:   decimal2("65"),
			NewDebtARS:         decimal2("350"),
			NewDebtUSD:         decimal2("35"),
			CarryOverDebtARS:   decimal2("300"),
			CarryOverDebtUSD:   decimal2("30"),
			NextMonthDebtARS:   decimal2("500"),
			NextMonthDebtUSD:   decimal2("50"),
			RemainingDebtARS:   decimal2("700"),
			RemainingDebtUSD:   decimal2("70"),
			TaxesARS:           decimal2("40"),
			TaxesUSD:           decimal2("4"),
			PastPaymentsARS:    decimal2("-60"),
			PastPaymentsUSD:    decimal2("-6"),
		},
		{
			StatementMonthDate: time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC),
			NetStatementARS:    decimal2("650"),
			NetStatementUSD:    decimal2("65"),
			CardMovementsARS:   decimal2("580"),
			CardMovementsUSD:   decimal2("58"),
			NewDebtARS:         decimal2("80"),
			NewDebtUSD:         decimal2("8"),
			CarryOverDebtARS:   decimal2("500"),
			CarryOverDebtUSD:   decimal2("50"),
			NextMonthDebtARS:   decimal2("0"),
			NextMonthDebtUSD:   decimal2("0"),
			RemainingDebtARS:   decimal2("0"),
			RemainingDebtUSD:   decimal2("0"),
			TaxesARS:           decimal2("70"),
			TaxesUSD:           decimal2("7"),
			PastPaymentsARS:    decimal2("0"),
			PastPaymentsUSD:    decimal2("0"),
		},
		{
			StatementMonthDate: time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC),
			NetStatementARS:    decimal2("-20"),
			NetStatementUSD:    decimal2("-2"),
			CardMovementsARS:   decimal2("0"),
			CardMovementsUSD:   decimal2("0"),
			NewDebtARS:         decimal2("0"),
			NewDebtUSD:         decimal2("0"),
			CarryOverDebtARS:   decimal2("0"),
			CarryOverDebtUSD:   decimal2("0"),
			NextMonthDebtARS:   decimal2("0"),
			NextMonthDebtUSD:   decimal2("0"),
			RemainingDebtARS:   decimal2("0"),
			RemainingDebtUSD:   decimal2("0"),
			TaxesARS:           decimal2("0"),
			TaxesUSD:           decimal2("0"),
			PastPaymentsARS:    decimal2("-20"),
			PastPaymentsUSD:    decimal2("-2"),
		},
	}

	// When
	actual := engine.OverviewMetricsByStatementMonth(rows)

	// Then
	require.Equal(t, expected, normalizeOverviewMetricsRows(actual))
}

func TestEngine_OverviewMetricsByStatementMonth_WhenMonthsHaveGap_ThenOnlyReturnsExistingMonths(t *testing.T) {
	// Given
	rows := []pdfcardsummary.MovementWithCardContext{
		makeFactRow(time.Date(2025, time.January, 3, 0, 0, 0, 0, time.UTC), pdfcardsummary.MovementTypeCard, "10", "1", nil, nil),
		makeFactRow(time.Date(2025, time.March, 5, 0, 0, 0, 0, time.UTC), pdfcardsummary.MovementTypeTax, "20", "2", nil, nil),
	}

	engine := New()

	// When
	actual := engine.OverviewMetricsByStatementMonth(rows)

	// Then
	require.Len(t, actual, 2)
	require.Equal(t, time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC), actual[0].StatementMonthDate)
	require.Equal(t, time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC), actual[1].StatementMonthDate)
}

func TestEngine_DebtMaturityScheduleByMonth_WhenLatestMonthHasInstallments_ThenExpandsFutureBuckets(t *testing.T) {
	// Given
	rows := []pdfcardsummary.MovementWithCardContext{
		makeFactRow(time.Date(2025, time.February, 3, 0, 0, 0, 0, time.UTC), pdfcardsummary.MovementTypeCard, "999", "99", intPtr(1), intPtr(12)), // not latest month
		makeFactRow(time.Date(2025, time.March, 10, 0, 0, 0, 0, time.UTC), pdfcardsummary.MovementTypeCard, "100", "1", intPtr(1), intPtr(3)),
		makeFactRow(time.Date(2025, time.March, 11, 0, 0, 0, 0, time.UTC), pdfcardsummary.MovementTypeCard, "50", "2", intPtr(2), intPtr(4)),
		makeFactRow(time.Date(2025, time.March, 12, 0, 0, 0, 0, time.UTC), pdfcardsummary.MovementTypeCard, "30", "0", intPtr(4), intPtr(4)), // no remaining installments
		makeFactRow(time.Date(2025, time.March, 13, 0, 0, 0, 0, time.UTC), pdfcardsummary.MovementTypeCard, "80", "8", intPtr(0), intPtr(3)), // invalid installment
		makeFactRow(time.Date(2025, time.March, 14, 0, 0, 0, 0, time.UTC), pdfcardsummary.MovementTypeTax, "700", "70", nil, nil),            // non-card movement
	}

	engine := New()
	expected := []DebtMaturityScheduleByMonthRow{
		{
			BaseStatementMonthDate: time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC),
			MaturityMonthDate:      time.Date(2025, time.April, 1, 0, 0, 0, 0, time.UTC),
			MonthOffset:            1,
			InstallmentCount:       2,
			MaturityTotalARS:       decimal2("150"),
			MaturityTotalUSD:       decimal2("3"),
		},
		{
			BaseStatementMonthDate: time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC),
			MaturityMonthDate:      time.Date(2025, time.May, 1, 0, 0, 0, 0, time.UTC),
			MonthOffset:            2,
			InstallmentCount:       2,
			MaturityTotalARS:       decimal2("150"),
			MaturityTotalUSD:       decimal2("3"),
		},
	}

	// When
	actual := engine.DebtMaturityScheduleByMonth(rows)

	// Then
	require.Equal(t, expected, normalizeDebtMaturityRows(actual))
}

func TestEngine_DebtMaturityScheduleByMonth_WhenLatestMonthHasNoFutureInstallments_ThenReturnsEmpty(t *testing.T) {
	// Given
	rows := []pdfcardsummary.MovementWithCardContext{
		makeFactRow(time.Date(2025, time.February, 3, 0, 0, 0, 0, time.UTC), pdfcardsummary.MovementTypeCard, "120", "12", intPtr(1), intPtr(3)),
		makeFactRow(time.Date(2025, time.March, 10, 0, 0, 0, 0, time.UTC), pdfcardsummary.MovementTypeCard, "200", "20", intPtr(1), intPtr(1)),
		makeFactRow(time.Date(2025, time.March, 11, 0, 0, 0, 0, time.UTC), pdfcardsummary.MovementTypeCard, "50", "5", nil, nil),
	}

	engine := New()

	// When
	actual := engine.DebtMaturityScheduleByMonth(rows)

	// Then
	require.Empty(t, actual)
}

func TestBuildCategoryBreakdownByMonthTable_WhenARSIsTied_ThenSortsByCategoryAscending(t *testing.T) {
	// Given
	rows := []SpendByCategoryRow{
		{
			Month:                time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC),
			Category:             "ZETA",
			CardMovementTotalARS: decimal2("100"),
			CardMovementTotalUSD: decimal2("0"),
		},
		{
			Month:                time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC),
			Category:             "ALFA",
			CardMovementTotalARS: decimal2("100"),
			CardMovementTotalUSD: decimal2("0"),
		},
	}

	// When
	tbl := buildCategoryBreakdownByMonthTable(rows)

	// Then
	require.Equal(t, "ALFA", tbl.Rows[0][1])
	require.Equal(t, "ZETA", tbl.Rows[1][1])
}

func TestBuildCategoryBreakdownByMonthTable_WhenMonthTotalARSIsZero_ThenShareIsZero(t *testing.T) {
	// Given
	rows := []SpendByCategoryRow{
		{
			Month:                time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC),
			Category:             "A",
			CardMovementTotalARS: decimal2("0"),
			CardMovementTotalUSD: decimal2("10"),
		},
		{
			Month:                time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC),
			Category:             "B",
			CardMovementTotalARS: decimal2("0"),
			CardMovementTotalUSD: decimal2("20"),
		},
	}

	// When
	tbl := buildCategoryBreakdownByMonthTable(rows)

	// Then
	require.Equal(t, "0.00", tbl.Rows[0][4], "share ARS is zero when month total ARS is zero")
	require.Equal(t, "0.00", tbl.Rows[1][4], "share ARS is zero when month total ARS is zero")
	require.Equal(t, "33.33", tbl.Rows[0][5], "share USD = 10/30*100")
	require.Equal(t, "66.67", tbl.Rows[1][5], "share USD = 20/30*100")
}

func TestEngine_MetaSummary_WhenUsingDemoDataset_ThenReturnsExpectedSummary(t *testing.T) {
	// Given
	parsedRows, err := pdfcardsummary.ParseMovementsWithCardContextCSV([]byte(demodataset.ExtractedCSV))
	require.NoError(t, err)

	engine := New()
	expectedSummary := MetaSummary{
		RowCount:            23,
		StatementMonthMin:   time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
		StatementMonthMax:   time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC),
		StatementMonthCount: 3,
	}

	// When
	actualSummary := engine.MetaSummary(parsedRows)

	// Then
	require.Equal(t, expectedSummary, actualSummary)
}

func TestEngine_DataQuality_WhenUsingDemoDatasetAndMappings_ThenReturnsExpectedIssuesAndSummary(t *testing.T) {
	// Given
	parsedRows, err := pdfcardsummary.ParseMovementsWithCardContextCSV([]byte(demodataset.ExtractedCSV))
	require.NoError(t, err)

	var mappings Mappings
	err = json.Unmarshal([]byte(demodataset.MappingsV1JSON), &mappings)
	require.NoError(t, err)

	engine := New()
	expectedIssues := []DQIssue{
		{
			RuleID:       DQRuleMissingCategoryMappingForCardMovement,
			Message:      "missing category mapping for CardMovement (detail=\"UNMAPPED MERCHANT\")",
			MovementType: pdfcardsummary.MovementTypeCard,
			CloseDate:    time.Date(2025, time.January, 25, 0, 0, 0, 0, time.UTC),
			Detail:       "UNMAPPED MERCHANT",
			CardOwner:    "OWNER B",
			CardNumber:   strPtr("1111"),
		},
		{
			RuleID:       DQRuleMissingOwnerMappingForCardMovement,
			Message:      "missing owner mapping for CardMovement (owner=\"OWNER B\", card=\"1111\")",
			MovementType: pdfcardsummary.MovementTypeCard,
			CloseDate:    time.Date(2025, time.January, 25, 0, 0, 0, 0, time.UTC),
			Detail:       "UNMAPPED MERCHANT",
			CardOwner:    "OWNER B",
			CardNumber:   strPtr("1111"),
		},
		{
			RuleID:       DQRuleMissingOwnerMappingForCardMovement,
			Message:      "missing owner mapping for CardMovement (owner=\"OWNER B\", card=\"1111\")",
			MovementType: pdfcardsummary.MovementTypeCard,
			CloseDate:    time.Date(2025, time.February, 25, 0, 0, 0, 0, time.UTC),
			Detail:       "SUPERMARKET DEMO",
			CardOwner:    "OWNER B",
			CardNumber:   strPtr("1111"),
		},
	}
	expectedSummary := []DQSummaryByRuleRow{
		{RuleID: DQRuleMissingCategoryMappingForCardMovement, Count: 1},
		{RuleID: DQRuleMissingOwnerMappingForCardMovement, Count: 2},
	}

	// When
	actualIssues, actualSummary := engine.DataQuality(parsedRows, mappings)

	// Then
	require.Equal(t, expectedIssues, actualIssues)
	require.Equal(t, expectedSummary, actualSummary)
}

func decimal2(s string) decimal.Decimal {
	d := decimal.RequireFromString(s)
	return decimal.RequireFromString(d.StringFixed(2))
}

func strPtr(s string) *string {
	return &s
}

func intPtr(v int) *int {
	return &v
}

func normalizeOverviewRows(rows []OverviewByStatementMonthRow) []OverviewByStatementMonthRow {
	out := make([]OverviewByStatementMonthRow, len(rows))
	for i, row := range rows {
		out[i] = OverviewByStatementMonthRow{
			StatementMonthDate:   row.StatementMonthDate,
			CardMovementTotalARS: decimal2(row.CardMovementTotalARS.String()),
			CardMovementTotalUSD: decimal2(row.CardMovementTotalUSD.String()),
		}
	}
	return out
}

func normalizeOverviewMetricsRows(rows []OverviewMetricsByStatementMonthRow) []OverviewMetricsByStatementMonthRow {
	out := make([]OverviewMetricsByStatementMonthRow, len(rows))
	for i, row := range rows {
		out[i] = OverviewMetricsByStatementMonthRow{
			StatementMonthDate: row.StatementMonthDate,
			NetStatementARS:    decimal2(row.NetStatementARS.String()),
			NetStatementUSD:    decimal2(row.NetStatementUSD.String()),
			CardMovementsARS:   decimal2(row.CardMovementsARS.String()),
			CardMovementsUSD:   decimal2(row.CardMovementsUSD.String()),
			NewDebtARS:         decimal2(row.NewDebtARS.String()),
			NewDebtUSD:         decimal2(row.NewDebtUSD.String()),
			CarryOverDebtARS:   decimal2(row.CarryOverDebtARS.String()),
			CarryOverDebtUSD:   decimal2(row.CarryOverDebtUSD.String()),
			NextMonthDebtARS:   decimal2(row.NextMonthDebtARS.String()),
			NextMonthDebtUSD:   decimal2(row.NextMonthDebtUSD.String()),
			RemainingDebtARS:   decimal2(row.RemainingDebtARS.String()),
			RemainingDebtUSD:   decimal2(row.RemainingDebtUSD.String()),
			TaxesARS:           decimal2(row.TaxesARS.String()),
			TaxesUSD:           decimal2(row.TaxesUSD.String()),
			PastPaymentsARS:    decimal2(row.PastPaymentsARS.String()),
			PastPaymentsUSD:    decimal2(row.PastPaymentsUSD.String()),
		}
	}
	return out
}

func normalizeDebtMaturityRows(rows []DebtMaturityScheduleByMonthRow) []DebtMaturityScheduleByMonthRow {
	out := make([]DebtMaturityScheduleByMonthRow, len(rows))
	for i, row := range rows {
		out[i] = DebtMaturityScheduleByMonthRow{
			BaseStatementMonthDate: row.BaseStatementMonthDate,
			MaturityMonthDate:      row.MaturityMonthDate,
			MonthOffset:            row.MonthOffset,
			InstallmentCount:       row.InstallmentCount,
			MaturityTotalARS:       decimal2(row.MaturityTotalARS.String()),
			MaturityTotalUSD:       decimal2(row.MaturityTotalUSD.String()),
		}
	}
	return out
}

func makeFactRow(
	closeDate time.Time,
	movementType pdfcardsummary.MovementType,
	amountARS string,
	amountUSD string,
	currentInstallment *int,
	totalInstallments *int,
) pdfcardsummary.MovementWithCardContext {
	return pdfcardsummary.MovementWithCardContext{
		StatementContext: pdfcardsummary.StatementContext{
			CloseDate: closeDate,
		},
		MovementType: movementType,
		Movement: pdfcardsummary.Movement{
			AmountARS:          decimal.RequireFromString(amountARS),
			AmountUSD:          decimal.RequireFromString(amountUSD),
			CurrentInstallment: currentInstallment,
			TotalInstallments:  totalInstallments,
		},
	}
}

func normalizeSpendByOwnerRows(rows []SpendByOwnerRow) []SpendByOwnerRow {
	out := make([]SpendByOwnerRow, len(rows))
	for i, row := range rows {
		out[i] = SpendByOwnerRow{
			Owner:                row.Owner,
			Month:                row.Month,
			CardMovementTotalARS: decimal2(row.CardMovementTotalARS.String()),
			CardMovementTotalUSD: decimal2(row.CardMovementTotalUSD.String()),
		}
	}
	return out
}
