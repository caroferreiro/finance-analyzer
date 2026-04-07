const DEFAULT_REQUIRED_WASM_FUNCTIONS = Object.freeze([
  "computeFromCSV",
  "exportTableCSVFromResult",
  "demoCSV",
  "demoMappingsJSON",
]);

const bootstrapPromises = new WeakMap();

export function getMissingWasmFunctions(
  scope = globalThis,
  requiredFns = DEFAULT_REQUIRED_WASM_FUNCTIONS
) {
  return requiredFns.filter((fn) => typeof scope[fn] !== "function");
}

export function okValue(result) {
  if (!result || result.ok !== true) {
    throw new Error(result?.error || "unknown wasm error");
  }
  return result.value;
}

function defaultLoadScript(scriptPath, scope = globalThis) {
  const doc = scope.document;
  if (!doc) {
    throw new Error("Cannot load wasm_exec.js: document is unavailable.");
  }

  const scriptUrl = new URL(scriptPath, doc.baseURI).href;
  const existing = doc.querySelector(`script[src="${scriptUrl}"]`);
  if (existing) {
    return new Promise((resolve, reject) => {
      if (typeof scope.Go === "function") {
        resolve();
        return;
      }
      existing.addEventListener("load", () => resolve(), { once: true });
      existing.addEventListener("error", () => reject(new Error(`Failed to load ${scriptPath}`)), {
        once: true,
      });
    });
  }

  return new Promise((resolve, reject) => {
    const script = doc.createElement("script");
    script.src = scriptPath;
    script.async = true;
    script.onload = () => resolve();
    script.onerror = () => reject(new Error(`Failed to load ${scriptPath}`));
    doc.head.appendChild(script);
  });
}

async function instantiateWasm({
  scope,
  wasmPath,
  go,
  fetchImpl,
  instantiateStreaming,
}) {
  if (typeof instantiateStreaming === "function") {
    try {
      const response = await fetchImpl(wasmPath);
      if (!response.ok) {
        throw new Error(`${wasmPath} could not be fetched (HTTP ${response.status}).`);
      }
      return instantiateStreaming(response, go.importObject);
    } catch {
      // Fallback to arrayBuffer instantiate for servers without proper wasm MIME type.
    }
  }

  const response = await fetchImpl(wasmPath);
  if (!response.ok) {
    throw new Error(`${wasmPath} could not be fetched (HTTP ${response.status}).`);
  }
  const bytes = await response.arrayBuffer();
  return scope.WebAssembly.instantiate(bytes, go.importObject);
}

export async function ensureWasmRuntime(options = {}) {
  const scope = options.scope || globalThis;
  const requiredFns = options.requiredFns || DEFAULT_REQUIRED_WASM_FUNCTIONS;

  if (getMissingWasmFunctions(scope, requiredFns).length === 0) {
    return;
  }
  const existingPromise = bootstrapPromises.get(scope);
  if (existingPromise) {
    return existingPromise;
  }

  const wasmExecPath = options.wasmExecPath || "./wasm_exec.js";
  const wasmPath = options.wasmPath || "./finance.wasm";
  const loadScript = options.loadScript || defaultLoadScript;
  const fetchImpl = options.fetchImpl || scope.fetch?.bind(scope);
  const instantiateStreaming =
    options.instantiateStreaming || scope.WebAssembly?.instantiateStreaming?.bind(scope.WebAssembly);

  if (!scope.WebAssembly || typeof scope.WebAssembly.instantiate !== "function") {
    throw new Error("WebAssembly is unavailable in this environment.");
  }
  if (typeof fetchImpl !== "function") {
    throw new Error("fetch is unavailable in this environment.");
  }

  const bootstrapPromise = (async () => {
    if (typeof scope.Go !== "function") {
      await loadScript(wasmExecPath, scope);
    }
    if (typeof scope.Go !== "function") {
      throw new Error(`Go runtime not found after loading ${wasmExecPath}.`);
    }

    const go = new scope.Go();
    scope.__financeWasmGo = go;

    const instanceResult = await instantiateWasm({
      scope,
      wasmPath,
      go,
      fetchImpl,
      instantiateStreaming,
    });

    go.run(instanceResult.instance);

    const missingAfter = getMissingWasmFunctions(scope, requiredFns);
    if (missingAfter.length > 0) {
      throw new Error(`WASM functions not found: ${missingAfter.join(", ")}`);
    }
  })().catch((err) => {
    bootstrapPromises.delete(scope);
    throw err;
  });
  bootstrapPromises.set(scope, bootstrapPromise);

  return bootstrapPromise;
}

export async function computeResultJSONFromCsv(csvText, mappingsJSON, options = {}) {
  const scope = options.scope || globalThis;
  await ensureWasmRuntime(options);
  return okValue(scope.computeFromCSV(csvText, mappingsJSON));
}

export async function computeResultFromCsv(csvText, mappingsJSON, options = {}) {
  return JSON.parse(await computeResultJSONFromCsv(csvText, mappingsJSON, options));
}

export async function exportTableCSVFromComputeResult(computeResultOrJSON, tableID, options = {}) {
  const scope = options.scope || globalThis;
  await ensureWasmRuntime(options);

  const computeResultJSON =
    typeof computeResultOrJSON === "string"
      ? computeResultOrJSON
      : JSON.stringify(computeResultOrJSON);

  return okValue(scope.exportTableCSVFromResult(computeResultJSON, tableID));
}

export async function loadDemoBundle(options = {}) {
  const scope = options.scope || globalThis;
  await ensureWasmRuntime(options);
  return {
    csvText: okValue(scope.demoCSV()),
    mappingsJSON: okValue(scope.demoMappingsJSON()),
  };
}

export { DEFAULT_REQUIRED_WASM_FUNCTIONS };
