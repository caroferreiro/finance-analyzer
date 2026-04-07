import test from "node:test";
import assert from "node:assert";

import {
  DEFAULT_REQUIRED_WASM_FUNCTIONS,
  getMissingWasmFunctions,
  okValue,
  ensureWasmRuntime,
  computeResultJSONFromCsv,
  computeResultFromCsv,
  exportTableCSVFromComputeResult,
  loadDemoBundle,
} from "./wasmRuntime.js";

test("getMissingWasmFunctions reports absent wasm exports", () => {
  const scope = {
    computeFromCSV: () => {},
    exportTableCSVFromResult: () => {},
  };
  const missing = getMissingWasmFunctions(scope, DEFAULT_REQUIRED_WASM_FUNCTIONS);
  assert.deepStrictEqual(missing, ["demoCSV", "demoMappingsJSON"]);
});

test("okValue unwraps success and throws on errors", () => {
  assert.strictEqual(okValue({ ok: true, value: "abc" }), "abc");
  assert.throws(() => okValue({ ok: false, error: "bad" }), /bad/);
  assert.throws(() => okValue(null), /unknown wasm error/);
});

test("ensureWasmRuntime bootstraps functions when missing", async () => {
  const scope = {
    WebAssembly: {
      instantiate: async () => ({ instance: { wasm: true } }),
    },
  };

  let fetchCalls = 0;
  const fetchImpl = async () => {
    fetchCalls += 1;
    return { ok: true, status: 200, arrayBuffer: async () => new ArrayBuffer(8) };
  };

  scope.Go = function Go() {
    this.importObject = {};
    this.run = () => {
      scope.computeFromCSV = () => ({ ok: true, value: JSON.stringify({ Tables: [] }) });
      scope.exportTableCSVFromResult = () => ({ ok: true, value: "csv" });
      scope.demoCSV = () => ({ ok: true, value: "demo" });
      scope.demoMappingsJSON = () => ({ ok: true, value: "{}" });
    };
  };

  await ensureWasmRuntime({
    scope,
    wasmExecPath: "./wasm_exec.js",
    wasmPath: "./finance.wasm",
    fetchImpl,
    instantiateStreaming: async () => ({ instance: {} }),
  });

  assert.strictEqual(fetchCalls, 1);
  assert.strictEqual(typeof scope.computeFromCSV, "function");
  assert.strictEqual(typeof scope.exportTableCSVFromResult, "function");
});

test("compute and export helpers use existing scope exports", async () => {
  const scope = {
    computeFromCSV: (csvText, mappingsJSON) => ({
      ok: true,
      value: JSON.stringify({ Tables: [{ TableID: csvText + mappingsJSON }] }),
    }),
    exportTableCSVFromResult: (resultJson, tableID) => ({
      ok: true,
      value: `${tableID}:${JSON.parse(resultJson).Tables.length}`,
    }),
    demoCSV: () => ({ ok: true, value: "demo-csv" }),
    demoMappingsJSON: () => ({ ok: true, value: "{}" }),
  };

  const resultJSON = await computeResultJSONFromCsv("csv", "{}", { scope });
  assert.ok(resultJSON.includes("Tables"));

  const result = await computeResultFromCsv("csv", "{}", { scope });
  assert.strictEqual(result.Tables.length, 1);

  const exported = await exportTableCSVFromComputeResult(result, "spend_by_owner", { scope });
  assert.strictEqual(exported, "spend_by_owner:1");

  const demo = await loadDemoBundle({ scope });
  assert.strictEqual(demo.csvText, "demo-csv");
  assert.strictEqual(demo.mappingsJSON, "{}");
});
