package main

import (
	"math"

	"github.com/ewaldhorn/godom/canvas"
	"github.com/ewaldhorn/godom/colour"
	"github.com/ewaldhorn/godom/dom"
)

// ------------------------------------------------------------------------------------------------
// Canvas Two: 2D Ball Physics Simulation
// ------------------------------------------------------------------------------------------------

const (
	CanvasTwoW = 600
	CanvasTwoH = 450
	MaxBalls   = 14
	TrailLen   = 10
)

type Ball struct {
	X, Y, Vx, Vy float32
	Radius       float32
	Col          colour.Colour
}

var (
	canvasTwoBuffer [CanvasTwoW * CanvasTwoH * 4]byte
	canvasTwo       *canvas.Canvas

	balls      [MaxBalls]Ball
	ballTrails [MaxBalls][TrailLen][2]float32
	trailHead  [MaxBalls]int

	// Gravity: 0=down, 1=left, 2=up, 3=right, 4=zero-g
	gravityMode      int
	gravityModeNames = [5]string{"DOWN", "LEFT", "UP", "RIGHT", "ZERO-G"}
)

func initCanvasTwo() {
	canvasTwo = canvas.NewCanvas(CanvasTwoW, CanvasTwoH, canvasTwoBuffer[:], "canvasTwoDiv")
	initBalls()
	updateGravityButtonText()
}

func initBalls() {
	w := float32(canvasTwo.Width())
	h := float32(canvasTwo.Height())

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

	for i := range MaxBalls {
		d := initData[i]
		balls[i] = Ball{
			X:      w * float32(d[0]) / 100.0,
			Y:      h * float32(d[1]) / 100.0,
			Vx:     float32(d[2]) / 100.0,
			Vy:     float32(d[3]) / 100.0,
			Radius: float32(d[4]),
			Col:    ballColours[i],
		}
		for t := range TrailLen {
			ballTrails[i][t][0] = balls[i].X
			ballTrails[i][t][1] = balls[i].Y
		}
		trailHead[i] = 0
	}
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
		impulse := min(180.0/dist, 12.0)
		balls[i].Vx += (dx / dist) * impulse
		balls[i].Vy += (dy / dist) * impulse
	}
}

func onCycleGravity() {
	if !isReady {
		return
	}
	gravityMode = (gravityMode + 1) % len(gravityModeNames)
	updateGravityButtonText()
}

func updateGravityButtonText() {
	btn := dom.GetElementByID("gravityButton")
	if !btn.IsValid() {
		return
	}
	btn.SetInnerText("Gravity: " + gravityModeNames[gravityMode])
}

func updateCanvasTwo() {
	w := float32(canvasTwo.Width())
	h := float32(canvasTwo.Height())

	dampen := float32(0.95)
	if gravityMode == 4 {
		dampen = 1.0 // elastic walls in zero-g
	}

	bg := colour.Colour{R: 8, G: 8, B: 15, A: 255}
	canvasTwo.ClearScreen(bg)

	// Record trail positions before updating physics
	for i := range balls {
		trailHead[i] = (trailHead[i] + 1) % TrailLen
		ballTrails[i][trailHead[i]][0] = balls[i].X
		ballTrails[i][trailHead[i]][1] = balls[i].Y
	}

	// Apply gravity based on mode
	var gx, gy float32
	switch gravityMode {
	case 0:
		gy = 0.08
	case 1:
		gx = -0.08
	case 2:
		gy = -0.08
	case 3:
		gx = 0.08
	}

	for i := range balls {
		ball := &balls[i]
		ball.Vx += gx
		ball.Vy += gy
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
	for i := range MaxBalls {
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

	// Draw trails
	for i := range balls {
		ball := &balls[i]
		for t := 1; t < TrailLen; t++ {
			idx := (trailHead[i] - t + TrailLen) % TrailLen
			tx := int32(ballTrails[i][idx][0])
			ty := int32(ballTrails[i][idx][1])
			alpha := float32(TrailLen-t) / float32(TrailLen)
			tr := colour.Colour{
				R: uint8(float32(ball.Col.R) * alpha * 0.35),
				G: uint8(float32(ball.Col.G) * alpha * 0.35),
				B: uint8(float32(ball.Col.B) * alpha * 0.35),
				A: 255,
			}
			trailR := max(int32(float32(ball.Radius)*alpha*0.75), 2)
			canvasTwo.ColourFilledCircle(tx, ty, trailR, tr)
		}
	}

	// Draw balls with velocity-based glow
	for i := range balls {
		ball := &balls[i]
		bx := int32(ball.X)
		by := int32(ball.Y)
		br := int32(ball.Radius)

		speed := float32(math.Sqrt(float64(ball.Vx*ball.Vx + ball.Vy*ball.Vy)))
		glowExtra := min(int32(speed*0.4), 10)

		glow := colour.Colour{
			R: max(ball.Col.R/3, 20),
			G: max(ball.Col.G/3, 20),
			B: max(ball.Col.B/3, 20),
			A: 255,
		}
		canvasTwo.ColourFilledCircle(bx, by, br+4+glowExtra, glow)
		canvasTwo.ColourFilledCircle(bx, by, br, ball.Col)

		hs := int32(ball.Radius / 3.0)
		canvasTwo.ColourPutPixel(bx-hs, by-hs, colour.White)

		if i == 0 {
			canvasTwo.ColourBorderCircle(bx, by, br+8+glowExtra, 2, colour.White)
		}
	}

	canvasTwo.Render()
}
