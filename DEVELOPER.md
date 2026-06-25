# Developer Guide — godom

Welcome to the `godom` developer guide! This document covers the package architecture, API reference details, and workflow for building, running, and testing the library.

> **Audience note:** This guide is primarily aimed at **contributors** to this repository and developers who want a deeper understanding of the internals. If you just want to use `godom` in your own project, start with [README.md](README.md) for installation, basic usage, and compilation instructions.

---

## Technical Overview

`godom` is a dependency-free, zero-allocation-focused Go package that wraps the browser DOM and Canvas 2D APIs.
- **Target**: `GOOS=js GOARCH=wasm`
- **Compiler**: Go 1.26 or later
- **Bridge Layer**: `syscall/js` (standard library)

---

## Keeping the Program Alive (`select {}`)

When Go compiles to WebAssembly, the runtime exits as soon as `main()` returns — this tears down all event listeners, animation callbacks, and goroutines. To prevent this, block the main goroutine indefinitely:

```go
func main() {
    dom.Init()
    // ... set up your UI and event handlers ...
    select {} // block forever; keeps WASM runtime and all callbacks alive
}
```

This is a standard pattern for Go WASM programs. Without it, any event handlers or `requestAnimationFrame` callbacks registered after `main()` returns will be silently dropped.

---

## Codebase Architecture & Packages

### 1. `dom`

The [`dom`](dom/dom.go) package manages global reference handles to JS objects (e.g. `document`, `body`, `head`) and provides simple wrapper functions for DOM CRUD operations, visibility (`Hide`/`Show`), events (`AddEventListener`), and focus handling.

#### The `Handle` type

`dom.Handle` is the core bridge type used throughout `godom`. It is a thin wrapper around `js.Value` and represents a live reference to any JavaScript object in the browser:

```go
type Handle js.Value
```

Key methods on `Handle`:

| Method | Description |
|---|---|
| `IsValid() bool` | Returns `true` if the handle is non-null and non-undefined |
| `Get(key string) string` | Reads a JS property as a string |
| `Set(key string, value any)` | Writes an arbitrary value to a JS property |
| `SetInnerText(text string)` | Sets `innerText` on the element |
| `SetInnerHTML(html string)` | Sets `innerHTML` on the element |
| `AddClassTo(class string)` | Adds a CSS class via `classList.add` |
| `RemoveClassFrom(class string)` | Removes a CSS class via `classList.remove` |
| `ReplaceClasses(classes []string)` | Replaces all classes on the element |

`Handle` is the glue between the `html` and `dom` packages — `html.Element.Build()` returns a `Handle`, and `html.Element.Child()` / `AppendTo()` both accept one.

#### Notable standalone functions

| Function | Description |
|---|---|
| `GetElementByID(id string) Handle` | Looks up an element; returns `Invalid` if not found |
| `AddNewStyleElement(css string)` | Injects a `<style>` block with raw CSS into `<head>` at runtime |
| `GetString(elemID, key string) string` | Reads a string property from an element by ID |
| `SetValue(elemID, key, value string)` | Writes a string property on an element by ID |
| `WrapElementWithNewDiv(element Handle, classes []string) Handle` | Wraps an existing element in a new `<div>` with the given CSS classes |
| `AddClass(elemID, class string)` | Adds a CSS class to an element looked up by ID |
| `RemoveClass(elemID, class string)` | Removes a CSS class from an element looked up by ID |
| `StartAnimationLoop(cb js.Value)` | Drives a 60 FPS render loop via `requestAnimationFrame` |

#### `Context2D`

`dom` also exports a `Context2D` struct that wraps the HTML5 canvas 2D rendering context for direct JS canvas API calls:

```go
type Context2D struct {
    Ctx Handle
}
```

Methods on `Context2D`:

| Method | Description |
|---|---|
| `BeginPath()` | Calls `beginPath()` on the context |
| `Fill()` | Fills the current path |
| `Arc(x, y, radius, startAngle, endAngle float64, ccw bool)` | Draws a circular arc |
| `FillStyle(style string)` | Sets the `fillStyle` property |

Obtain a `Context2D` from a `canvas.Canvas` instance via `canvas.GetContext2D()`.

---

### 2. `html`

The [`html`](html/html.go) package implements a fluent tag builder around the `Element` struct. It provides type-safe, chainable tag constructors for most common HTML elements (e.g., `Div()`, `Span()`, `Button()`, `Table()`).

Chainable methods on `*Element`:

| Method | Description |
|---|---|
| `.ID(val string)` | Sets the element's `id` attribute |
| `.Class(val string)` | Adds a CSS class name |
| `.Text(val string)` | Sets `innerText` |
| `.HTML(val string)` | Sets `innerHTML` |
| `.Attr(key, val string)` | Sets an arbitrary attribute |
| `.Child(childHandle dom.Handle)` | Appends a child element (accepts the result of `.Build()`) |
| `.AppendTo(parent dom.Handle)` | Appends this element to a parent handle |
| `.On(event string, cb js.Value)` | Registers an event listener |
| `.Build()` | Returns the underlying `dom.Handle` |

---

### 3. `canvas`

