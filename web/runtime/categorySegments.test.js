import test from "node:test";
import assert from "node:assert/strict";

import {
  DISCRETIONARY_CATEGORY_SEGMENT,
  ESSENTIAL_CATEGORY_SEGMENT,
  UNCLASSIFIED_CATEGORY_SEGMENT,
  categorySegmentFromMappedCategory,
  normalizeCategoryKey,
  normalizeCategorySegmentByCategory,
  normalizeCategorySegmentKey,
} from "./categorySegments.js";

test("normalizeCategoryKey lowercases and collapses whitespace", () => {
  assert.equal(normalizeCategoryKey("  Public   Transport "), "public transport");
  assert.equal(normalizeCategoryKey("Take\tout"), "take out");
});

test("normalizeCategorySegmentKey accepts only supported segments", () => {
  assert.equal(normalizeCategorySegmentKey(" Essential "), ESSENTIAL_CATEGORY_SEGMENT);
  assert.equal(normalizeCategorySegmentKey("discretionary"), DISCRETIONARY_CATEGORY_SEGMENT);
  assert.equal(normalizeCategorySegmentKey("unclassified"), "");
  assert.equal(normalizeCategorySegmentKey(""), "");
});

test("normalizeCategorySegmentByCategory normalizes keys and skips invalid rows", () => {
  var lookup = normalizeCategorySegmentByCategory({
    " Public Transport ": "essential",
    "Private   Transport": " discretionary ",
    Unknown: "unclassified",
    "": "essential",
  });

  assert.deepEqual(
    Object.fromEntries(Object.entries(lookup)),
    {
      "public transport": ESSENTIAL_CATEGORY_SEGMENT,
      "private transport": DISCRETIONARY_CATEGORY_SEGMENT,
    },
  );
});

test("categorySegmentFromMappedCategory returns unclassified when category has no explicit segment", () => {
  assert.equal(
    categorySegmentFromMappedCategory("Transport"),
    UNCLASSIFIED_CATEGORY_SEGMENT
  );
});

test("categorySegmentFromMappedCategory treats blank, unknown, and adjustments as unclassified", () => {
  assert.equal(categorySegmentFromMappedCategory(""), UNCLASSIFIED_CATEGORY_SEGMENT);
  assert.equal(categorySegmentFromMappedCategory("?"), UNCLASSIFIED_CATEGORY_SEGMENT);
  assert.equal(categorySegmentFromMappedCategory("Uncategorized"), UNCLASSIFIED_CATEGORY_SEGMENT);
  assert.equal(categorySegmentFromMappedCategory("Adjustments"), UNCLASSIFIED_CATEGORY_SEGMENT);
});

test("custom category segment lookup supports public/private transport split", () => {
  var lookup = normalizeCategorySegmentByCategory({
    "Public Transport": "essential",
    "Private Transport": "discretionary",
  });

  assert.equal(
    categorySegmentFromMappedCategory(" public transport ", lookup),
    ESSENTIAL_CATEGORY_SEGMENT
  );
  assert.equal(
    categorySegmentFromMappedCategory("PRIVATE   TRANSPORT", lookup),
    DISCRETIONARY_CATEGORY_SEGMENT
  );
});
