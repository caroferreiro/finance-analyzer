# Web App

This folder contains the static web app for `finance-analyzer`.

## What it does

The app loads CSV data, mappings, and workspace state in the browser and renders spending, debt, owner, and data-quality views. The public deployment boots with demo/public data by default.

## Does it keep data local?

Yes. This website is 100% local with respect to your finance data.

In this repo there is:
1. no application backend,
2. no server-side database,
3. no external persistence path for uploaded CSVs, mappings, or workspace state.

Working data lives in browser storage on the local machine so that:
1. sensitive financial data does not need to leave the device,
2. the app can be hosted as plain static files,
3. the public deployment model stays simple and reproducible.

The pinned Highcharts CDN is a runtime dependency for charting code only. It is not used to upload or store your finance data.

## Run it locally

From the repo root:

```sh
cd web
npm install
npm run build:wasm
python3 -m http.server 8080 -d .
```

Then open:

- `http://localhost:8080`

## Build and test

From `web/`:

```sh
npm run build:wasm
npm run test:unit
npm run test:smoke
```

## Highcharts runtime

The public web app loads Highcharts from the official pinned CDN instead of a vendored local snapshot:
1. `https://code.highcharts.com/12.5.0/highcharts.js`
2. `https://code.highcharts.com/12.5.0/themes/dark-unica.js`

Browser-based runs therefore need network access to `code.highcharts.com`.

## GitHub Pages deployment

The public site is deployed automatically from `main` through GitHub Actions.

See [../docs/DEPLOYMENT.md](../docs/DEPLOYMENT.md) for the deployment flow and post-deploy smoke checks.

## OSS sensitive artifact guard

From `web/`:

```sh
npm run guard:oss-sensitive
```

Optional full-tree scan:

```sh
npm run guard:oss-sensitive:all
```

Enable local pre-commit enforcement from the repo root:

```sh
./scripts/install-git-hooks.sh
```

## UX audit automation

From `web/`:

```sh
npm run audit:ux
```

Artifacts are written to:
1. `web/output/playwright/ux-audit/ux-audit-summary.json`
2. `web/output/playwright/ux-audit/ux-audit-*.png`
