package main

import (
	"math"

	"github.com/ewaldhorn/godom/dom"
)

// ------------------------------------------------------------------------------------------------
func drawTrunk(x, y, h float64, ctx dom.Context2D) {
	// (x, y) = bottom-center of trunk, h = trunk height
	baseHalf := h * 0.12 // half-width at the base
	topHalf := h * 0.06  // half-width at the top

	ctx.FillStyle("#775A29")
	ctx.BeginPath()

	// bottom-left
	ctx.MoveTo(x-baseHalf, y)

	// left edge curving inward toward the top
	ctx.QuadraticCurveTo(
		x-baseHalf*0.7, y-h/2, // control point
		x-topHalf, y-h, // top-left corner
	)

	// top edge
	ctx.LineTo(x+topHalf, y-h)

	// right edge curving back down
	ctx.QuadraticCurveTo(
		x+baseHalf*0.7, y-h/2, // control point
		x+baseHalf, y, // bottom-right
	)

	ctx.ClosePath()
	ctx.Fill()
}

// ------------------------------------------------------------------------------------------------
func drawLeaves(x, y, h float64, ctx dom.Context2D) {
	// The seed. Any change to h produces a different tree.
	r := newRNG(math.Float64bits(h))

	// Canopy geometry, expressed as fractions of h so it scales nicely.
	cx := x
	cy := y - h*1.05 // centre of canopy, just above the trunk's top
	spread := h * 0.55

	// Optional: same-h trees always pick the same shade.
	greens := []string{"#2E7D32", "#388E3C", "#43A047", "#4CAF50"}
	shade := greens[r.intn(len(greens))]

	n := 6 + r.intn(5) // 6..10 blobs, but stable for a given h
	for i := 0; i < n; i++ {
		// Uniform-ish point in a disc (sqrt for even area distribution).
		ang := r.float64() * 2 * math.Pi
		dist := math.Sqrt(r.float64()) * spread
		bx := cx + math.Cos(ang)*dist
		by := cy + math.Sin(ang)*dist*0.85 // squash vertically

		br := spread * (0.45 + r.float64()*0.40)

		ctx.FillStyle(shade)
		ctx.BeginPath()
		ctx.Arc(bx, by, br, 0, 2*math.Pi, false)
		ctx.Fill()
	}
}

// ------------------------------------------------------------------------------------------------
// drawTree draws a single tree at the xy coordinate, with the given height.
// NGL this took me a bit of experimenting to get right, I left comments for my future self.
// I eventually broke it into sub-functions, because it became a bit much to manage.
func drawTree(x, y, h float64, ctx dom.Context2D) {
	drawTrunk(x, y, h, ctx)
	drawLeaves(x, y, h, ctx)
}
