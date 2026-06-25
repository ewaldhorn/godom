package main

import (
	"syscall/js"

	"github.com/ewaldhorn/godom/canvas"
	"github.com/ewaldhorn/godom/colour"
	"github.com/ewaldhorn/godom/dom"
)

const (
	canvasW = 300
	canvasH = 300

	// Rectangle centered in the 300x300 canvas
	rectX = 100
	rectY = 100
	rectW = 100
	rectH = 100
)

var (
	blue    = colour.Colour{R: 30, G: 100, B: 255, A: 255}
	isWhite = false
)

func redraw(c *canvas.Canvas) {
	c.ClearScreen(colour.Black)
	if isWhite {
		c.ColourFilledRectangle(rectX, rectY, rectW, rectH, colour.White)
	} else {
		c.ColourFilledRectangle(rectX, rectY, rectW, rectH, blue)
	}
	c.Render()
}

func main() {
	dom.Init()

	pixels := make([]byte, canvasW*canvasH*4)
	c := canvas.NewCanvas(canvasW, canvasH, pixels, "app")

	redraw(c)

	// Click events on the canvas bubble up to its parent "app" div.
	// offsetX/offsetY are relative to the event target (the canvas),
	// so they map directly to pixel coordinates.
	// js.FuncOf returns a js.Func; pass .Value where js.Value is expected.
	clickFn := js.FuncOf(func(this js.Value, args []js.Value) any {
		x := args[0].Get("offsetX").Int()
		y := args[0].Get("offsetY").Int()

		// Only toggle if the click landed on the rectangle.
		if x >= rectX && x < rectX+rectW && y >= rectY && y < rectY+rectH {
			isWhite = !isWhite
			redraw(c)
		}
		return nil
	})
	dom.AddEventListenerByID("app", "click", clickFn.Value)

	// Block forever — Go's WASM runtime exits as soon as main() returns,
	// which would tear down all event listeners.
	select {}
}
