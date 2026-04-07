import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { parseCsvMatrix } from "./mockupsRuntime.js";
import { parseCategorySegmentsCsv } from "../mappingsCsvParse.js";
import { normalizeCategoryKey } from "./categorySegments.js";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const REPO_ROOT = path.resolve(__dirname, "..", "..");
const EXEMPT_CATEGORY_KEYS = new Set(["adjustments", "uncategorized", "?"]);

function activeCategoryKeysFromCategoryMap(csvText) {
  const matrix = parseCsvMatrix(csvText, ",");
  const keys = new Set();
  for (const row of matrix.slice(1)) {
    const category = normalizeCategoryKey(row?.[1] || "");
    if (!category || EXEMPT_CATEGORY_KEYS.has(category)) {
      continue;
    }
    keys.add(category);
  }
  return keys;
}

function activeCategoryKeysFromDemoMappings(parsed) {
  const keys = new Set();
  const categoryByDetail = parsed?.categoryByDetail || {};
  for (const value of Object.values(categoryByDetail)) {
    const category = normalizeCategoryKey(value);
    if (!category || EXEMPT_CATEGORY_KEYS.has(category)) {
      continue;
    }
    keys.add(category);
  }
  return keys;
}

function assertCoverage(activeCategoryKeys, segmentLookup, label) {
  const missing = [...activeCategoryKeys].filter((key) => !Object.prototype.hasOwnProperty.call(segmentLookup, key));
  assert.deepEqual(missing, [], `${label} is missing category segment mappings for: ${missing.join(", ")}`);
}

test("public category segment mapping covers active public categories", async () => {
  const categoryMapPath = path.join(
    REPO_ROOT,
    "web",
    "mockups_lab",
    "tmp_public_data",
    "current",
    "details_to_categories_map.csv"
  );
  const segmentMapPath = path.join(
    REPO_ROOT,
    "web",
    "mockups_lab",
    "tmp_public_data",
    "current",
    "category_segments_map.csv"
  );
  const [categoryMapText, segmentMapText] = await Promise.all([
    fs.readFile(categoryMapPath, "utf8"),
    fs.readFile(segmentMapPath, "utf8"),
  ]);

  assertCoverage(
    activeCategoryKeysFromCategoryMap(categoryMapText),
    parseCategorySegmentsCsv(segmentMapText),
    "public profile"
  );
});

test("embedded demo mappings cover active demo categories", async () => {
  const demoMappingsPath = path.join(REPO_ROOT, "pkg", "demo_dataset", "mappings.v1.json");
  const parsed = JSON.parse(await fs.readFile(demoMappingsPath, "utf8"));
  const segmentLookup = Object.fromEntries(
    Object.entries(parsed?.categorySegmentByCategory || {}).map(([category, segment]) => [
      normalizeCategoryKey(category),
      String(segment || "").trim(),
    ])
  );

  assertCoverage(activeCategoryKeysFromDemoMappings(parsed), segmentLookup, "embedded demo dataset");
});

test("sensitive category segment mapping covers active sensitive categories when overlay is present", async (t) => {
  const categoryMapPath = path.join(
    REPO_ROOT,
    "web",
    "mockups_lab",
    "tmp_sensitive_data",
    "current",
    "details_to_categories_map.csv"
  );
  const segmentMapPath = path.join(
    REPO_ROOT,
    "web",
    "mockups_lab",
    "tmp_sensitive_data",
    "current",
    "category_segments_map.csv"
  );

  try {
    await Promise.all([fs.access(categoryMapPath), fs.access(segmentMapPath)]);
  } catch {
    t.skip("Sensitive overlay is not present in this checkout.");
    return;
  }

  const [categoryMapText, segmentMapText] = await Promise.all([
    fs.readFile(categoryMapPath, "utf8"),
    fs.readFile(segmentMapPath, "utf8"),
  ]);

  assertCoverage(
    activeCategoryKeysFromCategoryMap(categoryMapText),
    parseCategorySegmentsCsv(segmentMapText),
    "sensitive profile"
  );
});
