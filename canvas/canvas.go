// Package canvas provides a 2D pixel buffer and drawing operations for HTML5 canvas.
package canvas

import (
	"math"
	"strconv"
	"syscall/js"

	"github.com/ewaldhorn/godom/colour"
	"github.com/ewaldhorn/godom/dom"
)

// ------------------------------------------------------------------------------------------------
// Canvas represents the in-memory pixel buffer and its associated HTML canvas.
type Canvas struct {
	width        int
	height       int
	pixels       []byte
	activeColour colour.Colour
	canvasHandle dom.Handle
	ctxHandle    dom.Handle

	// Pre-allocated JS arrays to speed up rendering without triggering GC.
	jsClampedArray js.Value
	jsImgData      js.Value
}

// ================================================================================================
// Constructors
// ================================================================================================

// ------------------------------------------------------------------------------------------------
// validateCanvasArgs panics with a clear message if the declared dimensions or the pixel buffer
// are inconsistent. In-browser, new ImageData would otherwise throw an obscure RangeError;
// offscreen canvases would silently misbehave instead.
func validateCanvasArgs(width, height int, pixels []byte) {
	if width <= 0 || height <= 0 {
		panic("godom: canvas: width and height must be positive, got " + strconv.Itoa(width) + "x" + strconv.Itoa(height))
	}
	if int64(len(pixels)) != int64(width)*int64(height)*4 {
		panic("godom: canvas: pixel buffer length " + strconv.Itoa(len(pixels)) + " does not match " + strconv.Itoa(width) + "x" + strconv.Itoa(height) + " (want " + strconv.FormatInt(int64(width)*int64(height)*4, 10) + ")")
	}
}

// ------------------------------------------------------------------------------------------------
// NewCanvas creates a canvas in-memory and associates it with a new canvas element appended to
// the DOM element identified by parentID.
func NewCanvas(width, height int, pixels []byte, parentID string) *Canvas {
	validateCanvasArgs(width, height, pixels)
	parentHandle := dom.GetElementByID(parentID)
	if !parentHandle.IsValid() {
		panic("godom: NewCanvas: no DOM element with id " + parentID)
	}
	canvasH := dom.CanvasCreate(parentHandle, width, height)
	ctxH := dom.CanvasGetContext(canvasH)

	// Pre-allocate typing variables in JavaScript.
	jsClampedArray := js.Global().Get("Uint8ClampedArray").New(len(pixels))
	jsImgData := js.Global().Get("ImageData").New(jsClampedArray, width, height)

	return &Canvas{
		width:          width,
		height:         height,
		pixels:         pixels,
		activeColour:   colour.Black,
		canvasHandle:   canvasH,
		ctxHandle:      ctxH,
		jsClampedArray: jsClampedArray,
		jsImgData:      jsImgData,
	}
}

// ------------------------------------------------------------------------------------------------
// NewOffscreenCanvas creates a Canvas backed only by a Go pixel buffer, with no DOM or JS
// objects attached. Useful for benchmarking and unit testing. Calling Render() on an offscreen
// canvas is a no-op.
func NewOffscreenCanvas(width, height int, pixels []byte) *Canvas {
	validateCanvasArgs(width, height, pixels)
	return &Canvas{
		width:        width,
		height:       height,
		pixels:       pixels,
		activeColour: colour.Black,
	}
}

// ================================================================================================
// Accessors
// ================================================================================================

// ------------------------------------------------------------------------------------------------
// Width returns the canvas width in pixels.
func (c *Canvas) Width() int { return c.width }

// ------------------------------------------------------------------------------------------------
// Height returns the canvas height in pixels.
func (c *Canvas) Height() int { return c.height }

// ------------------------------------------------------------------------------------------------
// Pixels returns the underlying pixel buffer slice.
// The slice is shared with the Canvas — writes to it are reflected immediately.
func (c *Canvas) Pixels() []byte { return c.pixels }

// ------------------------------------------------------------------------------------------------
// CanvasHandle returns the DOM handle for the underlying HTML canvas element.
func (c *Canvas) CanvasHandle() dom.Handle { return c.canvasHandle }

