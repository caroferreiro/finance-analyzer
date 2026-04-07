/**
 * Fixture builders for table models used in renderer tests.
 * Tables match the shape produced by the Go engine (TableID, Title, Columns, Rows).
 */

export function buildGenericTable({ columns = [], rows = [], description = "" } = {}) {
  return {
    TableID: "generic",
    Title: "Generic Table",
    Description: description,
    Columns: columns.length ? columns : [{ Key: "key", Label: "Key", Type: "string" }],
    Rows: rows,
  };
}

export function buildOwnerPivotTable({ owners = ["OWNER A", "OWNER B"], months = [] } = {}) {
  const columns = [
    { Key: "month", Label: "Month", Type: "date" },
    ...owners.flatMap((owner) => [
      { Key: `${owner}_ars`, Label: `${owner} (ARS)`, Type: "money_ars" },
      { Key: `${owner}_usd`, Label: `${owner} (USD)`, Type: "money_usd" },
    ]),
  ];
  const rows = months.map(([month, ...cells]) => [month, ...cells]);
  return {
    TableID: "spend_by_owner",
    Title: "Spend by Owner",
    Description: "",
    Columns: columns,
    Rows: rows,
  };
}

export function buildCategoryPivotTable({ categories = [], months = [] } = {}) {
  const columns = [
    { Key: "month", Label: "Month", Type: "date" },
    ...categories.flatMap((cat) => [
      { Key: `${cat}_ars`, Label: `${cat} (ARS)`, Type: "money_ars" },
      { Key: `${cat}_usd`, Label: `${cat} (USD)`, Type: "money_usd" },
    ]),
  ];
  const rows = months.map(([month, ...cells]) => [month, ...cells]);
  return {
    TableID: "spend_by_category",
    Title: "Spend by Category",
    Description: "",
    Columns: columns,
    Rows: rows,
  };
}

export function buildCategoryBreakdownTable({ rows = [] } = {}) {
  const columns = [
    { Key: "month", Label: "Month", Type: "date" },
    { Key: "category", Label: "Category", Type: "string" },
    { Key: "card_movement_total_ars", Label: "Card Movement Total ARS", Type: "money_ars" },
    { Key: "card_movement_total_usd", Label: "Card Movement Total USD", Type: "money_usd" },
    { Key: "share_of_month_ars_pct", Label: "Share of Month ARS", Type: "share" },
    { Key: "share_of_month_usd_pct", Label: "Share of Month USD", Type: "share" },
  ];
  return {
    TableID: "category_breakdown_by_month",
    Title: "Category Breakdown by Month",
    Description: "",
    Columns: columns,
    Rows: rows,
  };
}
