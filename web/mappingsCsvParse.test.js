import test from "node:test";
import assert from "node:assert/strict";
import {
  parseCategorySegmentsCsv,
  parseMappingsCsv,
  serializeCategorySegmentsCsv,
} from "./mappingsCsvParse.js";

test("parseMappingsCsv: ownersByCardOwner format", () => {
  const csv = "card_owner_label;canonical_owner\nVISA JOHN;OwnerA\nMASTERCARD JANE;OwnerB";
  const result = parseMappingsCsv(csv);
  assert.deepEqual(result, { "VISA JOHN": "OwnerA", "MASTERCARD JANE": "OwnerB" });
});

test("parseMappingsCsv: ownersByCardNumber format", () => {
  const csv = "card_number;canonical_owner\n****1234;OwnerA\n****5678;OwnerB";
  const result = parseMappingsCsv(csv);
  assert.deepEqual(result, { "****1234": "OwnerA", "****5678": "OwnerB" });
});

test("parseMappingsCsv: categoryByDetail format", () => {
  const csv = "detail;category\nSUPERMARKET DEMO;Groceries\nSTREAMING DEMO;Entertainment";
  const result = parseMappingsCsv(csv);
  assert.deepEqual(result, {
    "SUPERMARKET DEMO": "Groceries",
    "STREAMING DEMO": "Entertainment",
  });
});

test("parseMappingsCsv: empty or header-only returns empty object", () => {
  assert.deepEqual(parseMappingsCsv(""), {});
  assert.deepEqual(parseMappingsCsv("header;only"), {});
});

test("parseMappingsCsv: duplicate keys last wins", () => {
  const csv = "key;value\nA;1\nA;2";
  const result = parseMappingsCsv(csv);
  assert.deepEqual(result, { A: "2" });
});

test("parseMappingsCsv: skips empty rows", () => {
  const csv = "k;v\nA;1\n\nB;2\n";
  const result = parseMappingsCsv(csv);
  assert.deepEqual(result, { A: "1", B: "2" });
});

test("parseMappingsCsv: skips rows with empty key", () => {
  const csv = "k;v\n;skip\nA;1";
  const result = parseMappingsCsv(csv);
  assert.deepEqual(result, { A: "1" });
});

test("parseCategorySegmentsCsv: parses category->segment rows", () => {
  const csv = "category;segment\nPublic Transport;essential\nPrivate Transport;discretionary";
  const result = parseCategorySegmentsCsv(csv);
  assert.deepEqual(Object.fromEntries(Object.entries(result)), {
    "public transport": "essential",
    "private transport": "discretionary",
  });
});

test("parseCategorySegmentsCsv: duplicate categories last win after normalization", () => {
  const csv = "category;segment\nPublic Transport;essential\n public   transport ;discretionary";
  const result = parseCategorySegmentsCsv(csv);
  assert.deepEqual(Object.fromEntries(Object.entries(result)), {
    "public transport": "discretionary",
  });
});

test("parseCategorySegmentsCsv: empty or header-only returns empty object", () => {
  assert.deepEqual(Object.fromEntries(Object.entries(parseCategorySegmentsCsv(""))), {});
  assert.deepEqual(
    Object.fromEntries(Object.entries(parseCategorySegmentsCsv("category;segment"))),
    {}
  );
});

test("parseCategorySegmentsCsv: invalid header throws", () => {
  assert.throws(
    () => parseCategorySegmentsCsv("detail;segment\nFood;essential"),
    /header must be exactly: category;segment/
  );
});

test("parseCategorySegmentsCsv: invalid segment throws", () => {
  assert.throws(
    () => parseCategorySegmentsCsv("category;segment\nFuel;unclassified"),
    /unsupported segment/
  );
});

test("parseCategorySegmentsCsv: empty category throws", () => {
  assert.throws(
    () => parseCategorySegmentsCsv("category;segment\n;essential"),
    /empty category/
  );
});

test("serializeCategorySegmentsCsv: emits sorted two-column csv", () => {
  const csv = serializeCategorySegmentsCsv({
    "Private Transport": "discretionary",
    "Public Transport": "essential",
  });

  assert.equal(
    csv,
    "category;segment\nPrivate Transport;discretionary\nPublic Transport;essential\n"
  );
});

test("serializeCategorySegmentsCsv: skips invalid rows and keeps header", () => {
  const csv = serializeCategorySegmentsCsv({
    "": "essential",
    Fuel: "unclassified",
  });

  assert.equal(csv, "category;segment\n");
});
