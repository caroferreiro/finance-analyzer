import test from "node:test";
import assert from "node:assert";
import fs from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const WEB_ROOT = path.resolve(__dirname, "..");

test("non-mockup runtime modules do not import mockups_lab", async () => {
  const candidateFiles = [
    "csvCombine.js",
    "mappingsCsvParse.js",
    "runtime/computeTables.js",
    "runtime/mockupsRuntime.js",
    "runtime/wasmRuntime.js",
    "storage/db.js",
    "storage/utils.js",
    "storage/workspace.js",
  ];

  for (const relPath of candidateFiles) {
    const source = await fs.readFile(path.join(WEB_ROOT, relPath), "utf8");
    assert.equal(/from\s+["'][^"']*mockups_lab\//.test(source), false, relPath);
    assert.equal(/import\(\s*["'][^"']*mockups_lab\//.test(source), false, relPath);
  }
});
