#!/bin/bash
# test.sh — Run Godom unit tests via Go WebAssembly target using Node.js

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

echo "==> Resolving GOROOT ..."
GOROOT_DIR="$(go env GOROOT)"

# Locate wasm_exec_node.js
if [ -f "$GOROOT_DIR/lib/wasm/wasm_exec_node.js" ]; then
    WASM_EXEC="$GOROOT_DIR/lib/wasm/wasm_exec_node.js"
elif [ -f "$GOROOT_DIR/misc/wasm/wasm_exec_node.js" ]; then
    WASM_EXEC="$GOROOT_DIR/misc/wasm/wasm_exec_node.js"
else
    echo "Error: wasm_exec_node.js not found in GOROOT ($GOROOT_DIR)" >&2
    exit 1
fi

echo "==> Running unit tests for GOOS=js GOARCH=wasm ..."
GOOS=js GOARCH=wasm go test -exec="node $WASM_EXEC" ./...
