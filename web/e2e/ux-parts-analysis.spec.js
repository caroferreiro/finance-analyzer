import fs from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { test, expect } from "@playwright/test";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const OUTPUT_DIR = path.resolve(__dirname, "../output/playwright/ux-parts-analysis");
const SUMMARY_PATH = path.join(OUTPUT_DIR, "ux-parts-analysis.json");

const VIEW_PARTS = [
  { key: "view-overview", navId: "#fo-nav-overview", sectionId: "#overview", label: "Overview" },
  { key: "view-debt", navId: "#fo-nav-debt", sectionId: "#debt", label: "Debt" },
  { key: "view-owners", navId: "#fo-nav-owners", sectionId: "#owners", label: "Owners" },
  { key: "view-categories", navId: "#fo-nav-categories", sectionId: "#categories", label: "Categories" },
  { key: "view-dq", navId: "#fo-nav-dq", sectionId: "#dq", label: "Data Quality" },
  { key: "view-raw", navId: "#fo-nav-raw", sectionId: "#raw", label: "Raw Data" },
  { key: "view-settings", navId: "#fo-nav-settings", sectionId: "#settings", label: "Settings" },
];

function normalizeText(value) {
  return String(value || "").replace(/\s+/g, " ").trim();
}

function deriveGeneralSuggestions(metrics) {
  const items = [];
  if (metrics.topbarActionCount >= 4) {
    items.push("Topbar actions are dense; evaluate progressive disclosure for secondary actions.");
  }
  if (metrics.navButtonCount >= 7) {
    items.push("Sidebar has many primary entries; evaluate grouping labels or quick-jump affordances.");
  }
  if (metrics.placeholderTextHits > 0) {
    items.push("Residual placeholder copy found in shell-level text; replace with decision-oriented copy.");
  }
  if (!items.length) {
    items.push("Run one manual pass for spacing rhythm consistency between topbar, cards, and section starts.");
  }
  return items;
}

function deriveViewSuggestions(metrics) {
  const items = [];
  if (metrics.placeholderCount > 0) {
    items.push("Replace placeholder copy with production wording and examples.");
  }
  if (metrics.horizontalOverflowCount > 0) {
    items.push("Review overflow sources and reduce truncation/friction for first-screen scanning.");
  }
  if (metrics.tableCount >= 2) {
    items.push("Prioritize first table and defer secondary tables behind progressive disclosure.");
  }
  if (metrics.chartCount >= 2) {
    items.push("Confirm chart priority order and ensure first chart answers the main user question.");
  }
  if (metrics.detailsClosedCount > 0) {
    items.push("Validate summary labels of collapsed blocks to communicate value before expanding.");
  }
  if (!items.length) {
    items.push("No structural hotspots detected; run a focused copy/labels pass.");
  }
  return items.slice(0, 3);
}

function deriveTableSuggestions(metrics) {
  const items = [];
  if (metrics.wideTableCount > 0) {
    items.push("Wide tables detected; evaluate compact column presets and default ordering by decision priority.");
  }
  if (metrics.stickyHeaderTableCount < metrics.totalTableCount) {
    items.push("Not all tables use sticky headers; evaluate sticky context for long-scroll tables.");
  }
  if (metrics.ariaLabeledTableCount < metrics.totalTableCount) {
    items.push("Some tables miss aria-label; close accessibility gap in table semantics.");
  }
  if (!items.length) {
    items.push("Table layer is structurally stable; next step is tuning sort/filter discoverability.");
  }
  return items;
}

function deriveChartSuggestions(metrics) {
  const items = [];
  if (metrics.ariaLabeledCanvasCount < metrics.canvasCount) {
    items.push("Some canvases miss aria labels; complete chart accessibility contract.");
  }
  if (metrics.emptyHostCount > 0) {
    items.push("Some chart hosts are empty; improve empty-state messaging or hide until data is ready.");
  }
  if (metrics.canvasCount >= 6) {
    items.push("High chart density; evaluate chart hierarchy and whether secondary charts should be collapsed.");
  }
  if (!items.length) {
    items.push("Chart layer is stable; next iteration can focus on annotation and narrative cues.");
  }
  return items;
}

