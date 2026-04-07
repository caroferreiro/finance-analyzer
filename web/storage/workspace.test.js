import test from "node:test";
import assert from "node:assert/strict";
import { buildWorkspace, parseWorkspace } from "./workspace.js";

test("buildWorkspace and parseWorkspace: round-trip", () => {
  const csvFiles = [
    { name: "a.csv", content: "H1;H2\n1;2" },
    { name: "b.csv", content: "H1;H2\n3;4" },
  ];
  const mappingsObj = {
    ownersByCardOwner: { A: "X" },
    categoryByDetail: { D: "C" },
    categorySegmentByCategory: { Food: "essential" },
  };
  const workspace = buildWorkspace(csvFiles, mappingsObj);
  const json = JSON.stringify(workspace);
  const { csvFiles: parsedCsv, config } = parseWorkspace(json);
  assert.equal(parsedCsv.length, 2);
  assert.equal(parsedCsv[0].name, "a.csv");
  assert.equal(parsedCsv[0].content, "H1;H2\n1;2");
  assert.equal(parsedCsv[1].name, "b.csv");
  assert.equal(parsedCsv[1].content, "H1;H2\n3;4");
  assert.deepEqual(config, mappingsObj);
});

test("parseWorkspace: empty workspace", () => {
  const { csvFiles, config } = parseWorkspace('{"csvFiles":[],"config":{}}');
  assert.equal(csvFiles.length, 0);
  assert.deepEqual(config, {});
});

test("parseWorkspace: invalid JSON throws", () => {
  assert.throws(() => parseWorkspace("not json"), Error);
});

test("parseWorkspace: missing csvFiles uses empty array", () => {
  const { csvFiles, config } = parseWorkspace('{"config":{}}');
  assert.equal(csvFiles.length, 0);
  assert.deepEqual(config, {});
});
