import { combineLoadedCsvFiles } from "../csvCombine.js";
import { parseCategorySegmentsCsv } from "../mappingsCsvParse.js";
import { createStorage } from "../storage/db.js";
import { buildWorkspace, parseWorkspace } from "../storage/workspace.js";
import { computeResultFromCsv, exportTableCSVFromComputeResult } from "./wasmRuntime.js";
import { buildRuntimeSnapshot } from "./computeTables.js";

const DATA_HEADERS = [
  "Bank",
  "CardCompany",
  "CloseDate",
  "ExpirationDate",
  "TotalARS",
  "TotalUSD",
  "CardNumber",
  "CardOwner",
  "CardTotalARS",
  "CardTotalUSD",
  "MovementType",
  "OriginalDate",
  "ReceiptNumber",
  "Detail",
  "CurrentInstallment",
  "TotalInstallments",
  "AmountARS",
  "AmountUSD",
];
const OWNER_HEADERS = ["RawOwner", "OwnerNormalized"];
const CATEGORY_HEADERS = ["Detail", "Category"];

const REQUIRED_MAPPING_FILES = Object.freeze({
  owner: "owner_map.csv",
  category: "details_to_categories_map.csv",
});
const PREFERRED_PURE_DATA_FILES = Object.freeze([
  "tmp_sensitive_data/current/Santander_joined.csv",
  "tmp_sensitive_data/current/MercadoPago_joined.csv",
]);
const SENSITIVE_REQUIRED_FILES = Object.freeze({
  owner: "tmp_sensitive_data/current/owner_map.csv",
  category: "tmp_sensitive_data/current/details_to_categories_map.csv",
  categorySegments: "tmp_sensitive_data/current/category_segments_map.csv",
});
const PUBLIC_REQUIRED_FILES = Object.freeze({
  owner: "tmp_public_data/current/owner_map.csv",
  category: "tmp_public_data/current/details_to_categories_map.csv",
  categorySegments: "tmp_public_data/current/category_segments_map.csv",
});
const PUBLIC_DATA_FILES = Object.freeze(["tmp_public_data/current/demo_extracted.csv"]);
const LOAD_PROFILE_CONTRACTS = Object.freeze({
  public: Object.freeze({
    mappingFiles: PUBLIC_REQUIRED_FILES,
    dataFiles: PUBLIC_DATA_FILES,
  }),
  sensitive: Object.freeze({
    mappingFiles: SENSITIVE_REQUIRED_FILES,
    dataFiles: PREFERRED_PURE_DATA_FILES,
  }),
});
const DEFAULT_LOAD_PROFILE = "public";
const MOCKUPS_STORAGE_DB_NAME = "finance_dashboard_mockups_db";

function defaultStorage() {
  return createStorage(MOCKUPS_STORAGE_DB_NAME);
}

function trimCell(value) {
  return String(value == null ? "" : value).trim();
}

function stripBom(value) {
  return value.replace(/^\uFEFF/, "");
}

function isEmptyRow(cells) {
  return cells.every((value) => trimCell(value) === "");
}

function assertHeaders(actual, expected, fileName) {
  if (actual.length !== expected.length) {
    throw new Error(
      `${fileName} header length mismatch. Expected ${expected.length}, got ${actual.length}.`
    );
  }
  for (let i = 0; i < expected.length; i++) {
    if (actual[i] !== expected[i]) {
      throw new Error(
        `${fileName} header mismatch at column ${i + 1}. Expected "${expected[i]}", got "${actual[i]}".`
      );
    }
  }
}

function normalizeDetailKey(value) {
  return String(value || "")
    .normalize("NFKC")
    .trim()
    .replace(/\s+/g, " ")
    .toUpperCase();
}

export function normalizeBasePath(basePath) {
  const path = String(basePath || "./").trim();
  if (!path) {
    return "./";
  }
  return path.endsWith("/") ? path : `${path}/`;
}

export function resolveDataPath(basePath, fileName) {
  return `${normalizeBasePath(basePath)}${fileName}`;
}

function basename(path) {
  const value = String(path || "");
  const parts = value.split("/");
  return parts[parts.length - 1] || value;
}

function detectDelimiter(input) {
  const firstLine = String(input || "").split(/\r?\n/, 1)[0] || "";
  const semicolons = (firstLine.match(/;/g) || []).length;
  const commas = (firstLine.match(/,/g) || []).length;
  return semicolons > commas ? ";" : ",";
}

export async function fetchText(path, fetchImpl = fetch) {
  const response = await fetchImpl(path, { cache: "no-store" });
  if (!response.ok) {
    throw new Error(`${path} could not be fetched (HTTP ${response.status}).`);
  }
  return response.text();
}

