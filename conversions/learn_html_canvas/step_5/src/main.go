package main

import (
	"github.com/ewaldhorn/godom/dom"
)

// ------------------------------------------------------------------------------------------------
var balloonPos = BalloonStatus{X: 350, Y: 500}

// ------------------------------------------------------------------------------------------------
const (
	canvasWidth  = 800
	canvasHeight = 600
)

// ------------------------------------------------------------------------------------------------
func drawGameBackground(ctx dom.Context2D) {
	ctx.ClearRect(0, 0, canvasWidth, canvasHeight)
	ctx.FillStyle("yellow")
	ctx.BeginPath()
	ctx.Arc(10.0, 10.0, 150.0, 0.0, 180.0, false)
	ctx.Fill()
}

// ------------------------------------------------------------------------------------------------
func renderScene(ctx dom.Context2D) {
	drawGameBackground(ctx)
	drawBalloon(balloonPos, ctx)
	drawTree(200, 500, 100, ctx)
}

// ------------------------------------------------------------------------------------------------
func updateScene() {
	balloonPos.Update()
}

// ------------------------------------------------------------------------------------------------
func main() {
	// First step is to init the dom
	dom.Init()

	// create our canvas and set up click interaction
	myCanvas, context2d := createAndConfigureHTMLCanvas()
	registerCanvasEventHandlers(myCanvas)

	// draw once per browser animation frame
	stopLoop := dom.StartLoop(func() {
		updateScene()
		renderScene(context2d)
	})
	defer stopLoop()

	// need to block this from exiting, otherwise wasm terminates our app
	select {}
}
