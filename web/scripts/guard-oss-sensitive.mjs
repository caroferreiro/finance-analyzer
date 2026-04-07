#!/usr/bin/env node

import { execFileSync } from "node:child_process";

const MODE_VALUES = new Set(["delta", "all"]);

function parseMode(argv) {
  const modeArg = argv.find((arg) => arg.startsWith("--mode="));
  if (!modeArg) {
    return "delta";
  }
  const value = modeArg.slice("--mode=".length).trim().toLowerCase();
  if (!MODE_VALUES.has(value)) {
    throw new Error(`Invalid --mode value "${value}". Allowed: delta, all.`);
  }
  return value;
}

function git(cwd, args) {
  return execFileSync("git", args, {
    cwd,
    encoding: "utf8",
    stdio: ["ignore", "pipe", "pipe"],
  }).trim();
}

function toLines(text) {
  return String(text || "")
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean);
}

function normalizePath(filePath) {
  return String(filePath || "")
    .replace(/\\/g, "/")
    .replace(/^\.\//, "");
}

const RULES = Object.freeze([
  {
    id: "INV-001",
    reason: "Private integration PDF fixtures must stay out of OSS tree.",
    match: (filePath) => filePath.startsWith("pkg/integration_tests/") && filePath.endsWith(".pdf"),
  },
  {
    id: "INV-002",
    reason: "Sensitive joined CSV overlay is private-only.",
    match: (filePath) => filePath.startsWith("web/mockups_lab/tmp_sensitive_data/"),
  },
  {
    id: "INV-003",
    reason: "Legacy sensitive mapping files must not re-enter OSS-tracked paths.",
    match: (filePath) =>
      filePath === "web/mockups_lab/owner_map.csv" ||
      filePath === "web/mockups_lab/details_to_categories_map.csv",
  },
  {
    id: "INV-004",
    reason: "Legacy sensitive shortcut page must remain out of OSS launcher surface.",
    match: (filePath) => filePath === "web/mockups_lab/finance_analyzer_mockup_sensitive_shortcut.html",
  },
  {
    id: "INV-005",
    reason: "Snapshot HTML files are local artifacts and OSS-remove.",
    match: (filePath) => filePath.startsWith("web/__snapshots__/") && filePath.endsWith(".html"),
  },
  {
    id: "INV-006",
    reason: "Playwright result artifacts are local-only and OSS-remove.",
    match: (filePath) => filePath.startsWith("web/test-results/"),
  },
  {
    id: "INV-007",
    reason: "OS metadata files (.DS_Store) are OSS-remove.",
    match: (filePath) => /(^|\/)\.DS_Store$/u.test(filePath),
  },
]);

function collectCandidates(repoRoot, mode) {
  if (mode === "all") {
    return toLines(git(repoRoot, ["ls-files"]));
  }
  const addedTracked = toLines(git(repoRoot, ["diff", "--name-only", "--diff-filter=ACR", "HEAD"]));
  const untracked = toLines(git(repoRoot, ["ls-files", "--others", "--exclude-standard"]));
  return Array.from(new Set([...addedTracked, ...untracked]));
}

function findViolations(files) {
  const violations = [];
  for (const rawPath of files) {
    const filePath = normalizePath(rawPath);
    for (const rule of RULES) {
      if (rule.match(filePath)) {
        violations.push({
          filePath,
          ruleId: rule.id,
          reason: rule.reason,
        });
      }
    }
  }
  return violations;
}

function formatViolations(violations) {
  const byRule = new Map();
  for (const violation of violations) {
    const key = violation.ruleId;
    if (!byRule.has(key)) {
      byRule.set(key, {
        reason: violation.reason,
        files: [],
      });
    }
    byRule.get(key).files.push(violation.filePath);
  }

  const lines = [];
  for (const [ruleId, entry] of byRule.entries()) {
    lines.push(`- ${ruleId}: ${entry.reason}`);
    const uniqueFiles = Array.from(new Set(entry.files)).sort();
    for (const filePath of uniqueFiles) {
      lines.push(`  - ${filePath}`);
    }
  }
  return lines.join("\n");
}

function main() {
  let mode;
  try {
    mode = parseMode(process.argv.slice(2));
  } catch (err) {
    console.error(`[guard:oss-sensitive] ${err.message}`);
    process.exit(2);
  }

  let repoRoot;
  try {
    repoRoot = git(process.cwd(), ["rev-parse", "--show-toplevel"]);
  } catch (err) {
    console.error("[guard:oss-sensitive] Failed to locate git repository root.");
    process.exit(2);
  }

  const candidates = collectCandidates(repoRoot, mode);
  const violations = findViolations(candidates);

  if (!violations.length) {
    console.log(
      `[guard:oss-sensitive] PASS (${mode} mode): no blocked sensitive artifacts detected in candidate set.`
    );
    process.exit(0);
  }

  console.error(`[guard:oss-sensitive] FAIL (${mode} mode): detected blocked sensitive artifacts.`);
  console.error(formatViolations(violations));
  console.error("");
  console.error("Resolution:");
  console.error("1. Remove/move blocked files to private overlay/repo paths.");
  console.error("2. Keep the public tree aligned with the repository's public/private data split rules.");
  console.error("3. Re-run: npm run guard:oss-sensitive");
  process.exit(1);
}

main();
