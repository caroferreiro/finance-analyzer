import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";
import { JSDOM } from "jsdom";

import { renderTableMarkup } from "./tableRenderer.js";
import {
  buildGenericTable,
  buildOwnerPivotTable,
  buildCategoryPivotTable,
  buildCategoryBreakdownTable,
} from "./tableRenderer.fixtures.js";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

function parseTableMarkup(html) {
  const { document } = new JSDOM(html).window;
  return document;
}

/** Collapse whitespace so snapshot is stable across formatting changes. */
function normalizeHtml(html) {
  return html.trim().replace(/\s+/g, " ");
}

test("renderTableMarkup returns empty string for missing table", () => {
  assert.equal(renderTableMarkup(null), "");
});

test("renderTableMarkup renders header and no body rows for empty table", () => {
  const table = {
    TableID: "meta_summary",
    Title: "Meta Summary",
    Description: "",
    Columns: [
      { Key: "key", Label: "Key", Type: "string" },
      { Key: "value", Label: "Value", Type: "string" },
    ],
    Rows: [],
  };

  const html = renderTableMarkup(table);
  const doc = parseTableMarkup(html);

  const tableEl = doc.querySelector("table");
  assert.ok(tableEl, "table element exists");

  const thead = tableEl.querySelector("thead");
  assert.ok(thead, "thead exists");

  const headerCells = thead.querySelectorAll("th");
  assert.equal(headerCells.length, 2, "two header cells");
  assert.equal(headerCells[0].textContent.trim(), "Key");
  assert.equal(headerCells[1].textContent.trim(), "Value");

  const tbody = tableEl.querySelector("tbody");
  assert.ok(tbody, "tbody exists");
  assert.equal(tbody.querySelectorAll("tr").length, 0, "no body rows");
});

test("fixtures: generic table renders with expected header count", () => {
  const table = buildGenericTable({
    columns: [
      { Key: "a", Label: "A", Type: "string" },
      { Key: "b", Label: "B", Type: "string" },
    ],
    rows: [],
  });
  const html = renderTableMarkup(table);
  const doc = parseTableMarkup(html);
  const headerCells = doc.querySelectorAll("thead th");
  assert.equal(headerCells.length, 2, "two header cells");
  assert.equal(headerCells[0].textContent.trim(), "A");
  assert.equal(headerCells[1].textContent.trim(), "B");
});

test("fixtures: owner pivot table renders two-row header with correct column count", () => {
  const table = buildOwnerPivotTable({
    owners: ["OWNER A", "OWNER B"],
    months: [],
  });
  const html = renderTableMarkup(table);
  const doc = parseTableMarkup(html);
  const headerRows = doc.querySelectorAll("thead tr");
  assert.equal(headerRows.length, 2, "two header rows");
  const row1Cells = headerRows[0].querySelectorAll("th");
  const row2Cells = headerRows[1].querySelectorAll("th");
  assert.equal(row1Cells.length, 3, "row 1: Month + 2 owner groups");
  assert.equal(row2Cells.length, 4, "row 2: ARS, USD per owner");
  assert.equal(row1Cells[0].textContent.trim(), "Month");
  assert.equal(row2Cells[0].textContent.trim(), "ARS");
  assert.equal(row2Cells[1].textContent.trim(), "USD");
});

test("fixtures: category pivot table renders two-row header", () => {
  const table = buildCategoryPivotTable({
    categories: ["Groceries", "Transport"],
    months: [],
  });
  const html = renderTableMarkup(table);
  const doc = parseTableMarkup(html);
  const headerRows = doc.querySelectorAll("thead tr");
  assert.equal(headerRows.length, 2);
  assert.equal(headerRows[0].querySelectorAll("th").length, 3, "Month + 2 category groups");
  assert.equal(headerRows[1].querySelectorAll("th").length, 4, "ARS/USD per category");
});

test("spend_by_category two-row header: row 1 has rowspan on Month and colspan=2 on category groups", () => {
  const table = buildCategoryPivotTable({
    categories: ["Groceries", "Transport"],
    months: [],
  });
  const html = renderTableMarkup(table);
  const doc = parseTableMarkup(html);
  const row1 = doc.querySelectorAll("thead tr")[0];
  const cells = row1.querySelectorAll("th");

  assert.equal(cells[0].getAttribute("rowspan"), "2", "Month has rowspan=2");
  assert.equal(cells[0].textContent.trim(), "Month");

  assert.equal(cells[1].getAttribute("colspan"), "2", "first category group has colspan=2");
  assert.equal(cells[1].textContent.trim(), "Groceries");

  assert.equal(cells[2].getAttribute("colspan"), "2", "second category group has colspan=2");
  assert.equal(cells[2].textContent.trim(), "Transport");
});

test("spend_by_category two-row header: row 2 has ARS and USD labels under each group", () => {
  const table = buildCategoryPivotTable({
    categories: ["Groceries", "Transport"],
    months: [],
  });
  const html = renderTableMarkup(table);
  const doc = parseTableMarkup(html);
  const row2 = doc.querySelectorAll("thead tr")[1];
  const cells = row2.querySelectorAll("th");

  assert.equal(cells.length, 4);
  assert.equal(cells[0].textContent.trim(), "ARS");
  assert.equal(cells[1].textContent.trim(), "USD");
  assert.equal(cells[2].textContent.trim(), "ARS");
  assert.equal(cells[3].textContent.trim(), "USD");
});

