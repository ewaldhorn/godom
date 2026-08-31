// ------------------------------------------------------------------------------------------------
// Canvas Four — Grove & Pond (800×500, animated ducks + trees)
//
// Unlike canvases one/two/three, which rasterize into a pixel buffer and blit it via putImageData,
// this scene is drawn entirely with canvas_cmd: every shape, gradient, and sprite blit for the
// whole frame is recorded into one packed byte buffer and flushed to the real Canvas2D context
// with a SINGLE JS call (see canvasFourOpCount in the HUD, updated below). Trees and ducks are
// baked once into offscreen sprite canvases at init and then just re-blitted every frame — cheap
// regardless of how many draw ops went into the original artwork.
//
// Ported from OdinDOM's canvas_four.odin.
// ------------------------------------------------------------------------------------------------

package main

import (
	"fmt"
	"math"

	"github.com/ewaldhorn/godom/canvas_cmd"
	"github.com/ewaldhorn/godom/dom"
)

const (
	CanvasFourW = 800
	CanvasFourH = 500
	HorizonY    = 320

	// tau is a full turn in radians (2π); Go's math package has no Tau constant.
	tau = 2 * math.Pi

	PondCx = 400
	PondCy = 400
	PondRx = 320
	PondRy = 78

	SpriteTreeSmall = 0
	SpriteTreeBig   = 1
	SpriteDuck      = 2

	TreeSmallW = 70
	TreeSmallH = 110
	TreeBigW   = 100
	TreeBigH   = 150
	DuckW      = 48
	DuckH      = 36

	MaxTrees       = 6
	MaxDucks       = 4
	MaxRipples     = 6
	RippleLifetime = 1.2

	DuckMinX = PondCx - PondRx + 60
	DuckMaxX = PondCx + PondRx - 60
)

var (
	canvasFourCtx     dom.Handle
	canvasFourCmd     = canvas_cmd.NewBuffer()
	canvasFourOpCount int
	canvasFourTime    float32
)

type GroveTree struct {
	X, Y float32
	Big  bool
}

var trees [MaxTrees]GroveTree

type GroveDuck struct {
	X, BaseY float32
	Speed    float32
	Phase    float32
	LastCos  float32 // bob velocity sign from the previous frame, used to detect the low point of the bob
}

var ducks [MaxDucks]GroveDuck

type GroveRipple struct {
	X, Y float32
	Age  float32 // seconds since spawn; -1 = inactive
}

var ripples [MaxRipples]GroveRipple

func initCanvasFour() {
	parent := dom.GetElementByID("canvasFourDiv")
	canvas := dom.CanvasCreate(parent, CanvasFourW, CanvasFourH)
	canvasFourCtx = dom.CanvasGetContext(canvas)

	trees = [MaxTrees]GroveTree{
		{X: 40, Y: HorizonY, Big: false},
		{X: 130, Y: HorizonY, Big: true},
		{X: 690, Y: HorizonY, Big: true},
		{X: 760, Y: HorizonY, Big: false},
		{X: 220, Y: HorizonY - 6, Big: false},
		{X: 600, Y: HorizonY - 6, Big: false},
	}

	ducks = [MaxDucks]GroveDuck{
		{X: 150, BaseY: 400, Speed: 26, Phase: 0.0},
		{X: 420, BaseY: 420, Speed: -20, Phase: 1.4},
		{X: 300, BaseY: 440, Speed: 18, Phase: 2.8},
		{X: 560, BaseY: 405, Speed: -24, Phase: 4.2},
	}

	for i := range ripples {
		ripples[i].Age = -1
	}

	bakeCanvasFourSprites()
}

// bakeCanvasFourSprites records tree and duck artwork once into offscreen sprite canvases and
// flushes immediately. Every subsequent frame just re-blits these sprites (DrawSprite*) instead
// of re-encoding the underlying paths/gradients, keeping the per-frame buffer tiny.
func bakeCanvasFourSprites() {
	b := canvasFourCmd
	b.Reset()

	bakeTreeSprite(b, SpriteTreeSmall, TreeSmallW, TreeSmallH, 0.85)
	bakeTreeSprite(b, SpriteTreeBig, TreeBigW, TreeBigH, 1.0)
	bakeDuckSprite(b, SpriteDuck, DuckW, DuckH)

	b.Flush(canvasFourCtx)
}