// ================================================================================================
// Core rendering
// ================================================================================================

// ------------------------------------------------------------------------------------------------
// Render blits the Go-side pixel buffer into JavaScript and updates the canvas.
// It is a no-op if the canvas was created with NewOffscreenCanvas.
func (c *Canvas) Render() {
	if c.jsClampedArray.IsUndefined() {
		return
	}
	js.CopyBytesToJS(c.jsClampedArray, c.pixels)
	js.Value(c.ctxHandle).Call("putImageData", c.jsImgData, 0, 0)
}

// ------------------------------------------------------------------------------------------------
// ClearScreen fills the entire canvas with the specified color.
func (c *Canvas) ClearScreen(col colour.Colour) {
	if len(c.pixels) < 4 {
		return
	}
	c.pixels[0] = col.R
	c.pixels[1] = col.G
	c.pixels[2] = col.B
	c.pixels[3] = col.A

	for bp := 4; bp < len(c.pixels); bp *= 2 {
		copy(c.pixels[bp:], c.pixels[:bp])
	}
}

// ------------------------------------------------------------------------------------------------
// SetColour sets the active drawing color.
func (c *Canvas) SetColour(col colour.Colour) {
	c.activeColour = col
}

// ------------------------------------------------------------------------------------------------
// GetColour retrieves the active drawing color.
func (c *Canvas) GetColour() colour.Colour {
	return c.activeColour
}

// ================================================================================================
// Drawing primitives — plain forms use the active colour; Colour* forms accept a colour
// argument inline and do NOT mutate the active colour.
// ================================================================================================

// ------------------------------------------------------------------------------------------------
// clipLine clips a 2D line segment to the canvas rectangle [0, width-1] x [0, height-1] using
// the Cohen-Sutherland algorithm. Returns false if the line is completely outside.
func (c *Canvas) clipLine(x1, y1, x2, y2 int32) (int32, int32, int32, int32, bool) {
	const (
		inside = 0      // 0000
		left   = 1 << 0 // 0001
		right  = 1 << 1 // 0010
		bottom = 1 << 2 // 0100
		top    = 1 << 3 // 1000
	)

	xMin, yMin := int32(0), int32(0)
	xMax, yMax := int32(c.width-1), int32(c.height-1)

	computeOutCode := func(x, y int32) int {
		code := inside
		if x < xMin {
			code |= left
		} else if x > xMax {
			code |= right
		}
		if y < yMin {
			code |= bottom
		} else if y > yMax {
			code |= top
		}
		return code
	}

	code1 := computeOutCode(x1, y1)
	code2 := computeOutCode(x2, y2)

	for {
		if (code1 | code2) == 0 {
			// Both points inside
			return x1, y1, x2, y2, true
		}
		if (code1 & code2) != 0 {
			// Both points share an outside zone: trivially reject
			return 0, 0, 0, 0, false
		}

		// At least one point is outside
		var codeOut int
		if code1 != 0 {
			codeOut = code1
		} else {
			codeOut = code2
		}

		var x, y int32
		dx := int64(x2 - x1)
		dy := int64(y2 - y1)

		if (codeOut & top) != 0 {
			x = int32(int64(x1) + dx*int64(yMax-y1)/dy)
			y = yMax
		} else if (codeOut & bottom) != 0 {
			x = int32(int64(x1) + dx*int64(yMin-y1)/dy)
			y = yMin
		} else if (codeOut & right) != 0 {
			y = int32(int64(y1) + dy*int64(xMax-x1)/dx)
			x = xMax
		} else if (codeOut & left) != 0 {
			y = int32(int64(y1) + dy*int64(xMin-x1)/dx)
			x = xMin
		}

		if codeOut == code1 {
			x1, y1 = x, y
			code1 = computeOutCode(x1, y1)
		} else {
			x2, y2 = x, y
			code2 = computeOutCode(x2, y2)
		}
	}
}

