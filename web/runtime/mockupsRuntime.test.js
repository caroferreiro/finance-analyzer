import test from "node:test";
import assert from "node:assert";

import {
  normalizeBasePath,
  resolveDataPath,
  buildMappingsFromRequiredCsv,
  bundleFromDataFiles,
  bundleFromRequiredFiles,
  loadRequiredFiles,
  restoreBundleFromStorage,
  loadMockupBundle,
  computeBundle,
  exportTableCSV,
  exportWorkspaceArtifact,
  parseWorkspaceArtifact,
} from "./mockupsRuntime.js";

const VALID_DATA_CSV = [
  "Bank,CardCompany,CloseDate,ExpirationDate,TotalARS,TotalUSD,CardNumber,CardOwner,CardTotalARS,CardTotalUSD,MovementType,OriginalDate,ReceiptNumber,Detail,CurrentInstallment,TotalInstallments,AmountARS,AmountUSD",
  "Santander,VISA,01/02/2025,10/02/2025,1000,0,1234,Alice,1000,0,CardMovement,31/01/2025,1,Supermarket,1,1,1000,0",
].join("\n");

const VALID_OWNER_MAP = ["RawOwner,OwnerNormalized", "Alice,Alice Normalized"].join("\n");
const VALID_CATEGORY_MAP = ["Detail,Category", "Supermarket,Food"].join("\n");
const VALID_CATEGORY_SEGMENTS_MAP = ["category;segment", "Food;essential"].join("\n");
const SENSITIVE_EXTRA_DATA_FILE = "2025-03-21-Santander-AMEX.pdf.csv";

test("base path helpers normalize and resolve consistently", () => {
  assert.strictEqual(normalizeBasePath(""), "./");
  assert.strictEqual(normalizeBasePath("mockups_lab"), "mockups_lab/");
  assert.strictEqual(resolveDataPath("mockups_lab", "data.csv"), "mockups_lab/data.csv");
});

test("buildMappingsFromRequiredCsv parses strict CSV and normalizes detail keys", () => {
  const mappings = buildMappingsFromRequiredCsv(
    VALID_OWNER_MAP,
    VALID_CATEGORY_MAP,
    VALID_CATEGORY_SEGMENTS_MAP
  );
  assert.deepStrictEqual(mappings.ownersByCardNumber, {});
  assert.strictEqual(mappings.ownersByCardOwner.ALICE, "Alice Normalized");
  assert.strictEqual(mappings.categoryByDetail.SUPERMARKET, "Food");
  assert.deepStrictEqual(
    Object.fromEntries(Object.entries(mappings.categorySegmentByCategory)),
    { food: "essential" }
  );
});

test("buildMappingsFromRequiredCsv parses wildcard prefix entries into categoryByDetailPrefix", () => {
  const categoryMapWithPrefixes = [
    "Detail,Category",
    "Supermarket,Food",
    "JETSMART*,Travel",
    "FABRIC SUSHI*,Take out",
  ].join("\n");
  const mappings = buildMappingsFromRequiredCsv(
    VALID_OWNER_MAP,
    categoryMapWithPrefixes,
    VALID_CATEGORY_SEGMENTS_MAP
  );
  assert.strictEqual(mappings.categoryByDetail.SUPERMARKET, "Food");
  assert.ok(!("JETSMART*" in mappings.categoryByDetail), "wildcard entry must not appear in exact map");
  assert.ok(Array.isArray(mappings.categoryByDetailPrefix));
  assert.strictEqual(mappings.categoryByDetailPrefix.length, 2);
  assert.strictEqual(mappings.categoryByDetailPrefix[0].prefix, "FABRIC SUSHI");
  assert.strictEqual(mappings.categoryByDetailPrefix[0].category, "Take out");
  assert.strictEqual(mappings.categoryByDetailPrefix[1].prefix, "JETSMART");
  assert.strictEqual(mappings.categoryByDetailPrefix[1].category, "Travel");
});

test("buildMappingsFromRequiredCsv sorts prefixes longest-first", () => {
  const categoryMap = [
    "Detail,Category",
    "MER*,Generic",
    "MERPAGO*,Payments",
    "MERPAGO*CARREFOUR*,Supermarket",
  ].join("\n");
  const mappings = buildMappingsFromRequiredCsv(
    VALID_OWNER_MAP,
    categoryMap,
    VALID_CATEGORY_SEGMENTS_MAP
  );
  const prefixes = mappings.categoryByDetailPrefix.map(e => e.prefix);
  assert.deepStrictEqual(prefixes, ["MERPAGO*CARREFOUR", "MERPAGO", "MER"]);
});

