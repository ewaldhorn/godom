#!/bin/bash
# build.sh — Build the Godom WASM module via standard Go
# Output goes to docs/godom.wasm.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

mkdir -p docs

echo "==> Copying wasm_exec.js from GOROOT ..."
if [ -f "$(go env GOROOT)/lib/wasm/wasm_exec.js" ]; then
    cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" docs/
else
    cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" docs/
fi

echo "==> Building godom.wasm ..."
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o docs/godom.wasm demo/demo.go

echo "==> Done. WASM binary: $(wc -c < docs/godom.wasm) bytes"
