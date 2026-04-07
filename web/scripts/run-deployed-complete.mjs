#!/usr/bin/env node

import fs from "node:fs";
import path from "node:path";
import { spawn } from "node:child_process";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const webRoot = path.resolve(__dirname, "..");
const localUrlPath = path.join(webRoot, ".deployed-site-url");
const defaultBaseURL = "https://alechan.github.io/finance-analyzer/";

function normalizeBaseURL(value) {
  const normalized = String(value || "").trim();
  if (!normalized) {
    return "";
  }
  return normalized.endsWith("/") ? normalized : `${normalized}/`;
}

function readLocalBaseURL() {
  if (!fs.existsSync(localUrlPath)) {
    return "";
  }
  return normalizeBaseURL(fs.readFileSync(localUrlPath, "utf8"));
}

function resolveBaseURL() {
  const envBaseURL = normalizeBaseURL(process.env.PLAYWRIGHT_BASE_URL);
  if (envBaseURL) {
    return { baseURL: envBaseURL, source: "PLAYWRIGHT_BASE_URL" };
  }

  const fileBaseURL = readLocalBaseURL();
  if (fileBaseURL) {
    return { baseURL: fileBaseURL, source: localUrlPath };
  }

  return { baseURL: defaultBaseURL, source: "default deployed URL" };
}

const { baseURL, source } = resolveBaseURL();
const extraArgs = process.argv.slice(2);
const npxCommand = process.platform === "win32" ? "npx.cmd" : "npx";

// eslint-disable-next-line no-console
console.log(`[deployed-complete] Using ${baseURL} (${source})`);

const child = spawn(npxCommand, ["playwright", "test", "e2e/deployed-complete.spec.js", ...extraArgs], {
  cwd: webRoot,
  stdio: "inherit",
  env: {
    ...process.env,
    PLAYWRIGHT_BASE_URL: baseURL,
    PLAYWRIGHT_SKIP_WEBSERVER: "1",
  },
});

child.on("exit", (code, signal) => {
  if (signal) {
    process.kill(process.pid, signal);
    return;
  }
  process.exit(code ?? 1);
});
