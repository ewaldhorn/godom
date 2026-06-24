package main

import (
	_ "embed"
	"fmt"
	"math"
	"syscall/js"
	"unsafe"

	"github.com/ewaldhorn/godom/canvas"
	"github.com/ewaldhorn/godom/colour"
	"github.com/ewaldhorn/godom/dom"
	"github.com/ewaldhorn/godom/html"
	"github.com/ewaldhorn/godom/sound"
)

// ------------------------------------------------------------------------------------------------
// Resources & Constants
// ------------------------------------------------------------------------------------------------

//go:embed bodystyle.css
var bodyStyle string

//go:embed godom.txt
var dommieText string

const (
	Version = "0.0.1f"
	Name    = "Godom Demo"
)

// ------------------------------------------------------------------------------------------------
// Application State
// ------------------------------------------------------------------------------------------------

var (
	isReady              bool = false
	booCounter           uint32
	applicationContainer dom.Handle = dom.Invalid
	articleElement       dom.Handle = dom.Invalid
	asideElement         dom.Handle = dom.Invalid
	canvasOneTime        uint32
	gridOffset           float64
)

// Click sound buffer — pre-rendered 50ms UI click (2205 samples at 44.1kHz)
var clickBuffer [2205]float32

// CSS class / size options for aside element generators
var cssColours = []string{"red", "blue", "orange"}
var cssSizes = []string{"large", "larger", "xlarge"}

// Random text pools for aside content
var facts = []string{
	"WASM runs at near-native speed.",
	"Go has an excellent runtime and GC.",
	"First website: info.cern.ch (1991).",
	"Go has goroutines and channels.",
	"A kilobyte of RAM cost ~$3M in 1957.",
	"The first computer bug was a real moth.",
}
var quips = []string{
	"Stay awhile and listen.",
	"All these worlds are yours.",
	"It was a pleasure to burn.",
	"The sky above the port…",
	"So long, and thanks for all the fish.",
	"I'm sorry, Dave.",
}
var tags = []string{
	"go",
	"wasm",
	"syscall-js",
	"concurrent",
	"retro",
	"synthwave",
}

// Demo-local PRNG (xorshift64*) — independent of colour's PRNG
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

// ------------------------------------------------------------------------------------------------
// Canvas buffers — allocated in data segment.
// ------------------------------------------------------------------------------------------------

var canvasOneBuffer [800 * 600 * 4]byte
var canvasOne *canvas.Canvas

var canvasTwoBuffer [600 * 450 * 4]byte
var canvasTwo *canvas.Canvas

var canvasThreeBuffer [512 * 320 * 4]byte
var canvasThree *canvas.Canvas

// ------------------------------------------------------------------------------------------------
// Ball physics — fixed-size array.
// ------------------------------------------------------------------------------------------------

const MaxBalls = 14

type Ball struct {
	X, Y, Vx, Vy float32
	Radius       float32
	Col          colour.Colour
}

var balls [MaxBalls]Ball

// ------------------------------------------------------------------------------------------------
// Drum machine — fixed-size sequencer state.
// ------------------------------------------------------------------------------------------------

const (
	DrumTracks     = 6
	DrumSteps      = 16
	DrumCanvasW    = 512
	DrumCanvasH    = 320
	DrumGridX      = 74
	DrumGridY      = 82
	DrumCellW      = 24
	DrumCellH      = 24
	DrumCellGap    = 3
	DrumCellPitchX = DrumCellW + DrumCellGap
	DrumCellPitchY = DrumCellH + 6
)

var drumPattern = [DrumTracks][DrumSteps]bool{
	{true, false, false, false, true, false, false, false, true, false, false, false, true, false, false, false},
	{false, false, false, false, true, false, false, false, false, false, false, false, true, false, false, false},
	{true, false, true, false, true, false, true, false, true, false, true, false, true, false, true, false},
	{false, false, false, false, false, false, false, true, false, false, false, false, false, false, false, true},
	{false, false, false, false, false, false, true, false, false, false, false, false, false, false, true, false},
	{false, true, false, false, false, true, false, false, false, true, false, false, false, true, false, false},
}

var (
	drumCurrentStep uint32 = 0
	drumPlaying     bool   = false
	drumTick        uint32 = 0
	drumBpm         uint32 = 120
	drumStepAccum   uint32 = 0
)

var drumLabels = []string{"K", "SN", "HH", "OH", "CL", "RM"}

var drumNeon = [DrumTracks]colour.Colour{
	{R: 0, G: 240, B: 220, A: 255},
	{R: 255, G: 0, B: 180, A: 255},
	{R: 255, G: 230, B: 0, A: 255},
	{R: 255, G: 153, B: 0, A: 255},
	{R: 160, G: 0, B: 255, A: 255},
	{R: 0, G: 255, B: 100, A: 255},
}

var drumDim = [DrumTracks]colour.Colour{
	{R: 10, G: 42, B: 40, A: 255},
	{R: 42, G: 0, B: 30, A: 255},
	{R: 42, G: 37, B: 0, A: 255},
	{R: 42, G: 24, B: 0, A: 255},
	{R: 26, G: 0, B: 48, A: 255},
	{R: 0, G: 40, B: 16, A: 255},
}

// ------------------------------------------------------------------------------------------------
// Interaction coordinates
// ------------------------------------------------------------------------------------------------

