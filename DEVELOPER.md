# Developer Guide — godom

Welcome to the `godom` developer guide! This document explains the architecture, package structure, and guidelines for compiling, running, and testing the `godom` WebAssembly library.

---

## Technical Overview
`godom` is a dependency-free, zero-allocation-focused Go package that wraps the browser DOM and Canvas 2D APIs.
- **Target**: `GOOS=js GOARCH=wasm`
- **Compiler**: Go 1.26.4 or later
- **Bridge Layer**: `syscall/js` (standard library)

---

## Codebase Architecture & Packages

### 1. `dom`
The [`dom`](dom/dom.go) package manages global reference handles to JS objects (e.g. `document`, `body`, `head`) and provides simple wrapper functions for DOM CRUD operations, visibility (`Hide`/`Show`), events (`AddEventListener`), and focus handling.

### 2. `html`
The [`html`](html/html.go) package implements a fluent tag builder around the `Element` struct. It provides type-safe, chainable tag constructors for most common HTML elements (e.g., `Div()`, `Span()`, `Button()`, `Table()`).
- Chainable functions such as `.Class()`, `.ID()`, `.Text()`, `.HTML()`, `.Attr()`, `.Child()`, and `.On()` make UI building highly readable.

### 3. `canvas`
The [`canvas`](canvas/canvas.go) package is built for fast rendering loops.
- Avoids garbage collector overhead by pre-allocating JS objects (`Uint8ClampedArray`, `ImageData`).
- Provides drawing routines implemented in Go (Bresenham line and circle algorithms, rectangles, triangles, and pixel manipulation).
- Double-buffered updates: Draw to the Go-side `Pixels []byte` slice, then call `.Render()` to blit changes to JavaScript.

### 4. `colour`
The [`colour`](colour/colour.go) package provides standard RGBA structures, color constants, grayscale conversion, and a seedable Xorshift64 PRNG to generate random colors without external imports.

### 5. `sound`
The [`sound`](sound/sound.go) package contains mathematical algorithms to synthesize audio samples in memory (e.g., click sounds) to be played in the browser.

---

## Development Workflow

### Compiling the Demo
To compile the WebAssembly module:
```bash
./build.sh
```
This copies the required JS bridge (`wasm_exec.js`) and outputs the compiled binary to `docs/godom.wasm`.

### Running Locally
To launch a development server and try the included synthwave physics and sequencer demo:
```bash
./run.sh
```
Then open your browser and navigate to `http://localhost:9000`.

---

## Testing & Benchmarks

### Running Unit Tests
Since this library interacts directly with WebAssembly globals, tests must run under a WebAssembly environment (using Node.js).
Run the test runner:
```bash
./test.sh
```
The script locates `wasm_exec_node.js` inside your `GOROOT` and runs Go tests under `GOOS=js GOARCH=wasm`.

#### How Mocks Work
Because Node.js does not natively have browser DOM APIs, packages like `dom`, `canvas`, and `html` bootstrap minimal JS DOM mocks inside `init()` in their respective `*_test.go` files:
- Mocks simulate `document.createElement`, `classList`, `appendChild`, and event listeners in a minimal JS environment.
- At the end of asynchronous tests (e.g. `TestAnimationLoop`), remember to clean up timers/callbacks (e.g., setting `stopRAF = true`) to avoid memory leaks or panics.

### Running Benchmarks
To run the included canvas clear-screen benchmarks under Node.js:
```bash
./run_benchmarks.sh
```
