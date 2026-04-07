import test from "node:test";
import assert from "node:assert/strict";

import {
  categorySegmentFromMappedCategory,
  normalizeCategorySegmentByCategory,
} from "./categorySegments.js";

function summarizeSyntheticSpend(rows, categoryByDetail, categorySegmentLookup) {
  const summary = {
    essential: { rows: 0, amountARS: 0 },
    discretionary: { rows: 0, amountARS: 0 },
    unclassified: { rows: 0, amountARS: 0 },
  };

  for (const row of rows) {
    if (row.movementType !== "CardMovement") {
      continue;
    }
    const mappedCategory = categoryByDetail[row.detail] || "";
    const segment = categorySegmentFromMappedCategory(mappedCategory, categorySegmentLookup);
    summary[segment].rows += 1;
    summary[segment].amountARS += Number(row.amountARS || 0);
  }

  return summary;
}

test("synthetic public/private transport rows land in essential vs discretionary buckets", () => {
  const categoryByDetail = {
    "MERPAGO*EMOVASUBTE": "Public Transport",
    "PAYU*AR*UBER": "Private Transport",
  };
  const categorySegmentLookup = normalizeCategorySegmentByCategory({
    "Public Transport": "essential",
    "Private Transport": "discretionary",
  });
  const rows = [
    { movementType: "CardMovement", detail: "MERPAGO*EMOVASUBTE", amountARS: 1200 },
    { movementType: "CardMovement", detail: "PAYU*AR*UBER", amountARS: 3400 },
  ];

  const summary = summarizeSyntheticSpend(rows, categoryByDetail, categorySegmentLookup);

  assert.deepEqual(summary, {
    essential: { rows: 1, amountARS: 1200 },
    discretionary: { rows: 1, amountARS: 3400 },
    unclassified: { rows: 0, amountARS: 0 },
  });
});