// ------------------------------------------------------------------------------------------------
// Line draws a 1-pixel line from (x1, y1) to (x2, y2) using Bresenham's algorithm and the active
// color. Lines extending outside the canvas boundary are clipped first.
func (c *Canvas) Line(x1, y1, x2, y2 int32) {
	var ok bool
	x1, y1, x2, y2, ok = c.clipLine(x1, y1, x2, y2)
	if !ok {
		return
	}

	diffX := abs(x2 - x1)
	diffY := abs(y2 - y1)
	var slopeX, slopeY int32
	if x1 < x2 {
		slopeX = 1
	} else {
		slopeX = -1
	}
	if y1 < y2 {
		slopeY = 1
	} else {
		slopeY = -1
	}
	err := diffX - diffY
	for {
		c.PutPixel(x1, y1)
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -diffY {
			err -= diffY
			x1 += slopeX
		}
		if e2 < diffX {
			err += diffX
			y1 += slopeY
		}
	}
}

// ------------------------------------------------------------------------------------------------
// ColourLine draws a line from (x1, y1) to (x2, y2) with the specified color.
func (c *Canvas) ColourLine(x1, y1, x2, y2 int32, col colour.Colour) {
	old := c.activeColour
	c.activeColour = col
	c.Line(x1, y1, x2, y2)
	c.activeColour = old
}

// ------------------------------------------------------------------------------------------------
// LinePoint draws a line between two Point structures using the active color.
func (c *Canvas) LinePoint(p1, p2 Point) {
	c.Line(p1.X, p1.Y, p2.X, p2.Y)
}

// ------------------------------------------------------------------------------------------------
// ColourLinePoint draws a line between two Points with the specified color.
func (c *Canvas) ColourLinePoint(p1, p2 Point, col colour.Colour) {
	old := c.activeColour
	c.activeColour = col
	c.LinePoint(p1, p2)
	c.activeColour = old
}

// ------------------------------------------------------------------------------------------------
// Circle draws an outline circle at (midX, midY) with the specified radius using the active color.
// Uses the Bresenham midpoint circle algorithm — integer arithmetic only, no trigonometry.
func (c *Canvas) Circle(midX, midY, radius int32) {
	if radius <= 0 {
		return
	}
	x := radius
	y := int32(0)
	err := 1 - radius
	for x >= y {
		c.PutPixel(midX+x, midY+y)
		c.PutPixel(midX-x, midY+y)
		c.PutPixel(midX+x, midY-y)
		c.PutPixel(midX-x, midY-y)
		c.PutPixel(midX+y, midY+x)
		c.PutPixel(midX-y, midY+x)
		c.PutPixel(midX+y, midY-x)
		c.PutPixel(midX-y, midY-x)
		y++
		if err <= 0 {
			err += 2*y + 1
		} else {
			x--
			err += 2*(y-x) + 1
		}
	}
}

// ------------------------------------------------------------------------------------------------
// ColourCircle draws an outline circle at (midX, midY) with the specified radius and color.
func (c *Canvas) ColourCircle(midX, midY, radius int32, col colour.Colour) {
	old := c.activeColour
	c.activeColour = col
	c.Circle(midX, midY, radius)
	c.activeColour = old
}

// ------------------------------------------------------------------------------------------------
// FilledCircle draws a filled circle at (midX, midY) with the specified radius using the active
// color.
func (c *Canvas) FilledCircle(midX, midY, radius int32) {
	if radius <= 0 {
		return
	}
	r2 := int64(radius) * int64(radius)
	for dy := -radius; dy <= radius; dy++ {
		dy64 := int64(dy)
		chord := int32(math.Sqrt(float64(r2 - dy64*dy64)))
		for dx := -chord; dx <= chord; dx++ {
			c.PutPixel(midX+dx, midY+dy)
		}
	}
}

// ------------------------------------------------------------------------------------------------
// ColourFilledCircle draws a filled circle at (midX, midY) with the specified radius and color.
func (c *Canvas) ColourFilledCircle(midX, midY, radius int32, col colour.Colour) {
	old := c.activeColour
	c.activeColour = col
	c.FilledCircle(midX, midY, radius)
	c.activeColour = old
}

