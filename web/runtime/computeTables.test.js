import test from "node:test";
import assert from "node:assert";

import {
  indexTablesById,
  extractMetaSummary,
  latestMonthFromMeta,
  buildOverviewProjection,
  buildCardMovementTrend,
  buildTopOwnersFromCompute,
  buildTopCategoriesFromCompute,
  buildDebtMaturityScheduleFromCompute,
  buildRawExplorerRowsFromCompute,
  buildDqDiagnosticsFromCompute,
  buildRuntimeSnapshot,
} from "./computeTables.js";

function sampleComputeResult() {
  return {
    Tables: [
      {
        TableID: "meta_summary",
        Rows: [
          ["row_count", "120"],
          ["statement_month_min", "2024-12-01"],
          ["statement_month_max", "2025-02-01"],
        ],
      },
      {
        TableID: "overview_by_statement_month",
        Rows: [
          ["2025-02-01", "3000.00", "120.00"],
          ["2025-01-01", "2500.00", "80.00"],
        ],
      },
      {
        TableID: "overview_metrics_by_statement_month",
        Columns: [
          { Key: "statement_month_date", Label: "Statement Month Date", Type: "date" },
          { Key: "net_statement_ars", Label: "Net Statement ARS", Type: "money_ars" },
          { Key: "net_statement_usd", Label: "Net Statement USD", Type: "money_usd" },
          { Key: "card_movements_ars", Label: "Card Movements ARS", Type: "money_ars" },
          { Key: "card_movements_usd", Label: "Card Movements USD", Type: "money_usd" },
          { Key: "new_debt_ars", Label: "New Debt ARS", Type: "money_ars" },
          { Key: "new_debt_usd", Label: "New Debt USD", Type: "money_usd" },
          { Key: "carry_over_debt_ars", Label: "Carry Over Debt ARS", Type: "money_ars" },
          { Key: "carry_over_debt_usd", Label: "Carry Over Debt USD", Type: "money_usd" },
          { Key: "next_month_debt_ars", Label: "Next Month Debt ARS", Type: "money_ars" },
          { Key: "next_month_debt_usd", Label: "Next Month Debt USD", Type: "money_usd" },
          { Key: "remaining_debt_ars", Label: "Remaining Debt ARS", Type: "money_ars" },
          { Key: "remaining_debt_usd", Label: "Remaining Debt USD", Type: "money_usd" },
          { Key: "taxes_ars", Label: "Taxes ARS", Type: "money_ars" },
          { Key: "taxes_usd", Label: "Taxes USD", Type: "money_usd" },
          { Key: "past_payments_ars", Label: "Past Payments ARS", Type: "money_ars" },
          { Key: "past_payments_usd", Label: "Past Payments USD", Type: "money_usd" },
        ],
        Rows: [
          [
            "2025-01-01",
            "2600.00",
            "90.00",
            "2500.00",
            "80.00",
            "1800.00",
            "60.00",
            "700.00",
            "20.00",
            "500.00",
            "10.00",
            "1200.00",
            "30.00",
            "150.00",
            "15.00",
            "-50.00",
            "-5.00",
          ],
          [
            "2025-02-01",
            "3100.00",
            "132.00",
            "3000.00",
            "120.00",
            "2200.00",
            "90.00",
            "800.00",
            "30.00",
            "600.00",
            "12.00",
            "1100.00",
            "24.00",
            "180.00",
            "18.00",
            "-80.00",
            "-6.00",
          ],
        ],
      },
      {
        TableID: "spend_by_owner",
        Columns: [
          { Key: "month", Label: "Month", Type: "date" },
          { Key: "alice_ars", Label: "Alice (ARS)", Type: "money_ars" },
          { Key: "alice_usd", Label: "Alice (USD)", Type: "money_usd" },
          { Key: "bob_ars", Label: "Bob (ARS)", Type: "money_ars" },
          { Key: "bob_usd", Label: "Bob (USD)", Type: "money_usd" },
        ],
        Rows: [
          ["2025-01-01", "1200.00", "20.00", "900.00", "40.00"],
          ["2025-02-01", "400.00", "70.00", "1900.00", "30.00"],
        ],
      },
      {
        TableID: "spend_by_category",
        Columns: [
          { Key: "month", Label: "Month", Type: "date" },
          { Key: "travel_ars", Label: "Travel (ARS)", Type: "money_ars" },
          { Key: "travel_usd", Label: "Travel (USD)", Type: "money_usd" },
          { Key: "food_ars", Label: "Food (ARS)", Type: "money_ars" },
          { Key: "food_usd", Label: "Food (USD)", Type: "money_usd" },
        ],
        Rows: [
          ["2025-02-01", "600.00", "10.00", "1500.00", "50.00"],
        ],
      },
      {
        TableID: "debt_maturity_schedule_by_month_v1",
        Columns: [
          { Key: "base_statement_month_date", Label: "Base Statement Month Date", Type: "date" },
          { Key: "maturity_month_date", Label: "Maturity Month Date", Type: "date" },
          { Key: "month_offset", Label: "Month Offset", Type: "number" },
          { Key: "installment_count", Label: "Installment Count", Type: "number" },
          { Key: "maturity_total_ars", Label: "Maturity Total ARS", Type: "money_ars" },
          { Key: "maturity_total_usd", Label: "Maturity Total USD", Type: "money_usd" },
        ],
        Rows: [
          ["2025-02-01", "2025-03-01", "1", "2", "150.00", "3.00"],
          ["2025-02-01", "2025-04-01", "2", "1", "50.00", "1.00"],
        ],
      },
      {
        TableID: "raw_explorer_rows_v1",
        Columns: [
          { Key: "card_statement_close_date", Label: "Card Close Date", Type: "date" },
          { Key: "card_statement_due_date", Label: "Card Due Date", Type: "date" },
          { Key: "bank", Label: "Bank", Type: "string" },
          { Key: "card_company", Label: "Card Company", Type: "string" },
          { Key: "movement_date", Label: "Movement Date", Type: "date" },
          { Key: "card_number", Label: "Card Number", Type: "string" },
          { Key: "card_owner", Label: "Card Owner", Type: "string" },
          { Key: "movement_type", Label: "Movement Type", Type: "string" },
          { Key: "receipt_number", Label: "Receipt Number", Type: "string" },
          { Key: "detail", Label: "Detail", Type: "string" },
          { Key: "installment_current", Label: "Installment Current", Type: "number" },
          { Key: "installment_total", Label: "Installment Total", Type: "number" },
          { Key: "amount_ars", Label: "Amount ARS", Type: "money_ars" },
          { Key: "amount_usd", Label: "Amount USD", Type: "money_usd" },
        ],
        Rows: [
          [
            "2025-02-01",
            "2025-02-10",
            "Santander",
            "VISA",
            "2025-01-31",
            "1234",
            "Alice",
            "CardMovement",
            "A1",
            "Supermarket",
            "1",
            "3",
            "1000.00",
            "5.00",
          ],
          [
            "2025-02-01",
            "2025-02-10",
            "Santander",
            "VISA",
            "",
            "",
            "",
            "Tax",
            "",
            "Tax line",
            "",
            "",
            "100.00",
            "0.00",
          ],
        ],
      },
      {
        TableID: "dq_issues",
        Columns: [
          { Key: "rule_id", Label: "Rule ID", Type: "string" },
          { Key: "message", Label: "Message", Type: "string" },
          { Key: "movement_type", Label: "Movement Type", Type: "string" },
          { Key: "close_date", Label: "Close Date", Type: "date" },
          { Key: "detail", Label: "Detail", Type: "string" },
          { Key: "card_owner", Label: "Card Owner", Type: "string" },
          { Key: "card_number", Label: "Card Number", Type: "string" },
        ],
        Rows: [
          [
            "DQ003",
            "missing category mapping for CardMovement (detail=\"UNMAPPED\")",
            "CardMovement",
            "2025-02-01",
            "UNMAPPED",
            "Alice",
            "1111",
          ],
          [
            "DQ004",
            "missing owner mapping for CardMovement (owner=\"UNKNOWN\", card=\"2222\")",
            "CardMovement",
            "2025-02-01",
            "SUPERMARKET",
            "UNKNOWN",
            "2222",
          ],
        ],
      },
      {
        TableID: "dq_summary_by_rule",
        Columns: [
          { Key: "rule_id", Label: "Rule ID", Type: "string" },
          { Key: "count", Label: "Count", Type: "number" },
        ],
        Rows: [
          ["DQ003", "1"],
          ["DQ004", "1"],
        ],
      },
    ],
  };
}

