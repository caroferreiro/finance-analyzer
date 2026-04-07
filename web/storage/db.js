/**
 * IndexedDB storage for CSVs and configs.
 * Uses idb wrapper. See docs/INDEXEDDB_PLAN.md for schema conventions.
 *
 * @module storage/db
 */

import { openDB, deleteDB } from "../node_modules/idb/build/index.js";

const DB_VERSION = 1;
const CSV_FILES_STORE = "csv_files";
const CONFIGS_STORE = "configs";
const MAPPINGS_CONFIG_ID = "mappings-v1";

/**
 * Creates a storage instance bound to a database name.
 * Use default name for app; pass a test name for unit tests.
 *
 * @param {string} [dbName='finance_dashboard_db'] - Database name
 * @returns {object} Storage API
 */
export function createStorage(dbName = "finance_dashboard_db") {
  let dbPromise = null;

  function getDb() {
    if (!dbPromise) {
      dbPromise = openDB(dbName, DB_VERSION, {
        upgrade(db, _oldVersion, _newVersion, transaction) {
          if (!db.objectStoreNames.contains(CSV_FILES_STORE)) {
            db.createObjectStore(CSV_FILES_STORE, { keyPath: "id" });
          }
          if (!db.objectStoreNames.contains(CONFIGS_STORE)) {
            db.createObjectStore(CONFIGS_STORE, { keyPath: "id" });
          }
        },
      });
    }
    return dbPromise;
  }

  /**
   * Stores a CSV file. Generates id if not provided.
   *
   * @param {{ id?: string, name: string, content: string }} record
   * @returns {Promise<string>} The id used
   */
  async function putCsvFile(record) {
    const now = Math.floor(Date.now() / 1000);
    const id = record.id ?? `csv-${Date.now()}-${record.name}`;
    const full = {
      id,
      name: record.name,
      content: record.content,
      createdAt: record.createdAt ?? now,
      updatedAt: now,
    };
    const db = await getDb();
    await db.put(CSV_FILES_STORE, full);
    return id;
  }

  /**
   * Returns all CSV files from storage.
   *
   * @returns {Promise<Array<{ id: string, name: string, content: string, createdAt: number, updatedAt: number }>>}
   */
  async function getAllCsvFiles() {
    const db = await getDb();
    return db.getAll(CSV_FILES_STORE);
  }

  /**
   * Stores mappings config.
   *
   * @param {{ schemaVersion?: number, data: object }} record
   * @returns {Promise<void>}
   */
  async function putConfig(record) {
    const now = Math.floor(Date.now() / 1000);
    const full = {
      id: MAPPINGS_CONFIG_ID,
      schemaVersion: record.schemaVersion ?? 1,
      data: record.data,
      createdAt: record.createdAt ?? now,
    };
    const db = await getDb();
    await db.put(CONFIGS_STORE, full);
  }

  /**
   * Returns the mappings config, or null if not found.
   *
   * @returns {Promise<{ id: string, schemaVersion: number, data: object, createdAt: number } | null>}
   */
  async function getConfig() {
    const db = await getDb();
    const value = await db.get(CONFIGS_STORE, MAPPINGS_CONFIG_ID);
    return value ?? null;
  }

  /**
   * Deletes a CSV file by id.
   *
   * @param {string} id
   * @returns {Promise<void>}
   */
  async function deleteCsvFile(id) {
    const db = await getDb();
    await db.delete(CSV_FILES_STORE, id);
  }

  /**
   * Clears all data from both stores.
   *
   * @returns {Promise<void>}
   */
  async function clearAll() {
    const db = await getDb();
    const tx = db.transaction([CSV_FILES_STORE, CONFIGS_STORE], "readwrite");
    await tx.objectStore(CSV_FILES_STORE).clear();
    await tx.objectStore(CONFIGS_STORE).clear();
    await tx.done;
  }

  return {
    putCsvFile,
    getAllCsvFiles,
    putConfig,
    getConfig,
    deleteCsvFile,
    clearAll,
  };
}

/**
 * Deletes the database. Useful for tests.
 *
 * @param {string} dbName
 * @returns {Promise<void>}
 */
export async function deleteDatabase(dbName) {
  await deleteDB(dbName);
}
