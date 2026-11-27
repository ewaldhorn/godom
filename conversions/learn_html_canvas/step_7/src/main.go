package main

import "github.com/ewaldhorn/godom/dom"

// ------------------------------------------------------------------------------------------------
const (
	canvasWidth  = 800
	canvasHeight = 600
)

// ------------------------------------------------------------------------------------------------
func main() {
	// First step is to init the dom
	dom.Init()

	// create our canvas and set up click interaction
	myCanvas, context2d := createAndConfigureHTMLCanvas()
	game := NewGame(context2d)
	registerCanvasEventHandlers(myCanvas, game)

	// draw once per browser animation frame
	stopLoop := dom.StartLoop(func() {
		game.Tick()
	})
	defer stopLoop()

	// need to block this from exiting, otherwise wasm terminates our app
	select {}
}
