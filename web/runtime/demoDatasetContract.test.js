import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const REPO_ROOT = path.resolve(__dirname, "..", "..");
const DEMO_MAPPINGS_PATH = path.join(REPO_ROOT, "pkg", "demo_dataset", "mappings.v1.json");

test("demo dataset mappings include categorySegmentByCategory for embedded demo flow", async () => {
  const raw = await fs.readFile(DEMO_MAPPINGS_PATH, "utf8");
  const parsed = JSON.parse(raw);

  assert.ok(parsed && typeof parsed === "object");
  assert.deepEqual(parsed.categorySegmentByCategory, {
    Groceries: "essential",
    Subscriptions: "discretionary",
    Transport: "essential",
    Electronics: "discretionary",
    Entertainment: "discretionary",
  });
});
