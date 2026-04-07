#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "$repo_root"
GOOS=js GOARCH=wasm go build -o web/finance.wasm ./pkg/cmd/financewasm