test("indexTablesById indexes known tables", () => {
  const byId = indexTablesById(sampleComputeResult());
  assert.strictEqual(byId.size, 9);
  assert.ok(byId.has("meta_summary"));
  assert.ok(byId.has("spend_by_owner"));
  assert.ok(byId.has("dq_issues"));
});

test("extractMetaSummary + latestMonthFromMeta", () => {
  const meta = extractMetaSummary(sampleComputeResult());
  assert.strictEqual(meta.row_count, "120");
  assert.strictEqual(meta.statement_month_max, "2025-02-01");
  assert.strictEqual(latestMonthFromMeta(meta), "2025-02");
});

test("buildCardMovementTrend sorts months ascending", () => {
  const trend = buildCardMovementTrend(sampleComputeResult());
  assert.deepStrictEqual(trend.months, ["2025-01", "2025-02"]);
  assert.deepStrictEqual(trend.ARS, [2500, 3000]);
  assert.deepStrictEqual(trend.USD, [80, 120]);
});

test("buildOverviewProjection maps strict KPI/trend metrics", () => {
  const projection = buildOverviewProjection(sampleComputeResult());
  assert.strictEqual(projection.available, true);
  assert.deepStrictEqual(projection.months, ["2025-01", "2025-02"]);
  assert.strictEqual(projection.latestMonth, "2025-02");
  assert.strictEqual(projection.prevMonth, "2025-01");
  assert.strictEqual(projection.latest.currency.ARS.netStatement, 3100);
  assert.strictEqual(projection.prev.currency.USD.taxes, 15);
  assert.deepStrictEqual(projection.trend.ARS.newDebt, [1800, 2200]);
  assert.deepStrictEqual(projection.trend.USD.pastPayments, [-5, -6]);
});

