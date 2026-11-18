package main

import (
	"math"

	"github.com/ewaldhorn/godom/canvas"
	"github.com/ewaldhorn/godom/colour"
	"github.com/ewaldhorn/godom/dom"
)

// ------------------------------------------------------------------------------------------------
// Canvas One: Horizon Gallery and Vector Ship
// ------------------------------------------------------------------------------------------------

const (
	CanvasOneW       = 800
	CanvasOneH       = 600
	MaxCOneObstacles = 4
)

type COneObstacle struct {
	OriginX  float32
	Progress float32
	Speed    float32
	Active   bool
	Kind     uint8 // 0=asteroid, 1=bolt, 2=craft
}

type Theme struct {
	bg           colour.Colour
	panel        colour.Colour
	cyan         colour.Colour
	magenta      colour.Colour
	yellow       colour.Colour
	violet       colour.Colour
	sunStart     colour.Colour
	sunEnd       colour.Colour
	glowStartStr string
	glowMidStr   string
	glowEndStr   string
}

var (
	canvasOneBuffer [CanvasOneW * CanvasOneH * 4]byte
	canvasOne       *canvas.Canvas
	canvasOneTime   uint32
	gridOffset      float64
	currentThemeIdx int

	cOneObstacles    [MaxCOneObstacles]COneObstacle
	cOneSpawnTimer   uint32
	cOneShipX        float32 = 400
	cOneShipTargetX  float32 = 400
	cOneShipBank     float32
	cOneDecisionTick uint32
	cOneDodgeScore   uint32
	cOneInited       bool
)

var themes = []Theme{
	{
		bg:           colour.Colour{R: 5, G: 5, B: 16, A: 255},
		panel:        colour.Colour{R: 12, G: 12, B: 36, A: 255},
		cyan:         colour.Colour{R: 0, G: 240, B: 220, A: 255},
		magenta:      colour.Colour{R: 255, G: 0, B: 180, A: 255},
		yellow:       colour.Colour{R: 255, G: 230, B: 0, A: 255},
		violet:       colour.Colour{R: 160, G: 0, B: 255, A: 255},
		sunStart:     colour.Colour{R: 255, G: 230, B: 0, A: 255},
		sunEnd:       colour.Colour{R: 255, G: 0, B: 180, A: 255},
		glowStartStr: "rgba(255,0,180,0.12)",
		glowMidStr:   "rgba(255,180,30,0.22)",
		glowEndStr:   "rgba(255,230,0,0.32)",
	},
	{
		bg:           colour.Colour{R: 12, G: 4, B: 4, A: 255},
		panel:        colour.Colour{R: 24, G: 8, B: 8, A: 255},
		cyan:         colour.Colour{R: 255, G: 120, B: 0, A: 255},
		magenta:      colour.Colour{R: 255, G: 40, B: 0, A: 255},
		yellow:       colour.Colour{R: 255, G: 200, B: 0, A: 255},
		violet:       colour.Colour{R: 100, G: 0, B: 0, A: 255},
		sunStart:     colour.Colour{R: 255, G: 200, B: 0, A: 255},
		sunEnd:       colour.Colour{R: 255, G: 40, B: 0, A: 255},
		glowStartStr: "rgba(255,40,0,0.12)",
		glowMidStr:   "rgba(255,120,0,0.22)",
		glowEndStr:   "rgba(255,200,0,0.32)",
	},
	{
		bg:           colour.Colour{R: 20, G: 10, B: 30, A: 255},
		panel:        colour.Colour{R: 40, G: 20, B: 60, A: 255},
		cyan:         colour.Colour{R: 0, G: 255, B: 255, A: 255},
		magenta:      colour.Colour{R: 255, G: 105, B: 180, A: 255},
		yellow:       colour.Colour{R: 255, G: 255, B: 150, A: 255},
		violet:       colour.Colour{R: 138, G: 43, B: 226, A: 255},
		sunStart:     colour.Colour{R: 0, G: 255, B: 255, A: 255},
		sunEnd:       colour.Colour{R: 255, G: 105, B: 180, A: 255},
		glowStartStr: "rgba(255,105,180,0.12)",
		glowMidStr:   "rgba(138,43,226,0.22)",
		glowEndStr:   "rgba(0,255,255,0.32)",
	},
	{
		bg:           colour.Colour{R: 2, G: 8, B: 4, A: 255},
		panel:        colour.Colour{R: 4, G: 16, B: 8, A: 255},
		cyan:         colour.Colour{R: 0, G: 255, B: 100, A: 255},
		magenta:      colour.Colour{R: 0, G: 180, B: 50, A: 255},
		yellow:       colour.Colour{R: 200, G: 255, B: 200, A: 255},
		violet:       colour.Colour{R: 0, G: 80, B: 20, A: 255},
		sunStart:     colour.Colour{R: 200, G: 255, B: 200, A: 255},
		sunEnd:       colour.Colour{R: 0, G: 180, B: 50, A: 255},
		glowStartStr: "rgba(0,180,50,0.12)",
		glowMidStr:   "rgba(0,80,20,0.22)",
		glowEndStr:   "rgba(0,255,100,0.32)",
	},
}

