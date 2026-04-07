import { test, expect } from "@playwright/test";

const EXTERNAL_BASE_URL = process.env.PLAYWRIGHT_BASE_URL
  ? String(process.env.PLAYWRIGHT_BASE_URL)
  : "";

test.skip(!EXTERNAL_BASE_URL, "PLAYWRIGHT_BASE_URL is required for deployed complete checks.");

function unique(values) {
  return Array.from(new Set(values.filter(Boolean)));
}

function failureReport({ finalBootState, finalLoadProfile, sameOriginFailures, externalFailures, consoleErrors, pageErrors }) {
  const lines = [
    `Base URL: ${EXTERNAL_BASE_URL}`,
    `Final boot state: ${finalBootState || "<empty>"}`,
    `Final load profile: ${finalLoadProfile || "<empty>"}`,
    "",
    "Same-origin failures:",
    ...unique(sameOriginFailures),
    "",
    "External failures:",
    ...unique(externalFailures),
    "",
    "Console errors:",
    ...unique(consoleErrors),
    "",
    "Page errors:",
    ...unique(pageErrors),
  ];
  return lines.join("\n");
}

async function readRuntimeState(page) {
  return page.evaluate(() => ({
    bootState: String(document.body.dataset.foBootState || ""),
    loadProfile: String(document.body.dataset.foLoadProfile || ""),
    highchartsLoaded: Boolean(window.Highcharts && typeof window.Highcharts.chart === "function"),
  }));
}

async function waitForSettledBootState(page, timeoutMs) {
  const deadline = Date.now() + timeoutMs;
  let lastBootState = "";

  while (Date.now() < deadline) {
    const { bootState } = await readRuntimeState(page);
    lastBootState = bootState;
    if (bootState && bootState !== "loading") {
      return bootState;
    }
    await page.waitForTimeout(500);
  }

  throw new Error(`Timed out waiting for boot state to leave 'loading'. Last state: ${lastBootState || "<empty>"}`);
}

test("complete deployed site boots fully and renders charts from the real website", async ({ page }, testInfo) => {
  test.setTimeout(90_000);

  const sameOriginFailures = [];
  const externalFailures = [];
  const consoleErrors = [];
  const pageErrors = [];
  const siteOrigin = new URL(EXTERNAL_BASE_URL).origin;

  page.on("response", (response) => {
    const url = response.url();
    const failure = `${response.status()} ${url}`;
    if (url.startsWith(siteOrigin)) {
      if (response.status() >= 400) {
        sameOriginFailures.push(failure);
      }
      return;
    }
    if (response.status() >= 400) {
      externalFailures.push(failure);
    }
  });

  page.on("requestfailed", (request) => {
    const url = request.url();
    const failure = `${request.failure()?.errorText || "request failed"} ${url}`;
    if (url.startsWith(siteOrigin)) {
      sameOriginFailures.push(failure);
    } else {
      externalFailures.push(failure);
    }
  });

  page.on("console", (message) => {
    if (message.type() === "error") {
      consoleErrors.push(message.text());
    }
  });

  page.on("pageerror", (error) => {
    pageErrors.push(error?.stack || error?.message || String(error));
  });

  let finalBootState = "";
  let finalLoadProfile = "";

  try {
    await page.goto(EXTERNAL_BASE_URL, { waitUntil: "domcontentloaded" });

    await expect(page.locator("#fo-nav-overview")).toBeVisible();

    finalBootState = await waitForSettledBootState(page, 60_000);
    const runtimeState = await readRuntimeState(page);
    finalLoadProfile = runtimeState.loadProfile;

    expect(finalBootState).toBe("ready");
    expect(finalLoadProfile).toBe("public");
    expect(runtimeState.highchartsLoaded).toBe(true);
    expect(unique(sameOriginFailures)).toEqual([]);
    expect(unique(externalFailures)).toEqual([]);
    expect(unique(consoleErrors)).toEqual([]);
    expect(unique(pageErrors)).toEqual([]);

    await expect(page.locator(".fo-startup-card")).toHaveCount(0);
    await expect(page.locator("#fo-overview-new-table tbody tr")).toHaveCount(8);
    await expect(page.locator("#fo-overview-q8-host .highcharts-root").first()).toBeVisible();

    const overviewQ8Datasets = await page.evaluate(() => {
      const hosts = Array.from(document.querySelectorAll("#fo-overview-q8-host .fo-highcharts-host"));
      return hosts.map((host) => {
        const chart = host && host._foHighchartsInstance ? host._foHighchartsInstance : null;
        return chart && Array.isArray(chart.series) ? chart.series.map((series) => String(series.name || "")) : [];
      });
    });

    expect(overviewQ8Datasets[0]).toContain("Total remaining installment debt ARS");
    expect(overviewQ8Datasets[1]).toContain("Total remaining installment debt USD");
  } catch (error) {
    const runtimeState = await readRuntimeState(page).catch(() => ({
      bootState: finalBootState,
      loadProfile: finalLoadProfile,
      highchartsLoaded: false,
    }));
    finalBootState = runtimeState.bootState;
    finalLoadProfile = runtimeState.loadProfile;

    await testInfo.attach("deployed-complete-report.txt", {
      body: Buffer.from(
        failureReport({
          finalBootState,
          finalLoadProfile,
          sameOriginFailures,
          externalFailures,
          consoleErrors,
          pageErrors,
        }),
        "utf8"
      ),
      contentType: "text/plain",
    });

    const screenshot = await page.screenshot({ fullPage: true }).catch(() => null);
    if (screenshot) {
      await testInfo.attach("deployed-complete-screenshot.png", {
        body: screenshot,
        contentType: "image/png",
      });
    }

    throw error;
  }
});
