package dom

import (
	"syscall/js"
	"testing"
)

// ------------------------------------------------------------------------------------------------
func init() {
	js.Global().Get("eval").Invoke(`
		globalThis.Path2D = function(arg) {
			this.arg = arg || null;
		};

		globalThis.newMockContext2D = function() {
			const calls = [];
			const record = (name, args) => {
				calls.push({ name: name, args: Array.from(args) });
			};
			const canvas = {
				tagName: "canvas",
				width: 800,
				height: 600,
				getContext: function(kind) {
					if (kind === "2d") return ctx;
					return null;
				}
			};
			const ctx = {
				canvas: canvas,
				calls: calls,
				fillStyle: "#000000",
				strokeStyle: "#000000",
				lineWidth: 1.0,
				lineCap: "butt",
				lineJoin: "miter",
				miterLimit: 10.0,
				lineDashOffset: 0.0,
				lineDash: [],
				font: "10px sans-serif",
				textAlign: "start",
				textBaseline: "alphabetic",
				direction: "inherit",
				letterSpacing: "0px",
				wordSpacing: "0px",
				fontKerning: "auto",
				fontStretch: "normal",
				fontVariantCaps: "normal",
				textRendering: "auto",
				shadowBlur: 0.0,
				shadowColor: "rgba(0, 0, 0, 0)",
				shadowOffsetX: 0.0,
				shadowOffsetY: 0.0,
				globalAlpha: 1.0,
				globalCompositeOperation: "source-over",
				imageSmoothingEnabled: true,
				imageSmoothingQuality: "low",
				filter: "none",
				contextLost: false,
				transformMatrix: { a: 1, b: 0, c: 0, d: 1, e: 0, f: 0 },

				getContextAttributes: function() { record("getContextAttributes", arguments); return { alpha: true }; },
				isContextLost: function() { record("isContextLost", arguments); return this.contextLost; },
				save: function() { record("save", arguments); },
				restore: function() { record("restore", arguments); },
				reset: function() { record("reset", arguments); },

				clearRect: function() { record("clearRect", arguments); },
				fillRect: function() { record("fillRect", arguments); },
				strokeRect: function() { record("strokeRect", arguments); },

				fillText: function() { record("fillText", arguments); },
				strokeText: function() { record("strokeText", arguments); },
				measureText: function(txt) {
					record("measureText", arguments);
					return {
						width: txt.length * 8,
						actualBoundingBoxLeft: 1,
						actualBoundingBoxRight: txt.length * 8,
						fontBoundingBoxAscent: 12,
						fontBoundingBoxDescent: 3,
						actualBoundingBoxAscent: 10,
						actualBoundingBoxDescent: 2,
						emHeightAscent: 11,
						emHeightDescent: 2,
						hangingBaseline: 9,
						alphabeticBaseline: 0,
						ideographicBaseline: -2
					};
				},

				setLineDash: function(segs) {
					this.lineDash = Array.from(segs);
					record("setLineDash", arguments);
				},
				getLineDash: function() {
					record("getLineDash", arguments);
					return this.lineDash;
				},

				createLinearGradient: function() {
					record("createLinearGradient", arguments);
					const stops = [];
					return {
						kind: "linear",
						stops: stops,
						addColorStop: function(offset, color) { stops.push([offset, color]); }
					};
				},
				createRadialGradient: function() {
					record("createRadialGradient", arguments);
					const stops = [];
					return {
						kind: "radial",
						stops: stops,
						addColorStop: function(offset, color) { stops.push([offset, color]); }
					};
				},
				createConicGradient: function() {
					record("createConicGradient", arguments);
					const stops = [];
					return {
						kind: "conic",
						stops: stops,
						addColorStop: function(offset, color) { stops.push([offset, color]); }
					};
				},
				createPattern: function() {
					record("createPattern", arguments);
					return {
						kind: "pattern",
						matrix: null,
						setTransform: function(m) { this.matrix = m; }
					};
				},

				beginPath: function() { record("beginPath", arguments); },
				closePath: function() { record("closePath", arguments); },
				moveTo: function() { record("moveTo", arguments); },
				lineTo: function() { record("lineTo", arguments); },
				bezierCurveTo: function() { record("bezierCurveTo", arguments); },
				quadraticCurveTo: function() { record("quadraticCurveTo", arguments); },
				arc: function() { record("arc", arguments); },
				arcTo: function() { record("arcTo", arguments); },
				ellipse: function() { record("ellipse", arguments); },
				rect: function() { record("rect", arguments); },
				roundRect: function() { record("roundRect", arguments); },

				fill: function() { record("fill", arguments); },
				stroke: function() { record("stroke", arguments); },
				drawFocusIfNeeded: function() { record("drawFocusIfNeeded", arguments); },
				clip: function() { record("clip", arguments); },
				isPointInPath: function() { record("isPointInPath", arguments); return true; },
				isPointInStroke: function() { record("isPointInStroke", arguments); return true; },

				rotate: function() { record("rotate", arguments); },
				scale: function() { record("scale", arguments); },
				translate: function() { record("translate", arguments); },
				transform: function() { record("transform", arguments); },
				setTransform: function() { record("setTransform", arguments); },
				getTransform: function() { record("getTransform", arguments); return this.transformMatrix; },
				resetTransform: function() { record("resetTransform", arguments); },

				drawImage: function() { record("drawImage", arguments); },
				createImageData: function(w, h) {
					record("createImageData", arguments);
					if (typeof w === "object") {
						return { width: w.width, height: w.height };
					}
					return { width: w, height: h };
				},
				getImageData: function(sx, sy, sw, sh) {
					record("getImageData", arguments);
					return { width: sw, height: sh };
				},
				putImageData: function() { record("putImageData", arguments); }
			};
			return ctx;
		};
	`)
}

