package main

import (
	"github.com/ewaldhorn/godom/dom"
)

// ------------------------------------------------------------------------------------------------
const (
	BASKET_WIDTH      = 100
	BASKET_HALF_WIDTH = BASKET_WIDTH / 2
	BASKET_HEIGHT     = 50
)

// ------------------------------------------------------------------------------------------------
func drawBalloon(x, y float64, ctx dom.Context2D) {

	// first we render the basket
	ctx.FillStyle("brown")
	ctx.FillRect(x-BASKET_HALF_WIDTH, y, BASKET_WIDTH, BASKET_HEIGHT)

	// now the balloon
	ctx.FillStyle("#28DD28")
	ctx.BeginPath()
	ctx.MoveTo(x, y)                               // middle of the balloon
	ctx.QuadraticCurveTo(x-90, y-70, x-70, y-170)  // bottom-left
	ctx.QuadraticCurveTo(x-45, y-250, x, y-250)    // top-left to apex
	ctx.QuadraticCurveTo(x+45, y-250, x+70, y-170) // apex to right
	ctx.QuadraticCurveTo(x+90, y-70, x, y)         // right down to middle
	ctx.ClosePath()
	ctx.Fill()
}