func bakeTreeSprite(b *canvas_cmd.Buffer, id, w, h uint16, tint float32) {
	b.BakeBegin(id, w, h)

	fw := float32(w)
	fh := float32(h)
	trunkW := fw * 0.18
	trunkH := fh * 0.32

	b.SetFill("#5b3a22")
	b.FillRect(fw/2-trunkW/2, fh-trunkH, trunkW, trunkH)

	canopyColours := [3]string{"#1f5c2e", "#2c7a3d", "#3f9650"}
	cy := fh - trunkH
	for i := 0; i < 3; i++ {
		r := fw * (0.30 - float32(i)*0.03) * tint
		ox := fw/2 + float32(i-1)*fw*0.14
		oy := cy - float32(i)*fh*0.12 - r*0.5
		b.SetFill(canopyColours[i])
		b.BeginPath()
		b.Arc(ox, oy, r, 0, tau)
		b.Fill()
	}

	b.BakeEnd()
}

func bakeDuckSprite(b *canvas_cmd.Buffer, id, w, h uint16) {
	b.BakeBegin(id, w, h)

	fw := float32(w)
	fh := float32(h)

	b.SetFill("#ffd23d")
	b.BeginPath()
	b.Ellipse(fw*0.5, fh*0.62, fw*0.40, fh*0.34, 0, 0, tau)
	b.Fill()

	b.SetFill("#f2b705")
	b.BeginPath()
	b.Ellipse(fw*0.44, fh*0.60, fw*0.19, fh*0.18, 0.3, 0, tau)
	b.Fill()

	b.SetFill("#ffd23d")
	b.BeginPath()
	b.Arc(fw*0.78, fh*0.32, fh*0.24, 0, tau)
	b.Fill()

	b.SetFill("#f5a623")
	b.BeginPath()
	b.MoveTo(fw*0.94, fh*0.30)
	b.LineTo(fw*1.04, fh*0.34)
	b.LineTo(fw*0.94, fh*0.40)
	b.ClosePath()
	b.Fill()

	b.SetFill("#222222")
	b.BeginPath()
	b.Arc(fw*0.83, fh*0.28, fh*0.035, 0, tau)
	b.Fill()

	b.BakeEnd()
}

// updateCanvasFour re-records the whole scene into one buffer and flushes it with a single JS
// call. canvasFourOpCount tallies the commands that went into that one call, so the HUD can show
// what a naive per-op bridge crossing would have cost.
func updateCanvasFour() {
	if !isReady {
		return
	}

	b := canvasFourCmd
	b.Reset()
	canvasFourOpCount = 0
	canvasFourTime += 1.0 / 60.0

	drawCanvasFourSky(b)
	drawCanvasFourGround(b)
	drawCanvasFourPond(b)
	drawCanvasFourTrees(b)
	drawCanvasFourRipples(b)
	drawCanvasFourDucks(b)

	b.Flush(canvasFourCtx)
	updateCanvasFourHud()
}

func drawCanvasFourSky(b *canvas_cmd.Buffer) {
	b.LinearGradient(0, 0, 0, 0, HorizonY)
	b.AddColorStop(0, 0.0, "#ffd9a0")
	b.AddColorStop(0, 0.5, "#ffb6c9")
	b.AddColorStop(0, 1.0, "#9fd4ff")
	b.UseGradientFill(0)
	b.FillRect(0, 0, CanvasFourW, HorizonY)
	canvasFourOpCount += 5

	sunX, sunY := float32(640), float32(90)
	b.RadialGradient(1, sunX, sunY, 0, sunX, sunY, 120)
	b.AddColorStop(1, 0.0, "#fff6d8")
	b.AddColorStop(1, 1.0, "rgba(255,246,216,0)")
	b.UseGradientFill(1)
	b.BeginPath()
	b.Arc(sunX, sunY, 120, 0, tau)
	b.Fill()
	canvasFourOpCount += 6

	b.SetFill("#fff2c4")
	b.BeginPath()
	b.Arc(sunX, sunY, 34, 0, tau)
	b.Fill()
	canvasFourOpCount += 3
}

func drawCanvasFourGround(b *canvas_cmd.Buffer) {
	b.LinearGradient(2, 0, HorizonY, 0, CanvasFourH)
	b.AddColorStop(2, 0.0, "#4d8c3f")
	b.AddColorStop(2, 1.0, "#2f6329")
	b.UseGradientFill(2)
	b.FillRect(0, HorizonY, CanvasFourW, CanvasFourH-HorizonY)
	canvasFourOpCount += 4
}

