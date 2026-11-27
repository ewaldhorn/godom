package main

import (
	"github.com/ewaldhorn/godom/dom"
)

// ------------------------------------------------------------------------------------------------
var balloonPos = BalloonStatus{X: 200, Y: 200}

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
	ctx.Arc(10.0, 10.0, 60.0, 0.0, 180.0, false)
	ctx.Fill()
}

// ------------------------------------------------------------------------------------------------
func renderScene(ctx dom.Context2D) {
	drawGameBackground(ctx)
	drawBalloon(balloonPos, ctx)
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

	treeManager := NewTreeManager()

	// draw once per browser animation frame
	stopLoop := dom.StartLoop(func() {
		updateScene()
		treeManager.Update()
		renderScene(context2d)
		treeManager.Render(context2d)
	})
	defer stopLoop()

	// need to block this from exiting, otherwise wasm terminates our app
	select {}
}
