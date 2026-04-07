const TABLE_META_SUMMARY = "meta_summary";
const TABLE_OVERVIEW_BY_MONTH = "overview_by_statement_month";
const TABLE_OVERVIEW_METRICS_BY_MONTH = "overview_metrics_by_statement_month";
const TABLE_DEBT_MATURITY_SCHEDULE = "debt_maturity_schedule_by_month_v1";
const TABLE_SPEND_BY_OWNER = "spend_by_owner";
const TABLE_SPEND_BY_CATEGORY = "spend_by_category";
const TABLE_RAW_EXPLORER_ROWS = "raw_explorer_rows_v1";
const TABLE_DQ_ISSUES = "dq_issues";
const TABLE_DQ_SUMMARY_BY_RULE = "dq_summary_by_rule";
const OVERVIEW_METRIC_KEYS = [
  "netStatement",
  "cardMovements",
  "newDebt",
  "carryOverDebt",
  "nextMonthDebt",
  "remainingDebt",
  "taxes",
  "pastPayments",
];
const OVERVIEW_COLUMN_BY_METRIC = Object.freeze({
  netStatement: { ARS: "net_statement_ars", USD: "net_statement_usd" },
  cardMovements: { ARS: "card_movements_ars", USD: "card_movements_usd" },
  newDebt: { ARS: "new_debt_ars", USD: "new_debt_usd" },
  carryOverDebt: { ARS: "carry_over_debt_ars", USD: "carry_over_debt_usd" },
  nextMonthDebt: { ARS: "next_month_debt_ars", USD: "next_month_debt_usd" },
  remainingDebt: { ARS: "remaining_debt_ars", USD: "remaining_debt_usd" },
  taxes: { ARS: "taxes_ars", USD: "taxes_usd" },
  pastPayments: { ARS: "past_payments_ars", USD: "past_payments_usd" },
});

function asNumber(value) {
  const n = Number(value);
  return Number.isFinite(n) ? n : 0;
}

function monthKeyFromDate(dateString) {
  const value = String(dateString || "").trim();
  if (!value) {
    return "";
  }
  const match = /^(\d{4}-\d{2})/.exec(value);
  return match ? match[1] : value;
}

function emptyOverviewCurrencyMetrics() {
  return {
    netStatement: 0,
    cardMovements: 0,
    newDebt: 0,
    carryOverDebt: 0,
    nextMonthDebt: 0,
    remainingDebt: 0,
    taxes: 0,
    pastPayments: 0,
  };
}

function cloneOverviewCurrencyMetrics(source) {
  const out = emptyOverviewCurrencyMetrics();
  const input = source || {};
  for (const metric of OVERVIEW_METRIC_KEYS) {
    out[metric] = asNumber(input[metric]);
  }
  return out;
}

function emptyOverviewBucket() {
  return {
    currency: {
      ARS: emptyOverviewCurrencyMetrics(),
      USD: emptyOverviewCurrencyMetrics(),
    },
  };
}

function emptyOverviewTrend() {
  return {
    months: [],
    ARS: emptyOverviewCurrencyMetrics(),
    USD: emptyOverviewCurrencyMetrics(),
  };
}

function latestRowByMonth(rows) {
  if (!Array.isArray(rows) || rows.length === 0) {
    return null;
  }

  let latest = rows[0];
  let latestKey = monthKeyFromDate(rows[0]?.[0]);
  for (let i = 1; i < rows.length; i++) {
    const current = rows[i];
    const currentKey = monthKeyFromDate(current?.[0]);
    if (currentKey > latestKey) {
      latest = current;
      latestKey = currentKey;
    }
  }
  return latest;
}

function columnIndexByKey(table, key) {
  const columns = Array.isArray(table?.Columns) ? table.Columns : [];
  for (let i = 0; i < columns.length; i++) {
    if (String(columns[i]?.Key || "") === key) {
      return i;
    }
  }
  return -1;
}

function numberFromColumnKey(table, row, key) {
  const idx = columnIndexByKey(table, key);
  if (idx < 0) {
    return 0;
  }
  return asNumber(row?.[idx]);
}

function textFromColumnKey(table, row, key, fallbackIdx = -1) {
  const idx = columnIndexByKey(table, key);
  const resolvedIdx = idx >= 0 ? idx : fallbackIdx;
  if (resolvedIdx < 0) {
    return "";
  }
  return String(row?.[resolvedIdx] || "").trim();
}

