// Package colour provides structures, constants, and utilities for working with RGBA colors.
package colour

// ------------------------------------------------------------------------------------------------
// Colour represents an RGBA color structure.
type Colour struct {
	R, G, B, A uint8
}

// ------------------------------------------------------------------------------------------------
// Predefined colors.
var (
	White = Colour{R: 255, G: 255, B: 255, A: 255}
	Black = Colour{R: 0, G: 0, B: 0, A: 255}
	Empty = Colour{R: 0, G: 0, B: 0, A: 0}
)

// ------------------------------------------------------------------------------------------------
// IsEmpty returns true if the colour is fully transparent (all RGBA components are zero).
func (c Colour) IsEmpty() bool {
	return c.R == 0 && c.G == 0 && c.B == 0 && c.A == 0
}

// ------------------------------------------------------------------------------------------------
// ConvertToGrayscale converts the color to grayscale in-place using luminance weighting.
func (c *Colour) ConvertToGrayscale() {
	r := float32(c.R)
	g := float32(c.G)
	b := float32(c.B)
	shade := uint8(0.299*r + 0.587*g + 0.114*b)
	c.R = shade
	c.G = shade
	c.B = shade
}

// ------------------------------------------------------------------------------------------------
// rngState is the current state of the seedable Xorshift64 PRNG.
var rngState uint64 = 1337

// ------------------------------------------------------------------------------------------------
// Seed re-seeds the color PRNG with the specified seed value.
// The seed value s must be non-zero; a value of zero is silently ignored.
func Seed(s uint64) {
	if s != 0 {
		rngState = s
	}
}

// ------------------------------------------------------------------------------------------------
// nextRandom generates the next pseudorandom 32-bit unsigned integer using Xorshift64.
func nextRandom() uint32 {
	rngState ^= rngState << 13
	rngState ^= rngState >> 7
	rngState ^= rngState << 17
	return uint32(rngState)
}

// ------------------------------------------------------------------------------------------------
// RandomColour returns a random fully-opaque color.
func RandomColour() Colour {
	return Colour{
		R: uint8(nextRandom()),
		G: uint8(nextRandom()),
		B: uint8(nextRandom()),
		A: 255,
	}
}
