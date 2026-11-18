package main

import (
	_ "embed"
	"strconv"
	"syscall/js"
	"unsafe"

	"github.com/ewaldhorn/godom/dom"
	"github.com/ewaldhorn/godom/html"
	"github.com/ewaldhorn/godom/sound"
)

// ------------------------------------------------------------------------------------------------
// Embedded Resources and Constants
// ------------------------------------------------------------------------------------------------

//go:embed bodystyle.css
var bodyStyle string

//go:embed godom.txt
var dommieText string

const (
	Version = "0.0.5"
	Name    = "Godom Demo"
)

// Callback IDs for go_invoke_callback (JS to Go canvas pointer bridge).
// Keep the values in sync with CALLBACKS in docs/app.js.
const (
	cbCanvasInteract uint32 = iota
	cbDrumCanvasClick
	cbCanvasFourInteract
)

// ------------------------------------------------------------------------------------------------
// Shared Application State
// ------------------------------------------------------------------------------------------------

var (
	isReady              bool
	booCounter           uint32
	applicationContainer = dom.Invalid
	articleElement       = dom.Invalid
	asideElement         = dom.Invalid

	// Click sound buffer: pre-rendered 50ms UI click (2205 samples at 44.1kHz).
	clickBuffer [2205]float32

	// Latest canvas pointer coordinates set by JS before invoking a canvas callback.
	interactX int32 = -1
	interactY int32 = -1
)

// CSS class and size options for aside element generators.
var (
	cssColours = []string{"red", "blue", "orange"}
	cssSizes   = []string{"large", "larger", "xlarge"}

	facts = []string{
		"WASM runs at near native speed in modern browsers.",
		"Go includes an automatic garbage collector and goroutine scheduler.",
		"The first web page went live at CERN in 1991.",
		"Go channels coordinate concurrent work without shared locks.",
		"One kilobyte of magnetic core memory cost thousands of dollars in 1957.",
		"Early relay computers were debugged by removing physical moths from switches.",
	}

	quips = []string{
		"Pixels in motion, no framework required.",
		"Compiled from Go, rendered in the browser.",
		"Sixty frames per second of pure WebAssembly.",
		"Neon horizons and elastic collisions.",
		"Batching draw commands keeps the bridge fast.",
		"Zero external dependencies from top to bottom.",
	}

	tags = []string{
		"go",
		"wasm",
		"syscall-js",
		"concurrent",
		"retro",
		"synthwave",
	}
)

// Demo-local PRNG (xorshift64).
var rngState uint64 = 42

func nextRandom() uint32 {
	rngState ^= rngState << 13
	rngState ^= rngState >> 7
	rngState ^= rngState << 17
	return uint32(rngState)
}

func randomCssColour() string {
	return cssColours[int(nextRandom())%len(cssColours)]
}

func randomCssSize() string {
	return cssSizes[int(nextRandom())%len(cssSizes)]
}

