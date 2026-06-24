# Godom Benchmarks

This directory contains benchmarking code for Godom library operations.

Because Godom interacts directly with the browser's DOM and JavaScript execution via `syscall/js`, the library must be compiled and executed within a WebAssembly target.

## Running the Benchmarks

To run these benchmarks, you will need **Node.js** installed on your system.

Execute the following command from the project root:

```bash
GOOS=js GOARCH=wasm go test -exec="node $(go env GOROOT)/lib/wasm/wasm_exec_node.js" -bench=. -benchmem ./benchmarks/...
```

### Explanation of the Command

- `GOOS=js GOARCH=wasm`: Tells the Go toolchain to target WebAssembly.
- `-exec="node ..."`: Tells `go test` to use Node.js and the official Go WASM wrapper to execute the compiled test binary instead of running it as a native host binary.
- `-bench=.`: Runs all benchmarks.
- `-benchmem`: Includes memory allocation statistics.

---

## ClearScreen Benchmark Results

Clearing a standard `1920x1080` pixel buffer:

- **Naive byte-by-byte loop**: `~3.45 ms/op`
- **Optimized exponential copy loop**: `~0.15 ms/op` (runs **22x+ faster**!)

Result may vary, depending on your CPU architecture and type, but it should be a lot faster to do the native memory copy instead of iterating through the pixels and setting them individually.