var (
	interactX int32 = -1
	interactY int32 = -1
)

// ------------------------------------------------------------------------------------------------
// Callback dispatch
// ------------------------------------------------------------------------------------------------

func onAddSomethingClick() {
	if !isReady {
		return
	}
	roll := nextRandom() % 5
	switch roll {
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
	default:
		addBoo()
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

func onRefreshCanvasOneClick() {
	if !isReady {
		return
	}
	currentThemeIdx = (currentThemeIdx + 1) % len(themes)
}

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
}

func onCanvasInteraction() {
	if !isReady {
		return
	}
	if interactX < 0 || interactY < 0 {
		return
	}
	ix := float32(interactX)
	iy := float32(interactY)
	interactX = -1
	interactY = -1

	for i := range balls {
		dx := balls[i].X - ix
		dy := balls[i].Y - iy
		distSq := dx*dx + dy*dy
		if distSq < 1.0 {
			continue
		}
		dist := float32(math.Sqrt(float64(distSq)))
		impulse := 180.0 / dist
		if impulse > 12.0 {
			impulse = 12.0
		}
		balls[i].Vx += (dx / dist) * impulse
		balls[i].Vy += (dy / dist) * impulse
	}
}

func onDrumCanvasClick() {
	if !isReady {
		return
	}
	if interactX < 0 || interactY < 0 {
		return
	}
	rx := interactX - DrumGridX
	ry := interactY - DrumGridY
	interactX = -1
	interactY = -1
	if rx < 0 || ry < 0 {
		return
	}

	stepI := rx / DrumCellPitchX
	trackI := ry / DrumCellPitchY
	if stepI < 0 || stepI >= DrumSteps || trackI < 0 || trackI >= DrumTracks {
		return
	}
	if rx%DrumCellPitchX >= DrumCellW {
		return
	}
	if ry%DrumCellPitchY >= DrumCellH {
		return
	}

	drumPattern[trackI][stepI] = !drumPattern[trackI][stepI]
	if drumPattern[trackI][stepI] {
		js.Global().Call("drumHit", trackI)
	}
}

func onDrumPlayPause() {
	if !isReady {
		return
	}
	drumPlaying = !drumPlaying
	drumTick = 0
	drumStepAccum = 0
	setDrumButtonText()
	if drumPlaying {
		playCurrentDrumStep()
	}
}

func setDrumButtonText() {
	btn := dom.GetElementByID("drumPlayButton")
	if !btn.IsValid() {
		return
	}
	if drumPlaying {
		btn.SetInnerText("⏹ Drum Machine")
	} else {
		btn.SetInnerText("▶ Drum Machine")
	}
}

func updateDrumSequencer() {
	if !drumPlaying {
		return
	}
	drumTick++
	drumStepAccum += drumBpm
	for drumStepAccum >= 900 {
		drumStepAccum -= 900
		drumCurrentStep = (drumCurrentStep + 1) % DrumSteps
		playCurrentDrumStep()
	}
}

func playCurrentDrumStep() {
	step := int(drumCurrentStep)
	for track := 0; track < DrumTracks; track++ {
		if drumPattern[track][step] {
			js.Global().Call("drumHit", track)
		}
	}
}

// ------------------------------------------------------------------------------------------------
// DOM helpers
// ------------------------------------------------------------------------------------------------

func addBoo() {
	booCounter++
	text := fmt.Sprintf("Boo! (%d)", booCounter)
	id := fmt.Sprintf("boo-%d", booCounter)
	_ = html.P().
		ID(id).
		Attr("data-count", text).
		Text(text).
		AppendTo(asideElement)
}

func addRandomParagraph() {
	_ = html.Div().
		Class(randomCssColour()).
		Class(randomCssSize()).
		Attr("data-random", "yes").
		Child(html.P().ID("random-p").Text("This is some text using builder API").Build()).
		AppendTo(asideElement)
}

func addAsideNote() {
	idx := int(nextRandom()) % len(facts)
	id := fmt.Sprintf("note-%d", booCounter)
	_ = html.Div().
		Class("aside-note").
		Class(randomCssColour()).
		ID(id).
		Text(facts[idx]).
		AppendTo(asideElement)
}

func addAsideTag() {
	idx := int(nextRandom()) % len(tags)
	id := fmt.Sprintf("tag-%d", booCounter)
	_ = html.Span().
		Class("aside-tag").
		Class(randomCssColour()).
		ID(id).
		Text(tags[idx]).
		AppendTo(asideElement)
	// Also append space so tags don't stick
	_ = html.Span().
		Text(" ").
		AppendTo(asideElement)
}

func addAsideQuip() {
	idx := int(nextRandom()) % len(quips)
	id := fmt.Sprintf("quip-%d", booCounter)
	_ = html.Div().
		Class("aside-quip").
		ID(id).
		Text(quips[idx]).
		AppendTo(asideElement)
}

func toggleElements() {
	dom.Hide("loading")
	dom.Show("controls")
	dom.Show("information")
}

func createAppElements() {
	articleElement = html.Article().AppendTo(applicationContainer).Build()
	asideElement = html.Aside().AppendTo(applicationContainer).Build()
}

func populateArticleElement() {
	_ = html.P().ID("dommie-text").HTML(dommieText).AppendTo(articleElement)
}

func setTitle() {
	title := fmt.Sprintf("%s v%s", Name, Version)
	elem := dom.GetElementByID("title")
	if elem.IsValid() {
		elem.SetInnerText(title)
	}
}

// ------------------------------------------------------------------------------------------------
// CANVAS ONE — Dark Synthwave Gallery
// ------------------------------------------------------------------------------------------------

type Theme struct {
	bg             colour.Colour
	panel          colour.Colour
	cyan           colour.Colour
	magenta        colour.Colour
	yellow         colour.Colour
	violet         colour.Colour
	sun_start      colour.Colour
	sun_end        colour.Colour
	glow_start_str string
	glow_mid_str   string
	glow_end_str   string
}

var currentThemeIdx = 0

var themes = []Theme{
	{
		bg:             colour.Colour{R: 5, G: 5, B: 16, A: 255},
		panel:          colour.Colour{R: 12, G: 12, B: 36, A: 255},
		cyan:           colour.Colour{R: 0, G: 240, B: 220, A: 255},
		magenta:        colour.Colour{R: 255, G: 0, B: 180, A: 255},
		yellow:         colour.Colour{R: 255, G: 230, B: 0, A: 255},
		violet:         colour.Colour{R: 160, G: 0, B: 255, A: 255},
		sun_start:      colour.Colour{R: 255, G: 230, B: 0, A: 255},
		sun_end:        colour.Colour{R: 255, G: 0, B: 180, A: 255},
		glow_start_str: "rgba(255,0,180,0.12)",
		glow_mid_str:   "rgba(255,180,30,0.22)",
		glow_end_str:   "rgba(255,230,0,0.32)",
	},
	{
		bg:             colour.Colour{R: 12, G: 4, B: 4, A: 255},
		panel:          colour.Colour{R: 24, G: 8, B: 8, A: 255},
		cyan:           colour.Colour{R: 255, G: 120, B: 0, A: 255},
		magenta:        colour.Colour{R: 255, G: 40, B: 0, A: 255},
		yellow:         colour.Colour{R: 255, G: 200, B: 0, A: 255},
		violet:         colour.Colour{R: 100, G: 0, B: 0, A: 255},
		sun_start:      colour.Colour{R: 255, G: 200, B: 0, A: 255},
		sun_end:        colour.Colour{R: 255, G: 40, B: 0, A: 255},
		glow_start_str: "rgba(255,40,0,0.12)",
		glow_mid_str:   "rgba(255,120,0,0.22)",
		glow_end_str:   "rgba(255,200,0,0.32)",
	},
	{
		bg:             colour.Colour{R: 20, G: 10, B: 30, A: 255},
		panel:          colour.Colour{R: 40, G: 20, B: 60, A: 255},
		cyan:           colour.Colour{R: 0, G: 255, B: 255, A: 255},
		magenta:        colour.Colour{R: 255, G: 105, B: 180, A: 255},
		yellow:         colour.Colour{R: 255, G: 255, B: 150, A: 255},
		violet:         colour.Colour{R: 138, G: 43, B: 226, A: 255},
		sun_start:      colour.Colour{R: 0, G: 255, B: 255, A: 255},
		sun_end:        colour.Colour{R: 255, G: 105, B: 180, A: 255},
		glow_start_str: "rgba(255,105,180,0.12)",
		glow_mid_str:   "rgba(138,43,226,0.22)",
		glow_end_str:   "rgba(0,255,255,0.32)",
	},
	{
		bg:             colour.Colour{R: 2, G: 8, B: 4, A: 255},
		panel:          colour.Colour{R: 4, G: 16, B: 8, A: 255},
		cyan:           colour.Colour{R: 0, G: 255, B: 100, A: 255},
		magenta:        colour.Colour{R: 0, G: 180, B: 50, A: 255},
		yellow:         colour.Colour{R: 200, G: 255, B: 200, A: 255},
		violet:         colour.Colour{R: 0, G: 80, B: 20, A: 255},
		sun_start:      colour.Colour{R: 200, G: 255, B: 200, A: 255},
		sun_end:        colour.Colour{R: 0, G: 180, B: 50, A: 255},
		glow_start_str: "rgba(0,180,50,0.12)",
		glow_mid_str:   "rgba(0,80,20,0.22)",
		glow_end_str:   "rgba(0,255,100,0.32)",
	},
}

func performDemoOnCanvasOne() {
	w := int32(canvasOne.Width)
	h := int32(canvasOne.Height)
	horizon := h/2 + 50
	activeTheme := themes[currentThemeIdx]

	colour.Seed(uint64(nextRandom()) | 1)

	canvasOne.ClearScreen(activeTheme.bg)
	canvasOne.ColourRectangle(8, 8, w-16, h-16, 2, activeTheme.panel)

	canvasOne.SetColour(activeTheme.cyan)
	canvasOne.PutPixel(10, 10)
	canvasOne.PutPixel(w-11, 10)

	drawStarfield(w, horizon, canvasOneTime, activeTheme)
	drawRetroSun(w/2, horizon-20, 85, activeTheme)
	drawMountainSilhouettes(w, horizon, activeTheme)
	drawLaserGrid(w, h, horizon, gridOffset, activeTheme)
	drawVectorShip(w/2, h-55, canvasOneTime, activeTheme)

	canvasOne.Render()

	drawSunGlow(canvasOne.GetContext2D(), w, horizon-20, canvasOneTime, activeTheme)
}

func drawStarfield(w, horizon int32, time uint32, t Theme) {
	colour.Seed(999)
	for i := 0; i < 90; i++ {
		rx := int32(colour.RandomColour().R)*3 + int32(colour.RandomColour().G)%50
		ry := int32(colour.RandomColour().B) % (horizon - 30)
		if rx < 12 || rx >= w-12 || ry < 12 {
			continue
		}

		rVal := colour.RandomColour().G
		twinklePhase := (rVal + uint8(time%256)) % 12
		if twinklePhase == 0 {
			continue
		}

		dice := rVal % 8
		if dice == 0 {
			canvasOne.ColourPutPixel(rx, ry, colour.White)
			if twinklePhase > 2 {
				canvasOne.ColourPutPixel(rx-1, ry, t.cyan)
				canvasOne.ColourPutPixel(rx+1, ry, t.cyan)
				canvasOne.ColourPutPixel(rx, ry-1, t.magenta)
				canvasOne.ColourPutPixel(rx, ry+1, t.magenta)
			}
		} else if dice < 3 {
			canvasOne.ColourPutPixel(rx, ry, t.magenta)
		} else if dice < 6 {
			canvasOne.ColourPutPixel(rx, ry, t.cyan)
		} else {
			canvasOne.ColourPutPixel(rx, ry, colour.White)
		}
	}
}

func drawRetroSun(cx, cy, r int32, t Theme) {
	for dy := -r; dy <= r; dy++ {
		y := cy + dy
		rf := float64(r)
		dyf := float64(dy)
		chord := int32(math.Sqrt(rf*rf - dyf*dyf))

		if dy > 0 {
			val := dy % 12
			if val < dy/6+1 {
				continue
			}
		}

		ratio := float32(dy+r) / float32(2*r)
		red := uint8(float32(t.sun_start.R)*(1.0-ratio) + float32(t.sun_end.R)*ratio)
		green := uint8(float32(t.sun_start.G)*(1.0-ratio) + float32(t.sun_end.G)*ratio)
		blue := uint8(float32(t.sun_start.B)*(1.0-ratio) + float32(t.sun_end.B)*ratio)
		c := colour.Colour{R: red, G: green, B: blue, A: 255}

		canvasOne.ColourLine(cx-chord, y, cx+chord, y, c)
	}

	canvasOne.ColourCircle(cx, cy, r+4, t.cyan)
}

func drawMountainSilhouettes(w, horizon int32, t Theme) {
	bgPts := [][2]int32{
		{8, horizon},
		{120, horizon - 75},
		{240, horizon - 25},
		{350, horizon - 105},
		{480, horizon - 45},
		{620, horizon - 95},
		{710, horizon - 35},
		{w - 8, horizon},
	}

	fgPts := [][2]int32{
		{8, horizon},
		{90, horizon - 40},
		{180, horizon - 15},
		{290, horizon - 65},
		{390, horizon - 30},
		{510, horizon - 75},
		{640, horizon - 20},
		{730, horizon - 50},
		{w - 8, horizon},
	}

	for idx := 0; idx < len(bgPts)-1; idx++ {
		p1 := bgPts[idx]
		p2 := bgPts[idx+1]
		for x := p1[0]; x <= p2[0]; x++ {
			ratio := float32(x-p1[0]) / float32(p2[0]-p1[0])
			y := int32((1.0-ratio)*float32(p1[1]) + ratio*float32(p2[1]))
			canvasOne.ColourLine(x, y+1, x, horizon, t.bg)
		}
	}

	for idx := 0; idx < len(bgPts)-1; idx++ {
		canvasOne.ColourLine(bgPts[idx][0], bgPts[idx][1], bgPts[idx+1][0], bgPts[idx+1][1], t.violet)
	}

	for idx := 0; idx < len(fgPts)-1; idx++ {
		p1 := fgPts[idx]
		p2 := fgPts[idx+1]
		for x := p1[0]; x <= p2[0]; x++ {
			ratio := float32(x-p1[0]) / float32(p2[0]-p1[0])
			y := int32((1.0-ratio)*float32(p1[1]) + ratio*float32(p2[1]))
			canvasOne.ColourLine(x, y+1, x, horizon, t.bg)
		}
	}

	for idx := 0; idx < len(fgPts)-1; idx++ {
		canvasOne.ColourLine(fgPts[idx][0], fgPts[idx][1], fgPts[idx+1][0], fgPts[idx+1][1], t.cyan)
	}
}

func drawLaserGrid(w, h, horizon int32, scrollOffset float64, t Theme) {
	cxF := float64(w / 2)
	hF := float64(horizon)
	hRange := float64(h - horizon)
	curvePhase := float64(canvasOneTime) * 0.003
	curveAmp := 30.0
	segments := 12

	for x := int32(-120); x <= w+120; x += 35 {
		ex := float64(x)
		for s := 0; s < segments; s++ {
			t0 := float64(s) / float64(segments)
			t1 := float64(s+1) / float64(segments)

			y0F := hF + t0*hRange
			y1F := hF + t1*hRange

			c0 := curveAmp * t0 * math.Sin(t0*math.Pi*2.0+curvePhase)
			c1 := curveAmp * t1 * math.Sin(t1*math.Pi*2.0+curvePhase)

			x0F := cxF + (ex-cxF)*t0 + c0
			x1F := cxF + (ex-cxF)*t1 + c1

			canvasOne.ColourLine(int32(x0F), int32(y0F), int32(x1F), int32(y1F), t.violet)
			canvasOne.ColourLine(int32(x0F), int32(y0F+2), int32(x1F), int32(y1F+2), t.cyan)
		}
	}

	var i float64 = 0.0
	for {
		exponent := i + scrollOffset
		dist := 6.0 * math.Pow(1.25, exponent)
		y := horizon + int32(dist)
		if y >= h-8 {
			break
		}
		if y >= horizon+8 {
			yF := float64(y)
			prog := (yF - hF) / hRange
			c := curveAmp * prog * math.Sin(prog*math.Pi*2.0+curvePhase)
			left := int32(10.0 + c)
			right := int32(float64(w-10) + c)
			canvasOne.ColourLine(left, y, right, y, t.violet)
			canvasOne.ColourLine(left, y, right, y, t.magenta)
		}
		i += 1.0
	}
}

func drawVectorShip(cx, cy int32, time uint32, t Theme) {
	timeF := float64(time)
	bobY := int32(4.0 * math.Sin(timeF*0.06))
	swayX := int32(5.0 * math.Cos(timeF*0.04))

	scx := cx + swayX
	scy := cy + bobY

	flameLen := int32(26.0 + 4.0*math.Sin(timeF*0.25))

	canvasOne.ColourLine(scx-8, scy+12, scx, scy+flameLen, t.cyan)
	canvasOne.ColourLine(scx+8, scy+12, scx, scy+flameLen, t.cyan)
	canvasOne.ColourLine(scx-8, scy+12, scx+8, scy+12, t.cyan)

	canvasOne.ColourLine(scx-35, scy+10, scx+35, scy+10, t.magenta)
	canvasOne.ColourLine(scx-35, scy+10, scx-12, scy-20, t.magenta)
	canvasOne.ColourLine(scx+35, scy+10, scx+12, scy-20, t.magenta)

	canvasOne.ColourLine(scx-12, scy-20, scx, scy-40, t.yellow)
	canvasOne.ColourLine(scx+12, scy-20, scx, scy-40, t.yellow)
	canvasOne.ColourLine(scx-12, scy-20, scx+12, scy-20, t.yellow)

	canvasOne.ColourLine(scx-6, scy-10, scx, scy-25, t.cyan)
	canvasOne.ColourLine(scx+6, scy-10, scx, scy-25, t.cyan)
	canvasOne.ColourLine(scx-6, scy-10, scx+6, scy-10, t.cyan)

	shadow := t.cyan
	shadow.ConvertToGrayscale()
	tri1 := canvas.Point{X: scx - 4, Y: scy + flameLen + 4}
	tri2 := canvas.Point{X: scx + 4, Y: scy + flameLen + 4}
	tri3 := canvas.Point{X: scx, Y: scy + flameLen + 16}
	canvasOne.ColourLinePoint(tri1, tri2, shadow)
	canvasOne.ColourLinePoint(tri2, tri3, shadow)
	canvasOne.ColourLinePoint(tri1, tri3, shadow)
}

func drawSunGlow(ctx dom.Context2D, w, cy int32, time uint32, t Theme) {
	cxF := float64(w / 2)
	cyF := float64(cy)
	timeF := float64(time)
	pulse := 3.0 * math.Sin(timeF*0.05)

	ctx.BeginPath()
	ctx.FillStyle(t.glow_start_str)
	ctx.Arc(cxF, cyF, 130.0+pulse, 0.0, 6.2832, false)
	ctx.Fill()

	ctx.BeginPath()
	ctx.FillStyle(t.glow_mid_str)
	ctx.Arc(cxF, cyF, 95.0-pulse, 0.0, 6.2832, false)
	ctx.Fill()

	ctx.BeginPath()
	ctx.FillStyle(t.glow_end_str)
	ctx.Arc(cxF, cyF, 50.0+pulse, 0.0, 6.2832, false)
	ctx.Fill()
}

// ------------------------------------------------------------------------------------------------
// CANVAS TWO — Ball Physics Simulation
// ------------------------------------------------------------------------------------------------

func initBalls() {
	w := float32(canvasTwo.Width)
	h := float32(canvasTwo.Height)

	initData := [MaxBalls][5]int32{
		{20, 30, 270, -180, 12},
		{70, 20, -300, 120, 14},
		{50, 60, 180, 300, 11},
		{15, 70, -135, -270, 13},
		{80, 75, 330, -90, 9},
		{40, 40, -225, -225, 16},
		{60, 80, 150, 375, 12},
		{30, 15, -375, 150, 10},
		{85, 45, -165, -300, 17},
		{10, 50, 450, 135, 9},
		{55, 25, -270, -120, 13},
		{75, 60, 120, -330, 15},
		{25, 85, 300, 90, 11},
		{90, 10, -195, 255, 12},
	}

	ballColours := [MaxBalls]colour.Colour{
		{R: 0, G: 240, B: 220, A: 255},   // cyan
		{R: 255, G: 0, B: 180, A: 255},   // magenta
		{R: 0, G: 255, B: 100, A: 255},   // green
		{R: 255, G: 230, B: 0, A: 255},   // yellow
		{R: 160, G: 0, B: 255, A: 255},   // violet
		{R: 255, G: 100, B: 0, A: 255},   // orange
		{R: 0, G: 160, B: 255, A: 255},   // sky blue
		{R: 255, G: 0, B: 80, A: 255},    // hot red
		{R: 100, G: 255, B: 200, A: 255}, // mint
		{R: 255, G: 160, B: 0, A: 255},   // amber
		{R: 200, G: 0, B: 255, A: 255},   // purple
		{R: 0, G: 255, B: 255, A: 255},   // electric cyan
		{R: 255, G: 80, B: 180, A: 255},  // pink
		{R: 80, G: 255, B: 0, A: 255},    // lime
	}

	for i := 0; i < MaxBalls; i++ {
		d := initData[i]
		balls[i] = Ball{
			X:      w * float32(d[0]) / 100.0,
			Y:      h * float32(d[1]) / 100.0,
			Vx:     float32(d[2]) / 100.0,
			Vy:     float32(d[3]) / 100.0,
			Radius: float32(d[4]),
			Col:    ballColours[i],
		}
	}
}

func updateCanvasTwo() {
	w := float32(canvasTwo.Width)
	h := float32(canvasTwo.Height)
	var gravity float32 = 0.08
	var dampen float32 = 0.95
	bg := colour.Colour{R: 8, G: 8, B: 15, A: 255}

	canvasTwo.ClearScreen(bg)

	for i := range balls {
		ball := &balls[i]
		ball.Vy += gravity
		ball.X += ball.Vx
		ball.Y += ball.Vy

		r := ball.Radius

		if ball.X-r < 0.0 {
			ball.X = r
			ball.Vx = absF32(ball.Vx) * dampen
		}
		if ball.X+r > w {
			ball.X = w - r
			ball.Vx = -absF32(ball.Vx) * dampen
		}
		if ball.Y-r < 0.0 {
			ball.Y = r
			ball.Vy = absF32(ball.Vy) * dampen
		}
		if ball.Y+r > h {
			ball.Y = h - r
			ball.Vy = -absF32(ball.Vy) * dampen
		}
	}

	// Ball-to-ball collision
	for i := 0; i < MaxBalls; i++ {
		for j := i + 1; j < MaxBalls; j++ {
			ba := &balls[i]
			bb := &balls[j]
			dx := bb.X - ba.X
			dy := bb.Y - ba.Y
			distSq := dx*dx + dy*dy
			minDist := ba.Radius + bb.Radius + 1.0

			if distSq < minDist*minDist && distSq > 0.0001 {
				dist := float32(math.Sqrt(float64(distSq)))
				nx := dx / dist
				ny := dy / dist

				overlap := minDist - dist
				ba.X -= nx * overlap * 0.5
				ba.Y -= ny * overlap * 0.5
				bb.X += nx * overlap * 0.5
				bb.Y += ny * overlap * 0.5

				dvx := ba.Vx - bb.Vx
				dvy := ba.Vy - bb.Vy
				dvn := dvx*nx + dvy*ny
				if dvn > 0 {
					ba.Vx -= dvn * nx
					ba.Vy -= dvn * ny
					bb.Vx += dvn * nx
					bb.Vy += dvn * ny
				}
			}
		}
	}

	for i := range balls {
		ball := &balls[i]
		bx := int32(ball.X)
		by := int32(ball.Y)
		br := int32(ball.Radius)

		glow := colour.Colour{
			R: maxUint8(ball.Col.R/3, 20),
			G: maxUint8(ball.Col.G/3, 20),
			B: maxUint8(ball.Col.B/3, 20),
			A: 255,
		}
		canvasTwo.ColourFilledCircle(bx, by, br+4, glow)
		canvasTwo.ColourFilledCircle(bx, by, br, ball.Col)

		hs := int32(ball.Radius / 3.0)
		canvasTwo.ColourPutPixel(bx-hs, by-hs, colour.White)

		if i == 0 {
			canvasTwo.ColourBorderCircle(bx, by, br+8, 2, colour.White)
		}
	}

	canvasTwo.Render()
}

// ------------------------------------------------------------------------------------------------
// CANVAS THREE — Retro Drum Machine
// ------------------------------------------------------------------------------------------------

func performDemoOnCanvasThree() {
	updateDrumSequencer()

	bg := colour.Colour{R: 5, G: 5, B: 16, A: 255}
	panel := colour.Colour{R: 12, G: 12, B: 34, A: 255}
	border := colour.Colour{R: 0, G: 240, B: 220, A: 255}
	muted := colour.Colour{R: 68, G: 68, B: 112, A: 255}

	canvasThree.ClearScreen(bg)
	canvasThree.ColourRectangle(8, 8, DrumCanvasW-16, DrumCanvasH-16, 2, border)
	canvasThree.ColourRectangle(14, 14, DrumCanvasW-28, DrumCanvasH-28, 1, panel)

	drawTinyText("DRUM 120 BPM", 26, 26, 2, border)
	if drumPlaying {
		drawTinyText("RUN", 400, 26, 2, drumNeon[5])
	} else {
		drawTinyText("PAUSE", 400, 26, 2, muted)
	}

	for step := 0; step < DrumSteps; step++ {
		x := int32(DrumGridX + step*DrumCellPitchX)
		stepLabelY := int32(DrumGridY - 24)
		if step%4 == 0 {
			canvasThree.ColourLine(x-5, DrumGridY-6, x-5, DrumGridY+DrumTracks*DrumCellPitchY-2, muted)
		}
		drawTinyDigit(uint8((step+1)%10), x+8, stepLabelY, 1, muted)
	}

	for track := 0; track < DrumTracks; track++ {
		y := int32(DrumGridY + track*DrumCellPitchY)
		drawTinyText(drumLabels[track], 24, y+8, 2, drumNeon[track])
		canvasThree.ColourFilledCircle(60, y+12, 4, drumNeon[track])

		for step := 0; step < DrumSteps; step++ {
			x := int32(DrumGridX + step*DrumCellPitchX)
			var cellColour colour.Colour
			if drumPattern[track][step] {
				cellColour = drumNeon[track]
			} else {
				cellColour = drumDim[track]
			}
			canvasThree.ColourFilledRectangle(x, y, DrumCellW, DrumCellH, cellColour)
			canvasThree.ColourRectangle(x, y, DrumCellW, DrumCellH, 1, panel)

			if drumPattern[track][step] {
				canvasThree.ColourFilledRectangle(x+5, y+5, DrumCellW-10, DrumCellH-10, colour.White)
				canvasThree.ColourFilledRectangle(x+7, y+7, DrumCellW-14, DrumCellH-14, drumNeon[track])
			}
		}
	}

	playheadX := int32(DrumGridX + int(drumCurrentStep)*DrumCellPitchX)
	canvasThree.ColourRectangle(playheadX-2, DrumGridY-4, DrumCellW+4, DrumTracks*DrumCellPitchY-4, 2, colour.White)
	canvasThree.ColourFilledRectangle(playheadX+2, DrumGridY+DrumTracks*DrumCellPitchY+4, DrumCellW-4, 5, colour.White)
	canvasThree.Render()
}

func drawTinyText(text string, x, y, scale int32, c colour.Colour) {
	cursor := x
	for i := 0; i < len(text); i++ {
		ch := text[i]
		if ch == ' ' {
			cursor += 4 * scale
		} else {
			drawTinyGlyph(ch, cursor, y, scale, c)
			cursor += 4 * scale
		}
	}
}

func drawTinyDigit(digit uint8, x, y, scale int32, c colour.Colour) {
	drawTinyGlyph('0'+digit, x, y, scale, c)
}

func drawTinyGlyph(ch uint8, x, y, scale int32, c colour.Colour) {
	var rows [5]string
	switch ch {
	case '0':
		rows = [5]string{"111", "101", "101", "101", "111"}
	case '1':
		rows = [5]string{"010", "110", "010", "010", "111"}
	case '2':
		rows = [5]string{"111", "001", "111", "100", "111"}
	case '3':
		rows = [5]string{"111", "001", "111", "001", "111"}
	case '4':
		rows = [5]string{"101", "101", "111", "001", "001"}
	case '5':
		rows = [5]string{"111", "100", "111", "001", "111"}
	case '6':
		rows = [5]string{"111", "100", "111", "101", "111"}
	case '7':
		rows = [5]string{"111", "001", "010", "010", "010"}
	case '8':
		rows = [5]string{"111", "101", "111", "101", "111"}
	case '9':
		rows = [5]string{"111", "101", "111", "001", "111"}
	case 'A':
		rows = [5]string{"010", "101", "111", "101", "101"}
	case 'B':
		rows = [5]string{"110", "101", "110", "101", "110"}
	case 'C':
		rows = [5]string{"111", "100", "100", "100", "111"}
	case 'D':
		rows = [5]string{"110", "101", "101", "101", "110"}
	case 'E':
		rows = [5]string{"111", "100", "110", "100", "111"}
	case 'H':
		rows = [5]string{"101", "101", "111", "101", "101"}
	case 'I':
		rows = [5]string{"111", "010", "010", "010", "111"}
	case 'K':
		rows = [5]string{"101", "101", "110", "101", "101"}
	case 'L':
		rows = [5]string{"100", "100", "100", "100", "111"}
	case 'M':
		rows = [5]string{"101", "111", "111", "101", "101"}
	case 'N':
		rows = [5]string{"101", "111", "111", "111", "101"}
	case 'O':
		rows = [5]string{"111", "101", "101", "101", "111"}
	case 'P':
		rows = [5]string{"110", "101", "110", "100", "100"}
	case 'R':
		rows = [5]string{"110", "101", "110", "101", "101"}
	case 'S':
		rows = [5]string{"111", "100", "111", "001", "111"}
	case 'U':
		rows = [5]string{"101", "101", "101", "101", "111"}
	case ' ':
		rows = [5]string{"000", "000", "000", "000", "000"}
	default:
		rows = [5]string{"000", "000", "000", "000", "000"}
	}

	for rowIdx, row := range rows {
		for colIdx := 0; colIdx < len(row); colIdx++ {
			if row[colIdx] == '1' {
				canvasThree.ColourFilledRectangle(
					x+int32(colIdx)*scale,
					y+int32(rowIdx)*scale,
					scale,
					scale,
					c,
				)
			}
		}
	}
}

// ------------------------------------------------------------------------------------------------
// Entry Point
// ------------------------------------------------------------------------------------------------

func main() {
	dom.Init()
	dom.Log([]string{"Ok.", "Godom is starting.", "Here we go!"})

	toggleElements()
	setTitle()
	dom.AddNewStyleElement(bodyStyle)

	applicationContainer = dom.GetElementByID("application")
	createAppElements()
	populateArticleElement()

	// Register basic buttons
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

	// Create canvas rendering structures
	canvasOne = canvas.NewCanvas(800, 600, canvasOneBuffer[:], "canvasOneDiv")
	canvasTwo = canvas.NewCanvas(600, 450, canvasTwoBuffer[:], "canvasTwoDiv")
	canvasThree = canvas.NewCanvas(DrumCanvasW, DrumCanvasH, canvasThreeBuffer[:], "canvasThreeDiv")

	performDemoOnCanvasOne()
	initBalls()
	performDemoOnCanvasThree()
	setDrumButtonText()

	isReady = true

	// Canvas pixel readback demo
	px, ok := canvasOne.GetPixel(0, 0)
	if ok && !px.IsEmpty() {
		dom.Log([]string{"Canvas One pixel (0,0): non-empty after render."})
	}

	// DOM ID-based property access demo (setValue + getString)
	dom.SetValue("title", "lang", "en")
	lang := dom.GetString("title", "lang")
	dom.Log([]string{"dom.getString() on #title:", lang})

	// replaceClasses demo on title
	titleHandle := dom.GetElementByID("title")
	if titleHandle.IsValid() {
		titleHandle.ReplaceClasses([]string{"demo-title"})
	}

	// Builder .on() demo
	_ = html.Button().
		Text("Builder .on() demo").
		On("click", js.FuncOf(func(this js.Value, args []js.Value) any {
			onAddSomethingClick()
			return nil
		}).Value).
		AppendTo(applicationContainer)

	// Handle-based API showcase
	showcaseHeading := dom.CreateParagraphWithText("Handle-based API Showcase")
	showcaseHeading.AddClassTo("boo-header")
	dom.AddElementTo(asideElement, showcaseHeading)

	showcaseHeading.Set("title", "Created via handle-based dom.set() API")
	readback := showcaseHeading.Get("title")
	dom.Log([]string{"dom.get() on showcase heading:", readback})

	// Exercise dom.wrapElementWithNewDiv()
	showcaseP := dom.CreateParagraphWithText("This paragraph was wrapped via dom.wrapElementWithNewDiv()")
	wrapped := dom.WrapElementWithNewDiv(showcaseP, []string{"boo-wrapper"})
	dom.AddElementTo(asideElement, wrapped)

	// Exercise addClassTo on the aside element itself
	asideElement.AddClassTo("showcase-active")

	// Exercise addEventListener by handle
	dom.AddEventListener(articleElement, "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		dom.Log([]string{"Handle-based event listener fired on article element."})
		return nil
	}).Value)

	// Expose interaction variables and invoke_callback to JS
	js.Global().Set("go_set_interaction", js.FuncOf(func(this js.Value, args []js.Value) any {
		interactX = int32(args[0].Int())
		interactY = int32(args[1].Int())
		return nil
	}))

	js.Global().Set("go_invoke_callback", js.FuncOf(func(this js.Value, args []js.Value) any {
		id := uint32(args[0].Int())
		switch id {
		case 0:
			onAddSomethingClick()
		case 1:
			onClearAsideClick()
		case 2:
			onRefreshCanvasOneClick()
		case 3:
			onAnimationTick()
		case 4:
			onCanvasInteraction()
		case 6:
			onDrumCanvasClick()
		case 7:
			onDrumPlayPause()
		}
		return nil
	}))

	js.Global().Set("go_get_click_buffer", js.FuncOf(func(this js.Value, args []js.Value) any {
		byteSlice := unsafe.Slice((*byte)(unsafe.Pointer(&clickBuffer[0])), len(clickBuffer)*4)
		uint8Array := js.Global().Get("Uint8Array").New(len(byteSlice))
		js.CopyBytesToJS(uint8Array, byteSlice)
		float32Array := js.Global().Get("Float32Array").New(uint8Array.Get("buffer"), uint8Array.Get("byteOffset"), len(clickBuffer))
		return float32Array
	}))

	js.Global().Set("go_drum_is_beat_active", js.FuncOf(func(this js.Value, args []js.Value) any {
		track := args[0].Int()
		step := args[1].Int()
		if track >= DrumTracks || step >= DrumSteps {
			return 0
		}
		if drumPattern[track][step] {
			return 1
		}
		return 0
	}))
	js.Global().Set("go_drum_get_track_count", js.FuncOf(func(this js.Value, args []js.Value) any {
		return DrumTracks
	}))
	js.Global().Set("go_drum_get_step_count", js.FuncOf(func(this js.Value, args []js.Value) any {
		return DrumSteps
	}))

	// Pre-render UI click sound
	sound.FillClick(clickBuffer[:])

	// Start standard requestAnimationFrame tick loop in Go.
	// This calls onAnimationTick on every frame tick.
	animationLoopCb := js.FuncOf(func(this js.Value, args []js.Value) any {
		onAnimationTick()
		return nil
	})
	dom.StartAnimationLoop(animationLoopCb.Value)

	// Hold Go runtime scheduler alive forever
	select {}
}

// ------------------------------------------------------------------------------------------------
// Math and casting helpers
// ------------------------------------------------------------------------------------------------

func absF32(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

func maxUint8(a, b uint8) uint8 {
	if a > b {
		return a
	}
	return b
}