package benchmarks

import (
	"testing"

	"godom/canvas"
	"godom/colour"
)

// ------------------------------------------------------------------------------------------------
// Colour matches the internal RGBA structure.
var testColour = colour.Colour{R: 100, G: 150, B: 200, A: 255}

// ------------------------------------------------------------------------------------------------
// BenchmarkClearScreenNaive benchmarks clearing the canvas using the naive byte-by-byte loop.
func BenchmarkClearScreenNaive(b *testing.B) {
	pixels := make([]byte, 1920*1080*4)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(pixels); j += 4 {
			pixels[j] = testColour.R
			pixels[j+1] = testColour.G
			pixels[j+2] = testColour.B
			pixels[j+3] = testColour.A
		}
	}
}

// ------------------------------------------------------------------------------------------------
// BenchmarkClearScreenOptimized benchmarks clearing the canvas using the optimized exponential 
// copy loop.
func BenchmarkClearScreenOptimized(b *testing.B) {
	pixels := make([]byte, 1920*1080*4)
	c := &canvas.Canvas{
		Width:  1920,
		Height: 1080,
		Pixels: pixels,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.ClearScreen(testColour)
	}
}
