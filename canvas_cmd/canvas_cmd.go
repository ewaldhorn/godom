// Package canvas_cmd implements a retained draw-command buffer for driving a real JS
// CanvasRenderingContext2D from Go/WASM.
//
// WHY: godom's `canvas` package rasterizes into a pixel buffer and blits it via putImageData,
// which cannot reproduce gradients, text/emoji, or the full Canvas 2D drawing API. Games that
// need those instead want to issue real Canvas2D calls — but doing so one method at a time across
// the WASM<->JS boundary is expensive at 60fps.
//
// This package lets Go record an entire frame's worth of draw operations into one packed byte
// buffer, then flush it with a SINGLE JS call (domCanvasCmdFlush in docs/app.js). The JS side
// walks the buffer and replays each op onto the context. Encoding logic stays in Go; JS is a
// mechanical dispatcher.
//
// WIRE FORMAT (little-endian):
//
//	command := opcode:u8  followed by that opcode's typed args
//	f32 arg := 4 bytes IEEE-754 LE
//	str arg := len:u16 LE  followed by `len` UTF-8 bytes
//	u8  arg := 1 byte
//
// The Buffer is a fixed-capacity, preallocated arena reused every frame (Reset rewinds the
// cursor) — no per-frame heap churn. The wire format and opcode numbers are identical to
// OdinDOM's canvas_cmd package, so the JS dispatcher is shared nearly verbatim.
package canvas_cmd

import (
	"math"
	"syscall/js"

	"github.com/ewaldhorn/godom/dom"
)

// Capacity is the fixed buffer capacity — comfortably fits a full frame of a large city.
const Capacity = 1 << 20 // 1 MiB

// ------------------------------------------------------------------------------------------------
// Opcodes — keep in sync with the switch in docs/app.js `domCanvasCmdFlush`.
// ------------------------------------------------------------------------------------------------
const (
	OpSetFill         = 0x01 // str
	OpSetStroke       = 0x02 // str
	OpSetLineWidth    = 0x03 // f32
	OpSetFont         = 0x04 // str
	OpSetTextAlign    = 0x05 // str
	OpSetTextBaseline = 0x06 // str
	OpSetGlobalAlpha  = 0x07 // f32
	OpSetLineCap      = 0x08 // str

	OpFillRect   = 0x10 // f32 x4
	OpStrokeRect = 0x11 // f32 x4
	OpClearRect  = 0x12 // f32 x4

	OpBeginPath = 0x20
	OpMoveTo    = 0x21 // f32 x2
	OpLineTo    = 0x22 // f32 x2
	OpClosePath = 0x23
	OpArc       = 0x24 // x,y,r,a0,a1 (f32) + ccw (u8)
	OpEllipse   = 0x25 // x,y,rx,ry,rot,a0,a1 (f32) + ccw (u8)
	OpRect      = 0x26 // f32 x4
	OpBezier    = 0x27 // cx1,cy1,cx2,cy2,x,y (f32)
	OpFill      = 0x28
	OpStroke    = 0x29
	OpClip      = 0x2A

	OpFillText   = 0x30 // str, x, y (f32)
	OpStrokeText = 0x31 // str, x, y (f32)

	OpSave           = 0x40
	OpRestore        = 0x41
	OpTranslate      = 0x42 // f32 x2
	OpScale          = 0x43 // f32 x2
	OpRotate         = 0x44 // f32
	OpSetTransform   = 0x45 // a,b,c,d,e,f (f32)
	OpResetTransform = 0x46

	OpLinearGradient    = 0x50 // grad_id(u16), x0,y0,x1,y1 (f32)
	OpRadialGradient    = 0x51 // grad_id(u16), x0,y0,r0,x1,y1,r1 (f32)
	OpAddColorStop      = 0x52 // grad_id(u16), offset(f32), color(str)
	OpUseGradientFill   = 0x53 // grad_id(u16)
	OpUseGradientStroke = 0x54 // grad_id(u16)

	OpDrawSprite       = 0x60 // sprite_id(u16), dx,dy (f32)
	OpDrawSpriteScaled = 0x61 // sprite_id(u16), dx,dy,dw,dh (f32)
	OpDrawSpriteSub    = 0x62 // sprite_id(u16), sx,sy,sw,sh,dx,dy,dw,dh (f32)

	OpSetShadow   = 0x70 // color(str), blur(f32)
	OpClearShadow = 0x71

	OpBakeBegin = 0x80 // sprite_id(u16), w(u16), h(u16) — subsequent ops target the sprite canvas
	OpBakeEnd   = 0x81 // — subsequent ops target the main context again
)

