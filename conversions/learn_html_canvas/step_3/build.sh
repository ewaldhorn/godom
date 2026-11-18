#!/usr/bin/env bash

# let's create an output directory so we keep things tidy
mkdir -p dist

# have Go target wasm directly and output our binary in the dist directory
GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o ./dist/app.wasm ./src
