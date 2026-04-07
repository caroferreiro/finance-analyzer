/**
 * Storage error handling utilities.
 *
 * @module storage/utils
 */

/**
 * Formats storage usage for display. Handles undefined quota (e.g., private browsing).
 *
 * @param {number} usage - Usage in bytes
 * @param {number|undefined} quota - Quota in bytes, or undefined
 * @returns {string} e.g. "12.5 MB / 500 MB used" or "12.5 MB used"
 */
export function formatStorageUsage(usage, quota) {
  const usedMB = ((usage ?? 0) / 1024 / 1024).toFixed(2);
  if (quota == null || quota === undefined) {
    return `${usedMB} MB used`;
  }
  const totalMB = (quota / 1024 / 1024).toFixed(2);
  return `${usedMB} MB / ${totalMB} MB used`;
}

/**
 * Returns the persistent storage status message.
 *
 * @param {boolean} persisted - Result of navigator.storage.persisted()
 * @returns {string}
 */
export function getPersistentStorageStatus(persisted) {
  return persisted
    ? "Persistent storage enabled."
    : "May be cleared under storage pressure; export workspace recommended.";
}

export function isQuotaExceededError(err) {
  return err?.name === "QuotaExceededError" || err?.code === 22;
}

/**
 * Returns a user-facing message for storage failures.
 *
 * @param {Error} storageErr
 * @returns {string}
 */
export function getStorageWarning(storageErr) {
  return isQuotaExceededError(storageErr)
    ? "Storage full. Your data is in memory but won't persist. Export to save a copy."
    : "Storage unavailable. Data will not persist across reloads.";
}
