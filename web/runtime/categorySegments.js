export const ESSENTIAL_CATEGORY_SEGMENT = "essential";
export const DISCRETIONARY_CATEGORY_SEGMENT = "discretionary";
export const UNCLASSIFIED_CATEGORY_SEGMENT = "unclassified";

export const VALID_CATEGORY_SEGMENT_KEYS = Object.freeze([
  ESSENTIAL_CATEGORY_SEGMENT,
  DISCRETIONARY_CATEGORY_SEGMENT,
]);

export function normalizeCategoryKey(value) {
  return String(value || "")
    .trim()
    .toLowerCase()
    .replace(/\s+/g, " ");
}

export function normalizeCategorySegmentKey(value) {
  var key = normalizeCategoryKey(value);
  return VALID_CATEGORY_SEGMENT_KEYS.includes(key) ? key : "";
}

export function normalizeCategorySegmentByCategory(segmentByCategory) {
  var normalized = Object.create(null);
  if (!segmentByCategory || typeof segmentByCategory !== "object" || Array.isArray(segmentByCategory)) {
    return normalized;
  }

  Object.entries(segmentByCategory).forEach(function (entry) {
    var categoryKey = normalizeCategoryKey(entry[0]);
    var segmentKey = normalizeCategorySegmentKey(entry[1]);
    if (!categoryKey || !segmentKey) {
      return;
    }
    normalized[categoryKey] = segmentKey;
  });

  return normalized;
}
export function categorySegmentFromMappedCategory(category, segmentByCategoryLookup) {
  var key = normalizeCategoryKey(category);
  if (
    !key ||
    key === "?" ||
    key === "uncategorized" ||
    key === "adjustments"
  ) {
    return UNCLASSIFIED_CATEGORY_SEGMENT;
  }
  var lookup = segmentByCategoryLookup || {};
  return lookup[key] || UNCLASSIFIED_CATEGORY_SEGMENT;
}
