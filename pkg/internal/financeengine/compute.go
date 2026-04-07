package financeengine

import (
	"slices"
	"time"

	"github.com/Alechan/finance-analyzer/pkg/internal/pdfcardsummary"
	"github.com/shopspring/decimal"
)

const (
	TableIDMetaSummary              = "meta_summary"
	TableIDOverviewByMonth          = "overview_by_statement_month"
	TableIDOverviewMetricsByMonth   = "overview_metrics_by_statement_month"
	TableIDDebtMaturitySchedule     = "debt_maturity_schedule_by_month_v1"
	TableIDSpendByOwner             = "spend_by_owner"
	TableIDSpendByCategory          = "spend_by_category"
	TableIDCategoryBreakdownByMonth = "category_breakdown_by_month"
	TableIDRawExplorerRows          = "raw_explorer_rows_v1"
	TableIDDQIssues                 = "dq_issues"
	TableIDDQSummaryByRule          = "dq_summary_by_rule"
)

func (e *Engine) Compute(rows []pdfcardsummary.MovementWithCardContext, mappings Mappings) ComputeResult {
	overviewRows := e.OverviewByStatementMonth(rows)
	overviewMetricsRows := e.OverviewMetricsByStatementMonth(rows)
	debtMaturityRows := e.DebtMaturityScheduleByMonth(rows)
	ownerRows := e.SpendByOwner(rows)
	categoryRows := e.SpendByCategory(rows, mappings)
	rawExplorerRows := e.RawExplorerRows(rows)
	metaSummary := e.MetaSummary(rows)
	dqIssues, dqSummary := e.DataQuality(rows, mappings)

	tables := []Table{
		buildMetaSummaryTable(metaSummary),
		buildOverviewByMonthTable(overviewRows),
		buildOverviewMetricsByMonthTable(overviewMetricsRows),
		buildDebtMaturityScheduleByMonthTable(debtMaturityRows),
		buildSpendByOwnerTable(ownerRows),
		buildSpendByCategoryTable(categoryRows),
		buildCategoryBreakdownByMonthTable(categoryRows),
		buildRawExplorerRowsTable(rawExplorerRows),
		buildDQIssuesTable(dqIssues),
		buildDQSummaryTable(dqSummary),
	}

	return ComputeResult{Tables: tables}
}

// If you change the rows in meta_summary, update the Description in the Table below.
func buildMetaSummaryTable(meta MetaSummary) Table {
	return Table{
		TableID: TableIDMetaSummary,
		Title:   "Meta Summary",
		Description: "<p>High-level stats about your dataset. Use as a quick sanity check before diving into spend tables.</p>" +
			"<ul>" +
			"<li><strong>row_count</strong><ul><li>Total movements in the dataset. Confirm the expected volume was loaded.</li></ul></li>" +
			"<li><strong>statement_month_min</strong><ul><li>Earliest statement month. Start of the date range.</li></ul></li>" +
			"<li><strong>statement_month_max</strong><ul><li>Latest statement month. End of the date range.</li></ul></li>" +
			"<li><strong>statement_month_count</strong><ul><li>Distinct statement months. How many billing cycles are covered.</li></ul></li>" +
			"</ul>",
		Columns: []TableColumn{
			{Key: "key", Label: "Key", Type: ColumnTypeString},
			{Key: "value", Label: "Value", Type: ColumnTypeString},
		},
		Rows: [][]string{
			{"row_count", intToString(meta.RowCount)},
			{"statement_month_min", formatDate(meta.StatementMonthMin)},
			{"statement_month_max", formatDate(meta.StatementMonthMax)},
			{"statement_month_count", intToString(meta.StatementMonthCount)},
		},
	}
}

// If you change the structure of overview_by_statement_month, update the Description below.
func buildOverviewByMonthTable(rows []OverviewByStatementMonthRow) Table {
	tableRows := make([][]string, 0, len(rows))
	for _, row := range rows {
		tableRows = append(tableRows, []string{
			formatDate(row.StatementMonthDate),
			formatDecimal(row.CardMovementTotalARS),
			formatDecimal(row.CardMovementTotalUSD),
		})
	}

	return Table{
		TableID:     TableIDOverviewByMonth,
		Title:       "Overview by Statement Month",
		Description: "<p>Card movement totals (ARS and USD) grouped by statement month. One row per billing cycle. Use it to see how spending changes over time.</p>",
		Columns: []TableColumn{
			{Key: "statement_month_date", Label: "Statement Month Date", Type: ColumnTypeDate},
			{Key: "card_movement_total_ars", Label: "Card Movement Total ARS", Type: ColumnTypeMoneyARS},
			{Key: "card_movement_total_usd", Label: "Card Movement Total USD", Type: ColumnTypeMoneyUSD},
		},
		Rows: tableRows,
	}
}