The [`canvas`](canvas/canvas.go) package is built for fast rendering loops.
- Avoids garbage collector overhead by pre-allocating JS objects (`Uint8ClampedArray`, `ImageData`).
- Provides drawing routines implemented in Go (Bresenham line and circle algorithms, rectangles, triangles, and pixel manipulation).
- Double-buffered updates: draw to the `pixels []byte` slice you passed into `NewCanvas`, then call `.Render()` to blit changes to JavaScript in one shot.

#### `Point`

Many drawing methods accept a `Point` struct instead of raw coordinates:

```go
type Point struct {
    X, Y int32
}
```

#### Drawing primitives

Each primitive has two forms:
- A **plain** form that uses the currently active colour set via `SetColour`.
- A **`Colour*`** form that accepts a `colour.Colour` argument, letting you pass a colour inline without a separate `SetColour` call.

| Plain | Colour variant | Description |
|---|---|---|
| `PutPixel(x, y int32)` | `ColourPutPixel(x, y int32, col Colour)` | Draw a single pixel |
| `GetPixel(x, y int32) (Colour, bool)` | — | Read a pixel's colour; `false` if out of bounds |
| `Line(x1, y1, x2, y2 int32)` | `ColourLine(...)` | Bresenham line |
| `LinePoint(p1, p2 Point)` | `ColourLinePoint(...)` | Line between two `Point` values |
| `Circle(midX, midY, radius int32)` | `ColourCircle(...)` | Outline circle |
| `FilledCircle(midX, midY, radius int32)` | `ColourFilledCircle(...)` | Filled circle |
| `BorderCircle(midX, midY, radius, borderWidth int32)` | `ColourBorderCircle(...)` | Ring / annulus |
| `Rectangle(xStart, yStart, w, h, thickness int32)` | `ColourRectangle(...)` | Outline rectangle |
| `FilledRectangle(xStart, yStart, w, h int32)` | `ColourFilledRectangle(...)` | Filled rectangle |
| `Triangle(p1, p2, p3 Point)` | — | Wireframe triangle |

Other key methods:

| Method | Description |
|---|---|
| `NewCanvas(width, height int, pixels []byte, parentID string) *Canvas` | Creates and attaches a canvas to the DOM element with the given ID |
| `NewOffscreenCanvas(width, height int, pixels []byte) *Canvas` | Creates a pixel-buffer-only canvas with no DOM attachment; `Render()` is a no-op. Useful for tests and benchmarks. |
| `Width() int` | Returns the canvas width in pixels |
| `Height() int` | Returns the canvas height in pixels |
| `Pixels() []byte` | Returns the pixel buffer slice (shared reference — writes are reflected immediately) |
| `CanvasHandle() dom.Handle` | Returns the DOM handle for the underlying `<canvas>` element |
| `ClearScreen(col colour.Colour)` | Fills the entire pixel buffer with one colour using an O(n log n) copy doubling trick |
| `SetColour(col colour.Colour)` | Sets the active drawing colour |
| `GetColour() colour.Colour` | Returns the active drawing colour |
| `Render()` | Blits the pixel buffer to JavaScript via `CopyBytesToJS` + `putImageData`; no-op on an offscreen canvas |
| `GetContext2D() dom.Context2D` | Returns a `Context2D` wrapper for the underlying JS canvas context |

---

### 4. `colour`

The [`colour`](colour/colour.go) package provides standard RGBA structures, color constants, grayscale conversion, and a seedable Xorshift64 PRNG to generate random colors without external imports.

```go
type Colour struct {
    R, G, B, A uint8
}
```

Predefined constants: `White`, `Black`, `Empty` (fully transparent).

| Function / Method | Description |
|---|---|
| `IsEmpty() bool` | Returns `true` if all RGBA components are zero (fully transparent) |
| `ConvertToGrayscale()` | Converts the colour in-place to grayscale using luminance weighting (`0.299R + 0.587G + 0.114B`) |
| `RandomColour() Colour` | Returns a random fully-opaque colour |
| `Seed(s uint64)` | Re-seeds the PRNG; a zero value is silently ignored |

---

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

Use the provided script from the project root (it handles the `wasm_exec_node.js` path differences between Go versions automatically):

```bash
./run_benchmarks.sh
```

Or run the command manually:

```bash
GOOS=js GOARCH=wasm go test -exec="node $(go env GOROOT)/lib/wasm/wasm_exec_node.js" -bench=. -benchmem ./benchmarks/...
```

**Command flags explained:**
- `GOOS=js GOARCH=wasm` — targets the WebAssembly runtime
- `-exec="node ..."` — uses Node.js + the Go WASM wrapper to execute the compiled test binary
- `-bench=.` — runs all benchmarks
- `-benchmem` — includes memory allocation statistics in the output

#### ClearScreen benchmark results

Clearing a `1920×1080` pixel buffer:

| Approach | Time/op | Notes |
|---|---|---|
| Naive byte-by-byte loop | ~3.45 ms | Sets each RGBA pixel individually |
| Optimised exponential copy | ~0.15 ms | **~22× faster** — used in `ClearScreen()` |

Results will vary by CPU and Node.js version. These were captured on a standard developer laptop under Node.js with WASM.

#### Adding your own benchmarks

Add a `_test.go` file to the `benchmarks/` package following the standard Go benchmark signature:

```go
func BenchmarkMyOperation(b *testing.B) {
    // setup ...
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        // operation under test
    }
}
```

Then run `./run_benchmarks.sh` to include it.