test("spend_by_owner two-row header: row 1 has Month and owner groups with colspan=2", () => {
  const table = buildOwnerPivotTable({
    owners: ["OWNER A", "OWNER B"],
    months: [],
  });
  const html = renderTableMarkup(table);
  const doc = parseTableMarkup(html);
  const row1 = doc.querySelectorAll("thead tr")[0];
  const cells = row1.querySelectorAll("th");

  assert.equal(cells[0].getAttribute("rowspan"), "2", "Month has rowspan=2");
  assert.equal(cells[0].textContent.trim(), "Month");

  assert.equal(cells[1].getAttribute("colspan"), "2", "OWNER A group has colspan=2");
  assert.equal(cells[1].textContent.trim(), "OWNER A");

  assert.equal(cells[2].getAttribute("colspan"), "2", "OWNER B group has colspan=2");
  assert.equal(cells[2].textContent.trim(), "OWNER B");
});

test("spend_by_owner two-row header: row 2 has ARS and USD under each owner", () => {
  const table = buildOwnerPivotTable({
    owners: ["OWNER A", "OWNER B"],
    months: [],
  });
  const html = renderTableMarkup(table);
  const doc = parseTableMarkup(html);
  const row2 = doc.querySelectorAll("thead tr")[1];
  const cells = row2.querySelectorAll("th");

  assert.equal(cells.length, 4);
  assert.equal(cells[0].textContent.trim(), "ARS");
  assert.equal(cells[1].textContent.trim(), "USD");
  assert.equal(cells[2].textContent.trim(), "ARS");
  assert.equal(cells[3].textContent.trim(), "USD");
});

test("money cells: non-zero values show compact format and raw value in title", () => {
  const table = buildOwnerPivotTable({
    owners: ["OWNER A"],
    months: [["2025-01-01", "20277.00", "50"]],
  });
  const html = renderTableMarkup(table);
  const doc = parseTableMarkup(html);
  const row = doc.querySelector("tbody tr");
  const cells = row.querySelectorAll("td");

  assert.equal(cells[0].textContent.trim(), "2025-01-01");
  assert.equal(cells[1].textContent.trim(), "20.28k", "20277 -> 20.28k");
  assert.equal(cells[1].getAttribute("title"), "20277.00", "title has raw value");
  assert.equal(cells[2].textContent.trim(), "50", "50 stays as 50");
  assert.equal(cells[2].getAttribute("title"), "50", "title has raw value");
});

test("money cells: zero values show em dash", () => {
  const table = buildOwnerPivotTable({
    owners: ["OWNER A"],
    months: [["2025-01-01", "0", "0.00"]],
  });
  const html = renderTableMarkup(table);
  const doc = parseTableMarkup(html);
  const row = doc.querySelector("tbody tr");
  const cells = row.querySelectorAll("td");

  assert.equal(cells[1].textContent.trim(), "—", "zero ARS shows em dash");
  assert.equal(cells[1].getAttribute("title"), "0");
  assert.equal(cells[2].textContent.trim(), "—", "zero USD shows em dash");
  assert.equal(cells[2].getAttribute("title"), "0.00");
});

test("sticky first column: pivot table header and body first cells have sticky-col class", () => {
  const table = buildOwnerPivotTable({
    owners: ["OWNER A"],
    months: [["2025-01-01", "100", "0"]],
  });
  const html = renderTableMarkup(table);
  const doc = parseTableMarkup(html);

  const firstHeaderCell = doc.querySelector("thead tr th");
  assert.ok(firstHeaderCell.classList.contains("sticky-col"), "first header th has sticky-col");

  const firstBodyCell = doc.querySelector("tbody tr td");
  assert.ok(firstBodyCell.classList.contains("sticky-col"), "first body td has sticky-col");
});

test("share columns: values display with % suffix, zero shows em dash", () => {
  const table = buildCategoryBreakdownTable({
    rows: [
      ["2025-03-01", "Groceries", "5000.00", "50.00", "38.35", "100.00"],
      ["2025-03-01", "Transport", "0.00", "0.00", "0.00", "0.00"],
    ],
  });
  const html = renderTableMarkup(table);
  const doc = parseTableMarkup(html);
  const bodyRows = doc.querySelectorAll("tbody tr");

  assert.equal(bodyRows[0].querySelectorAll("td")[4].textContent.trim(), "38.35%");
  assert.equal(bodyRows[0].querySelectorAll("td")[5].textContent.trim(), "100.00%");
  assert.equal(bodyRows[1].querySelectorAll("td")[4].textContent.trim(), "—");
  assert.equal(bodyRows[1].querySelectorAll("td")[5].textContent.trim(), "—");
});

test("spend_by_category snapshot: normalized HTML matches approved contract", () => {
  const table = buildCategoryPivotTable({
    categories: ["Cat A", "Cat B"],
    months: [["2025-02-01", "100.00", "0", "50", "25.50"]],
  });
  const html = renderTableMarkup(table);
  const actual = normalizeHtml(html);

  const snapshotPath = path.join(__dirname, "__snapshots__", "spend_by_category.txt");
  const expected = fs.readFileSync(snapshotPath, "utf8").trim();
  assert.equal(actual, expected, "spend_by_category markup must match approved snapshot");
});
