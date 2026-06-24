package canvas

import (
	"syscall/js"
	"testing"

	"github.com/ewaldhorn/godom/colour"
)

func init() {
	// Bootstrap global mocks for testing under Node.js WASM runner.
	js.Global().Get("eval").Invoke(`
		globalThis.ImageData = class ImageData {
			constructor(data, width, height) {
				this.data = data;
				this.width = width;
				this.height = height;
			}
		};

		globalThis.document = {
			getElementById: function(id) {
				return {
					id: id,
					appendChild: function(child) {}
				};
			},
			createElement: function(tag) {
				let el = {
					tagName: tag,
					width: 0,
					height: 0,
					getContext: function(type) {
						return {
							putImageData: function(imgData, x, y) {}
						};
					},
					setAttribute: function(k, v) { this[k] = v; }
				};
				return el;
			}
		};
	`)
}

func TestNewCanvas(t *testing.T) {
	pixels := make([]byte, 100*100*4)
	c := NewCanvas(100, 100, pixels, "canvas-parent")

	if c.Width != 100 || c.Height != 100 {
		t.Errorf("expected 100x100 canvas, got %dx%d", c.Width, c.Height)
	}

	if len(c.Pixels) != 10000*4 {
		t.Errorf("expected pixel buffer of size 40000, got %d", len(c.Pixels))
	}

	if c.GetColour() != colour.Black {
		t.Errorf("expected default active color to be Black, got %+v", c.GetColour())
	}
}

func TestPixelOperations(t *testing.T) {
	pixels := make([]byte, 10*10*4)
	c := NewCanvas(10, 10, pixels, "parent")

	// Set colour
	c.SetColour(colour.White)
	if c.GetColour() != colour.White {
		t.Errorf("expected active color to be White, got %+v", c.GetColour())
	}

	// Put pixel
	c.PutPixel(2, 3)
	col, ok := c.GetPixel(2, 3)
	if !ok || col != colour.White {
		t.Errorf("expected pixel at (2,3) to be White, got %+v (ok=%t)", col, ok)
	}

	// Out of bounds check
	c.PutPixel(-1, 5)
	c.PutPixel(10, 5)
	c.PutPixel(5, -1)
	c.PutPixel(5, 10)

	_, ok = c.GetPixel(10, 5)
	if ok {
		t.Error("expected out of bounds pixel get to return ok=false")
	}

	// ColourPutPixel
	c.ColourPutPixel(4, 5, colour.Colour{R: 128, G: 64, B: 32, A: 255})
	col, ok = c.GetPixel(4, 5)
	if !ok || col.R != 128 || col.G != 64 || col.B != 32 {
		t.Errorf("expected custom color at (4,5), got %+v", col)
	}
}

func TestClearScreen(t *testing.T) {
	pixels := make([]byte, 5*5*4)
	c := NewCanvas(5, 5, pixels, "parent")

	c.ClearScreen(colour.White)

	for i := 0; i < len(pixels); i += 4 {
		if pixels[i] != 255 || pixels[i+1] != 255 || pixels[i+2] != 255 || pixels[i+3] != 255 {
			t.Errorf("expected cleared pixel at index %d to be White, got R=%d G=%d B=%d A=%d", i, pixels[i], pixels[i+1], pixels[i+2], pixels[i+3])
		}
	}
}

func TestDrawingShapes(t *testing.T) {
	pixels := make([]byte, 20*20*4)
	c := NewCanvas(20, 20, pixels, "parent")

	// Draw Line
	c.ColourLine(0, 0, 19, 19, colour.White)
	col, _ := c.GetPixel(10, 10)
	if col != colour.White {
		t.Error("expected diagonal line pixel at (10,10) to be White")
	}

	// Draw Rectangle
	c.ColourRectangle(2, 2, 10, 10, 1, colour.White)
	col, _ = c.GetPixel(2, 2)
	if col != colour.White {
		t.Error("expected rectangle border pixel at (2,2) to be White")
	}

	// Draw FilledRectangle
	c.ColourFilledRectangle(5, 5, 3, 3, colour.White)
	col, _ = c.GetPixel(6, 6)
	if col != colour.White {
		t.Error("expected filled rectangle internal pixel to be White")
	}

	// Draw Circle
	c.ColourCircle(10, 10, 5, colour.White)
	// Outer point (15, 10) should be set
	col, _ = c.GetPixel(15, 10)
	if col != colour.White {
		t.Error("expected circle outline pixel to be White")
	}

	// Draw FilledCircle
	c.ColourFilledCircle(10, 10, 4, colour.White)
	col, _ = c.GetPixel(10, 10)
	if col != colour.White {
		t.Error("expected filled circle center to be White")
	}

	// Draw BorderCircle
	c.ColourBorderCircle(10, 10, 6, 2, colour.White)

	// Draw Triangle
	c.Triangle(Point{X: 0, Y: 0}, Point{X: 19, Y: 0}, Point{X: 10, Y: 10})

	// Call Render to verify it executes without errors
	c.Render()

	// Verify GetContext2D returns non-nil context wrapper
	ctx := c.GetContext2D()
	if js.Value(ctx.Ctx).IsUndefined() {
		t.Error("expected context handle to be defined")
	}
}