func initCanvasOne() {
	canvasOne = canvas.NewCanvas(CanvasOneW, CanvasOneH, canvasOneBuffer[:], "canvasOneDiv")
	performDemoOnCanvasOne()
}

func onRefreshCanvasOneClick() {
	if !isReady {
		return
	}
	currentThemeIdx = (currentThemeIdx + 1) % len(themes)
}

func initCOneAI(w int32) {
	cOneShipX = float32(w) / 2
	cOneShipTargetX = float32(w) / 2
	cOneInited = true
}

func spawnCOneObstacle() {
	for i := range cOneObstacles {
		if !cOneObstacles[i].Active {
			side := float32(1)
			if nextRandom()%2 == 0 {
				side = -1
			}
			spread := float32(nextRandom()%260) + 40
			cOneObstacles[i] = COneObstacle{
				OriginX:  side * spread,
				Progress: 0.02,
				Speed:    0.0045 + float32(nextRandom()%10)*0.0003,
				Active:   true,
				Kind:     uint8(nextRandom() % 3),
			}
			return
		}
	}
}

func updateCOneObstacles(w int32) {
	cOneSpawnTimer++
	if cOneSpawnTimer > 280 {
		cOneSpawnTimer = 0
		spawnCOneObstacle()
	}

	cx := float32(w) / 2
	for i := range cOneObstacles {
		o := &cOneObstacles[i]
		if !o.Active {
			continue
		}
		o.Progress += o.Speed
		if o.Progress >= 1.05 {
			o.Active = false
			finalX := cx + o.OriginX
			if absF32(finalX-cOneShipX) > 45 {
				cOneDodgeScore++
			}
		}
	}
}

func updateCOneShipAI(w int32) {
	cx := float32(w) / 2

	var bestThreat float32
	var threatOrigX float32

	for i := range cOneObstacles {
		o := &cOneObstacles[i]
		if !o.Active || o.Progress < 0.15 {
			continue
		}
		finalX := cx + o.OriginX
		dist := absF32(finalX - cOneShipX)
		dangerZone := float32(100)
		threat := o.Progress * max(0, 1.0-dist/dangerZone)
		if threat > bestThreat {
			bestThreat = threat
			threatOrigX = o.OriginX
		}
	}

	cOneDecisionTick++

	if bestThreat > 0.28 && cOneDecisionTick > 30 {
		if threatOrigX > 0 {
			cOneShipTargetX = cx - 110 - float32(nextRandom()%50)
		} else {
			cOneShipTargetX = cx + 110 + float32(nextRandom()%50)
		}
		cOneDecisionTick = 0
	} else if cOneDecisionTick > 120 {
		cOneShipTargetX = cx + float32(int(nextRandom()%80)-40)
		cOneDecisionTick = 0
	}

	margin := float32(110)
	wF := float32(w)
	cOneShipTargetX = min(max(cOneShipTargetX, margin), wF-margin)

	diff := cOneShipTargetX - cOneShipX
	cOneShipX += diff * 0.018

	cOneShipBank = min(max(diff*0.003, -1.0), 1.0)
}

