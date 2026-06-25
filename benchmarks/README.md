# Benchmarks — What This Means for You

This directory contains benchmarking code used to validate the performance design decisions made inside `godom`.

You don't need to run any of this to use the library. The results are documented here so you understand **why** certain APIs are implemented the way they are.

---

## `ClearScreen` — Why It's Fast

The most common canvas operation in any game or animation loop is clearing the screen. A naive implementation iterates every pixel individually:

```go
// Naive: iterates every pixel — slow at 1920×1080
for j := 0; j < len(pixels); j += 4 {
    pixels[j]   = r
    pixels[j+1] = g
    pixels[j+2] = b
    pixels[j+3] = a
}
```

`godom`'s `ClearScreen()` instead uses an **exponential copy doubling** strategy — write 4 bytes once, then double the filled region on each pass using `copy()`:

```go
pixels[0..3] = colour   // seed the first pixel
copy(pixels[4:8],   pixels[0:4])   // 1 → 2 pixels
copy(pixels[8:16],  pixels[0:8])   // 2 → 4 pixels
// ... doubles until the full buffer is filled
```

### Benchmark results

Clearing a `1920×1080` pixel buffer (measured under Node.js / WebAssembly):

| Approach | Time/op | Relative speed |
|---|---|---|
| Naive byte-by-byte loop | ~3.45 ms | baseline |
| `ClearScreen()` exponential copy | ~0.15 ms | **~22× faster** |

### Practical takeaway

At 60 FPS you have ~16.7 ms per frame. A naive clear alone would consume ~21% of that budget. With `ClearScreen()` it drops to less than 1%. **Always use `ClearScreen()` rather than zeroing the `Pixels` slice manually.**

---

## Running the Benchmarks Yourself

If you want to reproduce these results or benchmark your own usage, see the **Running Benchmarks** section in [DEVELOPER.md](../DEVELOPER.md).
