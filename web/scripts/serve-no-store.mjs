#!/usr/bin/env node

import http from "node:http";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

function parseArgs(argv) {
  const args = { root: path.resolve(__dirname, ".."), port: 8787, host: "127.0.0.1" };
  for (let i = 0; i < argv.length; i++) {
    const value = argv[i];
    if (value === "--root" && argv[i + 1]) {
      args.root = path.resolve(argv[++i]);
      continue;
    }
    if (value === "--port" && argv[i + 1]) {
      args.port = Number(argv[++i]);
      continue;
    }
    if (value === "--host" && argv[i + 1]) {
      args.host = String(argv[++i]);
      continue;
    }
  }
  if (!Number.isFinite(args.port) || args.port <= 0 || args.port > 65535) {
    throw new Error("Invalid --port value.");
  }
  return args;
}

function contentTypeFor(filePath) {
  const ext = path.extname(filePath).toLowerCase();
  switch (ext) {
    case ".html":
      return "text/html; charset=utf-8";
    case ".js":
    case ".mjs":
      return "text/javascript; charset=utf-8";
    case ".css":
      return "text/css; charset=utf-8";
    case ".json":
      return "application/json; charset=utf-8";
    case ".wasm":
      return "application/wasm";
    case ".csv":
      return "text/csv; charset=utf-8";
    case ".svg":
      return "image/svg+xml";
    case ".png":
      return "image/png";
    case ".jpg":
    case ".jpeg":
      return "image/jpeg";
    default:
      return "application/octet-stream";
  }
}

function shouldDisableCache(filePath) {
  const ext = path.extname(filePath).toLowerCase();
  return ext === ".html" || ext === ".js" || ext === ".mjs" || ext === ".wasm";
}

function noStoreHeaders(headers) {
  headers["cache-control"] = "no-store, no-cache, max-age=0, must-revalidate";
  headers.pragma = "no-cache";
  headers.expires = "0";
}

function safeResolve(root, requestPath) {
  const normalized = path
    .normalize(String(requestPath || "/"))
    .replace(/^(\.\.[/\\])+/, "")
    .replace(/\\/g, "/");
  const candidate = path.resolve(root, "." + normalized);
  const rel = path.relative(root, candidate);
  if (rel.startsWith("..") || path.isAbsolute(rel)) {
    return null;
  }
  return candidate;
}

function notFound(res) {
  res.writeHead(404, { "content-type": "text/plain; charset=utf-8" });
  res.end("Not found");
}

const { root, port, host } = parseArgs(process.argv.slice(2));

const server = http.createServer((req, res) => {
  const method = String(req.method || "GET").toUpperCase();
  if (method !== "GET" && method !== "HEAD") {
    res.writeHead(405, { "content-type": "text/plain; charset=utf-8" });
    res.end("Method not allowed");
    return;
  }

  const url = new URL(req.url || "/", "http://local");
  let pathname = decodeURIComponent(url.pathname || "/");
  if (pathname.endsWith("/")) {
    pathname += "index.html";
  }
  const filePath = safeResolve(root, pathname);
  if (!filePath) {
    notFound(res);
    return;
  }

  fs.stat(filePath, (statErr, stat) => {
    if (statErr || !stat.isFile()) {
      notFound(res);
      return;
    }

    const headers = {
      "content-type": contentTypeFor(filePath),
      "content-length": String(stat.size),
    };
    if (shouldDisableCache(filePath)) {
      noStoreHeaders(headers);
    }

    res.writeHead(200, headers);
    if (method === "HEAD") {
      res.end();
      return;
    }
    fs.createReadStream(filePath).pipe(res);
  });
});

server.listen(port, host, () => {
  // eslint-disable-next-line no-console
  console.log(`[serve-no-store] Serving ${root} at http://${host}:${port} (HTML/JS/WASM no-store)`);
});

server.on("error", (err) => {
  // eslint-disable-next-line no-console
  console.error(
    `[serve-no-store] Failed to listen on http://${host}:${port}: ${err?.message || String(err)}`,
  );
  process.exit(1);
});