func drawCanvasFourPond(b *canvas_cmd.Buffer) {
	b.LinearGradient(3, 0, PondCy-PondRy, 0, PondCy+PondRy)
	b.AddColorStop(3, 0.0, "#bfe6ff")
	b.AddColorStop(3, 1.0, "#3f8fc4")
	b.UseGradientFill(3)
	b.BeginPath()
	b.Ellipse(PondCx, PondCy, PondRx, PondRy, 0, 0, tau)
	b.Fill()
	canvasFourOpCount += 5

	b.SetStroke("rgba(255,255,255,0.35)")
	b.SetLineWidth(2)
	b.BeginPath()
	b.Ellipse(PondCx, PondCy, PondRx, PondRy, 0, 0, tau)
	b.Stroke()
	canvasFourOpCount += 4
}

func drawCanvasFourTrees(b *canvas_cmd.Buffer) {
	for _, t := range trees {
		var id uint16
		var w, h float32
		if t.Big {
			id = SpriteTreeBig
			w, h = TreeBigW, TreeBigH
		} else {
			id = SpriteTreeSmall
			w, h = TreeSmallW, TreeSmallH
		}
		b.DrawSprite(id, t.X-w/2, t.Y-h)
		canvasFourOpCount++
	}
}

func drawCanvasFourDucks(b *canvas_cmd.Buffer) {
	dt := float32(1.0 / 60.0)
	for i := range ducks {
		d := &ducks[i]
		d.X += d.Speed * dt
		if d.X < DuckMinX {
			d.X = DuckMinX
			d.Speed = -d.Speed
		}
		if d.X > DuckMaxX {
			d.X = DuckMaxX
			d.Speed = -d.Speed
		}
		theta := canvasFourTime*2.4 + d.Phase
		bob := float32(math.Sin(float64(theta))) * 4.0

		// A duck "bounces" on the water at the low point of its bob — the moment its downward
		// velocity (sin's derivative, cos) crosses back to positive. Spawn a ripple right there.
		cosTheta := float32(math.Cos(float64(theta)))
		if d.LastCos < 0 && cosTheta >= 0 {
			spawnCanvasFourRipple(d.X, d.BaseY+4.0)
		}
		d.LastCos = cosTheta

		dw := float32(DuckW)
		if d.Speed < 0 {
			dw = -DuckW
		}
		dx := d.X - DuckW/2
		dy := d.BaseY + bob - DuckH/2

		b.DrawSpriteScaled(SpriteDuck, dx, dy, dw, DuckH)
		canvasFourOpCount++
	}
}

func drawCanvasFourRipples(b *canvas_cmd.Buffer) {
	dt := float32(1.0 / 60.0)
	for i := range ripples {
		r := &ripples[i]
		if r.Age < 0 {
			continue
		}
		r.Age += dt
		if r.Age > RippleLifetime {
			r.Age = -1
			continue
		}
		t := r.Age / RippleLifetime
		radius := 8.0 + t*46.0
		alpha := (1.0 - t) * 0.6
		b.SetStroke(fmt.Sprintf("rgba(255,255,255,%.2f)", alpha))
		b.SetLineWidth(2)
		b.BeginPath()
		b.Ellipse(r.X, r.Y, radius, radius*0.35, 0, 0, tau)
		b.Stroke()
		canvasFourOpCount += 4
	}
}

func updateCanvasFourHud() {
	elem := dom.GetElementByID("canvasFourHud")
	if !elem.IsValid() {
		return
	}
	elem.SetInnerText(fmt.Sprintf(
		"%d Canvas2D commands batched into 1 WASM→JS call this frame (canvas_cmd) — vs %d individual bridge crossings without it",
		canvasFourOpCount,
		canvasFourOpCount,
	))
}

// spawnCanvasFourRipple activates a ripple at (x, y), reusing the oldest ripple slot once all
// MaxRipples are in flight. Shared by click interaction and the ducks' bob-bounce trigger.
func spawnCanvasFourRipple(x, y float32) {
	for i := range ripples {
		r := &ripples[i]
		if r.Age < 0 {
			r.X = x
			r.Y = y
			r.Age = 0
			return
		}
	}

	oldest := 0
	oldestAge := float32(-1)
	for i := range ripples {
		if ripples[i].Age > oldestAge {
			oldestAge = ripples[i].Age
			oldest = i
		}
	}
	ripples[oldest] = GroveRipple{X: x, Y: y, Age: 0}
}

// onCanvasFourInteraction spawns a ripple at the click/tap point. Interaction coordinates are
// reset after consuming, same as canvas two, so a stray re-fire is a no-op.
func onCanvasFourInteraction() {
	if !isReady {
		return
	}
	if interactX < 0 || interactY < 0 {
		return
	}
	x := float32(interactX)
	y := float32(interactY)
	interactX = -1
	interactY = -1

	spawnCanvasFourRipple(x, y)
}
