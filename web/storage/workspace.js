/**
 * Workspace export/import: single artifact containing CSVs + configs.
 *
 * @module storage/workspace
 */

const WORKSPACE_VERSION = 1;

/**
 * Builds a workspace object for export.
 *
 * @param {Array<{ name: string, content: string }>} csvFiles
 * @param {object} mappingsObj - Parsed mappings (ownersByCardOwner, etc.)
 * @returns {object} Workspace object
 */
export function buildWorkspace(csvFiles, mappingsObj) {
  return {
    version: WORKSPACE_VERSION,
    csvFiles: csvFiles.map(({ name, content }) => ({ name, content })),
    config: mappingsObj || {},
  };
}

/**
 * Parses a workspace JSON string. Returns { csvFiles, config } or throws.
 *
 * @param {string} jsonText
 * @returns {{ csvFiles: Array<{ name: string, content: string }>, config: object }}
 */
export function parseWorkspace(jsonText) {
  const obj = JSON.parse(jsonText);
  if (!obj || typeof obj !== "object") {
    throw new Error("Invalid workspace: not an object");
  }
  const csvFiles = Array.isArray(obj.csvFiles) ? obj.csvFiles : [];
  const config = obj.config && typeof obj.config === "object" ? obj.config : {};
  for (const f of csvFiles) {
    if (!f || typeof f.name !== "string" || typeof f.content !== "string") {
      throw new Error("Invalid workspace: csvFiles must have name and content");
    }
  }
  return { csvFiles, config };
}