export function parseCsvMatrix(input, delimiter = ",") {
  const text = String(input || "").replace(/\r\n/g, "\n").replace(/\r/g, "\n");
  const rows = [];
  let row = [];
  let field = "";
  let inQuotes = false;

  for (let i = 0; i < text.length; i++) {
    const ch = text[i];
    if (inQuotes) {
      if (ch === '"') {
        const next = text[i + 1];
        if (next === '"') {
          field += '"';
          i += 1;
        } else {
          inQuotes = false;
        }
      } else {
        field += ch;
      }
    } else if (ch === '"') {
      inQuotes = true;
    } else if (ch === delimiter) {
      row.push(field);
      field = "";
    } else if (ch === "\n") {
      row.push(field);
      rows.push(row);
      row = [];
      field = "";
    } else {
      field += ch;
    }
  }

  row.push(field);
  rows.push(row);

  while (rows.length && isEmptyRow(rows[rows.length - 1])) {
    rows.pop();
  }

  return rows;
}

function csvEscape(value, delimiter) {
  const raw = String(value == null ? "" : value);
  if (raw.includes(delimiter) || raw.includes('"') || raw.includes("\n") || raw.includes("\r")) {
    return `"${raw.replace(/"/g, '""')}"`;
  }
  return raw;
}

function matrixToDelimitedCsv(matrix, delimiter) {
  const rows = matrix.map((row) => row.map((value) => csvEscape(value, delimiter)).join(delimiter));
  return `${rows.join("\n")}\n`;
}

function normalizeDateForWasm(value) {
  const raw = String(value == null ? "" : value).trim();
  if (!raw) {
    return "";
  }
  const ddmmyyyy = /^(\d{1,2})\/(\d{1,2})\/(\d{4})$/.exec(raw);
  if (!ddmmyyyy) {
    return raw;
  }
  const day = ddmmyyyy[1].padStart(2, "0");
  const month = ddmmyyyy[2].padStart(2, "0");
  return `${ddmmyyyy[3]}-${month}-${day}`;
}

function normalizeDatesForWasmMatrix(matrix) {
  if (!Array.isArray(matrix) || matrix.length === 0) {
    return matrix;
  }
  const headers = matrix[0].map((value) => String(value || ""));
  const dateColumns = ["CloseDate", "ExpirationDate", "OriginalDate"]
    .map((name) => headers.indexOf(name))
    .filter((idx) => idx >= 0);
  if (dateColumns.length === 0) {
    return matrix;
  }

  const out = matrix.map((row) => row.slice());
  for (let r = 1; r < out.length; r++) {
    for (const colIdx of dateColumns) {
      if (colIdx < out[r].length) {
        out[r][colIdx] = normalizeDateForWasm(out[r][colIdx]);
      }
    }
  }
  return out;
}

function maybeConvertToSemicolonCsv(csvText) {
  const firstLine = String(csvText || "").split(/\r?\n/, 1)[0] || "";
  const looksCommaDelimited = firstLine.includes(",") && !firstLine.includes(";");
  if (!looksCommaDelimited) {
    return csvText;
  }
  const matrix = parseCsvMatrix(csvText);
  return matrixToDelimitedCsv(normalizeDatesForWasmMatrix(matrix), ";");
}

export function parseCsvAsObjects(text, expectedHeaders, fileName) {
  const matrix = parseCsvMatrix(text, detectDelimiter(text));
  if (!matrix.length) {
    throw new Error(`${fileName} is empty.`);
  }

  const headers = matrix[0].map(trimCell).map(stripBom);
  assertHeaders(headers, expectedHeaders, fileName);

  const rows = [];
  for (let i = 1; i < matrix.length; i++) {
    const raw = matrix[i].map(trimCell);
    if (isEmptyRow(raw)) {
      continue;
    }

    for (let extra = expectedHeaders.length; extra < raw.length; extra++) {
      if ((raw[extra] || "").trim() !== "") {
        throw new Error(`${fileName} has unexpected non-empty extra data at row ${i + 1}.`);
      }
    }

    const obj = {};
    for (let c = 0; c < expectedHeaders.length; c++) {
      obj[expectedHeaders[c]] = raw[c] || "";
    }
    rows.push(obj);
  }

  return rows;
}