test("ux parts analysis emits per-part findings using code + playwright evidence", async ({ page }) => {
  await fs.mkdir(OUTPUT_DIR, { recursive: true });

  await page.goto("/?uxPartsAudit=1");
  await expect
    .poll(async () => page.evaluate(() => String(document.body.dataset.foBootState || "")))
    .toBe("ready");

  await page.addStyleTag({
    content: `
      *,
      *::before,
      *::after {
        animation: none !important;
        transition: none !important;
        scroll-behavior: auto !important;
      }
    `,
  });

  const report = {
    generatedAt: new Date().toISOString(),
    url: page.url(),
    runtimeMode: await page.evaluate(() => String(document.body.dataset.foRuntimeMode || "")),
    loadProfile: await page.evaluate(() => String(document.body.dataset.foLoadProfile || "")),
    parts: [],
  };

  const generalMetrics = await page.evaluate(() => {
    const navButtons = Array.from(document.querySelectorAll(".nav button"));
    const topbarActions = Array.from(document.querySelectorAll(".topbar .actions .btn"));
    const shellTextNodes = Array.from(document.querySelectorAll(".brand, .topbar, .sidebar .note, .nav button"));
    const shellText = shellTextNodes.map((node) => String(node.textContent || ""));
    const placeholderTextHits = shellText.filter((text) =>
      text.toLowerCase().includes("placeholder")
    ).length;
    return {
      navButtonCount: navButtons.length,
      activeNavCount: navButtons.filter((button) => button.classList.contains("active")).length,
      topbarActionCount: topbarActions.length,
      hiddenTopbarActionCount: topbarActions.filter((button) => button.classList.contains("is-hidden")).length,
      placeholderTextHits,
    };
  });
  await page.screenshot({ path: path.join(OUTPUT_DIR, "part-general-shell.png"), fullPage: true });
  report.parts.push({
    key: "general-shell",
    label: "General shell (sidebar + topbar + load lifecycle)",
    order: 1,
    metrics: generalMetrics,
    suggestions: deriveGeneralSuggestions(generalMetrics),
    screenshot: path.join(OUTPUT_DIR, "part-general-shell.png"),
  });

  for (const [index, view] of VIEW_PARTS.entries()) {
    await page.locator(view.navId).click();
    await expect(page.locator(view.sectionId)).toBeVisible();
    await expect(page.locator(view.sectionId)).toHaveAttribute("aria-hidden", "false");

    const metrics = await page.evaluate((sectionSelector) => {
      function cleanText(value) {
        return String(value || "").replace(/\s+/g, " ").trim();
      }
      const section = document.querySelector(sectionSelector);
      if (!section) {
        return null;
      }
      const details = Array.from(section.querySelectorAll("details"));
      const overflowNodes = Array.from(section.querySelectorAll("*"))
        .filter((node) => node.clientWidth > 0 && node.scrollWidth > node.clientWidth + 2)
        .slice(0, 12);
      const placeholderSamples = Array.from(
        section.querySelectorAll("h3, .note, .rightdesc, td, th, .placeholder")
      )
        .map((node) => cleanText(node.textContent))
        .filter((text) => text.toLowerCase().includes("placeholder"))
        .slice(0, 6);
      return {
        cardCount: section.querySelectorAll(".card").length,
        tableCount: section.querySelectorAll("table").length,
        chartCount: section.querySelectorAll("canvas").length,
        buttonCount: section.querySelectorAll("button").length,
        detailsCount: details.length,
        detailsClosedCount: details.filter((node) => !node.open).length,
        horizontalOverflowCount: overflowNodes.length,
        placeholderCount: placeholderSamples.length,
        placeholderSamples,
      };
    }, view.sectionId);

    expect(metrics, `Missing metrics for ${view.key}`).toBeTruthy();
    const screenshotPath = path.join(OUTPUT_DIR, `part-${view.key}.png`);
    await page.locator(view.sectionId).screenshot({ path: screenshotPath });
    report.parts.push({
      key: view.key,
      label: `${view.label} view`,
      order: 2 + index,
      metrics,
      suggestions: deriveViewSuggestions(metrics),
      screenshot: screenshotPath,
    });
  }

  const tableMetrics = await page.evaluate(() => {
    const tables = Array.from(document.querySelectorAll("section table"));
    const stickyHeaderTableCount = tables.filter((table) => {
      const th = table.querySelector("thead th");
      return th && getComputedStyle(th).position === "sticky";
    }).length;
    const ariaLabeledTableCount = tables.filter((table) => table.hasAttribute("aria-label")).length;
    const wideTableCount = tables.filter(
      (table) => table.clientWidth > 0 && table.scrollWidth > table.clientWidth + 2
    ).length;
    const maxHeaderColumns = Math.max(
      0,
      ...tables.map((table) => table.querySelectorAll("thead tr:first-child th").length)
    );
    return {
      totalTableCount: tables.length,
      stickyHeaderTableCount,
      ariaLabeledTableCount,
      wideTableCount,
      maxHeaderColumns,
    };
  });
  report.parts.push({
    key: "tables-crosscut",
    label: "Tables cross-cutting layer",
    order: 9,
    metrics: tableMetrics,
    suggestions: deriveTableSuggestions(tableMetrics),
  });

  const chartMetrics = await page.evaluate(() => {
    function cleanText(value) {
      return String(value || "").replace(/\s+/g, " ").trim();
    }
    const hosts = Array.from(document.querySelectorAll("[id$='-host']"));
    const chartHosts = hosts.filter(
      (host) => host.querySelector("canvas") || host.getAttribute("data-fo-chart-role")
    );
    const canvases = Array.from(document.querySelectorAll("canvas"));
    const ariaLabeledCanvasCount = canvases.filter(
      (canvas) => cleanText(canvas.getAttribute("aria-label")).length > 0
    ).length;
    const emptyHostCount = chartHosts.filter((host) => !host.querySelector("canvas")).length;
    return {
      hostCount: chartHosts.length,
      canvasCount: canvases.length,
      ariaLabeledCanvasCount,
      emptyHostCount,
    };
  });
  report.parts.push({
    key: "charts-crosscut",
    label: "Charts cross-cutting layer",
    order: 10,
    metrics: chartMetrics,
    suggestions: deriveChartSuggestions(chartMetrics),
  });

  report.parts.sort((a, b) => a.order - b.order);
  await fs.writeFile(SUMMARY_PATH, `${JSON.stringify(report, null, 2)}\n`, "utf8");

  expect(report.runtimeMode).toBe("strict");
  expect(report.loadProfile).toBe("public");
  expect(report.parts).toHaveLength(10);
  for (const part of report.parts) {
    expect(Array.isArray(part.suggestions)).toBeTruthy();
    expect(part.suggestions.length).toBeGreaterThan(0);
  }
});