// Buffer is a fixed-capacity command arena, reused across frames.
type Buffer struct {
	data       []byte
	overflowed bool
}

// NewBuffer returns an empty command arena with the package capacity.
func NewBuffer() *Buffer {
	return &Buffer{data: make([]byte, 0, Capacity)}
}

// Reset rewinds the buffer to empty. Call once at the start of every frame.
func (b *Buffer) Reset() {
	b.data = b.data[:0]
	b.overflowed = false
}

// Len returns the number of bytes currently recorded in the buffer.
func (b *Buffer) Len() int {
	return len(b.data)
}

// Flush replays the recorded commands onto ctx, then leaves the buffer as-is (call Reset next
// frame). No-op if nothing was recorded. The frame is copied into a JS-owned Uint8Array — Go
// cannot hand JS a pointer into its own heap — but this is still exactly ONE WASM->JS crossing
// per flush.
func (b *Buffer) Flush(ctx dom.Handle) {
	if len(b.data) == 0 {
		return
	}
	bytes := js.Global().Get("Uint8Array").New(len(b.data))
	js.CopyBytesToJS(bytes, b.data)
	js.Global().Get("domCanvasCmdFlush").Invoke(js.Value(ctx), bytes)
}

// ------------------------------------------------------------------------------------------------
// Low-level encoders (private). Overflowing the buffer sets overflowed and stops writing; the
// recorded data is discarded next frame via Reset.
// ------------------------------------------------------------------------------------------------

func (b *Buffer) putU8(v uint8) {
	if len(b.data)+1 > Capacity {
		b.overflowed = true
		return
	}
	b.data = append(b.data, v)
}

func (b *Buffer) putU16(v uint16) {
	if len(b.data)+2 > Capacity {
		b.overflowed = true
		return
	}
	b.data = append(b.data, uint8(v), uint8(v>>8))
}

func (b *Buffer) putF32(v float32) {
	if len(b.data)+4 > Capacity {
		b.overflowed = true
		return
	}
	bits := math.Float32bits(v)
	b.data = append(b.data, uint8(bits), uint8(bits>>8), uint8(bits>>16), uint8(bits>>24))
}

func (b *Buffer) putStr(s string) {
	n := len(s)
	if n > 0xFFFF {
		n = 0xFFFF
	}
	if len(b.data)+2+n > Capacity {
		b.overflowed = true
		return
	}
	b.data = append(b.data, uint8(n), uint8(n>>8))
	b.data = append(b.data, s[:n]...)
}

// ------------------------------------------------------------------------------------------------
// State setters.
// ------------------------------------------------------------------------------------------------

func (b *Buffer) SetFill(color string)   { b.putU8(OpSetFill); b.putStr(color) }
func (b *Buffer) SetStroke(color string) { b.putU8(OpSetStroke); b.putStr(color) }
func (b *Buffer) SetLineWidth(w float32) { b.putU8(OpSetLineWidth); b.putF32(w) }
func (b *Buffer) SetFont(font string)    { b.putU8(OpSetFont); b.putStr(font) }
func (b *Buffer) SetTextAlign(align string) {
	b.putU8(OpSetTextAlign)
	b.putStr(align)
}
func (b *Buffer) SetTextBaseline(baseline string) {
	b.putU8(OpSetTextBaseline)
	b.putStr(baseline)
}
func (b *Buffer) SetGlobalAlpha(a float32) { b.putU8(OpSetGlobalAlpha); b.putF32(a) }
func (b *Buffer) SetLineCap(cap string)    { b.putU8(OpSetLineCap); b.putStr(cap) }

// ------------------------------------------------------------------------------------------------
// Rectangles.
// ------------------------------------------------------------------------------------------------

