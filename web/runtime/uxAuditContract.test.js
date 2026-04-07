import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const CONTRACT_PATH = path.resolve(__dirname, "../e2e/fixtures/ux-audit-contract.json");
const contract = JSON.parse(fs.readFileSync(CONTRACT_PATH, "utf8"));

test("ux audit contract has canonical view set and unique keys", () => {
  const keys = contract.views.map((view) => view.key);
  assert.deepEqual(keys, ["overview", "debt", "owners", "categories", "dq", "raw", "settings"]);
  assert.equal(new Set(keys).size, keys.length);
});

test("ux audit contract uses non-negative integer exact counts", () => {
  for (const view of contract.views) {
    for (const field of ["expectedCardCount", "expectedTableCount", "expectedChartHostCount"]) {
      assert.equal(Number.isInteger(view[field]), true, `${view.key}.${field} must be integer`);
      assert.equal(view[field] >= 0, true, `${view.key}.${field} must be >= 0`);
    }
    assert.equal(
      typeof view.expectHorizontalOverflow,
      "boolean",
      `${view.key}.expectHorizontalOverflow must be boolean`
    );
  }
});

test("ux audit contract chart host lists match expected chart host counts", () => {
  for (const view of contract.views) {
    assert.equal(Array.isArray(view.requiredChartHosts), true, `${view.key}.requiredChartHosts must be array`);
    const uniqueHosts = new Set(view.requiredChartHosts);
    assert.equal(
      uniqueHosts.size,
      view.requiredChartHosts.length,
      `${view.key}.requiredChartHosts must not contain duplicates`
    );
    if (view.expectedChartHostCount === 0) {
      assert.equal(view.requiredChartHosts.length, 0, `${view.key} should not declare chart hosts`);
      continue;
    }
    assert.equal(
      view.requiredChartHosts.length,
      view.expectedChartHostCount,
      `${view.key}.requiredChartHosts length must match expectedChartHostCount`
    );
    for (const host of view.requiredChartHosts) {
      assert.equal(
        String(host).startsWith("fo-"),
        true,
        `${view.key} chart host '${host}' should use canonical fo-* id`
      );
    }
  }
});
