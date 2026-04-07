import { humanReadableAmount, isMoneyColumn, isShareColumn } from "./format.js";

function buildTwoRowHeader(columns) {
  if (columns.length === 0) return "<tr></tr><tr></tr>";
  const row1 = [];
  const row2 = [];
  let i = 0;

  while (i < columns.length && !(columns[i]?.Key ?? "").endsWith("_ars")) {
    const label = columns[i].Label ?? columns[i].Key ?? "";
    const stickyClass = i === 0 ? "sticky-col" : "";
    row1.push(`<th rowspan="2" class="${stickyClass}">${label}</th>`);
    i++;
  }

  while (i < columns.length) {
    const arsCol = columns[i];
    const groupLabel =
      (arsCol?.Label ?? "").replace(/\s*\(ARS\)$/i, "") ||
      (arsCol?.Key ?? "").replace(/_ars$/, "") ||
      "Totals";
    row1.push(`<th colspan="2">${groupLabel}</th>`);
    row2.push("<th>ARS</th>", "<th>USD</th>");
    i += 2;
  }

  return `<tr>${row1.join("")}</tr><tr>${row2.join("")}</tr>`;
}

export function renderTableMarkup(table) {
  if (!table) {
    return "";
  }

  const columns = table.Columns || [];
  const useTwoRowHeader = columns.some((c) => (c?.Key ?? "").endsWith("_ars"));
  const headerHtml = useTwoRowHeader
    ? buildTwoRowHeader(columns)
    : `<tr>${columns.map((c) => `<th>${c.Label ?? c.Key ?? ""}</th>`).join("")}</tr>`;

  const descriptionHtml = table.Description
    ? `<div class="table-description">${table.Description}</div>`
    : "";

  const rows = (table.Rows || [])
    .map((r) => {
      const cells = r.map((cell, j) => {
        const col = columns[j];
        const isMoney = col && isMoneyColumn(col.Type);
        const isShare = col && isShareColumn(col.Type);
        let display = isMoney ? humanReadableAmount(cell) : cell;
        if (isShare && cell !== "") {
          display = `${cell}%`;
        }
        if (isMoney && (parseFloat(String(cell).replace(/,/g, "")) === 0 || cell === "")) {
          display = "—";
        }
        if (isShare && (parseFloat(String(cell).replace(/,/g, "")) === 0 || cell === "")) {
          display = "—";
        }
        const title = isMoney ? cell : "";
        const titleAttr = title ? ` title="${String(title).replace(/"/g, "&quot;")}"` : "";
        const stickyClass = useTwoRowHeader && j === 0 ? "sticky-col" : "";
        return `<td${titleAttr}${stickyClass ? ` class="${stickyClass}"` : ""}>${display}</td>`;
      });
      return `<tr>${cells.join("")}</tr>`;
    })
    .join("");

  const tableClass = useTwoRowHeader ? " table-scroll" : "";
  return `
    ${descriptionHtml}
    <div class="table-scroll-wrap${tableClass}">
      <table>
        <thead>${headerHtml}</thead>
        <tbody>${rows}</tbody>
      </table>
    </div>
  `;
}