test("bundleFromRequiredFiles validates data.csv headers", () => {
  const badData = VALID_DATA_CSV.replace("AmountUSD", "AmountUsd");
  assert.throws(
    () => bundleFromRequiredFiles(badData, VALID_OWNER_MAP, VALID_CATEGORY_MAP),
    /header mismatch/
  );
});

test("loadRequiredFiles uses public-profile files by default", async () => {
  const requested = [];
  const fetchImpl = async (path) => {
    requested.push(path);
    return {
      ok: true,
      text: async () => {
        if (path.endsWith("demo_extracted.csv")) return VALID_DATA_CSV;
        if (path.endsWith("tmp_public_data/current/owner_map.csv")) return VALID_OWNER_MAP;
        if (path.endsWith("tmp_public_data/current/details_to_categories_map.csv"))
          return VALID_CATEGORY_MAP;
        if (path.endsWith("tmp_public_data/current/category_segments_map.csv"))
          return VALID_CATEGORY_SEGMENTS_MAP;
        throw new Error(`Unexpected path in fetchImpl: ${path}`);
      },
    };
  };

  const files = await loadRequiredFiles("/mockups/", fetchImpl);
  assert.strictEqual(files.loadProfile, "public");
  assert.strictEqual(files.dataCsvFiles.length, 1);
  assert.ok(files.dataCsvFiles[0].name.includes("demo_extracted"));
  assert.ok(files.dataCsvFiles[0].content.includes("CardCompany"));
  assert.strictEqual(files.categorySegmentsCsvText, VALID_CATEGORY_SEGMENTS_MAP);
  assert.ok(!requested.includes("/mockups/data.csv"));
  assert.deepStrictEqual(requested, [
    "/mockups/tmp_public_data/current/owner_map.csv",
    "/mockups/tmp_public_data/current/details_to_categories_map.csv",
    "/mockups/tmp_public_data/current/category_segments_map.csv",
    "/mockups/tmp_public_data/current/demo_extracted.csv",
  ]);
});

test("loadRequiredFiles supports sensitive profile and loads every configured transaction CSV", async () => {
  const requested = [];
  const fetchImpl = async (path) => {
    requested.push(path);
    return {
      ok: true,
      text: async () => {
        if (path.endsWith("Santander_joined.csv")) return VALID_DATA_CSV;
        if (path.endsWith("MercadoPago_joined.csv")) return VALID_DATA_CSV;
        if (path.endsWith("owner_map.csv")) return VALID_OWNER_MAP;
        if (path.endsWith("details_to_categories_map.csv")) return VALID_CATEGORY_MAP;
        if (path.endsWith("category_segments_map.csv")) return VALID_CATEGORY_SEGMENTS_MAP;
        throw new Error(`Unexpected path in fetchImpl: ${path}`);
      },
    };
  };

  const files = await loadRequiredFiles("/mockups/", fetchImpl, { loadProfile: "sensitive" });
  assert.strictEqual(files.loadProfile, "sensitive");
  assert.strictEqual(files.dataCsvFiles.length, 2);
  assert.ok(files.dataCsvFiles[0].name.includes("Santander"));
  assert.ok(files.dataCsvFiles[1].name.includes("MercadoPago"));
  assert.strictEqual(files.categorySegmentsCsvText, VALID_CATEGORY_SEGMENTS_MAP);
  assert.deepStrictEqual(requested, [
    "/mockups/tmp_sensitive_data/current/owner_map.csv",
    "/mockups/tmp_sensitive_data/current/details_to_categories_map.csv",
    "/mockups/tmp_sensitive_data/current/category_segments_map.csv",
    "/mockups/tmp_sensitive_data/current/Santander_joined.csv",
    "/mockups/tmp_sensitive_data/current/MercadoPago_joined.csv",
  ]);
});