// ------------------------------------------------------------------------------------------------
func newTestContext2D() Context2D {
	return Context2D{Ctx: Handle(js.Global().Call("newMockContext2D"))}
}

// ------------------------------------------------------------------------------------------------
func assertCall(t *testing.T, ctx Context2D, idx int, expectedName string, expectedArgCount int) js.Value {
	t.Helper()
	calls := js.Value(ctx.Ctx).Get("calls")
	if idx >= calls.Length() {
		t.Fatalf("expected call %d (%s), only %d calls recorded", idx, expectedName, calls.Length())
	}
	entry := calls.Index(idx)
	if got := entry.Get("name").String(); got != expectedName {
		t.Errorf("call %d: expected %s, got %s", idx, expectedName, got)
	}
	args := entry.Get("args")
	if args.Length() != expectedArgCount {
		t.Errorf("call %d (%s): expected %d args, got %d", idx, expectedName, expectedArgCount, args.Length())
	}
	return args
}

// ------------------------------------------------------------------------------------------------
func TestNewContext2DAndState(t *testing.T) {
	raw := js.Global().Call("newMockContext2D")
	canvasHandle := Handle(raw.Get("canvas"))
	ctx := NewContext2D(canvasHandle)

	if !ctx.Canvas().IsValid() {
		t.Error("expected Canvas() to return a valid handle")
	}
	attrs := ctx.GetContextAttributes()
	if !attrs.Get("alpha").Bool() {
		t.Error("expected alpha context attribute to be true")
	}
	if ctx.IsContextLost() {
		t.Error("expected IsContextLost() to be false")
	}
	ctx.Save()
	ctx.Restore()
	ctx.Reset()

	assertCall(t, ctx, 0, "getContextAttributes", 0)
	assertCall(t, ctx, 1, "isContextLost", 0)
	assertCall(t, ctx, 2, "save", 0)
	assertCall(t, ctx, 3, "restore", 0)
	assertCall(t, ctx, 4, "reset", 0)
}

// ------------------------------------------------------------------------------------------------
func TestContext2DRectsAndText(t *testing.T) {
	ctx := newTestContext2D()

	ctx.ClearRect(0, 0, 100, 100)
	ctx.FillRect(10, 20, 30, 40)
	ctx.StrokeRect(5, 5, 50, 50)
	ctx.FillText("hi", 10, 20)
	ctx.FillTextMaxWidth("hi", 10, 20, 80)
	ctx.StrokeText("yo", 15, 25)
	ctx.StrokeTextMaxWidth("yo", 15, 25, 90)
	metrics := ctx.MeasureText("hello")

	assertCall(t, ctx, 0, "clearRect", 4)
	assertCall(t, ctx, 1, "fillRect", 4)
	assertCall(t, ctx, 2, "strokeRect", 4)
	assertCall(t, ctx, 3, "fillText", 3)
	assertCall(t, ctx, 4, "fillText", 4)
	assertCall(t, ctx, 5, "strokeText", 3)
	assertCall(t, ctx, 6, "strokeText", 4)
	assertCall(t, ctx, 7, "measureText", 1)

	if metrics.Width != 40 {
		t.Errorf("expected measured width 40, got %v", metrics.Width)
	}
	if metrics.ActualBoundingBoxAscent != 10 || metrics.IdeographicBaseline != -2 {
		t.Errorf("unexpected text metrics fields: %+v", metrics)
	}
}