function overviewMetricRows(computeResult) {
  const byId = indexTablesById(computeResult);
  const table = byId.get(TABLE_OVERVIEW_METRICS_BY_MONTH);
  if (!table) {
    return [];
  }

  const monthIdx = columnIndexByKey(table, "statement_month_date");
  const resolvedMonthIdx = monthIdx >= 0 ? monthIdx : 0;
  const rows = [];
  for (const rawRow of table.Rows || []) {
    const month = monthKeyFromDate(rawRow?.[resolvedMonthIdx]);
    if (!month) {
      continue;
    }

    const currency = {
      ARS: emptyOverviewCurrencyMetrics(),
      USD: emptyOverviewCurrencyMetrics(),
    };
    for (const metric of OVERVIEW_METRIC_KEYS) {
      const mapping = OVERVIEW_COLUMN_BY_METRIC[metric];
      currency.ARS[metric] = numberFromColumnKey(table, rawRow, mapping.ARS);
      currency.USD[metric] = numberFromColumnKey(table, rawRow, mapping.USD);
    }
    rows.push({ month, currency });
  }

  rows.sort((a, b) => a.month.localeCompare(b.month));
  return rows;
}

function trendFromOverviewMetricRows(rows) {
  const trend = emptyOverviewTrend();
  trend.months = rows.map((row) => row.month);
  for (const metric of OVERVIEW_METRIC_KEYS) {
    trend.ARS[metric] = rows.map((row) => row.currency.ARS[metric]);
    trend.USD[metric] = rows.map((row) => row.currency.USD[metric]);
  }
  return trend;
}

export function buildOverviewProjection(computeResult) {
  const rows = overviewMetricRows(computeResult);
  if (rows.length === 0) {
    return {
      available: false,
      months: [],
      latestMonth: "",
      prevMonth: "",
      latest: emptyOverviewBucket(),
      prev: emptyOverviewBucket(),
      trend: emptyOverviewTrend(),
    };
  }

  const latestRow = rows[rows.length - 1];
  const prevRow = rows.length > 1 ? rows[rows.length - 2] : null;
  const latest = {
    currency: {
      ARS: cloneOverviewCurrencyMetrics(latestRow.currency.ARS),
      USD: cloneOverviewCurrencyMetrics(latestRow.currency.USD),
    },
  };
  const prev = prevRow
    ? {
        currency: {
          ARS: cloneOverviewCurrencyMetrics(prevRow.currency.ARS),
          USD: cloneOverviewCurrencyMetrics(prevRow.currency.USD),
        },
      }
    : emptyOverviewBucket();

  return {
    available: true,
    months: rows.map((row) => row.month),
    latestMonth: latestRow.month,
    prevMonth: prevRow ? prevRow.month : null,
    latest,
    prev,
    trend: trendFromOverviewMetricRows(rows),
  };
}

export function indexTablesById(computeResult) {
  const out = new Map();
  const tables = Array.isArray(computeResult?.Tables) ? computeResult.Tables : [];
  for (const table of tables) {
    if (table && table.TableID) {
      out.set(table.TableID, table);
    }
  }
  return out;
}

export function extractMetaSummary(computeResult) {
  const byId = indexTablesById(computeResult);
  const table = byId.get(TABLE_META_SUMMARY);
  if (!table) {
    return {};
  }

  const meta = {};
  for (const row of table.Rows || []) {
    const key = row?.[0];
    const value = row?.[1];
    if (key) {
      meta[String(key)] = String(value || "");
    }
  }
  return meta;
}

export function latestMonthFromMeta(metaSummary) {
  const raw = metaSummary?.statement_month_max || "";
  return monthKeyFromDate(raw);
}

export function buildCardMovementTrend(computeResult) {
  const byId = indexTablesById(computeResult);
  const table = byId.get(TABLE_OVERVIEW_BY_MONTH);
  if (!table) {
    const strictRows = overviewMetricRows(computeResult);
    return {
      months: strictRows.map((row) => row.month),
      ARS: strictRows.map((row) => row.currency.ARS.cardMovements),
      USD: strictRows.map((row) => row.currency.USD.cardMovements),
    };
  }

  const rows = (table.Rows || [])
    .map((row) => ({
      month: monthKeyFromDate(row?.[0]),
      ars: asNumber(row?.[1]),
      usd: asNumber(row?.[2]),
    }))
    .filter((row) => row.month)
    .sort((a, b) => a.month.localeCompare(b.month));

  return {
    months: rows.map((row) => row.month),
    ARS: rows.map((row) => row.ars),
    USD: rows.map((row) => row.usd),
  };
}