// If you change the structure of overview_metrics_by_statement_month, update the Description below.
func buildOverviewMetricsByMonthTable(rows []OverviewMetricsByStatementMonthRow) Table {
	tableRows := make([][]string, 0, len(rows))
	for _, row := range rows {
		tableRows = append(tableRows, []string{
			formatDate(row.StatementMonthDate),
			formatDecimal(row.NetStatementARS),
			formatDecimal(row.NetStatementUSD),
			formatDecimal(row.CardMovementsARS),
			formatDecimal(row.CardMovementsUSD),
			formatDecimal(row.NewDebtARS),
			formatDecimal(row.NewDebtUSD),
			formatDecimal(row.CarryOverDebtARS),
			formatDecimal(row.CarryOverDebtUSD),
			formatDecimal(row.NextMonthDebtARS),
			formatDecimal(row.NextMonthDebtUSD),
			formatDecimal(row.RemainingDebtARS),
			formatDecimal(row.RemainingDebtUSD),
			formatDecimal(row.TaxesARS),
			formatDecimal(row.TaxesUSD),
			formatDecimal(row.PastPaymentsARS),
			formatDecimal(row.PastPaymentsUSD),
		})
	}

	return Table{
		TableID: TableIDOverviewMetricsByMonth,
		Title:   "Overview Metrics by Statement Month",
		Description: "<p>Monthly overview metrics (ARS and USD): net statement, card movements, debt split, taxes, and past payments. " +
			"Designed for strict mockup KPI/trend projections with one row per statement month.</p>",
		Columns: []TableColumn{
			{Key: "statement_month_date", Label: "Statement Month Date", Type: ColumnTypeDate},
			{Key: "net_statement_ars", Label: "Net Statement ARS", Type: ColumnTypeMoneyARS},
			{Key: "net_statement_usd", Label: "Net Statement USD", Type: ColumnTypeMoneyUSD},
			{Key: "card_movements_ars", Label: "Card Movements ARS", Type: ColumnTypeMoneyARS},
			{Key: "card_movements_usd", Label: "Card Movements USD", Type: ColumnTypeMoneyUSD},
			{Key: "new_debt_ars", Label: "New Debt ARS", Type: ColumnTypeMoneyARS},
			{Key: "new_debt_usd", Label: "New Debt USD", Type: ColumnTypeMoneyUSD},
			{Key: "carry_over_debt_ars", Label: "Carry Over Debt ARS", Type: ColumnTypeMoneyARS},
			{Key: "carry_over_debt_usd", Label: "Carry Over Debt USD", Type: ColumnTypeMoneyUSD},
			{Key: "next_month_debt_ars", Label: "Next Month Debt ARS", Type: ColumnTypeMoneyARS},
			{Key: "next_month_debt_usd", Label: "Next Month Debt USD", Type: ColumnTypeMoneyUSD},
			{Key: "remaining_debt_ars", Label: "Remaining Debt ARS", Type: ColumnTypeMoneyARS},
			{Key: "remaining_debt_usd", Label: "Remaining Debt USD", Type: ColumnTypeMoneyUSD},
			{Key: "taxes_ars", Label: "Taxes ARS", Type: ColumnTypeMoneyARS},
			{Key: "taxes_usd", Label: "Taxes USD", Type: ColumnTypeMoneyUSD},
			{Key: "past_payments_ars", Label: "Past Payments ARS", Type: ColumnTypeMoneyARS},
			{Key: "past_payments_usd", Label: "Past Payments USD", Type: ColumnTypeMoneyUSD},
		},
		Rows: tableRows,
	}
}