func (b *Buffer) FillRect(x, y, w, h float32) {
	b.putU8(OpFillRect)
	b.putF32(x)
	b.putF32(y)
	b.putF32(w)
	b.putF32(h)
}
func (b *Buffer) StrokeRect(x, y, w, h float32) {
	b.putU8(OpStrokeRect)
	b.putF32(x)
	b.putF32(y)
	b.putF32(w)
	b.putF32(h)
}
func (b *Buffer) ClearRect(x, y, w, h float32) {
	b.putU8(OpClearRect)
	b.putF32(x)
	b.putF32(y)
	b.putF32(w)
	b.putF32(h)
}

// ------------------------------------------------------------------------------------------------
// Paths.
// ------------------------------------------------------------------------------------------------

func (b *Buffer) BeginPath() { b.putU8(OpBeginPath) }
func (b *Buffer) ClosePath() { b.putU8(OpClosePath) }
func (b *Buffer) Fill()      { b.putU8(OpFill) }
func (b *Buffer) Stroke()    { b.putU8(OpStroke) }
func (b *Buffer) Clip()      { b.putU8(OpClip) }

func (b *Buffer) MoveTo(x, y float32) {
	b.putU8(OpMoveTo)
	b.putF32(x)
	b.putF32(y)
}
func (b *Buffer) LineTo(x, y float32) {
	b.putU8(OpLineTo)
	b.putF32(x)
	b.putF32(y)
}
func (b *Buffer) Rect(x, y, w, h float32) {
	b.putU8(OpRect)
	b.putF32(x)
	b.putF32(y)
	b.putF32(w)
	b.putF32(h)
}

// Arc appends an arc to the current path. ccw is optional (default false).
func (b *Buffer) Arc(x, y, r, a0, a1 float32, ccw ...bool) {
	b.putU8(OpArc)
	b.putF32(x)
	b.putF32(y)
	b.putF32(r)
	b.putF32(a0)
	b.putF32(a1)
	b.putU8(boolToU8(ccw))
}

// Ellipse appends an ellipse to the current path. ccw is optional (default false).
func (b *Buffer) Ellipse(x, y, rx, ry, rot, a0, a1 float32, ccw ...bool) {
	b.putU8(OpEllipse)
	b.putF32(x)
	b.putF32(y)
	b.putF32(rx)
	b.putF32(ry)
	b.putF32(rot)
	b.putF32(a0)
	b.putF32(a1)
	b.putU8(boolToU8(ccw))
}

// ------------------------------------------------------------------------------------------------
// Text.
// ------------------------------------------------------------------------------------------------

func (b *Buffer) FillText(text string, x, y float32) {
	b.putU8(OpFillText)
	b.putStr(text)
	b.putF32(x)
	b.putF32(y)
}
func (b *Buffer) StrokeText(text string, x, y float32) {
	b.putU8(OpStrokeText)
	b.putStr(text)
	b.putF32(x)
	b.putF32(y)
}

// ------------------------------------------------------------------------------------------------
// Transforms.
// ------------------------------------------------------------------------------------------------

func (b *Buffer) Save()    { b.putU8(OpSave) }
func (b *Buffer) Restore() { b.putU8(OpRestore) }
func (b *Buffer) Translate(x, y float32) {
	b.putU8(OpTranslate)
	b.putF32(x)
	b.putF32(y)
}
func (b *Buffer) Scale(x, y float32) {
	b.putU8(OpScale)
	b.putF32(x)
	b.putF32(y)
}
func (b *Buffer) Rotate(a float32) { b.putU8(OpRotate); b.putF32(a) }

// BezierCurveTo appends a cubic Bézier curve to the current path.
func (b *Buffer) BezierCurveTo(cx1, cy1, cx2, cy2, x, y float32) {
	b.putU8(OpBezier)
	b.putF32(cx1)
	b.putF32(cy1)
	b.putF32(cx2)
	b.putF32(cy2)
	b.putF32(x)
	b.putF32(y)
}

// SetTransform replaces the current transform matrix: [a c e / b d f / 0 0 1].
func (b *Buffer) SetTransform(a, bb, c, d, e, f float32) {
	b.putU8(OpSetTransform)
	b.putF32(a)
	b.putF32(bb)
	b.putF32(c)
	b.putF32(d)
	b.putF32(e)
	b.putF32(f)
}
func (b *Buffer) ResetTransform() { b.putU8(OpResetTransform) }

