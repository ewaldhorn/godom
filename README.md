# Godom — A Go WebAssembly DOM Utility Library

`godom` is a header-less, dependency-free wrapper around the browser DOM and Canvas 2D APIs written in pure Go. It targets standard WebAssembly (`GOOS=js GOARCH=wasm`) and operates using Go 1.26.4's standard `syscall/js` layer.

It was created to prove the feasibility of replicating Zig-based DOM manipulation libraries (like `zigdom`) in Go with zero-allocation canvas updates and low-latency audio pre-rendering.

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

1. Copy `godom.js` and `synth-worklet.js` from the repository's `docs/` folder into your web deployment directory.
2. Locate and copy `wasm_exec.js` from your Go installation:
   ```bash
   cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" .
   ```
3. Load the WebAssembly module inside your `index.html` page:

```html
<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <script src="wasm_exec.js"></script>
    <script src="godom.js"></script>
  </head>
  <body>
    <div id="loading">Loading App...</div>
    <script>
      GoDom.instantiate("app.wasm").then(() => {
        document.getElementById("loading").remove();
      }).catch(err => {
        console.error("Failed to load WASM:", err);
      });
    </script>
  </body>
</html>
```

---

## Running the Included Demo

This repository includes a feature-rich, high-performance synthwave physics and audio sequencer demo (migrated from `zigdom`).

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