package main

import "github.com/ewaldhorn/godom/dom"

// ------------------------------------------------------------------------------------------------
type BalloonGame struct {
	balloonStatus BalloonStatus
	treeManager   TreeManager
	ctx           dom.Context2D
}

// ------------------------------------------------------------------------------------------------
func NewGame(ctx dom.Context2D) *BalloonGame {
	game := &BalloonGame{
		balloonStatus: BalloonStatus{X: 200, Y: 200},
		treeManager:   *NewTreeManager(),
		ctx:           ctx,
	}
	return game
}

// ------------------------------------------------------------------------------------------------
func (b *BalloonGame) IsRising(state bool) {
	b.balloonStatus.IsRising = state
}

// ------------------------------------------------------------------------------------------------
func (b *BalloonGame) drawGameBackground() {
	b.ctx.ClearRect(0, 0, canvasWidth, canvasHeight)
	b.ctx.FillStyle("yellow")
	b.ctx.BeginPath()
	b.ctx.Arc(0.0, 0.0, 60.0, 0.0, 180.0, false)
	b.ctx.Fill()
}

// ------------------------------------------------------------------------------------------------
func (b *BalloonGame) render() {
	b.drawGameBackground()
	drawBalloon(b.balloonStatus, b.ctx)
	b.treeManager.Render(b.ctx)
}

// ------------------------------------------------------------------------------------------------
func (b *BalloonGame) Tick() {
	b.balloonStatus.Update()
	b.treeManager.Update()
	b.render()
}
