#!/bin/bash
# run.sh — Build godom.wasm and start the dev server

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

echo "==> Building WASM ..."
./build.sh

echo "==> Starting dev server on http://localhost:9000 ..."
npx http-server ./docs/ -p 9000 -c-1