test("loadRequiredFiles fails when sensitive transaction CSVs are unavailable", async () => {
  const requested = [];
  const fetchImpl = async (path) => {
    requested.push(path);
    if (
      path.endsWith("Santander_joined.csv") ||
      path.endsWith(SENSITIVE_EXTRA_DATA_FILE) ||
      path.endsWith("VISA-PRISMA_joined.csv")
    ) {
      return { ok: false, status: 404, text: async () => "" };
    }
    return {
      ok: true,
      text: async () => {
        if (path.endsWith("owner_map.csv")) return VALID_OWNER_MAP;
        if (path.endsWith("details_to_categories_map.csv")) return VALID_CATEGORY_MAP;
        if (path.endsWith("category_segments_map.csv")) return VALID_CATEGORY_SEGMENTS_MAP;
        throw new Error(`Unexpected path in fetchImpl: ${path}`);
      },
    };
  };

  await assert.rejects(
    () => loadRequiredFiles("/mockups/", fetchImpl, { loadProfile: "sensitive" }),
    /Strict mockups mode \(sensitive\) requires/
  );
  assert.ok(!requested.includes("/mockups/data.csv"));
});

test("bundleFromDataFiles validates every source file", () => {
  const badData = VALID_DATA_CSV.replace("AmountUSD", "AmountUsd");
  assert.throws(
    () =>
      bundleFromDataFiles(
        [
          { name: "Santander_joined.csv", content: VALID_DATA_CSV },
          { name: "VISA-PRISMA_joined.csv", content: badData },
        ],
        VALID_OWNER_MAP,
        VALID_CATEGORY_MAP
      ),
    /header mismatch/
  );
});

test("loadMockupBundle prefers storage and skips network when data exists", async () => {
  const storage = {
    getAllCsvFiles: async () => [{ name: "data.csv", content: VALID_DATA_CSV }],
    getConfig: async () => ({
      data: {
        ownersByCardOwner: { Alice: "A" },
        categoryByDetail: {},
        ownersByCardNumber: {},
        categorySegmentByCategory: { Food: "essential" },
      },
    }),
    clearAll: async () => {
      throw new Error("should not clear");
    },
    putCsvFile: async () => {
      throw new Error("should not persist csv");
    },
    putConfig: async () => {
      throw new Error("should not persist config");
    },
  };

  const bundle = await loadMockupBundle({
    basePath: "/ignored/",
    storage,
    fetchImpl: async () => {
      throw new Error("network should not be used");
    },
  });

  assert.strictEqual(bundle.source, "storage");
  assert.strictEqual(bundle.csvFiles.length, 1);
  assert.strictEqual(bundle.mappingsObj.ownersByCardOwner.Alice, "A");
  assert.strictEqual(bundle.mappingsObj.categorySegmentByCategory.Food, "essential");
});

test("restoreBundleFromStorage returns an explicit empty bundle when config exists without CSV files", async () => {
  const storage = {
    getAllCsvFiles: async () => [],
    getConfig: async () => ({
      data: {
        ownersByCardOwner: {},
        categoryByDetail: { SUPERMARKET: "Food" },
        ownersByCardNumber: {},
        categorySegmentByCategory: { Food: "essential" },
      },
    }),
  };

  const bundle = await restoreBundleFromStorage(storage);

  assert.ok(bundle);
  assert.strictEqual(bundle.source, "storage");
  assert.deepStrictEqual(bundle.csvFiles, []);
  assert.strictEqual(bundle.mappingsObj.categoryByDetail.SUPERMARKET, "Food");
});

test("loadMockupBundle prefers explicit empty storage state and skips network", async () => {
  const storage = {
    getAllCsvFiles: async () => [],
    getConfig: async () => ({
      data: {
        ownersByCardOwner: {},
        categoryByDetail: {},
        ownersByCardNumber: {},
        categorySegmentByCategory: {},
      },
    }),
    clearAll: async () => {
      throw new Error("should not clear");
    },
    putCsvFile: async () => {
      throw new Error("should not persist csv");
    },
    putConfig: async () => {
      throw new Error("should not persist config");
    },
  };

  const bundle = await loadMockupBundle({
    basePath: "/ignored/",
    storage,
    fetchImpl: async () => {
      throw new Error("network should not be used");
    },
  });

  assert.strictEqual(bundle.source, "storage");
  assert.deepStrictEqual(bundle.csvFiles, []);
  assert.deepStrictEqual(bundle.mappingsObj.ownersByCardOwner, {});
});