// If you change the structure of debt_maturity_schedule_by_month_v1, update the Description below.
func buildDebtMaturityScheduleByMonthTable(rows []DebtMaturityScheduleByMonthRow) Table {
	tableRows := make([][]string, 0, len(rows))
	for _, row := range rows {
		tableRows = append(tableRows, []string{
			formatDate(row.BaseStatementMonthDate),
			formatDate(row.MaturityMonthDate),
			intToString(row.MonthOffset),
			intToString(row.InstallmentCount),
			formatDecimal(row.MaturityTotalARS),
			formatDecimal(row.MaturityTotalUSD),
		})
	}

	return Table{
		TableID: TableIDDebtMaturitySchedule,
		Title:   "Debt Maturity Schedule by Month v1",
		Description: "<p>Future installment obligations projected from the latest statement month only. " +
			"Each active installment row contributes one payment amount to every future month up to its remaining installments. " +
			"Rows are grouped by maturity month and include installment counts with ARS/USD totals.</p>",
		Columns: []TableColumn{
			{Key: "base_statement_month_date", Label: "Base Statement Month Date", Type: ColumnTypeDate},
			{Key: "maturity_month_date", Label: "Maturity Month Date", Type: ColumnTypeDate},
			{Key: "month_offset", Label: "Month Offset", Type: ColumnTypeNumber},
			{Key: "installment_count", Label: "Installment Count", Type: ColumnTypeNumber},
			{Key: "maturity_total_ars", Label: "Maturity Total ARS", Type: ColumnTypeMoneyARS},
			{Key: "maturity_total_usd", Label: "Maturity Total USD", Type: ColumnTypeMoneyUSD},
		},
		Rows: tableRows,
	}
}

// If you change the structure of spend_by_owner, update the Description below.
func buildSpendByOwnerTable(rows []SpendByOwnerRow) Table {
	type amountPair struct {
		ars decimal.Decimal
		usd decimal.Decimal
	}
	monthToOwner := make(map[time.Time]map[string]amountPair)
	monthSet := make(map[time.Time]struct{})
	ownerSet := make(map[string]struct{})

	for _, row := range rows {
		monthSet[row.Month] = struct{}{}
		ownerSet[row.Owner] = struct{}{}
		if monthToOwner[row.Month] == nil {
			monthToOwner[row.Month] = make(map[string]amountPair)
		}
		prev := monthToOwner[row.Month][row.Owner]
		monthToOwner[row.Month][row.Owner] = amountPair{
			ars: prev.ars.Add(row.CardMovementTotalARS),
			usd: prev.usd.Add(row.CardMovementTotalUSD),
		}
	}

	months := make([]time.Time, 0, len(monthSet))
	for m := range monthSet {
		months = append(months, m)
	}
	slices.SortFunc(months, func(a, b time.Time) int { return b.Compare(a) })

	owners := make([]string, 0, len(ownerSet))
	for o := range ownerSet {
		owners = append(owners, o)
	}
	slices.Sort(owners)

	columns := []TableColumn{{Key: "month", Label: "Month", Type: ColumnTypeDate}}
	for _, owner := range owners {
		columns = append(columns,
			TableColumn{Key: owner + "_ars", Label: owner + " (ARS)", Type: ColumnTypeMoneyARS},
			TableColumn{Key: owner + "_usd", Label: owner + " (USD)", Type: ColumnTypeMoneyUSD},
		)
	}

	tableRows := make([][]string, 0, len(months))
	for _, month := range months {
		row := []string{formatDate(month)}
		for _, owner := range owners {
			amt := monthToOwner[month][owner]
			row = append(row, formatDecimal(amt.ars), formatDecimal(amt.usd))
		}
		tableRows = append(tableRows, row)
	}

	return Table{
		TableID:     TableIDSpendByOwner,
		Title:       "Spend by Owner",
		Description: "<p>Totals (ARS and USD) by month, with owners as columns. Use it to see who spends what in each billing cycle.</p>",
		Columns:     columns,
		Rows:        tableRows,
	}
}

// If you change the structure of spend_by_category, update the Description below.
func buildSpendByCategoryTable(rows []SpendByCategoryRow) Table {
	type amountPair struct {
		ars decimal.Decimal
		usd decimal.Decimal
	}
	monthToCategory := make(map[time.Time]map[string]amountPair)
	monthSet := make(map[time.Time]struct{})
	categorySet := make(map[string]struct{})

	for _, row := range rows {
		monthSet[row.Month] = struct{}{}
		categorySet[row.Category] = struct{}{}
		if monthToCategory[row.Month] == nil {
			monthToCategory[row.Month] = make(map[string]amountPair)
		}
		prev := monthToCategory[row.Month][row.Category]
		monthToCategory[row.Month][row.Category] = amountPair{
			ars: prev.ars.Add(row.CardMovementTotalARS),
			usd: prev.usd.Add(row.CardMovementTotalUSD),
		}
	}

	months := make([]time.Time, 0, len(monthSet))
	for m := range monthSet {
		months = append(months, m)
	}
	slices.SortFunc(months, func(a, b time.Time) int { return b.Compare(a) })

	categories := make([]string, 0, len(categorySet))
	for c := range categorySet {
		categories = append(categories, c)
	}
	slices.Sort(categories)

	columns := []TableColumn{{Key: "month", Label: "Month", Type: ColumnTypeDate}}
	for _, cat := range categories {
		columns = append(columns,
			TableColumn{Key: cat + "_ars", Label: cat + " (ARS)", Type: ColumnTypeMoneyARS},
			TableColumn{Key: cat + "_usd", Label: cat + " (USD)", Type: ColumnTypeMoneyUSD},
		)
	}

	tableRows := make([][]string, 0, len(months))
	for _, month := range months {
		row := []string{formatDate(month)}
		for _, cat := range categories {
			amt := monthToCategory[month][cat]
			row = append(row, formatDecimal(amt.ars), formatDecimal(amt.usd))
		}
		tableRows = append(tableRows, row)
	}

	return Table{
		TableID:     TableIDSpendByCategory,
		Title:       "Spend by Category",
		Description: "<p>Card movement totals (ARS and USD) by month, with categories as columns. Use it to see where your money goes in each billing cycle.</p>",
		Columns:     columns,
		Rows:        tableRows,
	}
}

