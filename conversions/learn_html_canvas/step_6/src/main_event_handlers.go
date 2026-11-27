package main

import (
	"github.com/ewaldhorn/godom/dom"
	"github.com/ewaldhorn/godom/html"
)

// ------------------------------------------------------------------------------------------------
func isSpaceKey(e dom.Event) bool {
	return e.Code() == "Space" || e.Key() == " "
}

// ------------------------------------------------------------------------------------------------
func registerCanvasEventHandlers(canvas html.Element) {
	var pointerDown, spaceDown bool

	syncRising := func() {
		balloonPos.IsRising = pointerDown || spaceDown
	}

	stopPointer := func() {
		pointerDown = false
		syncRising()
	}

	canvas.OnEvent("pointerdown", func(e dom.Event) {
		if e.Button() != 0 {
			return
		}
		e.PreventDefault()
		pointerDown = true
		syncRising()
	})

	canvas.OnEvent("contextmenu", func(e dom.Event) {
		e.PreventDefault()
		stopPointer()
	})

	canvas.Handle.On("pointerleave", stopPointer)
	dom.Window.On("pointerup", stopPointer)
	dom.Window.On("pointercancel", stopPointer)

	dom.Window.OnEvent("keydown", func(e dom.Event) {
		if !isSpaceKey(e) {
			return
		}
		e.PreventDefault()
		spaceDown = true
		syncRising()
	})

	dom.Window.OnEvent("keyup", func(e dom.Event) {
		if !isSpaceKey(e) {
			return
		}
		e.PreventDefault()
		spaceDown = false
		syncRising()
	})

	dom.Window.On("blur", func() {
		pointerDown = false
		spaceDown = false
		syncRising()
	})
}
