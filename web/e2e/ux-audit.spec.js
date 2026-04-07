import fs from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { test, expect } from "@playwright/test";
import uxContract from "./fixtures/ux-audit-contract.json" with { type: "json" };

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const OUTPUT_DIR = path.resolve(__dirname, "../output/playwright/ux-audit");
const SUMMARY_PATH = path.join(OUTPUT_DIR, "ux-audit-summary.json");

const VIEWS = [
  { key: "overview", navId: "#fo-nav-overview", sectionId: "#overview" },
  { key: "debt", navId: "#fo-nav-debt", sectionId: "#debt" },
  { key: "owners", navId: "#fo-nav-owners", sectionId: "#owners" },
  { key: "categories", navId: "#fo-nav-categories", sectionId: "#categories" },
  { key: "dq", navId: "#fo-nav-dq", sectionId: "#dq" },
  { key: "raw", navId: "#fo-nav-raw", sectionId: "#raw" },
  { key: "settings", navId: "#fo-nav-settings", sectionId: "#settings" },
];

function normalizeText(value) {
  return String(value || "").replace(/\s+/g, " ").trim();
}

test("ux audit emits per-view screenshots + structural summary and enforces contract", async ({ page }) => {
  await fs.mkdir(OUTPUT_DIR, { recursive: true });

  await page.goto("/?uxAudit=1");
  await expect
    .poll(async () => page.evaluate(() => String(document.body.dataset.foBootState || "")))
    .toBe("ready");

  const runtimeMeta = await page.evaluate(() => ({
    url: window.location.href,
    runtimeMode: String(document.body.dataset.foRuntimeMode || ""),
    loadProfile: String(document.body.dataset.foLoadProfile || ""),
    bootState: String(document.body.dataset.foBootState || ""),
  }));

  expect(runtimeMeta.runtimeMode).toBe(uxContract.expectedRuntimeMode);
  expect(runtimeMeta.loadProfile).toBe(uxContract.expectedLoadProfile);
  expect(runtimeMeta.bootState).toBe("ready");

  // Remove animation jitter so screenshots and layout checks are more stable.
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

  const views = [];
  for (const view of VIEWS) {
    await page.locator(view.navId).click();
    await expect(page.locator(view.sectionId)).toBeVisible();
    await expect(page.locator(view.sectionId)).toHaveAttribute("aria-hidden", "false");

    const screenshotPath = path.join(OUTPUT_DIR, `ux-audit-${view.key}.png`);
    await page.locator(view.sectionId).screenshot({ path: screenshotPath });

    const summary = await page.evaluate((sectionSelector) => {
      const section = document.querySelector(sectionSelector);
      if (!section) {
        return null;
      }

      const chartHostSet = new Set();
      const chartCandidates = section.querySelectorAll("[id$='-host']");
      for (const candidate of chartCandidates) {
        const id = String(candidate.id || "");
        if (!id) {
          continue;
        }
        if (
          candidate.querySelector("canvas, .fo-highcharts-host, .highcharts-root") ||
          candidate.getAttribute("data-fo-chart-role")
        ) {
          chartHostSet.add(id);
        }
      }

      const tables = Array.from(section.querySelectorAll("table"));
      const sampleHeaders = tables.slice(0, 3).map((table) => {
        const row = table.querySelector("thead tr");
        const cells = row ? Array.from(row.querySelectorAll("th")) : [];
        return cells.map((cell) => String(cell.textContent || "").replace(/\s+/g, " ").trim());
      });

      const text = String(section.textContent || "").replace(/\s+/g, " ").trim();
      const loadingFragments = [];
      for (const fragment of ["Loading runtime metrics...", "Loading strict runtime data..."]) {
        if (text.includes(fragment)) {
          loadingFragments.push(fragment);
        }
      }

      return {
        cardCount: section.querySelectorAll(".card").length,
        tableCount: tables.length,
        chartHostCount: chartHostSet.size,
        chartHosts: Array.from(chartHostSet).sort(),
        sampleHeaders,
        loadingFragments,
        hasHorizontalOverflow: section.scrollWidth > section.clientWidth + 1,
      };
    }, view.sectionId);

    expect(summary).toBeTruthy();
    views.push({
      key: view.key,
      ...summary,
      screenshot: screenshotPath,
    });
  }

  const report = {
    contractVersion: uxContract.version,
    ...runtimeMeta,
    views,
  };
  await fs.writeFile(SUMMARY_PATH, `${JSON.stringify(report, null, 2)}\n`, "utf8");

  const viewByKey = new Map(views.map((view) => [view.key, view]));
  expect(viewByKey.size).toBe(uxContract.views.length);
  for (const expectedView of uxContract.views) {
    const actual = viewByKey.get(expectedView.key);
    expect(actual, `Missing view in audit summary: ${expectedView.key}`).toBeTruthy();
    expect(
      actual.cardCount,
      `${expectedView.key} cardCount drifted from contract`
    ).toBe(expectedView.expectedCardCount);
    expect(
      actual.tableCount,
      `${expectedView.key} tableCount drifted from contract`
    ).toBe(expectedView.expectedTableCount);
    expect(
      actual.chartHostCount,
      `${expectedView.key} chartHostCount drifted from contract`
    ).toBe(expectedView.expectedChartHostCount);
    expect(
      actual.hasHorizontalOverflow,
      `${expectedView.key} horizontal overflow drifted from contract`
    ).toBe(expectedView.expectHorizontalOverflow);

    for (const requiredChartHost of expectedView.requiredChartHosts) {
      expect(
        actual.chartHosts,
        `${expectedView.key} must include chart host ${normalizeText(requiredChartHost)}`
      ).toContain(requiredChartHost);
    }
    expect(
      actual.loadingFragments,
      `${expectedView.key} should not keep loading placeholders after ready`
    ).toEqual([]);
  }
});