func absF32(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

// ------------------------------------------------------------------------------------------------
// Animation Tick and DOM Showcase Actions
// ------------------------------------------------------------------------------------------------

func onAnimationTick() {
	if !isReady {
		return
	}
	canvasOneTime++
	gridOffset += 0.025
	if gridOffset >= 1.0 {
		gridOffset -= 1.0
	}

	performDemoOnCanvasOne()
	updateCanvasTwo()
	performDemoOnCanvasThree()
	updateCanvasFour()
}

func onAddSomethingClick() {
	if !isReady {
		return
	}
	switch nextRandom() % 5 {
	case 0:
		addBoo()
	case 1:
		addRandomParagraph()
	case 2:
		addAsideNote()
	case 3:
		addAsideTag()
	case 4:
		addAsideQuip()
	}
}

func onClearAsideClick() {
	if !isReady {
		return
	}
	asideElement.RemoveClassFrom("showcase-active")
	dom.RemoveAllChildElementsFrom(asideElement)
	dom.SetFocus("addSomethingButton")
}

func addBoo() {
	booCounter++
	countStr := strconv.FormatUint(uint64(booCounter), 10)
	text := "Boo! (" + countStr + ")"
	id := "boo-" + countStr
	html.P().
		ID(id).
		Attr("data-count", text).
		Text(text).
		AppendTo(asideElement)
}

func addRandomParagraph() {
	html.Div().
		Class(randomCssColour()).
		Class(randomCssSize()).
		Attr("data-random", "yes").
		Child(html.P().ID("random-p").Text("Built with the chainable html.Element builder").Build()).
		AppendTo(asideElement)
}

func addAsideNote() {
	booCounter++
	idx := int(nextRandom()) % len(facts)
	id := "note-" + strconv.FormatUint(uint64(booCounter), 10)
	html.Div().
		Class("aside-note").
		Class(randomCssColour()).
		ID(id).
		Text(facts[idx]).
		AppendTo(asideElement)
}

func addAsideTag() {
	booCounter++
	idx := int(nextRandom()) % len(tags)
	id := "tag-" + strconv.FormatUint(uint64(booCounter), 10)
	html.Span().
		Class("aside-tag").
		Class(randomCssColour()).
		ID(id).
		Text(tags[idx]).
		AppendTo(asideElement)
	html.Span().
		Text(" ").
		AppendTo(asideElement)
}

func addAsideQuip() {
	booCounter++
	idx := int(nextRandom()) % len(quips)
	id := "quip-" + strconv.FormatUint(uint64(booCounter), 10)
	html.Div().
		Class("aside-quip").
		ID(id).
		Text(quips[idx]).
		AppendTo(asideElement)
}

func initDOMLayout() {
	dom.Hide("loading")
	dom.Show("controls")
	dom.Show("information")

	titleElem := dom.GetElementByID("title")
	if titleElem.IsValid() {
		titleElem.SetInnerText(Name + " v" + Version)
		titleElem.ReplaceClasses([]string{"demo-title"})
	}
	dom.SetValue("title", "lang", "en")

	dom.AddNewStyleElement(bodyStyle)

	applicationContainer = dom.GetElementByID("application")
	articleElement = html.Article().AppendTo(applicationContainer).Build()
	asideElement = html.Aside().Class("showcase-active").AppendTo(applicationContainer).Build()

	html.P().ID("dommie-text").HTML(dommieText).AppendTo(articleElement)

	// Populate initial aside header using the handle-based API.
	showcaseHeading := dom.CreateParagraphWithText("DOM Builder Showcase")
	showcaseHeading.AddClassTo("boo-header")
	showcaseHeading.Set("title", "Click Add Something to append elements")
	dom.AddElementTo(asideElement, showcaseHeading)

	showcaseHint := dom.CreateParagraphWithText("Use Add Something and Clear Aside above to test live DOM manipulation from Go.")
	wrappedHint := dom.WrapElementWithNewDiv(showcaseHint, []string{"boo-wrapper"})
	dom.AddElementTo(asideElement, wrappedHint)
}

func registerEventListeners() {
	dom.AddEventListenerByID("addSomethingButton", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		onAddSomethingClick()
		return nil
	}).Value)

	dom.AddEventListenerByID("clearAsideButton", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		onClearAsideClick()
		return nil
	}).Value)

	dom.AddEventListenerByID("refreshButton", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		onRefreshCanvasOneClick()
		return nil
	}).Value)

	dom.AddEventListenerByID("drumPlayButton", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		onDrumPlayPause()
		return nil
	}).Value)

	dom.AddEventListenerByID("gravityButton", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		onCycleGravity()
		return nil
	}).Value)

	// Expose canvas pointer interaction and audio buffer helpers to JS.
	js.Global().Set("go_set_interaction", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 2 {
			return nil
		}
		interactX = int32(args[0].Int())
		interactY = int32(args[1].Int())
		return nil
	}))

	js.Global().Set("go_invoke_callback", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 1 {
			return nil
		}
		if len(args) >= 3 {
			interactX = int32(args[1].Int())
			interactY = int32(args[2].Int())
		}
		id := uint32(args[0].Int())
		switch id {
		case cbCanvasInteract:
			onCanvasInteraction()
		case cbDrumCanvasClick:
			onDrumCanvasClick()
		case cbCanvasFourInteract:
			onCanvasFourInteraction()
		default:
			dom.Log("go_invoke_callback: unknown callback id " + strconv.FormatUint(uint64(id), 10))
		}
		return nil
	}))

	js.Global().Set("go_get_click_buffer", js.FuncOf(func(this js.Value, args []js.Value) any {
		byteSlice := unsafe.Slice((*byte)(unsafe.Pointer(unsafe.SliceData(clickBuffer[:]))), len(clickBuffer)*4)
		uint8Array := js.Global().Get("Uint8Array").New(len(byteSlice))
		js.CopyBytesToJS(uint8Array, byteSlice)
		return js.Global().Get("Float32Array").New(uint8Array.Get("buffer"), uint8Array.Get("byteOffset"), len(clickBuffer))
	}))
}

// ------------------------------------------------------------------------------------------------
// Entry Point
// ------------------------------------------------------------------------------------------------

func main() {
	dom.Init()
	dom.Log("Ok.", "Godom is starting.", "Here we go!")

	initDOMLayout()
	registerEventListeners()

	initCanvasOne()
	initCanvasTwo()
	initCanvasThree()
	initCanvasFour()

	sound.FillClick(clickBuffer[:])

	isReady = true

	animationLoopCb := js.FuncOf(func(this js.Value, args []js.Value) any {
		onAnimationTick()
		return nil
	})
	dom.StartAnimationLoop(animationLoopCb.Value)

	// Block forever so the Go WASM runtime and all registered callbacks stay alive.
	select {}
}