// ------------------------------------------------------------------------------------------------
// Gradients. Create with an app-chosen `id` (small integers, reused each frame), add stops, then
// activate as the fill or stroke style. Gradients are recreated per-frame since they bake in
// absolute coordinates.
// ------------------------------------------------------------------------------------------------

func (b *Buffer) LinearGradient(id uint16, x0, y0, x1, y1 float32) {
	b.putU8(OpLinearGradient)
	b.putU16(id)
	b.putF32(x0)
	b.putF32(y0)
	b.putF32(x1)
	b.putF32(y1)
}
func (b *Buffer) RadialGradient(id uint16, x0, y0, r0, x1, y1, r1 float32) {
	b.putU8(OpRadialGradient)
	b.putU16(id)
	b.putF32(x0)
	b.putF32(y0)
	b.putF32(r0)
	b.putF32(x1)
	b.putF32(y1)
	b.putF32(r1)
}
func (b *Buffer) AddColorStop(id uint16, offset float32, color string) {
	b.putU8(OpAddColorStop)
	b.putU16(id)
	b.putF32(offset)
	b.putStr(color)
}
func (b *Buffer) UseGradientFill(id uint16) {
	b.putU8(OpUseGradientFill)
	b.putU16(id)
}
func (b *Buffer) UseGradientStroke(id uint16) {
	b.putU8(OpUseGradientStroke)
	b.putU16(id)
}

// ------------------------------------------------------------------------------------------------
// Shadows.
// ------------------------------------------------------------------------------------------------

func (b *Buffer) SetShadow(color string, blur float32) {
	b.putU8(OpSetShadow)
	b.putStr(color)
	b.putF32(blur)
}
func (b *Buffer) ClearShadow() { b.putU8(OpClearShadow) }

// ------------------------------------------------------------------------------------------------
// Sprites — bake once into an offscreen canvas, then blit cheaply with DrawSprite.
//
// Usage:
//
//	b.BakeBegin(id, w, h)   // subsequent draw ops render into sprite `id`'s canvas
//	... draw the sprite art ...
//	b.BakeEnd()             // back to the main context
//
// then per frame:
//
//	b.DrawSprite(id, x, y)
// ------------------------------------------------------------------------------------------------

func (b *Buffer) BakeBegin(id uint16, w, h uint16) {
	b.putU8(OpBakeBegin)
	b.putU16(id)
	b.putU16(w)
	b.putU16(h)
}
func (b *Buffer) BakeEnd() { b.putU8(OpBakeEnd) }

func (b *Buffer) DrawSprite(id uint16, dx, dy float32) {
	b.putU8(OpDrawSprite)
	b.putU16(id)
	b.putF32(dx)
	b.putF32(dy)
}
func (b *Buffer) DrawSpriteScaled(id uint16, dx, dy, dw, dh float32) {
	b.putU8(OpDrawSpriteScaled)
	b.putU16(id)
	b.putF32(dx)
	b.putF32(dy)
	b.putF32(dw)
	b.putF32(dh)
}
func (b *Buffer) DrawSpriteSub(id uint16, sx, sy, sw, sh, dx, dy, dw, dh float32) {
	b.putU8(OpDrawSpriteSub)
	b.putU16(id)
	b.putF32(sx)
	b.putF32(sy)
	b.putF32(sw)
	b.putF32(sh)
	b.putF32(dx)
	b.putF32(dy)
	b.putF32(dw)
	b.putF32(dh)
}

// ------------------------------------------------------------------------------------------------
// Text measurement (synchronous, does not use the buffer).
// ------------------------------------------------------------------------------------------------

var measureCtx js.Value

// MeasureText returns the rendered width (px) of text in font, measured on a dedicated offscreen
// context so it never disturbs the render context's state.
func MeasureText(font, text string) float32 {
	if measureCtx.IsUndefined() {
		measureCtx = js.Global().Get("document").Call("createElement", "canvas").Call("getContext", "2d")
	}
	measureCtx.Set("font", font)
	return float32(measureCtx.Call("measureText", text).Get("width").Float())
}

// ------------------------------------------------------------------------------------------------
// Helpers.
// ------------------------------------------------------------------------------------------------

func boolToU8(ccw []bool) uint8 {
	if len(ccw) > 0 && ccw[0] {
		return 1
	}
	return 0
}
