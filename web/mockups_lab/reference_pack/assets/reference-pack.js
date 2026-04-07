(function (global) {
  "use strict";

  function financeOverview() {
    if (!global.FinanceOverview) {
      throw new Error("FinanceOverview runtime is missing.");
    }
    if (typeof global.FinanceOverview.loadModel !== "function") {
      throw new Error("FinanceOverview.loadModel is not available.");
    }
    return global.FinanceOverview;
  }

  function helpers() {
    var fo = financeOverview();
    if (!fo.helpers) {
      throw new Error("FinanceOverview.helpers is not available.");
    }
    return fo.helpers;
  }

  function clearBlockingError() {
    var overlay = document.getElementById("fo-error-overlay");
    if (overlay) {
      overlay.remove();
    }
  }

  async function loadModel(basePath) {
    clearBlockingError();
    return financeOverview().loadModel({ basePath: basePath || "./" });
  }

  function setText(id, value) {
    var node = document.getElementById(id);
    if (!node) {
      return;
    }
    node.textContent = value;
  }

  function monthRange(model) {
    if (!model || !model.months || !model.months.length) {
      return "No months available";
    }
    if (model.months.length === 1) {
      return model.months[0];
    }
    return model.months[0] + " to " + model.months[model.months.length - 1];
  }

  function fmtCompact(value) {
    return helpers().formatCompact(value || 0);
  }

  function fmtFull(value) {
    return helpers().formatFull(value || 0);
  }

  function metricDiff(model, currency, key) {
    return helpers().computeDiff(model, currency, key);
  }

  function rankedRows(map, prefix, limit) {
    var rows = helpers().rankMap(map, prefix, limit || 12);
    return rows.map(function (row) {
      return {
        key: String(row.label || "").replace(prefix + ": ", ""),
        label: row.label,
        value: row.value
      };
    });
  }

  function renderRankTable(tbody, rows, currency, options) {
    if (!tbody) {
      return;
    }
    var maxRows = options && options.maxRows ? options.maxRows : 8;
    var finalRows = (rows || []).slice(0, maxRows);
    tbody.innerHTML = "";

    if (!finalRows.length) {
      var empty = document.createElement("tr");
      empty.innerHTML = '<td colspan="3">No data</td>';
      tbody.appendChild(empty);
      return;
    }

    finalRows.forEach(function (row, index) {
      var tr = document.createElement("tr");
      var key = row.key || row.label || "Unknown";
      tr.innerHTML =
        "<td>" +
        (index + 1) +
        "</td><td>" +
        escapeHtml(key) +
        "</td><td class=\"num\" title=\"" +
        escapeHtml(currency + " " + fmtFull(row.value)) +
        "\">" +
        fmtCompact(row.value) +
        "</td>";
      tbody.appendChild(tr);
    });
  }

  function renderHeatGrid(host, rows, currency, options) {
    if (!host) {
      return;
    }
    var maxRows = options && options.maxRows ? options.maxRows : 12;
    var finalRows = (rows || []).slice(0, maxRows);
    host.innerHTML = "";

    if (!finalRows.length) {
      host.textContent = "No data";
      return;
    }

    var maxAbs = finalRows.reduce(function (acc, row) {
      return Math.max(acc, Math.abs(row.value || 0));
    }, 1);

    finalRows.forEach(function (row) {
      var normalized = Math.abs(row.value || 0) / maxAbs;
      var alpha = 0.18 + normalized * 0.65;
      var positive = (row.value || 0) >= 0;
      var cell = document.createElement("article");
      cell.className = "rp-map-cell";
      cell.style.background = positive
        ? "rgba(15, 118, 110, " + alpha.toFixed(3) + ")"
        : "rgba(190, 24, 93, " + alpha.toFixed(3) + ")";
      var key = row.key || row.label || "Unknown";
      cell.innerHTML =
        "<div class=\"name\">" +
        escapeHtml(key) +
        "</div><div class=\"amt\" title=\"" +
        escapeHtml(currency + " " + fmtFull(row.value)) +
        "\">" +
        fmtCompact(row.value) +
        "</div>";
      host.appendChild(cell);
    });
  }

  function renderKpiStrip(host, items) {
    if (!host) {
      return;
    }
    host.innerHTML = "";
    (items || []).forEach(function (item) {
      var tile = document.createElement("article");
      tile.className = "rp-kpi";
      var tone = item.deltaClass || "";
      tile.innerHTML =
        "<div class=\"label\">" +
        escapeHtml(item.label) +
        "</div><div class=\"value\" title=\"" +
        escapeHtml(item.currency + " " + fmtFull(item.value)) +
        "\">" +
        fmtCompact(item.value) +
        "</div><div class=\"delta " +
        tone +
        "\">" +
        escapeHtml(item.deltaText || "") +
        "</div>";
      host.appendChild(tile);
    });
  }

  function escapeHtml(value) {
    return String(value)
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#39;");
  }

  function filteredRows(processedRows, filters) {
    var data = processedRows || [];
    var month = filters && filters.month ? filters.month : "all";
    var movement = filters && filters.movement ? filters.movement : "all";
    var query = filters && filters.query ? String(filters.query).trim().toUpperCase() : "";

    return data.filter(function (row) {
      if (month !== "all" && row.statementMonth !== month) {
        return false;
      }
      if (movement !== "all" && row.movementType !== movement) {
        return false;
      }
      if (!query) {
        return true;
      }
      var haystack = [row.owner, row.bank, row.category, row.detail]
        .join(" ")
        .toUpperCase();
      return haystack.indexOf(query) >= 0;
    });
  }

  function summarizeRows(rows) {
    return rows.reduce(
      function (acc, row) {
        acc.count += 1;
        acc.ars += row.amountARS || 0;
        acc.usd += row.amountUSD || 0;
        return acc;
      },
      { count: 0, ars: 0, usd: 0 }
    );
  }

  global.ReferencePack = {
    filteredRows: filteredRows,
    fmtCompact: fmtCompact,
    fmtFull: fmtFull,
    loadModel: loadModel,
    metricDiff: metricDiff,
    monthRange: monthRange,
    rankedRows: rankedRows,
    renderHeatGrid: renderHeatGrid,
    renderKpiStrip: renderKpiStrip,
    renderRankTable: renderRankTable,
    setText: setText,
    summarizeRows: summarizeRows,
    clearBlockingError: clearBlockingError
  };
})(window);