// ------------------------------------------------------------------------------------------------
func TestContext2DStyleProperties(t *testing.T) {
	ctx := newTestContext2D()

	ctx.LineWidth(3.5)
	ctx.LineCap("round")
	ctx.LineJoin("bevel")
	ctx.MiterLimit(5.0)
	ctx.SetLineDash([]float64{4, 2})
	ctx.LineDashOffset(1.5)
	if ctx.GetLineWidth() != 3.5 || ctx.GetLineCap() != "round" || ctx.GetLineJoin() != "bevel" || ctx.GetMiterLimit() != 5.0 || ctx.GetLineDashOffset() != 1.5 {
		t.Error("line style getters did not return set values")
	}
	dash := ctx.GetLineDash()
	if len(dash) != 2 || dash[0] != 4 || dash[1] != 2 {
		t.Errorf("expected dash [4 2], got %v", dash)
	}

	ctx.Font("16px monospace")
	ctx.TextAlign("center")
	ctx.TextBaseline("middle")
	ctx.Direction("ltr")
	ctx.LetterSpacing("2px")
	ctx.WordSpacing("4px")
	ctx.FontKerning("none")
	ctx.FontStretch("condensed")
	ctx.FontVariantCaps("small-caps")
	ctx.TextRendering("geometricPrecision")
	if ctx.GetFont() != "16px monospace" || ctx.GetTextAlign() != "center" || ctx.GetTextBaseline() != "middle" ||
		ctx.GetDirection() != "ltr" || ctx.GetLetterSpacing() != "2px" || ctx.GetWordSpacing() != "4px" ||
		ctx.GetFontKerning() != "none" || ctx.GetFontStretch() != "condensed" ||
		ctx.GetFontVariantCaps() != "small-caps" || ctx.GetTextRendering() != "geometricPrecision" {
		t.Error("text style getters did not return set values")
	}

	ctx.ShadowBlur(8)
	ctx.ShadowColor("rgba(0,0,0,0.5)")
	ctx.ShadowOffsetX(2)
	ctx.ShadowOffsetY(4)
	if ctx.GetShadowBlur() != 8 || ctx.GetShadowColor() != "rgba(0,0,0,0.5)" ||
		ctx.GetShadowOffsetX() != 2 || ctx.GetShadowOffsetY() != 4 {
		t.Error("shadow getters did not return set values")
	}

	ctx.GlobalAlpha(0.75)
	ctx.GlobalCompositeOperation("multiply")
	ctx.ImageSmoothingEnabled(false)
	ctx.ImageSmoothingQuality("high")
	ctx.Filter("blur(4px)")
	if ctx.GetGlobalAlpha() != 0.75 || ctx.GetGlobalCompositeOperation() != "multiply" ||
		ctx.GetImageSmoothingEnabled() || ctx.GetImageSmoothingQuality() != "high" ||
		ctx.GetFilter() != "blur(4px)" {
		t.Error("compositing or filter getters did not return set values")
	}
}

// ------------------------------------------------------------------------------------------------
func TestContext2DGradientsAndPatterns(t *testing.T) {
	ctx := newTestContext2D()

	ctx.FillStyle("blue")
	if ctx.GetFillStyle().String() != "blue" {
		t.Errorf("expected fillStyle 'blue', got %v", ctx.GetFillStyle())
	}
	ctx.StrokeStyle("red")
	if ctx.GetStrokeStyle().String() != "red" {
		t.Errorf("expected strokeStyle 'red', got %v", ctx.GetStrokeStyle())
	}

	lg := ctx.CreateLinearGradient(0, 0, 100, 0)
	lg.AddColorStop(0, "red")
	lg.AddColorStop(1, "blue")
	ctx.FillStyleValue(lg)
	if ctx.GetFillStyle().Get("kind").String() != "linear" {
		t.Error("expected fillStyle to hold linear gradient")
	}

	rg := ctx.CreateRadialGradient(50, 50, 5, 50, 50, 40)
	rg.AddColorStop(0, "white")
	ctx.StrokeStyleValue(rg)
	if ctx.GetStrokeStyle().Get("kind").String() != "radial" {
		t.Error("expected strokeStyle to hold radial gradient")
	}

	cg := ctx.CreateConicGradient(0, 50, 50)
	cg.AddColorStop(0.5, "green")

	img := Handle(js.ValueOf(map[string]any{"tagName": "img"}))
	pat := ctx.CreatePattern(img, "repeat")
	pat.SetTransform(Handle(js.ValueOf(map[string]any{"a": 2})))
	ctx.FillStyleValue(pat)
	if ctx.GetFillStyle().Get("kind").String() != "pattern" {
		t.Error("expected fillStyle to hold pattern")
	}
}