export function buildMappingsFromRequiredCsv(ownerMapCsvText, categoryMapCsvText, categorySegmentsCsvText = "") {
  const ownerRows = parseCsvAsObjects(ownerMapCsvText, OWNER_HEADERS, REQUIRED_MAPPING_FILES.owner);
  const categoryRows = parseCsvAsObjects(
    categoryMapCsvText,
    CATEGORY_HEADERS,
    REQUIRED_MAPPING_FILES.category
  );
  const categorySegmentByCategory = parseCategorySegmentsCsv(categorySegmentsCsvText);

  const ownersByCardOwner = {};
  for (const row of ownerRows) {
    const raw = trimCell(row.RawOwner);
    const normalized = trimCell(row.OwnerNormalized);
    if (raw && normalized) {
      ownersByCardOwner[raw.toUpperCase()] = normalized;
    }
  }

  const categoryByDetail = {};
  const categoryByDetailPrefix = [];
  for (const row of categoryRows) {
    const raw = trimCell(row.Detail);
    const category = trimCell(row.Category);
    if (!raw || !category) continue;
    if (raw.endsWith("*")) {
      categoryByDetailPrefix.push({
        prefix: normalizeDetailKey(raw.slice(0, -1)),
        category,
      });
    } else {
      categoryByDetail[normalizeDetailKey(raw)] = category;
    }
  }
  // Longest prefix first so more specific patterns win.
  categoryByDetailPrefix.sort((a, b) => b.prefix.length - a.prefix.length);

  return {
    ownersByCardOwner,
    ownersByCardNumber: {},
    categoryByDetail,
    categoryByDetailPrefix,
    categorySegmentByCategory,
  };
}

function normalizeLoadProfile(loadProfile) {
  return String(loadProfile || "").trim().toLowerCase() === "sensitive"
    ? "sensitive"
    : DEFAULT_LOAD_PROFILE;
}

function loadProfileContract(loadProfile) {
  const normalized = normalizeLoadProfile(loadProfile);
  return {
    loadProfile: normalized,
    contract: LOAD_PROFILE_CONTRACTS[normalized],
  };
}

async function loadProfileDataFiles(basePath, fetchImpl, dataFiles, loadProfile) {
  const loaded = [];
  for (const relativePath of dataFiles || []) {
    const fullPath = resolveDataPath(basePath, relativePath);
    try {
      const csvText = await fetchText(fullPath, fetchImpl);
      parseCsvAsObjects(csvText, DATA_HEADERS, relativePath);
      loaded.push({
        name: basename(relativePath),
        content: csvText,
      });
    } catch (err) {
      const msg = err && err.message ? err.message : String(err);
      if (/HTTP 404/.test(msg)) {
        continue;
      }
      throw err;
    }
  }
  if (!loaded.length) {
    throw new Error(`No data CSV files configured for load profile "${loadProfile}".`);
  }
  return loaded;
}

export async function loadRequiredFiles(
  basePath = "./",
  fetchImpl = fetch,
  { loadProfile = DEFAULT_LOAD_PROFILE } = {}
) {
  const normalizedBase = normalizeBasePath(basePath);
  const { loadProfile: activeProfile, contract } = loadProfileContract(loadProfile);
  const mappingFiles = contract.mappingFiles || REQUIRED_MAPPING_FILES;
  const requiredDataFiles = Array.isArray(contract.dataFiles) ? contract.dataFiles : [];
  const [ownerMapCsvText, categoryMapCsvText, categorySegmentsCsvText] = await Promise.all([
    fetchText(resolveDataPath(normalizedBase, mappingFiles.owner), fetchImpl),
    fetchText(resolveDataPath(normalizedBase, mappingFiles.category), fetchImpl),
    mappingFiles.categorySegments
      ? fetchText(resolveDataPath(normalizedBase, mappingFiles.categorySegments), fetchImpl)
      : Promise.resolve(""),
  ]);

  let dataCsvFiles;
  try {
    dataCsvFiles = await loadProfileDataFiles(
      normalizedBase,
      fetchImpl,
      requiredDataFiles,
      activeProfile
    );
  } catch (err) {
    const reason = err && err.message ? err.message : String(err);
    const requiredList = requiredDataFiles.concat([mappingFiles.owner, mappingFiles.category]);
    throw new Error(
      `Strict mockups mode (${activeProfile}) requires: ` +
        requiredList.join(", ") +
        ". " +
        reason
    );
  }

  return {
    dataCsvText: String(dataCsvFiles[0]?.content || ""),
    dataCsvFiles,
    ownerMapCsvText,
    categoryMapCsvText,
    categorySegmentsCsvText,
    loadProfile: activeProfile,
  };
}

function toRuntimeBundle(csvFiles, mappingsObj, source) {
  const normalizedCsvFiles = csvFiles.map((file) => ({
    id: file.id,
    name: file.name,
    content: file.content,
    createdAt: file.createdAt,
  }));
  return {
    source,
    csvFiles: normalizedCsvFiles,
    mappingsObj: mappingsObj || {},
  };
}

export async function restoreBundleFromStorage(storage = defaultStorage()) {
  const [csvFiles, config] = await Promise.all([storage.getAllCsvFiles(), storage.getConfig()]);
  const hasStoredConfig = !!config;
  if (!Array.isArray(csvFiles) || csvFiles.length === 0) {
    if (!hasStoredConfig) {
      return null;
    }
    const mappingsObj = config?.data && typeof config.data === "object" ? config.data : {};
    return toRuntimeBundle([], mappingsObj, "storage");
  }

  if (!hasStoredConfig) {
    return null;
  }

  const mappingsObj = config?.data && typeof config.data === "object" ? config.data : {};
  return toRuntimeBundle(csvFiles, mappingsObj, "storage");
}

