package main

import (
	"syscall/js"

	"github.com/ewaldhorn/godom/canvas"
	"github.com/ewaldhorn/godom/colour"
	"github.com/ewaldhorn/godom/dom"
)

// ------------------------------------------------------------------------------------------------
// Canvas Three: 16 Step Drum Machine Sequencer
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

var (
	canvasThreeBuffer [DrumCanvasW * DrumCanvasH * 4]byte
	canvasThree       *canvas.Canvas

	drumCurrentStep uint32
	drumPlaying     bool
	drumBpm         uint32 = 120
	drumLastStepMs  float64

	drumLabels = [DrumTracks]string{"K", "SN", "HH", "OH", "CL", "RM"}
)

var drumPattern = [DrumTracks][DrumSteps]bool{
	{true, false, false, false, true, false, false, false, true, false, false, false, true, false, false, false},
	{false, false, false, false, true, false, false, false, false, false, false, false, true, false, false, false},
	{true, false, true, false, true, false, true, false, true, false, true, false, true, false, true, false},
	{false, false, false, false, false, false, false, true, false, false, false, false, false, false, false, true},
	{false, false, false, false, false, false, true, false, false, false, false, false, false, false, true, false},
	{false, true, false, false, false, true, false, false, false, true, false, false, false, true, false, false},
}

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

func initCanvasThree() {
	canvasThree = canvas.NewCanvas(DrumCanvasW, DrumCanvasH, canvasThreeBuffer[:], "canvasThreeDiv")
	performDemoOnCanvasThree()
	setDrumButtonText()
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
	if rx%DrumCellPitchX >= DrumCellW || ry%DrumCellPitchY >= DrumCellH {
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
	setDrumButtonText()
	if drumPlaying {
		drumLastStepMs = js.Global().Get("performance").Call("now").Float()
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
	stepMs := 60000.0 / float64(drumBpm) / 4.0
	now := js.Global().Get("performance").Call("now").Float()
	for now-drumLastStepMs >= stepMs {
		drumLastStepMs += stepMs
		drumCurrentStep = (drumCurrentStep + 1) % DrumSteps
		playCurrentDrumStep()
	}
}

func playCurrentDrumStep() {
	step := int(drumCurrentStep)
	for track := range DrumTracks {
		if drumPattern[track][step] {
			js.Global().Call("drumHit", track)
		}
	}
}

func performDemoOnCanvasThree() {
	updateDrumSequencer()

	bg := colour.Colour{R: 5, G: 5, B: 16, A: 255}
	panel := colour.Colour{R: 12, G: 12, B: 34, A: 255}
	border := colour.Colour{R: 0, G: 240, B: 220, A: 255}
	muted := colour.Colour{R: 68, G: 68, B: 112, A: 255}

	canvasThree.ClearScreen(bg)
	canvasThree.ColourRectangle(8, 8, DrumCanvasW-16, DrumCanvasH-16, 2, border)
	canvasThree.ColourRectangle(14, 14, DrumCanvasW-28, DrumCanvasH-28, 1, panel)

	drawBitmapText(canvasThree, "DRUM 120 BPM", 26, 26, 2, border)
	if drumPlaying {
		drawBitmapText(canvasThree, "RUN", 400, 26, 2, drumNeon[5])
	} else {
		drawBitmapText(canvasThree, "PAUSE", 400, 26, 2, muted)
	}

	for step := range DrumSteps {
		x := int32(DrumGridX + step*DrumCellPitchX)
		stepLabelY := int32(DrumGridY - 24)
		if step%4 == 0 {
			canvasThree.ColourLine(x-5, DrumGridY-6, x-5, DrumGridY+DrumTracks*DrumCellPitchY-2, muted)
		}
		drawBitmapDigit(canvasThree, uint8((step+1)%10), x+8, stepLabelY, 1, muted)
	}

	for track := range DrumTracks {
		y := int32(DrumGridY + track*DrumCellPitchY)
		drawBitmapText(canvasThree, drumLabels[track], 24, y+8, 2, drumNeon[track])
		canvasThree.ColourFilledCircle(60, y+12, 4, drumNeon[track])

		for step := range DrumSteps {
			x := int32(DrumGridX + step*DrumCellPitchX)
			cellColour := drumDim[track]
			if drumPattern[track][step] {
				cellColour = drumNeon[track]
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
