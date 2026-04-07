#!/usr/bin/env node
import { readFileSync, writeFileSync } from "node:fs";
import { resolve, basename } from "node:path";

function parseCsv(text, separator = ",") {
  return text
    .trim()
    .split("\n")
    .slice(1)
    .filter((line) => line.trim())
    .map((line) => line.split(separator).map((col) => col.trim()));
}

function toObject(rows) {
  const obj = {};
  for (const [key, value] of rows) {
    if (key && value) obj[key] = value;
  }
  return obj;
}

const args = process.argv.slice(2);
if (args.includes("--help") || args.includes("-h")) {
  console.log(`Usage: node csvs-to-mappings-json.mjs [options]

Converts CSV mapping files into a single mappings JSON importable by Finance Analyzer.

Options:
  --owner <file>      owner_map.csv           (header: RawOwner,OwnerNormalized)
  --category <file>   details_to_categories.csv (header: Detail,Category)
  --segments <file>   category_segments.csv    (header: category;segment  — semicolon-separated)
  --out <file>        output path (default: mappings.json)
  -h, --help          show this help

At least one CSV must be provided.`);
  process.exit(0);
}

function flag(name) {
  const idx = args.indexOf(name);
  return idx !== -1 && idx + 1 < args.length ? args[idx + 1] : null;
}

const ownerPath = flag("--owner");
const categoryPath = flag("--category");
const segmentsPath = flag("--segments");
const outPath = flag("--out") || "mappings.json";

if (!ownerPath && !categoryPath && !segmentsPath) {
  console.error("Error: provide at least one of --owner, --category, or --segments. Use --help for usage.");
  process.exit(1);
}

const mappings = {
  ownersByCardOwner: {},
  ownersByCardNumber: {},
  categoryByDetail: {},
  categorySegmentByCategory: {},
};

if (ownerPath) {
  const text = readFileSync(resolve(ownerPath), "utf8");
  mappings.ownersByCardOwner = toObject(parseCsv(text, ","));
  console.log(`  owner rules: ${Object.keys(mappings.ownersByCardOwner).length} (from ${basename(ownerPath)})`);
}

if (categoryPath) {
  const text = readFileSync(resolve(categoryPath), "utf8");
  mappings.categoryByDetail = toObject(parseCsv(text, ","));
  console.log(`  category rules: ${Object.keys(mappings.categoryByDetail).length} (from ${basename(categoryPath)})`);
}

if (segmentsPath) {
  const text = readFileSync(resolve(segmentsPath), "utf8");
  mappings.categorySegmentByCategory = toObject(parseCsv(text, ";"));
  console.log(`  segment rules: ${Object.keys(mappings.categorySegmentByCategory).length} (from ${basename(segmentsPath)})`);
}

const dest = resolve(outPath);
writeFileSync(dest, JSON.stringify(mappings, null, 2) + "\n");
console.log(`\nWrote ${dest}`);