// If you change the structure of category_breakdown_by_month, update the Description below.
func buildCategoryBreakdownByMonthTable(rows []SpendByCategoryRow) Table {
	sortedRows := make([]SpendByCategoryRow, len(rows))
	copy(sortedRows, rows)

	monthTotalARS := make(map[time.Time]decimal.Decimal)
	monthTotalUSD := make(map[time.Time]decimal.Decimal)
	for _, row := range sortedRows {
		monthTotalARS[row.Month] = monthTotalARS[row.Month].Add(row.CardMovementTotalARS)
		monthTotalUSD[row.Month] = monthTotalUSD[row.Month].Add(row.CardMovementTotalUSD)
	}

	slices.SortFunc(sortedRows, func(a, b SpendByCategoryRow) int {
		if c := b.Month.Compare(a.Month); c != 0 {
			return c
		}
		if c := b.CardMovementTotalARS.Cmp(a.CardMovementTotalARS); c != 0 {
			return c
		}
		if a.Category < b.Category {
			return -1
		}
		if a.Category > b.Category {
			return 1
		}
		return 0
	})

	tableRows := make([][]string, 0, len(sortedRows))
	for _, row := range sortedRows {
		shareOfMonthARS := decimal.Zero
		totalARS := monthTotalARS[row.Month]
		if !totalARS.IsZero() {
			shareOfMonthARS = row.CardMovementTotalARS.Div(totalARS).Mul(decimal.NewFromInt(100))
		}

		shareOfMonthUSD := decimal.Zero
		totalUSD := monthTotalUSD[row.Month]
		if !totalUSD.IsZero() {
			shareOfMonthUSD = row.CardMovementTotalUSD.Div(totalUSD).Mul(decimal.NewFromInt(100))
		}

		tableRows = append(tableRows, []string{
			formatDate(row.Month),
			row.Category,
			formatDecimal(row.CardMovementTotalARS),
			formatDecimal(row.CardMovementTotalUSD),
			formatDecimal(shareOfMonthARS),
			formatDecimal(shareOfMonthUSD),
		})
	}

	return Table{
		TableID:     TableIDCategoryBreakdownByMonth,
		Title:       "Category Breakdown by Month",
		Description: "<p>Month-focused category breakdown. Select a month to see categories ranked by spend and their share of the month total.</p>",
		Columns: []TableColumn{
			{Key: "month", Label: "Month", Type: ColumnTypeDate},
			{Key: "category", Label: "Category", Type: ColumnTypeString},
			{Key: "card_movement_total_ars", Label: "Card Movement Total ARS", Type: ColumnTypeMoneyARS},
			{Key: "card_movement_total_usd", Label: "Card Movement Total USD", Type: ColumnTypeMoneyUSD},
			{Key: "share_of_month_ars_pct", Label: "Share of Month ARS", Type: ColumnTypeShare},
			{Key: "share_of_month_usd_pct", Label: "Share of Month USD", Type: ColumnTypeShare},
		},
		Rows: tableRows,
	}
}