func drawCOneObstacle(o *COneObstacle, cx, horizon, h int32, t Theme) {
	screenX := int32(float32(cx) + o.OriginX*o.Progress)
	screenY := int32(float32(horizon) + o.Progress*float32(h-10-horizon))
	sz := int32(3 + int32(16*o.Progress))

	if screenY <= horizon || screenY >= h-5 {
		return
	}

	finalX := float32(cx) + o.OriginX
	isNearMiss := o.Progress > 0.6 && absF32(finalX-cOneShipX) < 85

	switch o.Kind {
	case 0: // asteroid
		col := t.magenta
		if isNearMiss {
			col = colour.White
		}
		canvasOne.ColourCircle(screenX, screenY, sz, col)
		canvasOne.ColourLine(screenX-sz-2, screenY, screenX+sz+2, screenY, col)
		canvasOne.ColourLine(screenX, screenY-sz, screenX, screenY+sz, col)
		canvasOne.ColourLine(screenX-sz/2, screenY-sz/2, screenX+sz/2, screenY+sz/2, col)
	case 1: // energy bolt
		col := t.yellow
		if isNearMiss {
			col = colour.White
		}
		canvasOne.ColourLine(screenX, screenY-sz*2, screenX, screenY+sz, col)
		canvasOne.ColourLine(screenX-sz, screenY-sz, screenX+sz, screenY-sz, col)
		canvasOne.ColourLine(screenX-sz/2, screenY, screenX+sz/2, screenY, col)
	case 2: // oncoming craft
		col := t.violet
		ec := t.cyan
		if isNearMiss {
			col = colour.White
			ec = colour.White
		}
		canvasOne.ColourLine(screenX-sz*2, screenY-sz/2, screenX, screenY+sz, col)
		canvasOne.ColourLine(screenX+sz*2, screenY-sz/2, screenX, screenY+sz, col)
		canvasOne.ColourLine(screenX-sz*2, screenY-sz/2, screenX+sz*2, screenY-sz/2, ec)
		canvasOne.ColourLine(screenX-sz/2, screenY-sz/4, screenX+sz/2, screenY-sz/4, ec)
	}
}

func drawCOneObstacles(cx, horizon, h int32, t Theme) {
	for i := range cOneObstacles {
		if cOneObstacles[i].Active {
			drawCOneObstacle(&cOneObstacles[i], cx, horizon, h, t)
		}
	}
}

func drawCOneHUD(t Theme) {
	score := cOneDodgeScore
	hundreds := uint8((score / 100) % 10)
	tens := uint8((score / 10) % 10)
	ones := uint8(score % 10)

	drawBitmapText(canvasOne, "SCORE", 20, 20, 3, t.cyan)
	drawBitmapDigit(canvasOne, hundreds, 86, 20, 3, t.yellow)
	drawBitmapDigit(canvasOne, tens, 98, 20, 3, t.yellow)
	drawBitmapDigit(canvasOne, ones, 110, 20, 3, t.yellow)
}

func performDemoOnCanvasOne() {
	w := int32(canvasOne.Width())
	h := int32(canvasOne.Height())
	horizon := h/2 + 50
	activeTheme := themes[currentThemeIdx]

	if !cOneInited {
		initCOneAI(w)
	}

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

	updateCOneObstacles(w)
	updateCOneShipAI(w)

	drawCOneObstacles(w/2, horizon, h, activeTheme)
	drawVectorShip(int32(cOneShipX), h-55, canvasOneTime, activeTheme)
	drawCOneHUD(activeTheme)

	canvasOne.Render()

	drawSunGlow(canvasOne.GetContext2D(), w, horizon-20, canvasOneTime, activeTheme)
}

