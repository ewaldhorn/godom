package canvas

import (
	"godom/colour"
)

// ------------------------------------------------------------------------------------------------
// pixelOffset returns the offset in the Pixels slice for the pixel at (x, y).
// It returns false if the coordinates are out of bounds.
func (c *Canvas) pixelOffset(x, y int32) (int, bool) {
	if x < 0 || y < 0 {
		return 0, false
	}
	ux := int(x)
	uy := int(y)
	if ux >= c.Width || uy >= c.Height {
		return 0, false
	}
	return (uy*c.Width + ux) * 4, true
}

// ------------------------------------------------------------------------------------------------
// PutPixel draws a pixel at (x, y) using the active drawing color.
func (c *Canvas) PutPixel(x, y int32) {
	off, ok := c.pixelOffset(x, y)
	if !ok {
		return
	}
	c.Pixels[off] = c.ActiveColour.R
	c.Pixels[off+1] = c.ActiveColour.G
	c.Pixels[off+2] = c.ActiveColour.B
	c.Pixels[off+3] = c.ActiveColour.A
}

// ------------------------------------------------------------------------------------------------
// ColourPutPixel draws a pixel at (x, y) with the specified color.
func (c *Canvas) ColourPutPixel(x, y int32, col colour.Colour) {
	off, ok := c.pixelOffset(x, y)
	if !ok {
		return
	}
	c.Pixels[off] = col.R
	c.Pixels[off+1] = col.G
	c.Pixels[off+2] = col.B
	c.Pixels[off+3] = col.A
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
		R: c.Pixels[off],
		G: c.Pixels[off+1],
		B: c.Pixels[off+2],
		A: c.Pixels[off+3],
	}, true
}