// ------------------------------------------------------------------------------------------------
// BorderCircle draws a ring (annulus) at (midX, midY) with the specified radius and border width
// using the active color.
func (c *Canvas) BorderCircle(midX, midY, radius, borderWidth int32) {
	if radius <= 0 {
		return
	}
	if borderWidth <= 0 {
		return
	}
	innerRadius := radius - borderWidth
	if innerRadius <= 0 {
		c.FilledCircle(midX, midY, radius)
		return
	}
	outerR2 := int64(radius) * int64(radius)
	innerR2 := int64(innerRadius) * int64(innerRadius)
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			d2 := int64(dx)*int64(dx) + int64(dy)*int64(dy)
			if d2 <= outerR2 && d2 > innerR2 {
				c.PutPixel(midX+dx, midY+dy)
			}
		}
	}
}

// ------------------------------------------------------------------------------------------------
// ColourBorderCircle draws a ring (annulus) at (midX, midY) with the specified radius, border
// width, and color.
func (c *Canvas) ColourBorderCircle(midX, midY, radius, borderWidth int32, col colour.Colour) {
	old := c.activeColour
	c.activeColour = col
	c.BorderCircle(midX, midY, radius, borderWidth)
	c.activeColour = old
}

// ------------------------------------------------------------------------------------------------
// FilledRectangle draws a filled rectangle starting at (xStart, yStart) with the specified width
// and height using the active color.
func (c *Canvas) FilledRectangle(xStart, yStart, width, height int32) {
	for y := int32(0); y < height; y++ {
		for x := int32(0); x < width; x++ {
			c.PutPixel(xStart+x, yStart+y)
		}
	}
}

// ------------------------------------------------------------------------------------------------
// ColourFilledRectangle draws a filled rectangle starting at (xStart, yStart) with the specified
// width, height, and color.
func (c *Canvas) ColourFilledRectangle(xStart, yStart, width, height int32, col colour.Colour) {
	old := c.activeColour
	c.activeColour = col
	c.FilledRectangle(xStart, yStart, width, height)
	c.activeColour = old
}

// ------------------------------------------------------------------------------------------------
// rectangleOutline draws the 1-pixel outline of a rectangle starting at (xStart, yStart) with the
// specified width and height using the active color.
func (c *Canvas) rectangleOutline(xStart, yStart, width, height int32) {
	x2 := xStart + width - 1
	y2 := yStart + height - 1
	c.Line(xStart, yStart, x2, yStart)
	c.Line(x2, yStart, x2, y2)
	c.Line(xStart, y2, x2, y2)
	c.Line(xStart, yStart, xStart, y2)
}

// ------------------------------------------------------------------------------------------------
// Rectangle draws an outline rectangle starting at (xStart, yStart) with the specified width,
// height, and border thickness using the active color.
func (c *Canvas) Rectangle(xStart, yStart, width, height, thickness int32) {
	for t := int32(0); t < thickness; t++ {
		if width-t*2 <= 0 || height-t*2 <= 0 {
			break
		}
		c.rectangleOutline(xStart+t, yStart+t, width-t*2, height-t*2)
	}
}

// ------------------------------------------------------------------------------------------------
// ColourRectangle draws an outline rectangle starting at (xStart, yStart) with the specified width,
// height, thickness, and color.
func (c *Canvas) ColourRectangle(xStart, yStart, width, height, thickness int32, col colour.Colour) {
	old := c.activeColour
	c.activeColour = col
	c.Rectangle(xStart, yStart, width, height, thickness)
	c.activeColour = old
}

// ------------------------------------------------------------------------------------------------
// GetContext2D returns a Context2D wrapper for the HTML canvas context.
// Panics if called on a canvas created with NewOffscreenCanvas.
func (c *Canvas) GetContext2D() dom.Context2D {
	if js.Value(c.ctxHandle).IsUndefined() {
		panic("godom: GetContext2D called on an offscreen canvas; only valid for canvases created with NewCanvas")
	}
	return dom.Context2D{Ctx: c.ctxHandle}
}
