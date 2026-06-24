# Godom — A Go WebAssembly DOM Utility Library

`godom` is a header-less, dependency-free wrapper around the browser DOM and Canvas 2D APIs written in pure Go. It targets standard WebAssembly (`GOOS=js GOARCH=wasm`) and operates using Go 1.26.4's standard `syscall/js` layer.

It was created to prove the feasibility of zero-allocation canvas updates and low-latency audio pre-rendering with DOM manipulation in Go WebAssembly.

---

## Prerequisites

- **Go 1.26** or later
- **Node.js** (for running tests under a WASM environment — see [`DEVELOPER.md`](DEVELOPER.md))

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

	// Block the main thread to keep event listeners alive
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
2. Load the WebAssembly module inside your `index.html` page:

```html
<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <script src="wasm_exec.js"></script>
  </head>
  <body>
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

The `canvas` package provides a pixel buffer and drawing primitives implemented in pure Go. Draw to the Go-side `Pixels` slice, then call `Render()` to blit to the browser in one shot:

```go
import (
	"github.com/ewaldhorn/godom/canvas"
	"github.com/ewaldhorn/godom/colour"
)

pixels := make([]byte, 320*240*4)
c := canvas.NewCanvas(320, 240, pixels, "app")

c.ClearScreen(colour.Black)
c.SetColour(colour.White)
c.Line(10, 10, 100, 100)
c.FilledCircle(160, 120, 50)
c.Rectangle(200, 50, 80, 60, 2)
c.Render()
```

Drawing primitives: `Line`, `Circle`, `FilledCircle`, `BorderCircle`, `Rectangle`, `FilledRectangle`, `Triangle`, `PutPixel`, `GetPixel`.

---

## Events

Register event listeners on any element with the fluent `On()` method:

```go
import (
	"syscall/js"

	"github.com/ewaldhorn/godom/dom"
	"github.com/ewaldhorn/godom/html"
)

btn := html.Button().
	Text("Click me").
	On("click", js.FuncOf(func(this js.Value, args []js.Value) any {
		dom.Alert("clicked!")
		return nil
	})).
	AppendTo(dom.Body)
```

The `dom` package also provides `AddEventListener` and `AddEventListenerByID` for lower-level event wiring, plus `StartAnimationLoop` for `requestAnimationFrame`-driven rendering.

---

## Packages

| Package | Purpose |
|---------|---------|
| **`dom`** | Global handles (`Document`, `Body`, `Head`), element CRUD (`CreateElement`, `GetElementByID`, `AddElementTo`), visibility (`Hide`, `Show`), events (`AddEventListener`, `StartAnimationLoop`), canvas bootstrap (`CanvasCreate`, `CanvasGetContext`), logging (`Log`, `Alert`). |
| **`html`** | Fluent tag builder: `Div`, `Span`, `Button`, `Input`, `Form`, `Table`, `H1`–`H6`, `Ul`/`Ol`/`Li`, `A`, `Img`, `Br`, `Hr`, and 20+ more. Chainable methods: `.Class()`, `.ID()`, `.Text()`, `.HTML()`, `.Attr()`, `.Child()`, `.On()`, `.AppendTo()`, `.Build()`. |
| **`canvas`** | Double-buffered pixel canvas: `NewCanvas`, `ClearScreen`, `SetColour`, `Line`, `Circle`, `FilledCircle`, `BorderCircle`, `Rectangle`, `FilledRectangle`, `Triangle`, `PutPixel`, `GetPixel`, `Render`, `GetContext2D`. |
| **`colour`** | RGBA struct and presets (`White`, `Black`, `Empty`), grayscale conversion, and a seedable Xorshift64 PRNG (`RandomColour`, `Seed`). |
| **`sound`** | Audio synthesis utilities: `FillClick` pre-renders a triangle-wave click buffer for browser playback. |

---

## Running the Included Demo

This repository includes a feature-rich, high-performance synthwave physics and audio sequencer demo.

### Prerequisites
* Go 1.26.4
* Node.js (for `npx http-server` serving)

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


## Benchmarking

While working on this library, I learned a thing or two, and I document that learning by leaving behind the benchmarking code I used to figure out a better way of doing something. Partly because I find it interesting but also so that people can point out where I missed something or failed to understand a nuance I didn't know existed.

---

## Further Reading

For testing instructions, package architecture, and benchmark details, see [`DEVELOPER.md`](DEVELOPER.md).
