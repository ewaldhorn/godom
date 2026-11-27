package main

import (
	"github.com/ewaldhorn/godom/dom"
)

// ------------------------------------------------------------------------------------------------
type BalloonStatus struct {
	X, Y     float64
	IsRising bool
}

// ------------------------------------------------------------------------------------------------
func (b *BalloonStatus) Update() {
	if b.IsRising {
		b.Y -= 1
	} else {
		b.Y += 1
	}
}

// ------------------------------------------------------------------------------------------------
const (
	BASKET_WIDTH      = 80
	BASKET_HALF_WIDTH = BASKET_WIDTH / 2
	BASKET_HEIGHT     = 40
)

// ------------------------------------------------------------------------------------------------
func drawBalloon(pos BalloonStatus, ctx dom.Context2D) {
	// first we render the basket
	ctx.FillStyle("brown")
	ctx.FillRect(pos.X-BASKET_HALF_WIDTH, pos.Y, BASKET_WIDTH, BASKET_HEIGHT)

	// now the balloon
	ctx.FillStyle("#28DD28")
	ctx.BeginPath()
	ctx.MoveTo(pos.X, pos.Y)                                       // middle of the balloon
	ctx.QuadraticCurveTo(pos.X-60, pos.Y-50, pos.X-50, pos.Y-100)  // bottom-left
	ctx.QuadraticCurveTo(pos.X-25, pos.Y-170, pos.X, pos.Y-170)    // top-left to apex
	ctx.QuadraticCurveTo(pos.X+25, pos.Y-170, pos.X+50, pos.Y-100) // apex to right
	ctx.QuadraticCurveTo(pos.X+60, pos.Y-50, pos.X, pos.Y)         // right down to middle
	ctx.ClosePath()
	ctx.Fill()
}
