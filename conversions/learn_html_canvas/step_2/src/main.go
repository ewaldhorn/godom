package main

import (
	"github.com/ewaldhorn/godom/dom"
	"github.com/ewaldhorn/godom/html"
)

// ------------------------------------------------------------------------------------------------
func getContext2DfromCanvasHandle(canvasHandle dom.Handle) dom.Context2D {
	if !canvasHandle.IsValid() {
		panic("godom: getContext2DfromCanvasHandle received an invalid canvas handle")
	}

	ctx := dom.CanvasGetContext(canvasHandle)
	if !ctx.IsValid() {
		panic("godom: getContext2DfromCanvasHandle failed to retrieve Canvas2D context")
	}

	return dom.Context2D{Ctx: ctx}
}

// ------------------------------------------------------------------------------------------------
func createAndConfigureHTMLCanvas() dom.Context2D {
	// create a new Div container for our canvas
	canvasContainer := html.Div().Class("canvas-container")

	// create the actual canvas, set the attributes and background colour
	myCanvas := html.Canvas()
	myCanvas.Attr("width", "800")
	myCanvas.Attr("height", "600")
	myCanvas.Style("background-color", "blue")

	// apped the canvas to the container, then the container to the html body
	myCanvas.AppendTo(canvasContainer.Handle)
	canvasContainer.AppendTo(dom.Body)

	// return the Context2D handle, which is what we need to draw to the canvas
	return getContext2DfromCanvasHandle(myCanvas.Handle)
}

// ------------------------------------------------------------------------------------------------
func drawToTheCanvas(ctx dom.Context2D) {
	ctx.FillStyle("yellow")
	// ctx.BeginPath()
	ctx.Arc(10.0, 10.0, 150.0, 0.0, 180.0, false)
	ctx.Fill()

}

// ------------------------------------------------------------------------------------------------
func main() {
	// First step is to init the dom
	dom.Init()

	// now we can create elements
	// we won't need the actual container so we ignore the variable
	_ = html.Div().Class("container").Child(html.H1().Text("Hello from GoDom").Build()).AppendTo(dom.Body)

	// create our canvas and draw to it
	context2d := createAndConfigureHTMLCanvas()
	drawToTheCanvas(context2d)

	// need to block this from exiting, otherwise wasm terminates our app
	select {}
}
