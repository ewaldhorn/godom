# Godom — A Go WebAssembly DOM Utility Library

`godom` is a header-less, dependency-free wrapper around the browser DOM and Canvas 2D APIs written in pure Go. It targets standard WebAssembly (`GOOS=js GOARCH=wasm`) and operates using Go 1.26's standard `syscall/js` layer.

It was created to prove the feasibility of zero-allocation canvas updates and low-latency audio pre-rendering with DOM manipulation in Go WebAssembly.

---

## Prerequisites

- **Go 1.26** or later
- **Node.js** (for running tests and the included demo — see [`DEVELOPER.md`](DEVELOPER.md))

---

## Installation

To import the library packages in your own Go WebAssembly project:

```bash
go get github.com/ewaldhorn/godom
```

---

## Basic Usage

Create a `main.go` file inside your project:

```go
package main

import (
	"github.com/ewaldhorn/godom/dom"
	"github.com/ewaldhorn/godom/html"
)

func main() {
	// Initialize the DOM handles (captures document, body, and head)
	dom.Init()

	// Declaratively build elements in a chainable hierarchy
	_ = html.Div().
		Class("container").
		Child(html.H1().Text("Hello World!").Build()).
		Child(html.P().Text("Rendering using Go WebAssembly.").Build()).
		AppendTo(dom.Body)

	// Block forever — Go's WASM runtime exits as soon as main() returns,
	// which would tear down all event listeners and animation callbacks.
	select {}
}
```

---

## Compilation

Build your application to WebAssembly using the standard Go compiler target:

```bash
GOOS=js GOARCH=wasm go build -o app.wasm main.go
```

---

## Serving & Loading in JS

1. Locate and copy `wasm_exec.js` from your Go installation:
   ```bash
   cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" .
   ```
   > **Note:** On older Go installations the file may be at `$(go env GOROOT)/misc/wasm/wasm_exec.js` instead. If the first path gives an error, try that one.
2. Load the WebAssembly module inside your `index.html` page:

```html
<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <script src="wasm_exec.js"></script>
  </head>
  <body>
    <!-- Container for the canvas — the id must match the parentID passed to canvas.NewCanvas -->
    <div id="app"></div>
    <div id="loading">Loading App...</div>
    <script>
      const go = new Go();
      WebAssembly.instantiateStreaming(fetch("app.wasm"), go.importObject).then((result) => {
        document.getElementById("loading").remove();
        go.run(result.instance);
      }).catch(err => {
        console.error("Failed to load WASM:", err);
      });
    </script>
  </body>
</html>
```

> **Note:** The `synth-worklet.js` file in `docs/` is only required for the demo's audio synthesizer. It is not needed for basic DOM or canvas usage.

---

## Canvas

The `canvas` package provides a pixel buffer and drawing primitives implemented in pure Go. Draw to the `pixels []byte` slice you passed into `NewCanvas`, then call `Render()` to blit to the browser in one shot:

```go
import (
	"github.com/ewaldhorn/godom/canvas"
	"github.com/ewaldhorn/godom/colour"
)

pixels := make([]byte, 320*240*4)
// The last argument is the ID of an existing HTML element to attach the canvas to.
c := canvas.NewCanvas(320, 240, pixels, "app")

c.ClearScreen(colour.Black)
c.SetColour(colour.White)
c.Line(10, 10, 100, 100)
c.FilledCircle(160, 120, 50)
c.Rectangle(200, 50, 80, 60, 2)
c.Render()
```

> **Note:** `NewCanvas` appends a `<canvas>` element to the DOM element whose `id` matches the `parentID` argument (`"app"` above). Make sure that element exists in your `index.html` before the WASM module runs.

#### Drawing primitives

Each primitive has two forms. The **plain** form uses the active colour set by `SetColour`. The **`Colour*`** form takes a `colour.Colour` argument, letting you pass a colour inline without a separate `SetColour` call:

| Plain | Colour variant | Description |
|---|---|---|
| `PutPixel(x, y int32)` | `ColourPutPixel(x, y int32, col Colour)` | Single pixel |
| `GetPixel(x, y int32) (Colour, bool)` | — | Read pixel; `false` if out of bounds |
| `Line(x1, y1, x2, y2 int32)` | `ColourLine(...)` | Bresenham line |
| `LinePoint(p1, p2 Point)` | `ColourLinePoint(...)` | Line between two `Point` values |
| `Circle(midX, midY, radius int32)` | `ColourCircle(...)` | Outline circle |
| `FilledCircle(midX, midY, radius int32)` | `ColourFilledCircle(...)` | Filled circle |
| `BorderCircle(midX, midY, radius, borderWidth int32)` | `ColourBorderCircle(...)` | Ring / annulus |
| `Rectangle(x, y, w, h, thickness int32)` | `ColourRectangle(...)` | Outline rectangle |
| `FilledRectangle(x, y, w, h int32)` | `ColourFilledRectangle(...)` | Filled rectangle |
| `Triangle(p1, p2, p3 Point)` | — | Wireframe triangle |

`Point` is a simple struct `{ X, Y int32 }` defined in the `canvas` package.

---

## Events

Register event listeners on any element with the fluent `On()` method:

```go
import (
	"syscall/js"

	"github.com/ewaldhorn/godom/dom"
	"github.com/ewaldhorn/godom/html"
)

// js.FuncOf returns a js.Func. Pass .Value where a js.Value is expected.
clickFn := js.FuncOf(func(this js.Value, args []js.Value) any {
	dom.Alert("clicked!")
	return nil
})
btn := html.Button().
	Text("Click me").
	On("click", clickFn.Value).
	AppendTo(dom.Body)
```

