import test from "node:test";
import assert from "node:assert/strict";
import { combineLoadedCsvFiles } from "./csvCombine.js";

test("combineLoadedCsvFiles: three separate entries produce header once + data from all", () => {
  // T13.X: multiple CSVs are stored/combined as separate entries (header once, data from all).
  const entries = [
    { name: "a.csv", content: "H1;H2\n1;2" },
    { name: "b.csv", content: "H1;H2\n3;4" },
    { name: "c.csv", content: "H1;H2\n5;6" },
  ];
  const result = combineLoadedCsvFiles(entries);
  assert.equal(result, "H1;H2\n1;2\n3;4\n5;6\n");
});

test("combineLoadedCsvFiles: empty array returns empty string", () => {
  assert.equal(combineLoadedCsvFiles([]), "");
});

test("combineLoadedCsvFiles: single file returns header and data", () => {
  const entries = [{ name: "only.csv", content: "A;B\n1;2" }];
  assert.equal(combineLoadedCsvFiles(entries), "A;B\n1;2\n");
});

test("combineLoadedCsvFiles: file with only header returns header and newline", () => {
  const entries = [{ name: "header-only.csv", content: "H1;H2" }];
  assert.equal(combineLoadedCsvFiles(entries), "H1;H2\n");
});

test("combineLoadedCsvFiles: skips empty content file, uses next for header", () => {
  const entries = [
    { name: "empty.csv", content: "" },
    { name: "b.csv", content: "H1;H2\n3;4" },
  ];
  assert.equal(combineLoadedCsvFiles(entries), "H1;H2\n3;4\n");
});
