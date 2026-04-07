import test from "node:test";
import assert from "node:assert/strict";
import {
  formatStorageUsage,
  getPersistentStorageStatus,
  getStorageWarning,
  isQuotaExceededError,
} from "./utils.js";

test("getPersistentStorageStatus: granted", () => {
  assert.equal(getPersistentStorageStatus(true), "Persistent storage enabled.");
});

test("getPersistentStorageStatus: not granted", () => {
  assert.equal(
    getPersistentStorageStatus(false),
    "May be cleared under storage pressure; export workspace recommended."
  );
});

test("formatStorageUsage: with quota", () => {
  assert.equal(formatStorageUsage(13_107_200, 524_288_000), "12.50 MB / 500.00 MB used");
});

test("formatStorageUsage: quota undefined", () => {
  assert.equal(formatStorageUsage(13_107_200, undefined), "12.50 MB used");
});

test("formatStorageUsage: zero usage", () => {
  assert.equal(formatStorageUsage(0, 524_288_000), "0.00 MB / 500.00 MB used");
});

test("isQuotaExceededError: QuotaExceededError name", () => {
  assert.equal(isQuotaExceededError({ name: "QuotaExceededError" }), true);
});

test("isQuotaExceededError: code 22", () => {
  assert.equal(isQuotaExceededError({ code: 22 }), true);
});

test("isQuotaExceededError: generic error returns false", () => {
  assert.equal(isQuotaExceededError(new Error("foo")), false);
});

test("getStorageWarning: QuotaExceededError returns storage full message", () => {
  const err = { name: "QuotaExceededError" };
  const msg = getStorageWarning(err);
  assert.ok(msg.includes("Storage full"));
  assert.ok(msg.includes("Export to save a copy"));
});

test("getStorageWarning: generic error returns storage unavailable message", () => {
  const msg = getStorageWarning(new Error("UnknownError"));
  assert.ok(msg.includes("Storage unavailable"));
  assert.ok(msg.includes("will not persist"));
});

test("QuotaExceededError: catch and getStorageWarning shows message, app does not crash", async () => {
  const quotaExceededError = new DOMException("QuotaExceeded", "QuotaExceededError");
  const mockStorage = {
    putConfig: async () => {
      throw quotaExceededError;
    },
  };
  let warning = null;
  try {
    await mockStorage.putConfig({ data: {} });
  } catch (err) {
    warning = getStorageWarning(err);
  }
  assert.ok(warning !== null);
  assert.ok(warning.includes("Storage full"));
  assert.ok(warning.includes("Export to save a copy"));
});