func drawStarfield(w, horizon int32, time uint32, t Theme) {
	colour.Seed(999)
	for range 90 {
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
		red := uint8(float32(t.sunStart.R)*(1.0-ratio) + float32(t.sunEnd.R)*ratio)
		green := uint8(float32(t.sunStart.G)*(1.0-ratio) + float32(t.sunEnd.G)*ratio)
		blue := uint8(float32(t.sunStart.B)*(1.0-ratio) + float32(t.sunEnd.B)*ratio)
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

	for idx := range len(bgPts) - 1 {
		p1 := bgPts[idx]
		p2 := bgPts[idx+1]
		for x := p1[0]; x <= p2[0]; x++ {
			ratio := float32(x-p1[0]) / float32(p2[0]-p1[0])
			y := int32((1.0-ratio)*float32(p1[1]) + ratio*float32(p2[1]))
			canvasOne.ColourLine(x, y+1, x, horizon, t.bg)
		}
	}

	for idx := range len(bgPts) - 1 {
		canvasOne.ColourLine(bgPts[idx][0], bgPts[idx][1], bgPts[idx+1][0], bgPts[idx+1][1], t.violet)
	}

	for idx := range len(fgPts) - 1 {
		p1 := fgPts[idx]
		p2 := fgPts[idx+1]
		for x := p1[0]; x <= p2[0]; x++ {
			ratio := float32(x-p1[0]) / float32(p2[0]-p1[0])
			y := int32((1.0-ratio)*float32(p1[1]) + ratio*float32(p2[1]))
			canvasOne.ColourLine(x, y+1, x, horizon, t.bg)
		}
	}

	for idx := range len(fgPts) - 1 {
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
		for s := range segments {
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

	var i float64
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
			mid := int32(float64(w/2) + c)
			canvasOne.ColourLine(left, y, mid, y, t.violet)
			canvasOne.ColourLine(mid+1, y, right, y, t.magenta)
		}
		i += 1.0
	}
}

func drawVectorShip(cx, cy int32, time uint32, t Theme) {
	timeF := float64(time)
	bobY := int32(4.0 * math.Sin(timeF*0.06))
	swayX := int32(2.0 * math.Cos(timeF*0.04))

	scx := cx + swayX
	scy := cy + bobY

	bankOff := int32(cOneShipBank * 14)
	dodging := absF32(cOneShipBank) > 0.15

	flameLen := int32(26.0 + 4.0*math.Sin(timeF*0.25))
	if dodging {
		flameLen += 8
	}

	canvasOne.ColourLine(scx-8, scy+12, scx, scy+flameLen, t.cyan)
	canvasOne.ColourLine(scx+8, scy+12, scx, scy+flameLen, t.cyan)
	canvasOne.ColourLine(scx-8, scy+12, scx+8, scy+12, t.cyan)
	if dodging {
		canvasOne.ColourLine(scx-4, scy+14, scx, scy+flameLen+6, t.yellow)
		canvasOne.ColourLine(scx+4, scy+14, scx, scy+flameLen+6, t.yellow)
	}

	canvasOne.ColourLine(scx-35, scy+10+bankOff, scx+35, scy+10-bankOff, t.magenta)
	canvasOne.ColourLine(scx-35, scy+10+bankOff, scx-12, scy-20, t.magenta)
	canvasOne.ColourLine(scx+35, scy+10-bankOff, scx+12, scy-20, t.magenta)

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
	canvasOne.SetColour(shadow)
	canvasOne.Triangle(tri1, tri2, tri3)
}

func drawSunGlow(ctx dom.Context2D, w, cy int32, time uint32, t Theme) {
	cxF := float64(w / 2)
	cyF := float64(cy)
	timeF := float64(time)
	pulse := 3.0 * math.Sin(timeF*0.05)

	ctx.BeginPath()
	ctx.FillStyle(t.glowStartStr)
	ctx.Arc(cxF, cyF, 130.0+pulse, 0.0, tau, false)
	ctx.Fill()

	ctx.BeginPath()
	ctx.FillStyle(t.glowMidStr)
	ctx.Arc(cxF, cyF, 95.0-pulse, 0.0, tau, false)
	ctx.Fill()

	ctx.BeginPath()
	ctx.FillStyle(t.glowEndStr)
	ctx.Arc(cxF, cyF, 50.0+pulse, 0.0, tau, false)
	ctx.Fill()
}
