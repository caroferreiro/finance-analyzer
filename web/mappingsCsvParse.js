import {
  normalizeCategoryKey,
  normalizeCategorySegmentKey,
} from "./runtime/categorySegments.js";

/**
 * Parses a two-column mappings CSV into a key→value object.
 *
 * Format (same for all mapping types):
 * - Delimiter: semicolon (;) — consistent with extracted CSV; supports values with commas.
 * - First row: header (e.g. card_owner_label;canonical_owner) — skipped.
 * - Data rows: key;value — col0 = source key, col1 = target value. Trimmed. Empty rows skipped.
 * - Duplicate keys: last wins.
 *
 * @param {string} text - Raw CSV text
 * @returns {{ [key: string]: string }} Object mapping source keys to target values
 */
export function parseMappingsCsv(text) {
  const result = {};
  const lines = text.trim().split("\n").filter((l) => l.trim());
  if (lines.length < 2) return result;
  for (const line of lines.slice(1)) {
    const parts = line.split(";").map((s) => s.trim());
    if (parts.length >= 2 && parts[0]) {
      result[parts[0]] = parts[1] ?? "";
    }
  }
  return result;
}

function splitCsvRow(line) {
  return String(line || "")
    .split(";")
    .map((cell) => cell.trim());
}

function assertCategorySegmentsHeader(headerLine) {
  const header = splitCsvRow(String(headerLine || "").replace(/^\uFEFF/, ""));
  if (header.length < 2 || header[0].toLowerCase() !== "category" || header[1].toLowerCase() !== "segment") {
    throw new Error("Category segments CSV header must be exactly: category;segment");
  }
  for (let i = 2; i < header.length; i++) {
    if (header[i] !== "") {
      throw new Error("Category segments CSV must have exactly two columns: category;segment");
    }
  }
}

export function parseCategorySegmentsCsv(text) {
  const result = Object.create(null);
  const lines = String(text || "")
    .split("\n")
    .map((line) => line.replace(/\r$/, ""))
    .filter((line) => line.trim() !== "");

  if (lines.length === 0) {
    return result;
  }

  assertCategorySegmentsHeader(lines[0]);
  if (lines.length === 1) {
    return result;
  }

  for (const line of lines.slice(1)) {
    const parts = splitCsvRow(line);
    const rawCategory = parts[0] || "";
    const rawSegment = parts[1] || "";
    const categoryKey = normalizeCategoryKey(rawCategory);
    const segmentKey = normalizeCategorySegmentKey(rawSegment);

    for (let i = 2; i < parts.length; i++) {
      if (parts[i] !== "") {
        throw new Error("Category segments CSV rows must have exactly two columns.");
      }
    }
    if (!categoryKey && !rawSegment) {
      continue;
    }
    if (!categoryKey) {
      throw new Error("Category segments CSV contains a row with an empty category.");
    }
    if (!segmentKey) {
      throw new Error(
        'Category segments CSV contains unsupported segment "' + rawSegment + '" for category "' + rawCategory + '".'
      );
    }

    result[categoryKey] = segmentKey;
  }

  return result;
}

export function serializeCategorySegmentsCsv(segmentByCategory) {
  const rows = [["category", "segment"]];
  const entries = Object.entries(segmentByCategory || {})
    .map(([category, segment]) => [String(category || "").trim(), normalizeCategorySegmentKey(segment)])
    .filter(([category, segment]) => category && segment)
    .sort((a, b) => a[0].localeCompare(b[0], undefined, { sensitivity: "base" }));

  for (const [category, segment] of entries) {
    rows.push([category, segment]);
  }

  return `${rows.map((row) => row.join(";")).join("\n")}\n`;
}
