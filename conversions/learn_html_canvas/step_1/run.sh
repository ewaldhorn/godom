#!/usr/bin/env bash

# we want the run to fail if anything fails, this exits on any error
set -euo pipefail

# let's first call our build.sh file, we can the wasm build to succeed before we even try to keep going.
./build.sh

# if we get here, the build worked
# so now copy index.html to dist
cp ./html/index.html ./dist/

# copy done, let's serve that directory on port 9000
http-server ./dist/ -p 9000