test("strict projection keeps card-movement consistency with legacy overview table", () => {
  const projection = buildOverviewProjection(sampleComputeResult());
  const legacyTrend = buildCardMovementTrend(sampleComputeResult());
  assert.deepStrictEqual(projection.months, legacyTrend.months);
  assert.deepStrictEqual(projection.trend.ARS.cardMovements, legacyTrend.ARS);
  assert.deepStrictEqual(projection.trend.USD.cardMovements, legacyTrend.USD);
});

test("top rankings use latest month values and abs sort", () => {
  const topOwnersArs = buildTopOwnersFromCompute(sampleComputeResult(), "ARS", 2);
  assert.deepStrictEqual(topOwnersArs.labels, ["Bob", "Alice"]);
  assert.deepStrictEqual(topOwnersArs.values, [1900, 400]);

  const topCatsUsd = buildTopCategoriesFromCompute(sampleComputeResult(), "USD", 2);
  assert.deepStrictEqual(topCatsUsd.labels, ["Food", "Travel"]);
  assert.deepStrictEqual(topCatsUsd.values, [50, 10]);
});

test("buildDqDiagnosticsFromCompute maps summary and issue rows", () => {
  const diagnostics = buildDqDiagnosticsFromCompute(sampleComputeResult());
  assert.strictEqual(diagnostics.totalIssues, 2);
  assert.strictEqual(diagnostics.missingCategoryCount, 1);
  assert.strictEqual(diagnostics.missingOwnerCount, 1);
  assert.strictEqual(diagnostics.warningSummary.total, 2);
  assert.strictEqual(diagnostics.warningSummary.uncategorized, 1);
  assert.strictEqual(diagnostics.warningSummary.unmappedOwners, 1);
  assert.strictEqual(diagnostics.summary.length, 2);
  assert.strictEqual(diagnostics.issues[0].ruleId, "DQ003");
  assert.strictEqual(diagnostics.issues[1].cardOwner, "UNKNOWN");
});

