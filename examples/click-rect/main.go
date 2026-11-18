package main

import (
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
	isWhite bool
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
	// OffsetX and OffsetY are relative to the event target (the canvas),
	// so they map directly to pixel coordinates.
	dom.OnEventByID("app", "click", func(e dom.Event) {
		x, y := e.OffsetX(), e.OffsetY()
		if x >= rectX && x < rectX+rectW && y >= rectY && y < rectY+rectH {
			isWhite = !isWhite
			redraw(c)
		}
	})

	// Block forever so Go's WASM runtime and event listeners stay alive.
	select {}
}