test("loadMockupBundle falls back to files and persists bundle", async () => {
  let persistedCsv = 0;
  let persistedConfig = 0;
  const storage = {
    getAllCsvFiles: async () => [],
    getConfig: async () => null,
    clearAll: async () => {},
    putCsvFile: async () => {
      persistedCsv += 1;
    },
    putConfig: async () => {
      persistedConfig += 1;
    },
  };

  const fetchImpl = async (path) => ({
    ok: true,
    text: async () => {
      if (path.endsWith("demo_extracted.csv")) return VALID_DATA_CSV;
      if (path.endsWith("tmp_public_data/current/owner_map.csv")) return VALID_OWNER_MAP;
      if (path.endsWith("tmp_public_data/current/details_to_categories_map.csv"))
        return VALID_CATEGORY_MAP;
      if (path.endsWith("tmp_public_data/current/category_segments_map.csv"))
        return VALID_CATEGORY_SEGMENTS_MAP;
      throw new Error(`Unexpected path in fetchImpl: ${path}`);
    },
  });

  const bundle = await loadMockupBundle({
    basePath: "/mockups_lab/",
    storage,
    fetchImpl,
  });

  assert.strictEqual(bundle.source, "files");
  assert.strictEqual(bundle.csvFiles.length, 1);
  assert.ok(bundle.csvFiles[0].name.includes("demo_extracted"));
  assert.strictEqual(bundle.mappingsObj.categoryByDetail.SUPERMARKET, "Food");
  assert.deepStrictEqual(Object.fromEntries(Object.entries(bundle.mappingsObj.categorySegmentByCategory)), {
    food: "essential",
  });
  assert.strictEqual(persistedCsv, 1);
  assert.strictEqual(persistedConfig, 1);
});

test("computeBundle uses WASM scope and builds runtime snapshot", async () => {
  const computeResult = {
    Tables: [
      {
        TableID: "meta_summary",
        Rows: [["statement_month_max", "2025-02-01"]],
      },
      {
        TableID: "overview_by_statement_month",
        Rows: [["2025-02-01", "1000.00", "10.00"]],
      },
    ],
  };

  const scope = {
    computeFromCSV: (csvText, mappingsJSON) => {
      assert.ok(csvText.includes("CardCompany"));
      assert.ok(csvText.split("\n")[0].includes(";"));
      assert.ok(csvText.includes("2025-02-01"));
      assert.ok(mappingsJSON.includes("ownersByCardOwner"));
      return { ok: true, value: JSON.stringify(computeResult) };
    },
    exportTableCSVFromResult: () => ({ ok: true, value: "col1;col2\n1;2\n" }),
    demoCSV: () => ({ ok: true, value: "" }),
    demoMappingsJSON: () => ({ ok: true, value: "{}" }),
  };

  const bundle = bundleFromRequiredFiles(VALID_DATA_CSV, VALID_OWNER_MAP, VALID_CATEGORY_MAP);
  const computed = await computeBundle(bundle, { scope, wasmExecPath: "ignored", wasmPath: "ignored" });

  assert.strictEqual(computed.runtime.tableCount, 2);
  assert.strictEqual(computed.runtime.latestMonth, "2025-02");
});

test("exportTableCSV forwards compute result and table id", async () => {
  const scope = {
    computeFromCSV: () => ({ ok: true, value: JSON.stringify({ Tables: [] }) }),
    exportTableCSVFromResult: (_resultJson, tableID) => ({ ok: true, value: `table=${tableID}` }),
    demoCSV: () => ({ ok: true, value: "" }),
    demoMappingsJSON: () => ({ ok: true, value: "{}" }),
  };

  const csv = await exportTableCSV(
    { computeResult: { Tables: [] } },
    "spend_by_category",
    { scope, wasmExecPath: "ignored", wasmPath: "ignored" }
  );
  assert.strictEqual(csv, "table=spend_by_category");
});

test("workspace artifact helpers round-trip", () => {
  const bundle = {
    csvFiles: [{ name: "data.csv", content: VALID_DATA_CSV }],
    mappingsObj: {
      ownersByCardOwner: { Alice: "Alice" },
      ownersByCardNumber: {},
      categoryByDetail: {},
      categorySegmentByCategory: { Food: "essential" },
    },
  };

  const workspace = exportWorkspaceArtifact(bundle);
  const roundTrip = parseWorkspaceArtifact(JSON.stringify(workspace));
  assert.strictEqual(roundTrip.csvFiles.length, 1);
  assert.strictEqual(roundTrip.config.ownersByCardOwner.Alice, "Alice");
  assert.strictEqual(roundTrip.config.categorySegmentByCategory.Food, "essential");
});
