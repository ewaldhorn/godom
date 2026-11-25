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

	return dom.NewContext2D(canvasHandle)
}

// ------------------------------------------------------------------------------------------------
func createAndConfigureHTMLCanvas() (html.Element, dom.Context2D) {
	// create a new Div container for our canvas
	canvasContainer := html.Div().Class("canvas-container")

	// create the actual canvas, set the attributes and background colour
	myCanvas := html.Canvas()
	myCanvas.Attr("width", "800")
	myCanvas.Attr("height", "600")
	myCanvas.Style("background-color", "blue")
	myCanvas.Style("touch-action", "none")

	// apped the canvas to the container, then the container to the html body
	myCanvas.AppendTo(canvasContainer.Handle)
	canvasContainer.AppendTo(dom.Body)

	// return the canvas and Context2D handles
	return myCanvas, getContext2DfromCanvasHandle(myCanvas.Handle)
}
