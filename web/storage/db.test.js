import "fake-indexeddb/auto";
import test from "node:test";
import assert from "node:assert/strict";
import { createStorage } from "./db.js";

function uniqueDbName() {
  return `test_finance_db_${Date.now()}_${Math.random().toString(36).slice(2)}`;
}

test("putCsvFile and getAllCsvFiles: round-trip", async () => {
  // Given: fresh storage with unique db name
  const storage = createStorage(uniqueDbName());

  // When: put one CSV
  await storage.putCsvFile({ name: "extracted.csv", content: "H1;H2\n1;2" });

  // Then: getAll returns it
  const files = await storage.getAllCsvFiles();
  assert.equal(files.length, 1);
  assert.equal(files[0].name, "extracted.csv");
  assert.equal(files[0].content, "H1;H2\n1;2");
  assert.ok(typeof files[0].id === "string");
  assert.ok(typeof files[0].createdAt === "number");
  assert.ok(typeof files[0].updatedAt === "number");
});

test("putCsvFile: multiple files", async () => {
  const storage = createStorage(uniqueDbName());

  await storage.putCsvFile({ name: "a.csv", content: "A;B\n1;2" });
  await storage.putCsvFile({ name: "b.csv", content: "A;B\n3;4" });

  const files = await storage.getAllCsvFiles();
  assert.equal(files.length, 2);
  const names = files.map((f) => f.name).sort();
  assert.deepEqual(names, ["a.csv", "b.csv"]);
});

test("putConfig and getConfig: round-trip", async () => {
  const storage = createStorage(uniqueDbName());

  const mappings = {
    ownersByCardOwner: { A: "X" },
    categoryByDetail: {},
    categorySegmentByCategory: { Food: "essential" },
  };
  await storage.putConfig({ data: mappings });

  const config = await storage.getConfig();
  assert.ok(config !== null);
  assert.equal(config.schemaVersion, 1);
  assert.deepEqual(config.data, mappings);
  assert.ok(typeof config.createdAt === "number");
});

test("getConfig: returns null when empty", async () => {
  const storage = createStorage(uniqueDbName());

  const config = await storage.getConfig();
  assert.equal(config, null);
});

test("deleteCsvFile: removes by id", async () => {
  const storage = createStorage(uniqueDbName());

  const id = await storage.putCsvFile({ name: "x.csv", content: "data" });
  assert.equal((await storage.getAllCsvFiles()).length, 1);

  await storage.deleteCsvFile(id);
  assert.equal((await storage.getAllCsvFiles()).length, 0);
});

test("clearAll: empties both stores", async () => {
  const storage = createStorage(uniqueDbName());

  await storage.putCsvFile({ name: "a.csv", content: "x" });
  await storage.putConfig({ data: {} });

  await storage.clearAll();

  assert.equal((await storage.getAllCsvFiles()).length, 0);
  assert.equal(await storage.getConfig(), null);
});

test("list and delete: put multiple, list, delete one, list again", async () => {
  // Given: storage with multiple CSVs
  const storage = createStorage(uniqueDbName());
  const idA = await storage.putCsvFile({ name: "a.csv", content: "A" });
  const idB = await storage.putCsvFile({ name: "b.csv", content: "B" });
  const idC = await storage.putCsvFile({ name: "c.csv", content: "C" });

  // When: list all
  let files = await storage.getAllCsvFiles();
  assert.equal(files.length, 3);
  const names = files.map((f) => f.name).sort();
  assert.deepEqual(names, ["a.csv", "b.csv", "c.csv"]);

  // When: delete b by id
  await storage.deleteCsvFile(idB);

  // Then: list has a and c only
  files = await storage.getAllCsvFiles();
  assert.equal(files.length, 2);
  assert.deepEqual(
    files.map((f) => f.name).sort(),
    ["a.csv", "c.csv"]
  );
});