export async function persistBundleToStorage(bundle, storage = defaultStorage()) {
  await storage.clearAll();

  for (const entry of bundle.csvFiles || []) {
    await storage.putCsvFile({
      id: entry.id,
      name: entry.name,
      content: entry.content,
      createdAt: entry.createdAt,
    });
  }

  await storage.putConfig({ data: bundle.mappingsObj || {} });
}

export function bundleFromDataFiles(
  dataCsvFiles,
  ownerMapCsvText,
  categoryMapCsvText,
  categorySegmentsCsvText = ""
) {
  const normalizedDataFiles = Array.isArray(dataCsvFiles) ? dataCsvFiles : [];
  if (normalizedDataFiles.length === 0) {
    throw new Error("At least one data CSV file is required.");
  }
  for (const file of normalizedDataFiles) {
    parseCsvAsObjects(file.content, DATA_HEADERS, file.name || "data.csv");
  }
  const mappingsObj = buildMappingsFromRequiredCsv(
    ownerMapCsvText,
    categoryMapCsvText,
    categorySegmentsCsvText
  );
  return toRuntimeBundle(
    normalizedDataFiles.map((file) => ({
      name: file.name || "data.csv",
      content: file.content,
    })),
    mappingsObj,
    "files"
  );
}

export function bundleFromRequiredFiles(
  dataCsvText,
  ownerMapCsvText,
  categoryMapCsvText,
  categorySegmentsCsvText = ""
) {
  return bundleFromDataFiles(
    [
      {
        name: "data.csv",
        content: dataCsvText,
      },
    ],
    ownerMapCsvText,
    categoryMapCsvText,
    categorySegmentsCsvText
  );
}

export async function loadMockupBundle({
  basePath = "./",
  storage = defaultStorage(),
  preferStorage = true,
  fetchImpl = fetch,
  loadProfile = DEFAULT_LOAD_PROFILE,
} = {}) {
  if (preferStorage) {
    try {
      const stored = await restoreBundleFromStorage(storage);
      if (stored) {
        return stored;
      }
    } catch {
      // Ignore storage restore failures; fallback to required files.
    }
  }

  const files = await loadRequiredFiles(basePath, fetchImpl, {
    loadProfile,
  });
  const bundle = bundleFromDataFiles(
    files.dataCsvFiles,
    files.ownerMapCsvText,
    files.categoryMapCsvText,
    files.categorySegmentsCsvText
  );

  try {
    await persistBundleToStorage(bundle, storage);
  } catch {
    // Keep runtime functional even when IndexedDB is unavailable.
  }

  return bundle;
}

export async function computeBundle(bundle, {
  wasmExecPath,
  wasmPath,
  scope = globalThis,
} = {}) {
  const csvFiles = bundle.csvFiles || [];
  const csvText =
    csvFiles.length === 1
      ? String(csvFiles[0].content || "")
      : combineLoadedCsvFiles(csvFiles);
  const computeCsvText = maybeConvertToSemicolonCsv(csvText);
  const mappingsJSON = JSON.stringify(bundle.mappingsObj || {});
  const computeResult = await computeResultFromCsv(computeCsvText, mappingsJSON, {
    wasmExecPath,
    wasmPath,
    scope,
  });

  return {
    sourceCsvText: csvText,
    computeCsvText,
    mappingsJSON,
    computeResult,
    runtime: buildRuntimeSnapshot(computeResult, bundle.source),
  };
}

export async function exportTableCSV(bundleCompute, tableID, {
  wasmExecPath,
  wasmPath,
  scope = globalThis,
} = {}) {
  return exportTableCSVFromComputeResult(bundleCompute.computeResult, tableID, {
    wasmExecPath,
    wasmPath,
    scope,
  });
}

export function exportWorkspaceArtifact(bundle) {
  return buildWorkspace(bundle.csvFiles || [], bundle.mappingsObj || {});
}

export function parseWorkspaceArtifact(jsonText) {
  return parseWorkspace(jsonText);
}

export {
  REQUIRED_MAPPING_FILES,
  SENSITIVE_REQUIRED_FILES,
  PUBLIC_REQUIRED_FILES,
  PUBLIC_DATA_FILES,
  PREFERRED_PURE_DATA_FILES,
  LOAD_PROFILE_CONTRACTS,
  DEFAULT_LOAD_PROFILE,
  DATA_HEADERS,
  OWNER_HEADERS,
  CATEGORY_HEADERS,
};
