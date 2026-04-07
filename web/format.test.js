import test from "node:test";
import assert from "node:assert";
import { humanReadableAmount, isMoneyColumn } from "./format.js";

test("humanReadableAmount - under 1000", () => {
  assert.strictEqual(humanReadableAmount("20"), "20");
  assert.strictEqual(humanReadableAmount("123.45"), "123.45");
  assert.strictEqual(humanReadableAmount("0"), "0");
  assert.strictEqual(humanReadableAmount("999"), "999");
  assert.strictEqual(humanReadableAmount("999.99"), "999.99");
});

test("humanReadableAmount - thousands (k)", () => {
  assert.strictEqual(humanReadableAmount("2000"), "2k");
  assert.strictEqual(humanReadableAmount("20000"), "20k");
  assert.strictEqual(humanReadableAmount("200000"), "200k");
  assert.strictEqual(humanReadableAmount("20277.00"), "20.28k");
  assert.strictEqual(humanReadableAmount("1234567"), "1.23M");
});

test("humanReadableAmount - millions (M)", () => {
  assert.strictEqual(humanReadableAmount("2000000"), "2M");
  assert.strictEqual(humanReadableAmount("1234567"), "1.23M");
  assert.strictEqual(humanReadableAmount("1000000"), "1M");
});

test("humanReadableAmount - negatives", () => {
  assert.strictEqual(humanReadableAmount("-20"), "-20");
  assert.strictEqual(humanReadableAmount("-2000"), "-2k");
  assert.strictEqual(humanReadableAmount("-1234567"), "-1.23M");
});

test("humanReadableAmount - edge cases", () => {
  assert.strictEqual(humanReadableAmount(""), "");
  assert.strictEqual(humanReadableAmount("  "), "");
  assert.strictEqual(humanReadableAmount("abc"), "abc");
  assert.strictEqual(humanReadableAmount("1,234.56"), "1.23k");
});

test("isMoneyColumn", () => {
  assert.strictEqual(isMoneyColumn("money_ars"), true);
  assert.strictEqual(isMoneyColumn("money_usd"), true);
  assert.strictEqual(isMoneyColumn("string"), false);
  assert.strictEqual(isMoneyColumn("number"), false);
  assert.strictEqual(isMoneyColumn("date"), false);
  assert.strictEqual(isMoneyColumn(""), false);
  assert.strictEqual(isMoneyColumn(undefined), false);
});