// ------------------------------------------------------------------------------------------------
func TestContext2DPathsAndTransforms(t *testing.T) {
	ctx := newTestContext2D()
	p := NewPath2D("M0 0 L10 10")
	pCopy := NewPath2DFromPath(p)
	el := Handle(js.ValueOf(map[string]any{"id": "focus-el"}))

	ctx.BeginPath()
	ctx.MoveTo(0, 0)
	ctx.LineTo(10, 10)
	ctx.BezierCurveTo(1, 2, 3, 4, 5, 6)
	ctx.QuadraticCurveTo(1, 2, 3, 4)
	ctx.Arc(10, 10, 5, 0, 3.14, false)
	ctx.ArcTo(0, 0, 10, 10, 5)
	ctx.Ellipse(10, 10, 5, 8, 0, 0, 6.28, false)
	ctx.Rect(0, 0, 20, 20)
	ctx.RoundRect(0, 0, 20, 20)
	ctx.RoundRect(0, 0, 20, 20, 4)
	ctx.RoundRect(0, 0, 20, 20, 2, 4, 6, 8)
	ctx.ClosePath()

	ctx.Fill()
	ctx.FillRule("evenodd")
	ctx.FillPath(p)
	ctx.FillPath(pCopy, "evenodd")
	ctx.Stroke()
	ctx.StrokePath(p)
	ctx.DrawFocusIfNeeded(el)
	ctx.DrawFocusIfNeededPath(p, el)
	ctx.Clip()
	ctx.ClipRule("evenodd")
	ctx.ClipPath(p)
	ctx.ClipPath(p, "nonzero")
	if !ctx.IsPointInPath(5, 5) || !ctx.IsPointInPath(5, 5, "evenodd") {
		t.Error("expected IsPointInPath to return true")
	}
	if !ctx.IsPointInPathObj(p, 5, 5) || !ctx.IsPointInPathObj(p, 5, 5, "evenodd") {
		t.Error("expected IsPointInPathObj to return true")
	}
	if !ctx.IsPointInStroke(5, 5) || !ctx.IsPointInStrokeObj(p, 5, 5) {
		t.Error("expected IsPointInStroke helpers to return true")
	}

	ctx.Rotate(0.5)
	ctx.Scale(2, 2)
	ctx.Translate(10, 20)
	ctx.Transform(1, 0, 0, 1, 5, 5)
	ctx.SetTransform(1, 0, 0, 1, 0, 0)
	m := ctx.GetTransform()
	ctx.SetTransformMatrix(m)
	ctx.ResetTransform()

	assertCall(t, ctx, 0, "beginPath", 0)
	assertCall(t, ctx, 10, "roundRect", 5)
	argsMulti := assertCall(t, ctx, 11, "roundRect", 5)
	if argsMulti.Index(4).Length() != 4 {
		t.Errorf("expected 4 corner radii in array, got %d", argsMulti.Index(4).Length())
	}
	assertCall(t, ctx, 13, "fill", 0)
	assertCall(t, ctx, 14, "fill", 1)
	assertCall(t, ctx, 15, "fill", 1)
	assertCall(t, ctx, 16, "fill", 2)
	assertCall(t, ctx, 37, "setTransform", 1)
	assertCall(t, ctx, 38, "resetTransform", 0)
}

// ------------------------------------------------------------------------------------------------
func TestContext2DImagesAndPixels(t *testing.T) {
	ctx := newTestContext2D()
	img := Handle(js.ValueOf(map[string]any{"tagName": "img"}))

	ctx.DrawImage(img, 10, 20)
	ctx.DrawImageScaled(img, 10, 20, 100, 50)
	ctx.DrawImageSubrect(img, 0, 0, 16, 16, 10, 20, 32, 32)

	imgData := ctx.CreateImageData(64, 32)
	imgData2 := ctx.CreateImageDataFrom(imgData)
	gotData := ctx.GetImageData(0, 0, 64, 32)
	ctx.PutImageData(gotData, 0, 0)
	ctx.PutImageDataDirty(imgData2, 0, 0, 4, 4, 16, 16)

	assertCall(t, ctx, 0, "drawImage", 3)
	assertCall(t, ctx, 1, "drawImage", 5)
	assertCall(t, ctx, 2, "drawImage", 9)
	assertCall(t, ctx, 3, "createImageData", 2)
	assertCall(t, ctx, 4, "createImageData", 1)
	assertCall(t, ctx, 5, "getImageData", 4)
	assertCall(t, ctx, 6, "putImageData", 3)
	assertCall(t, ctx, 7, "putImageData", 7)
}