The `dom` package also provides `AddEventListener` and `AddEventListenerByID` for lower-level event wiring, plus `StartAnimationLoop` for `requestAnimationFrame`-driven rendering:

```go
import (
	"syscall/js"

	"github.com/ewaldhorn/godom/canvas"
	"github.com/ewaldhorn/godom/colour"
	"github.com/ewaldhorn/godom/dom"
)

func main() {
	dom.Init()

	pixels := make([]byte, 300*300*4)
	c := canvas.NewCanvas(300, 300, pixels, "app")

	var x, y int32 = 150, 150
	var dx, dy int32 = 2, 2
	red := colour.Colour{R: 255, G: 0, B: 0, A: 255}

	// js.FuncOf returns a js.Func. Pass .Value where a js.Value is expected.
	loopFn := js.FuncOf(func(this js.Value, args []js.Value) any {
		// Update position and bounce off walls
		x += dx
		y += dy
		if x <= 20 || x >= 280 { dx = -dx }
		if y <= 20 || y >= 280 { dy = -dy }

		// Draw
		c.ClearScreen(colour.Black)
		c.ColourFilledCircle(x, y, 20, red)
		c.Render()
		return nil
	})
	dom.StartAnimationLoop(loopFn.Value)

	select {}
}
```

---

## Packages

| Package | Purpose |
|---------|---------|
| **`dom`** | Global handles (`Document`, `Body`, `Head`), element CRUD (`CreateElement`, `GetElementByID`, `AddElementTo`), visibility (`Hide`, `Show`), events (`AddEventListener`, `StartAnimationLoop`), CSS injection (`AddNewStyleElement`), canvas bootstrap (`CanvasCreate`, `CanvasGetContext`), `Context2D` wrapper, logging (`Log`, `Alert`). |
| **`html`** | Fluent tag builder: `Div`, `Span`, `Button`, `Input`, `Form`, `Table`, `H1`–`H6`, `Ul`/`Ol`/`Li`, `A`, `Img`, `Br`, `Hr`, and 20+ more. Chainable methods: `.Class()`, `.ID()`, `.Text()`, `.HTML()`, `.Attr()`, `.Child()`, `.On()`, `.AppendTo()`, `.Build()`. |
| **`canvas`** | Double-buffered pixel canvas: `NewCanvas`, `ClearScreen`, `SetColour`, `Line`, `Circle`, `FilledCircle`, `BorderCircle`, `Rectangle`, `FilledRectangle`, `Triangle`, `PutPixel`, `GetPixel`, `Render`, `GetContext2D`. Each drawing primitive also has a `Colour*` variant that accepts a colour directly. |
| **`colour`** | RGBA struct and presets (`White`, `Black`, `Empty`). Methods: `IsEmpty()`, `ConvertToGrayscale()`. Seedable Xorshift64 PRNG: `RandomColour()`, `Seed()`. Custom colours are constructed directly: `colour.Colour{R: 255, G: 0, B: 0, A: 255}`. |
| **`sound`** | Audio synthesis utilities: `FillClick` pre-renders a triangle-wave click buffer for browser playback. |

---

## Running the Included Demo

This repository includes a feature-rich, high-performance synthwave physics and audio sequencer demo.

### Prerequisites
* Go 1.26 or later
* Node.js (used by `run.sh` via `npx http-server` to serve the demo)

### Build & Run
1. Make the scripts executable (if not already):
   ```bash
   chmod +x build.sh run.sh
   ```
2. Build the Go WASM demo:
   ```bash
   ./build.sh
   ```
3. Start the local development server:
   ```bash
   ./run.sh
   ```
4. Open your browser and navigate to `http://localhost:9000`.

### Demo Features
* **Canvas One:** Interactive synthwave render running Bresenham line and circle algorithms.
* **Canvas Two:** Real-time elastic ball physics simulation updating at 60 FPS.
* **Canvas Three:** Neo-neon 16-step drum machine sequencer with AudioWorklet synthesizer loops.


## Examples

The [`examples/`](examples/) directory contains self-contained projects showing how to use `godom` in practice.

### `click-rect` — Interactive canvas with click detection

A 300×300 canvas with a blue rectangle centered in it. Clicking or tapping the rectangle toggles it between blue and white. Demonstrates:
- Setting up a canvas with `canvas.NewCanvas`
- Drawing with `ColourFilledRectangle`
- Registering a click handler with `dom.AddEventListenerByID`
- Reading mouse coordinates from the JS event (`args[0].Get("offsetX").Int()`)
- Hit-testing pixel coordinates against a rectangle

**Build and run:**
```bash
cd examples/click-rect
chmod +x build.sh && ./build.sh
npx http-server . -p 9001
```
Then open `http://localhost:9001`.

---

## Benchmarking

The `benchmarks/` directory contains the code used to validate key performance decisions in this library — for example, the exponential copy strategy behind `ClearScreen()` which is ~22× faster than a naive pixel loop at 1920×1080.

See [`benchmarks/README.md`](benchmarks/README.md) for the results and what they mean for your usage.

---

## Further Reading

| Document | What's in it |
|---|---|
| [`DEVELOPER.md`](DEVELOPER.md) | Full API reference, package internals, testing & benchmark workflow for contributors |
| [`benchmarks/README.md`](benchmarks/README.md) | Performance results and practical guidance for library users |
| [`examples/`](examples/) | Self-contained runnable example projects |
