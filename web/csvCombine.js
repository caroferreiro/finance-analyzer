/**
 * Combines multiple stored CSV entries into a single CSV string.
 * Header is taken from the first file; data rows from all files are concatenated.
 * Used when calling the compute engine (CSVs stored separately, joined at compute).
 */

export function combineLoadedCsvFiles(files) {
  if (files.length === 0) return "";
  let header = "";
  const dataRows = [];
  for (const { content } of files) {
    const lines = content.trim().split("\n").filter((l) => l);
    if (lines.length === 0) continue;
    if (!header) header = lines[0];
    dataRows.push(...lines.slice(1));
  }
  if (dataRows.length === 0) return header + "\n";
  return header + "\n" + dataRows.join("\n") + "\n";
}