test("buildDqDiagnosticsFromCompute keeps total from summary when issues table is partial", () => {
  const compute = sampleComputeResult();
  const dqIssues = compute.Tables.find((table) => table.TableID === "dq_issues");
  const dqSummary = compute.Tables.find((table) => table.TableID === "dq_summary_by_rule");
  dqIssues.Rows = dqIssues.Rows.slice(0, 1);
  dqSummary.Rows = [
    ["DQ003", "4"],
    ["DQ004", "2"],
  ];

  const diagnostics = buildDqDiagnosticsFromCompute(compute);
  assert.strictEqual(diagnostics.issues.length, 1);
  assert.strictEqual(diagnostics.totalIssues, 6);
  assert.strictEqual(diagnostics.warningSummary.total, 6);
  assert.strictEqual(diagnostics.warningSummary.uncategorized, 4);
  assert.strictEqual(diagnostics.warningSummary.unmappedOwners, 2);
});

test("buildDebtMaturityScheduleFromCompute maps strict maturity schedule rows", () => {
  const maturity = buildDebtMaturityScheduleFromCompute(sampleComputeResult());
  assert.strictEqual(maturity.available, true);
  assert.strictEqual(maturity.columns.length, 6);
  assert.strictEqual(maturity.rows.length, 2);
  assert.strictEqual(maturity.baseStatementMonth, "2025-02");
  assert.deepStrictEqual(maturity.months, ["2025-03", "2025-04"]);
  assert.deepStrictEqual(maturity.ARS, [150, 50]);
  assert.deepStrictEqual(maturity.USD, [3, 1]);
  assert.deepStrictEqual(maturity.installmentCount, [2, 1]);
  assert.strictEqual(maturity.totalARS, 200);
  assert.strictEqual(maturity.totalUSD, 4);
  assert.strictEqual(maturity.horizonMonths, 2);
});

test("buildRawExplorerRowsFromCompute maps row-level baseline fields", () => {
  const rawExplorer = buildRawExplorerRowsFromCompute(sampleComputeResult());
  assert.strictEqual(rawExplorer.available, true);
  assert.strictEqual(rawExplorer.columns.length, 14);
  assert.strictEqual(rawExplorer.columns[0].key, "card_statement_close_date");
  assert.strictEqual(rawExplorer.columns[0].label, "Card Close Date");
  assert.strictEqual(rawExplorer.columns[13].key, "amount_usd");
  assert.strictEqual(rawExplorer.columns.some((column) => column.key === "total_ars"), false);
  assert.strictEqual(rawExplorer.columns.some((column) => column.key === "card_total_ars"), false);
  assert.strictEqual(rawExplorer.rows.length, 2);
  assert.strictEqual(rawExplorer.rows[0].statementMonth, "2025-02");
  assert.strictEqual(rawExplorer.rows[0].cardOwner, "Alice");
  assert.strictEqual(rawExplorer.rows[0].installmentCurrent, 1);
  assert.strictEqual(rawExplorer.rows[1].movementType, "Tax");
  assert.strictEqual(rawExplorer.rows[1].installmentCurrent, null);
});

test("buildRuntimeSnapshot composes runtime-friendly summary", () => {
  const snapshot = buildRuntimeSnapshot(sampleComputeResult(), "files");
  assert.strictEqual(snapshot.source, "files");
  assert.strictEqual(snapshot.tableCount, 9);
  assert.strictEqual(snapshot.mode, "strict");
  assert.strictEqual(snapshot.latestMonth, "2025-02");
  assert.strictEqual(snapshot.overviewProjection.latest.currency.ARS.netStatement, 3100);
  assert.deepStrictEqual(snapshot.cardMovementTrend.months, ["2025-01", "2025-02"]);
  assert.strictEqual(snapshot.topOwners.ARS.labels[0], "Bob");
  assert.strictEqual(snapshot.debtMaturity.available, true);
  assert.deepStrictEqual(snapshot.debtMaturity.months, ["2025-03", "2025-04"]);
  assert.strictEqual(snapshot.rawExplorer.available, true);
  assert.strictEqual(snapshot.rawExplorer.rows.length, 2);
  assert.strictEqual(snapshot.dq.totalIssues, 2);
});