function buildLatestBreakdown(computeResult, tableId, currency, limit) {
  const byId = indexTablesById(computeResult);
  const table = byId.get(tableId);
  if (!table) {
    return { labels: [], values: [] };
  }

  const latestRow = latestRowByMonth(table.Rows || []);
  if (!latestRow) {
    return { labels: [], values: [] };
  }

  const targetSuffix = currency === "USD" ? "(USD)" : "(ARS)";
  const values = [];
  for (let i = 1; i < (table.Columns || []).length; i++) {
    const column = table.Columns[i];
    const label = String(column?.Label || "");
    if (!label.endsWith(targetSuffix)) {
      continue;
    }
    const amount = asNumber(latestRow[i]);
    values.push({
      label: label.slice(0, label.length - targetSuffix.length).trim(),
      value: amount,
    });
  }

  values.sort((a, b) => Math.abs(b.value) - Math.abs(a.value));
  const selected = values.slice(0, Math.max(1, limit || 6));

  return {
    labels: selected.map((item) => item.label),
    values: selected.map((item) => item.value),
  };
}

export function buildTopOwnersFromCompute(computeResult, currency = "ARS", limit = 6) {
  return buildLatestBreakdown(computeResult, TABLE_SPEND_BY_OWNER, currency, limit);
}

export function buildTopCategoriesFromCompute(computeResult, currency = "ARS", limit = 6) {
  return buildLatestBreakdown(computeResult, TABLE_SPEND_BY_CATEGORY, currency, limit);
}

export function buildDebtMaturityScheduleFromCompute(computeResult) {
  const byId = indexTablesById(computeResult);
  const table = byId.get(TABLE_DEBT_MATURITY_SCHEDULE);
  if (!table) {
    return {
      available: false,
      columns: [],
      rows: [],
      baseStatementMonth: "",
      months: [],
      ARS: [],
      USD: [],
      installmentCount: [],
      totalARS: 0,
      totalUSD: 0,
      horizonMonths: 0,
    };
  }

  const columns = (table.Columns || []).map((column) => ({
    key: String(column?.Key || ""),
    label: String(column?.Label || ""),
    type: String(column?.Type || ""),
  }));
  const rows = (table.Rows || [])
    .map((row) => ({
      baseStatementMonthDate: textFromColumnKey(table, row, "base_statement_month_date", 0),
      maturityMonthDate: textFromColumnKey(table, row, "maturity_month_date", 1),
      monthOffset: asOptionalInteger(textFromColumnKey(table, row, "month_offset", 2)),
      installmentCount: asOptionalInteger(textFromColumnKey(table, row, "installment_count", 3)) || 0,
      amountARS: asNumber(textFromColumnKey(table, row, "maturity_total_ars", 4)),
      amountUSD: asNumber(textFromColumnKey(table, row, "maturity_total_usd", 5)),
    }))
    .filter((row) => row.maturityMonthDate)
    .sort((a, b) => {
      const maturityCmp = String(a.maturityMonthDate).localeCompare(String(b.maturityMonthDate));
      if (maturityCmp !== 0) {
        return maturityCmp;
      }
      return (a.monthOffset || 0) - (b.monthOffset || 0);
    });

  const months = rows.map((row) => monthKeyFromDate(row.maturityMonthDate));
  const ars = rows.map((row) => row.amountARS);
  const usd = rows.map((row) => row.amountUSD);
  const installmentCount = rows.map((row) => row.installmentCount);
  const totalARS = ars.reduce((acc, value) => acc + value, 0);
  const totalUSD = usd.reduce((acc, value) => acc + value, 0);

  return {
    available: true,
    columns,
    rows,
    baseStatementMonth: rows.length > 0 ? monthKeyFromDate(rows[0].baseStatementMonthDate) : "",
    months,
    ARS: ars,
    USD: usd,
    installmentCount,
    totalARS,
    totalUSD,
    horizonMonths: rows.length,
  };
}

function asOptionalInteger(value) {
  const text = String(value == null ? "" : value).trim();
  if (!text) {
    return null;
  }
  const n = Number(text);
  if (!Number.isFinite(n)) {
    return null;
  }
  return Math.trunc(n);
}

