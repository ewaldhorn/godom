#!/bin/bash
# build.sh — Build the click-rect example

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

echo "==> Copying wasm_exec.js from GOROOT ..."
if [ -f "$(go env GOROOT)/lib/wasm/wasm_exec.js" ]; then
    cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" .
else
    cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .
fi

echo "==> Building app.wasm ..."
GOOS=js GOARCH=wasm go build -o app.wasm main.go

echo "==> Done. Serve this directory with any static HTTP server."
echo "    e.g.: npx http-server . -p 9001"
