package canvas

import (
	"github.com/ewaldhorn/godom/colour"
)

// ------------------------------------------------------------------------------------------------
// pixelOffset returns the offset in the pixels slice for the pixel at (x, y).
// It returns false if the coordinates are out of bounds.
func (c *Canvas) pixelOffset(x, y int32) (int, bool) {
	if x < 0 || y < 0 {
		return 0, false
	}
	ux := int(x)
	uy := int(y)
	if ux >= c.width || uy >= c.height {
		return 0, false
	}
	return (uy*c.width + ux) * 4, true
}

// ------------------------------------------------------------------------------------------------
// PutPixel draws a pixel at (x, y) using the active drawing color.
func (c *Canvas) PutPixel(x, y int32) {
	off, ok := c.pixelOffset(x, y)
	if !ok {
		return
	}
	c.pixels[off] = c.activeColour.R
	c.pixels[off+1] = c.activeColour.G
	c.pixels[off+2] = c.activeColour.B
	c.pixels[off+3] = c.activeColour.A
}

// ------------------------------------------------------------------------------------------------
// ColourPutPixel draws a pixel at (x, y) with the specified color.
func (c *Canvas) ColourPutPixel(x, y int32, col colour.Colour) {
	off, ok := c.pixelOffset(x, y)
	if !ok {
		return
	}
	c.pixels[off] = col.R
	c.pixels[off+1] = col.G
	c.pixels[off+2] = col.B
	c.pixels[off+3] = col.A
}

// ------------------------------------------------------------------------------------------------
// GetPixel retrieves the color of the pixel at (x, y).
// It returns false if the coordinates are out of bounds.
func (c *Canvas) GetPixel(x, y int32) (colour.Colour, bool) {
	off, ok := c.pixelOffset(x, y)
	if !ok {
		return colour.Empty, false
	}
	return colour.Colour{
		R: c.pixels[off],
		G: c.pixels[off+1],
		B: c.pixels[off+2],
		A: c.pixels[off+3],
	}, true
}