export function buildRawExplorerRowsFromCompute(computeResult) {
  const byId = indexTablesById(computeResult);
  const table = byId.get(TABLE_RAW_EXPLORER_ROWS);
  if (!table) {
    return {
      available: false,
      columns: [],
      rows: [],
    };
  }

  const columns = (table.Columns || []).map((column) => ({
    key: String(column?.Key || ""),
    label: String(column?.Label || ""),
    type: String(column?.Type || ""),
  }));
  const rows = (table.Rows || []).map((row) => {
    const cardStatementCloseDate = textFromColumnKey(table, row, "card_statement_close_date", 0);
    const cardStatementDueDate = textFromColumnKey(table, row, "card_statement_due_date", 1);
    return {
      cardStatementCloseDate,
      cardStatementDueDate,
      statementMonth: monthKeyFromDate(cardStatementCloseDate),
      bank: textFromColumnKey(table, row, "bank", 2),
      cardCompany: textFromColumnKey(table, row, "card_company", 3),
      movementDate: textFromColumnKey(table, row, "movement_date", 4),
      cardNumber: textFromColumnKey(table, row, "card_number", 5),
      cardOwner: textFromColumnKey(table, row, "card_owner", 6),
      movementType: textFromColumnKey(table, row, "movement_type", 7),
      receiptNumber: textFromColumnKey(table, row, "receipt_number", 8),
      detail: textFromColumnKey(table, row, "detail", 9),
      installmentCurrent: asOptionalInteger(textFromColumnKey(table, row, "installment_current", 10)),
      installmentTotal: asOptionalInteger(textFromColumnKey(table, row, "installment_total", 11)),
      amountARS: asNumber(textFromColumnKey(table, row, "amount_ars", 12)),
      amountUSD: asNumber(textFromColumnKey(table, row, "amount_usd", 13)),
    };
  });

  return {
    available: true,
    columns,
    rows,
  };
}

export function buildDqDiagnosticsFromCompute(computeResult) {
  const byId = indexTablesById(computeResult);
  const issuesTable = byId.get(TABLE_DQ_ISSUES);
  const summaryTable = byId.get(TABLE_DQ_SUMMARY_BY_RULE);

  const byRule = {};
  const summary = [];
  for (const row of summaryTable?.Rows || []) {
    const ruleId = textFromColumnKey(summaryTable, row, "rule_id", 0);
    if (!ruleId) {
      continue;
    }
    const countRaw = numberFromColumnKey(summaryTable, row, "count");
    const count = Math.max(0, Math.trunc(countRaw));
    summary.push({ ruleId, count });
    byRule[ruleId] = count;
  }
  summary.sort((a, b) => b.count - a.count || a.ruleId.localeCompare(b.ruleId));

  const issues = [];
  for (const row of issuesTable?.Rows || []) {
    const ruleId = textFromColumnKey(issuesTable, row, "rule_id", 0);
    if (!ruleId) {
      continue;
    }
    issues.push({
      ruleId,
      message: textFromColumnKey(issuesTable, row, "message", 1),
      movementType: textFromColumnKey(issuesTable, row, "movement_type", 2),
      closeDate: textFromColumnKey(issuesTable, row, "close_date", 3),
      detail: textFromColumnKey(issuesTable, row, "detail", 4),
      cardOwner: textFromColumnKey(issuesTable, row, "card_owner", 5),
      cardNumber: textFromColumnKey(issuesTable, row, "card_number", 6),
    });
  }

  const totalBySummary = summary.reduce((acc, item) => acc + item.count, 0);
  const missingCategoryCount = byRule.DQ003 || 0;
  const missingOwnerCount = byRule.DQ004 || 0;
  const totalIssues = totalBySummary > 0 ? totalBySummary : issues.length;

  return {
    totalIssues,
    byRule,
    summary,
    issues,
    missingCategoryCount,
    missingOwnerCount,
    warningSummary: {
      total: totalIssues,
      uncategorized: missingCategoryCount,
      unmappedOwners: missingOwnerCount,
    },
  };
}

export function buildRuntimeSnapshot(computeResult, source = "unknown") {
  const meta = extractMetaSummary(computeResult);
  const overviewProjection = buildOverviewProjection(computeResult);
  return {
    source,
    tableCount: Array.isArray(computeResult?.Tables) ? computeResult.Tables.length : 0,
    mode: overviewProjection.available ? "strict" : "hybrid",
    computeResult,
    meta,
    latestMonth: latestMonthFromMeta(meta),
    overviewProjection,
    cardMovementTrend: buildCardMovementTrend(computeResult),
    topOwners: {
      ARS: buildTopOwnersFromCompute(computeResult, "ARS"),
      USD: buildTopOwnersFromCompute(computeResult, "USD"),
    },
    topCategories: {
      ARS: buildTopCategoriesFromCompute(computeResult, "ARS"),
      USD: buildTopCategoriesFromCompute(computeResult, "USD"),
    },
    debtMaturity: buildDebtMaturityScheduleFromCompute(computeResult),
    rawExplorer: buildRawExplorerRowsFromCompute(computeResult),
    dq: buildDqDiagnosticsFromCompute(computeResult),
  };
}