// If you change the structure of raw_explorer_rows_v1, update the Description below.
func buildRawExplorerRowsTable(rows []RawExplorerRow) Table {
	tableRows := make([][]string, 0, len(rows))
	for _, row := range rows {
		tableRows = append(tableRows, []string{
			formatDate(row.CardStatementCloseDate),
			formatDate(row.CardStatementDueDate),
			row.Bank,
			row.CardCompany,
			formatDatePtr(row.MovementDate),
			row.CardNumber,
			row.CardOwner,
			row.MovementType,
			row.ReceiptNumber,
			row.Detail,
			formatIntPtr(row.InstallmentCurrent),
			formatIntPtr(row.InstallmentTotal),
			formatDecimal(row.AmountARS),
			formatDecimal(row.AmountUSD),
		})
	}

	return Table{
		TableID: TableIDRawExplorerRows,
		Title:   "Raw Explorer Rows v1",
		Description: "<p>Row-level movement table for the integrated Raw Explorer baseline view. " +
			"Columns are ordered for readability (statement dates, card context, movement context, installments, amounts) " +
			"and intentionally exclude summary/card total columns to reduce noise.</p>",
		Columns: []TableColumn{
			{Key: "card_statement_close_date", Label: "Card Close Date", Type: ColumnTypeDate},
			{Key: "card_statement_due_date", Label: "Card Due Date", Type: ColumnTypeDate},
			{Key: "bank", Label: "Bank", Type: ColumnTypeString},
			{Key: "card_company", Label: "Card Company", Type: ColumnTypeString},
			{Key: "movement_date", Label: "Movement Date", Type: ColumnTypeDate},
			{Key: "card_number", Label: "Card Number", Type: ColumnTypeString},
			{Key: "card_owner", Label: "Card Owner", Type: ColumnTypeString},
			{Key: "movement_type", Label: "Movement Type", Type: ColumnTypeString},
			{Key: "receipt_number", Label: "Receipt Number", Type: ColumnTypeString},
			{Key: "detail", Label: "Detail", Type: ColumnTypeString},
			{Key: "installment_current", Label: "Installment Current", Type: ColumnTypeNumber},
			{Key: "installment_total", Label: "Installment Total", Type: ColumnTypeNumber},
			{Key: "amount_ars", Label: "Amount ARS", Type: ColumnTypeMoneyARS},
			{Key: "amount_usd", Label: "Amount USD", Type: ColumnTypeMoneyUSD},
		},
		Rows: tableRows,
	}
}

// If you change the structure of dq_issues, update the Description below.
func buildDQIssuesTable(rows []DQIssue) Table {
	tableRows := make([][]string, 0, len(rows))
	for _, row := range rows {
		cardNumber := ""
		if row.CardNumber != nil {
			cardNumber = *row.CardNumber
		}
		tableRows = append(tableRows, []string{
			row.RuleID,
			row.Message,
			string(row.MovementType),
			formatDate(row.CloseDate),
			row.Detail,
			row.CardOwner,
			cardNumber,
		})
	}

	return Table{
		TableID:     TableIDDQIssues,
		Title:       "DQ Issues",
		Description: "<p>Data quality issues: unmapped owners, unmapped categories, and other rule violations. Each row is one issue with context. Fix these to improve your analysis.</p>",
		Columns: []TableColumn{
			{Key: "rule_id", Label: "Rule ID", Type: ColumnTypeString},
			{Key: "message", Label: "Message", Type: ColumnTypeString},
			{Key: "movement_type", Label: "Movement Type", Type: ColumnTypeString},
			{Key: "close_date", Label: "Close Date", Type: ColumnTypeDate},
			{Key: "detail", Label: "Detail", Type: ColumnTypeString},
			{Key: "card_owner", Label: "Card Owner", Type: ColumnTypeString},
			{Key: "card_number", Label: "Card Number", Type: ColumnTypeString},
		},
		Rows: tableRows,
	}
}

// If you change the structure of dq_summary_by_rule, update the Description below.
func buildDQSummaryTable(rows []DQSummaryByRuleRow) Table {
	tableRows := make([][]string, 0, len(rows))
	for _, row := range rows {
		tableRows = append(tableRows, []string{
			row.RuleID,
			intToString(row.Count),
		})
	}

	return Table{
		TableID:     TableIDDQSummaryByRule,
		Title:       "DQ Summary by Rule",
		Description: "<p>Count of data quality issues per rule. Use it to prioritize which rules to fix first.</p>",
		Columns: []TableColumn{
			{Key: "rule_id", Label: "Rule ID", Type: ColumnTypeString},
			{Key: "count", Label: "Count", Type: ColumnTypeNumber},
		},
		Rows: tableRows,
	}
}

func formatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

func formatDatePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return formatDate(*t)
}

func formatIntPtr(v *int) string {
	if v == nil {
		return ""
	}
	return intToString(*v)
}
