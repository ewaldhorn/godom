package dom

import "syscall/js"

// ------------------------------------------------------------------------------------------------
// Context2D wraps the HTML5 CanvasRenderingContext2D interface.
type Context2D struct {
	Ctx Handle
}

// ------------------------------------------------------------------------------------------------
// NewContext2D retrieves the 2D rendering context from a canvas element handle.
func NewContext2D(canvas Handle) Context2D {
	return Context2D{Ctx: CanvasGetContext(canvas)}
}

// ------------------------------------------------------------------------------------------------
// CanvasGradient wraps a CanvasGradient JS object returned by gradient constructors.
type CanvasGradient struct {
	Handle Handle
}

// ------------------------------------------------------------------------------------------------
// AddColorStop adds a new color stop defined by an offset and a CSS color string.
func (g CanvasGradient) AddColorStop(offset float64, color string) {
	js.Value(g.Handle).Call("addColorStop", offset, color)
}

// ------------------------------------------------------------------------------------------------
// CanvasPattern wraps a CanvasPattern JS object returned by CreatePattern.
type CanvasPattern struct {
	Handle Handle
}

// ------------------------------------------------------------------------------------------------
// SetTransform applies a DOMMatrix transformation matrix to the pattern.
func (p CanvasPattern) SetTransform(matrix Handle) {
	js.Value(p.Handle).Call("setTransform", js.Value(matrix))
}

// ------------------------------------------------------------------------------------------------
// TextMetrics holds the measurements returned by MeasureText.
type TextMetrics struct {
	Width                    float64
	ActualBoundingBoxLeft    float64
	ActualBoundingBoxRight   float64
	FontBoundingBoxAscent    float64
	FontBoundingBoxDescent   float64
	ActualBoundingBoxAscent  float64
	ActualBoundingBoxDescent float64
	EmHeightAscent           float64
	EmHeightDescent          float64
	HangingBaseline          float64
	AlphabeticBaseline       float64
	IdeographicBaseline      float64
}

// ------------------------------------------------------------------------------------------------
// NewPath2D constructs a browser Path2D object, optionally initialized from an SVG path string.
func NewPath2D(svgPath ...string) Handle {
	ctor := js.Global().Get("Path2D")
	if len(svgPath) > 0 {
		return Handle(ctor.New(svgPath[0]))
	}
	return Handle(ctor.New())
}

// ------------------------------------------------------------------------------------------------
// NewPath2DFromPath constructs a browser Path2D object copied from an existing Path2D handle.
func NewPath2DFromPath(path Handle) Handle {
	return Handle(js.Global().Get("Path2D").New(js.Value(path)))
}

// ------------------------------------------------------------------------------------------------
func (c Context2D) call(method string, args ...any) js.Value {
	return js.Value(c.Ctx).Call(method, args...)
}

// ------------------------------------------------------------------------------------------------
func (c Context2D) setProp(prop string, val any) {
	js.Value(c.Ctx).Set(prop, val)
}

// ------------------------------------------------------------------------------------------------
func (c Context2D) getStringProp(prop string) string {
	return readStringProp(js.Value(c.Ctx), prop)
}

// ------------------------------------------------------------------------------------------------
func (c Context2D) getFloatProp(prop string) float64 {
	return readFloatProp(js.Value(c.Ctx), prop)
}

// ------------------------------------------------------------------------------------------------
func (c Context2D) getBoolProp(prop string) bool {
	v := js.Value(c.Ctx).Get(prop)
	if v.IsUndefined() || v.IsNull() {
		return false
	}
	return v.Bool()
}

// ------------------------------------------------------------------------------------------------
func readStringProp(v js.Value, key string) string {
	prop := v.Get(key)
	if prop.IsUndefined() || prop.IsNull() {
		return ""
	}
	return prop.String()
}

// ------------------------------------------------------------------------------------------------
func readFloatProp(v js.Value, key string) float64 {
	prop := v.Get(key)
	if prop.IsUndefined() || prop.IsNull() {
		return 0
	}
	return prop.Float()
}

// ------------------------------------------------------------------------------------------------
func floatSliceToJSArray(vals []float64) js.Value {
	arr := js.Global().Get("Array").New(len(vals))
	for i, v := range vals {
		arr.SetIndex(i, v)
	}
	return arr
}

// ------------------------------------------------------------------------------------------------
func unwrapStyleValue(style any) any {
	switch v := style.(type) {
	case CanvasGradient:
		return js.Value(v.Handle)
	case CanvasPattern:
		return js.Value(v.Handle)
	case Handle:
		return js.Value(v)
	default:
		return style
	}
}

// ================================================================================================
// Context metadata and state
// ================================================================================================

// ------------------------------------------------------------------------------------------------
// Canvas returns the handle to the HTMLCanvasElement associated with this context.
func (c Context2D) Canvas() Handle {
	return Handle(js.Value(c.Ctx).Get("canvas"))
}

// ------------------------------------------------------------------------------------------------
// GetContextAttributes returns the JS object containing the context creation attributes.
func (c Context2D) GetContextAttributes() js.Value {
	return c.call("getContextAttributes")
}

// ------------------------------------------------------------------------------------------------
// IsContextLost reports whether the rendering context has been lost.
func (c Context2D) IsContextLost() bool {
	return c.call("isContextLost").Bool()
}

// ------------------------------------------------------------------------------------------------
// Save pushes the current drawing state onto the state stack.
func (c Context2D) Save() {
	c.call("save")
}

// ------------------------------------------------------------------------------------------------
// Restore pops the most recently saved drawing state from the state stack.
func (c Context2D) Restore() {
	c.call("restore")
}

// ------------------------------------------------------------------------------------------------
// Reset resets the rendering context to its default state and clears the backing buffer.
func (c Context2D) Reset() {
	c.call("reset")
}

// ================================================================================================
// Rectangles
// ================================================================================================

// ------------------------------------------------------------------------------------------------
// ClearRect erases the pixels in a rectangular area by setting them to transparent black.
func (c Context2D) ClearRect(x, y, width, height float64) {
	c.call("clearRect", x, y, width, height)
}

// ------------------------------------------------------------------------------------------------
// FillRect draws a filled rectangle using the current fill style.
func (c Context2D) FillRect(x, y, width, height float64) {
	c.call("fillRect", x, y, width, height)
}

// ------------------------------------------------------------------------------------------------
// StrokeRect strokes a rectangle outline using the current stroke style.
func (c Context2D) StrokeRect(x, y, width, height float64) {
	c.call("strokeRect", x, y, width, height)
}

// ================================================================================================
// Text drawing and measurement
// ================================================================================================

// ------------------------------------------------------------------------------------------------
// FillText draws filled text at the specified coordinates.
func (c Context2D) FillText(text string, x, y float64) {
	c.call("fillText", text, x, y)
}

// ------------------------------------------------------------------------------------------------
// FillTextMaxWidth draws filled text constrained to the specified maximum width.
func (c Context2D) FillTextMaxWidth(text string, x, y, maxWidth float64) {
	c.call("fillText", text, x, y, maxWidth)
}

// ------------------------------------------------------------------------------------------------
// StrokeText strokes text outlines at the specified coordinates.
func (c Context2D) StrokeText(text string, x, y float64) {
	c.call("strokeText", text, x, y)
}

// ------------------------------------------------------------------------------------------------
// StrokeTextMaxWidth strokes text outlines constrained to the specified maximum width.
func (c Context2D) StrokeTextMaxWidth(text string, x, y, maxWidth float64) {
	c.call("strokeText", text, x, y, maxWidth)
}

// ------------------------------------------------------------------------------------------------
// MeasureText returns a TextMetrics struct with dimensions for the given text string.
func (c Context2D) MeasureText(text string) TextMetrics {
	m := c.call("measureText", text)
	return TextMetrics{
		Width:                    readFloatProp(m, "width"),
		ActualBoundingBoxLeft:    readFloatProp(m, "actualBoundingBoxLeft"),
		ActualBoundingBoxRight:   readFloatProp(m, "actualBoundingBoxRight"),
		FontBoundingBoxAscent:    readFloatProp(m, "fontBoundingBoxAscent"),
		FontBoundingBoxDescent:   readFloatProp(m, "fontBoundingBoxDescent"),
		ActualBoundingBoxAscent:  readFloatProp(m, "actualBoundingBoxAscent"),
		ActualBoundingBoxDescent: readFloatProp(m, "actualBoundingBoxDescent"),
		EmHeightAscent:           readFloatProp(m, "emHeightAscent"),
		EmHeightDescent:          readFloatProp(m, "emHeightDescent"),
		HangingBaseline:          readFloatProp(m, "hangingBaseline"),
		AlphabeticBaseline:       readFloatProp(m, "alphabeticBaseline"),
		IdeographicBaseline:      readFloatProp(m, "ideographicBaseline"),
	}
}

// ================================================================================================
// Line styles
// ================================================================================================

// ------------------------------------------------------------------------------------------------
// LineWidth sets the thickness of lines.
func (c Context2D) LineWidth(width float64) { c.setProp("lineWidth", width) }

// ------------------------------------------------------------------------------------------------
// GetLineWidth returns the current line thickness.
func (c Context2D) GetLineWidth() float64 { return c.getFloatProp("lineWidth") }

// ------------------------------------------------------------------------------------------------
// LineCap sets the shape used to draw the end points of lines ("butt", "round", or "square").
func (c Context2D) LineCap(cap string) { c.setProp("lineCap", cap) }

// ------------------------------------------------------------------------------------------------
// GetLineCap returns the current line cap style.
func (c Context2D) GetLineCap() string { return c.getStringProp("lineCap") }

// ------------------------------------------------------------------------------------------------
// LineJoin sets the shape used to join two line segments ("round", "bevel", or "miter").
func (c Context2D) LineJoin(join string) { c.setProp("lineJoin", join) }

// ------------------------------------------------------------------------------------------------
// GetLineJoin returns the current line join style.
func (c Context2D) GetLineJoin() string { return c.getStringProp("lineJoin") }

// ------------------------------------------------------------------------------------------------
// MiterLimit sets the miter limit ratio.
func (c Context2D) MiterLimit(limit float64) { c.setProp("miterLimit", limit) }

// ------------------------------------------------------------------------------------------------
// GetMiterLimit returns the current miter limit ratio.
func (c Context2D) GetMiterLimit() float64 { return c.getFloatProp("miterLimit") }

// ------------------------------------------------------------------------------------------------
// SetLineDash sets the line dash pattern used when stroking lines.
func (c Context2D) SetLineDash(segments []float64) {
	c.call("setLineDash", floatSliceToJSArray(segments))
}

// ------------------------------------------------------------------------------------------------
// GetLineDash returns the current line dash pattern.
func (c Context2D) GetLineDash() []float64 {
	arr := c.call("getLineDash")
	n := arr.Length()
	out := make([]float64, n)
	for i := range n {
		out[i] = arr.Index(i).Float()
	}
	return out
}

// ------------------------------------------------------------------------------------------------
// LineDashOffset sets the line dash offset (phase).
func (c Context2D) LineDashOffset(offset float64) { c.setProp("lineDashOffset", offset) }

// ------------------------------------------------------------------------------------------------
// GetLineDashOffset returns the current line dash offset.
func (c Context2D) GetLineDashOffset() float64 { return c.getFloatProp("lineDashOffset") }

// ================================================================================================
// Text styles
// ================================================================================================

// ------------------------------------------------------------------------------------------------
// Font sets the current CSS font specification used when drawing text.
func (c Context2D) Font(font string) { c.setProp("font", font) }

// ------------------------------------------------------------------------------------------------
// GetFont returns the current font specification.
func (c Context2D) GetFont() string { return c.getStringProp("font") }

// ------------------------------------------------------------------------------------------------
// TextAlign sets the current text alignment ("left", "right", "center", "start", or "end").
func (c Context2D) TextAlign(align string) { c.setProp("textAlign", align) }

// ------------------------------------------------------------------------------------------------
// GetTextAlign returns the current text alignment.
func (c Context2D) GetTextAlign() string { return c.getStringProp("textAlign") }

// ------------------------------------------------------------------------------------------------
// TextBaseline sets the current text baseline ("top", "hanging", "middle", "alphabetic", "ideographic", or "bottom").
func (c Context2D) TextBaseline(baseline string) { c.setProp("textBaseline", baseline) }

// ------------------------------------------------------------------------------------------------
// GetTextBaseline returns the current text baseline.
func (c Context2D) GetTextBaseline() string { return c.getStringProp("textBaseline") }

// ------------------------------------------------------------------------------------------------
// Direction sets the current text direction ("ltr", "rtl", or "inherit").
func (c Context2D) Direction(dir string) { c.setProp("direction", dir) }

// ------------------------------------------------------------------------------------------------
// GetDirection returns the current text direction.
func (c Context2D) GetDirection() string { return c.getStringProp("direction") }

// ------------------------------------------------------------------------------------------------
// LetterSpacing sets the spacing between letters as a CSS length string.
func (c Context2D) LetterSpacing(spacing string) { c.setProp("letterSpacing", spacing) }

// ------------------------------------------------------------------------------------------------
// GetLetterSpacing returns the current letter spacing.
func (c Context2D) GetLetterSpacing() string { return c.getStringProp("letterSpacing") }

// ------------------------------------------------------------------------------------------------
// WordSpacing sets the spacing between words as a CSS length string.
func (c Context2D) WordSpacing(spacing string) { c.setProp("wordSpacing", spacing) }

// ------------------------------------------------------------------------------------------------
// GetWordSpacing returns the current word spacing.
func (c Context2D) GetWordSpacing() string { return c.getStringProp("wordSpacing") }

// ------------------------------------------------------------------------------------------------
// FontKerning sets font kerning ("auto", "normal", or "none").
func (c Context2D) FontKerning(kerning string) { c.setProp("fontKerning", kerning) }

// ------------------------------------------------------------------------------------------------
// GetFontKerning returns the current font kerning setting.
func (c Context2D) GetFontKerning() string { return c.getStringProp("fontKerning") }

// ------------------------------------------------------------------------------------------------
// FontStretch sets font stretching (for example "normal", "condensed", or "expanded").
func (c Context2D) FontStretch(stretch string) { c.setProp("fontStretch", stretch) }

// ------------------------------------------------------------------------------------------------
// GetFontStretch returns the current font stretch setting.
func (c Context2D) GetFontStretch() string { return c.getStringProp("fontStretch") }

// ------------------------------------------------------------------------------------------------
// FontVariantCaps sets alternate capitalization for text rendering.
func (c Context2D) FontVariantCaps(caps string) { c.setProp("fontVariantCaps", caps) }

// ------------------------------------------------------------------------------------------------
// GetFontVariantCaps returns the current font variant caps setting.
func (c Context2D) GetFontVariantCaps() string { return c.getStringProp("fontVariantCaps") }

// ------------------------------------------------------------------------------------------------
// TextRendering sets text rendering hints ("auto", "optimizeSpeed", "optimizeLegibility", or "geometricPrecision").
func (c Context2D) TextRendering(rendering string) { c.setProp("textRendering", rendering) }

// ------------------------------------------------------------------------------------------------
// GetTextRendering returns the current text rendering setting.
func (c Context2D) GetTextRendering() string { return c.getStringProp("textRendering") }

// ================================================================================================
// Fill, stroke, gradients, and patterns
// ================================================================================================

// ------------------------------------------------------------------------------------------------
// FillStyle sets the fill style or color of the rendering context.
func (c Context2D) FillStyle(style string) { c.setProp("fillStyle", style) }

// ------------------------------------------------------------------------------------------------
// FillStyleValue sets the fill style from a CSS string, CanvasGradient, CanvasPattern, or Handle.
func (c Context2D) FillStyleValue(style any) { c.setProp("fillStyle", unwrapStyleValue(style)) }

// ------------------------------------------------------------------------------------------------
// GetFillStyle returns the raw JS fillStyle value.
func (c Context2D) GetFillStyle() js.Value { return js.Value(c.Ctx).Get("fillStyle") }

// ------------------------------------------------------------------------------------------------
// StrokeStyle sets the stroke style or color of the rendering context.
func (c Context2D) StrokeStyle(style string) { c.setProp("strokeStyle", style) }

// ------------------------------------------------------------------------------------------------
// StrokeStyleValue sets the stroke style from a CSS string, CanvasGradient, CanvasPattern, or Handle.
func (c Context2D) StrokeStyleValue(style any) { c.setProp("strokeStyle", unwrapStyleValue(style)) }

// ------------------------------------------------------------------------------------------------
// GetStrokeStyle returns the raw JS strokeStyle value.
func (c Context2D) GetStrokeStyle() js.Value { return js.Value(c.Ctx).Get("strokeStyle") }

// ------------------------------------------------------------------------------------------------
// CreateLinearGradient creates a linear gradient along the line given by the coordinates.
func (c Context2D) CreateLinearGradient(x0, y0, x1, y1 float64) CanvasGradient {
	return CanvasGradient{Handle: Handle(c.call("createLinearGradient", x0, y0, x1, y1))}
}

// ------------------------------------------------------------------------------------------------
// CreateRadialGradient creates a radial gradient using the size and coordinates of two circles.
func (c Context2D) CreateRadialGradient(x0, y0, r0, x1, y1, r1 float64) CanvasGradient {
	return CanvasGradient{Handle: Handle(c.call("createRadialGradient", x0, y0, r0, x1, y1, r1))}
}

// ------------------------------------------------------------------------------------------------
// CreateConicGradient creates a conic gradient around a point starting at startAngle radians.
func (c Context2D) CreateConicGradient(startAngle, x, y float64) CanvasGradient {
	return CanvasGradient{Handle: Handle(c.call("createConicGradient", startAngle, x, y))}
}

// ------------------------------------------------------------------------------------------------
// CreatePattern creates a pattern using the specified image source and repetition mode.
func (c Context2D) CreatePattern(image Handle, repetition string) CanvasPattern {
	return CanvasPattern{Handle: Handle(c.call("createPattern", js.Value(image), repetition))}
}

// ================================================================================================
// Shadows
// ================================================================================================

// ------------------------------------------------------------------------------------------------
// ShadowBlur sets the level of blurring applied to shadows.
func (c Context2D) ShadowBlur(blur float64) { c.setProp("shadowBlur", blur) }

// ------------------------------------------------------------------------------------------------
// GetShadowBlur returns the current shadow blur level.
func (c Context2D) GetShadowBlur() float64 { return c.getFloatProp("shadowBlur") }

// ------------------------------------------------------------------------------------------------
// ShadowColor sets the CSS color used for shadows.
func (c Context2D) ShadowColor(color string) { c.setProp("shadowColor", color) }

// ------------------------------------------------------------------------------------------------
// GetShadowColor returns the current shadow color.
func (c Context2D) GetShadowColor() string { return c.getStringProp("shadowColor") }

// ------------------------------------------------------------------------------------------------
// ShadowOffsetX sets the horizontal distance shadows are offset.
func (c Context2D) ShadowOffsetX(offset float64) { c.setProp("shadowOffsetX", offset) }

// ------------------------------------------------------------------------------------------------
// GetShadowOffsetX returns the current horizontal shadow offset.
func (c Context2D) GetShadowOffsetX() float64 { return c.getFloatProp("shadowOffsetX") }

// ------------------------------------------------------------------------------------------------
// ShadowOffsetY sets the vertical distance shadows are offset.
func (c Context2D) ShadowOffsetY(offset float64) { c.setProp("shadowOffsetY", offset) }

// ------------------------------------------------------------------------------------------------
// GetShadowOffsetY returns the current vertical shadow offset.
func (c Context2D) GetShadowOffsetY() float64 { return c.getFloatProp("shadowOffsetY") }

// ================================================================================================
// Paths
// ================================================================================================

// ------------------------------------------------------------------------------------------------
// BeginPath starts a new path by emptying the list of subpaths.
func (c Context2D) BeginPath() { c.call("beginPath") }

// ------------------------------------------------------------------------------------------------
// ClosePath adds a straight line from the current point back to the start of the current subpath.
func (c Context2D) ClosePath() { c.call("closePath") }

// ------------------------------------------------------------------------------------------------
// MoveTo begins a new subpath at the given coordinates.
func (c Context2D) MoveTo(x, y float64) { c.call("moveTo", x, y) }

// ------------------------------------------------------------------------------------------------
// LineTo connects the last point in the current subpath to the given coordinates with a straight line.
func (c Context2D) LineTo(x, y float64) { c.call("lineTo", x, y) }

// ------------------------------------------------------------------------------------------------
// BezierCurveTo adds a cubic Bezier curve to the current subpath.
func (c Context2D) BezierCurveTo(cp1x, cp1y, cp2x, cp2y, x, y float64) {
	c.call("bezierCurveTo", cp1x, cp1y, cp2x, cp2y, x, y)
}

// ------------------------------------------------------------------------------------------------
// QuadraticCurveTo adds a quadratic Bezier curve to the current subpath.
func (c Context2D) QuadraticCurveTo(cpx, cpy, x, y float64) {
	c.call("quadraticCurveTo", cpx, cpy, x, y)
}

// ------------------------------------------------------------------------------------------------
// Arc adds a circular arc to the current subpath.
func (c Context2D) Arc(x, y, radius, startAngle, endAngle float64, ccw bool) {
	c.call("arc", x, y, radius, startAngle, endAngle, ccw)
}

// ------------------------------------------------------------------------------------------------
// ArcTo adds a circular arc using two control points and a radius.
func (c Context2D) ArcTo(x1, y1, x2, y2, radius float64) {
	c.call("arcTo", x1, y1, x2, y2, radius)
}

// ------------------------------------------------------------------------------------------------
// Ellipse adds an elliptical arc to the current subpath.
func (c Context2D) Ellipse(x, y, radiusX, radiusY, rotation, startAngle, endAngle float64, ccw bool) {
	c.call("ellipse", x, y, radiusX, radiusY, rotation, startAngle, endAngle, ccw)
}

// ------------------------------------------------------------------------------------------------
// Rect adds a rectangle to the current path.
func (c Context2D) Rect(x, y, width, height float64) {
	c.call("rect", x, y, width, height)
}

// ------------------------------------------------------------------------------------------------
// RoundRect adds a rounded rectangle to the current path with zero, one, or multiple corner radii.
func (c Context2D) RoundRect(x, y, width, height float64, radii ...float64) {
	switch len(radii) {
	case 0:
		c.call("roundRect", x, y, width, height, 0)
	case 1:
		c.call("roundRect", x, y, width, height, radii[0])
	default:
		c.call("roundRect", x, y, width, height, floatSliceToJSArray(radii))
	}
}

// ================================================================================================
// Drawing paths and hit testing
// ================================================================================================

// ------------------------------------------------------------------------------------------------
// Fill fills the current path with the current fill style.
func (c Context2D) Fill() { c.call("fill") }

// ------------------------------------------------------------------------------------------------
// FillRule fills the current path using the specified fill rule ("nonzero" or "evenodd").
func (c Context2D) FillRule(fillRule string) { c.call("fill", fillRule) }

// ------------------------------------------------------------------------------------------------
// FillPath fills the given Path2D handle, optionally using a fill rule ("nonzero" or "evenodd").
func (c Context2D) FillPath(path Handle, fillRule ...string) {
	if len(fillRule) > 0 {
		c.call("fill", js.Value(path), fillRule[0])
		return
	}
	c.call("fill", js.Value(path))
}

// ------------------------------------------------------------------------------------------------
// Stroke strokes the current path with the current stroke style.
func (c Context2D) Stroke() { c.call("stroke") }

// ------------------------------------------------------------------------------------------------
// StrokePath strokes the given Path2D handle with the current stroke style.
func (c Context2D) StrokePath(path Handle) { c.call("stroke", js.Value(path)) }

// ------------------------------------------------------------------------------------------------
// DrawFocusIfNeeded draws a focus ring around the current path if the given element has focus.
func (c Context2D) DrawFocusIfNeeded(element Handle) {
	c.call("drawFocusIfNeeded", js.Value(element))
}

// ------------------------------------------------------------------------------------------------
// DrawFocusIfNeededPath draws a focus ring around the given Path2D if the element has focus.
func (c Context2D) DrawFocusIfNeededPath(path, element Handle) {
	c.call("drawFocusIfNeeded", js.Value(path), js.Value(element))
}

// ------------------------------------------------------------------------------------------------
// Clip turns the current path into the current clipping region.
func (c Context2D) Clip() { c.call("clip") }

// ------------------------------------------------------------------------------------------------
// ClipRule turns the current path into the clipping region using the specified fill rule.
func (c Context2D) ClipRule(fillRule string) { c.call("clip", fillRule) }

// ------------------------------------------------------------------------------------------------
// ClipPath turns the given Path2D handle into the clipping region, optionally using a fill rule.
func (c Context2D) ClipPath(path Handle, fillRule ...string) {
	if len(fillRule) > 0 {
		c.call("clip", js.Value(path), fillRule[0])
		return
	}
	c.call("clip", js.Value(path))
}

// ------------------------------------------------------------------------------------------------
// IsPointInPath reports whether the specified point is contained in the current path.
func (c Context2D) IsPointInPath(x, y float64, fillRule ...string) bool {
	if len(fillRule) > 0 {
		return c.call("isPointInPath", x, y, fillRule[0]).Bool()
	}
	return c.call("isPointInPath", x, y).Bool()
}

// ------------------------------------------------------------------------------------------------
// IsPointInPathObj reports whether the specified point is contained in the given Path2D handle.
func (c Context2D) IsPointInPathObj(path Handle, x, y float64, fillRule ...string) bool {
	if len(fillRule) > 0 {
		return c.call("isPointInPath", js.Value(path), x, y, fillRule[0]).Bool()
	}
	return c.call("isPointInPath", js.Value(path), x, y).Bool()
}

// ------------------------------------------------------------------------------------------------
// IsPointInStroke reports whether the specified point is inside the stroke of the current path.
func (c Context2D) IsPointInStroke(x, y float64) bool {
	return c.call("isPointInStroke", x, y).Bool()
}

// ------------------------------------------------------------------------------------------------
// IsPointInStrokeObj reports whether the specified point is inside the stroke of the given Path2D handle.
func (c Context2D) IsPointInStrokeObj(path Handle, x, y float64) bool {
	return c.call("isPointInStroke", js.Value(path), x, y).Bool()
}

// ================================================================================================
// Transformations
// ================================================================================================

// ------------------------------------------------------------------------------------------------
// Rotate adds a rotation of angle radians to the transformation matrix.
func (c Context2D) Rotate(angle float64) { c.call("rotate", angle) }

// ------------------------------------------------------------------------------------------------
// Scale adds a scaling transformation to the canvas units horizontally and vertically.
func (c Context2D) Scale(x, y float64) { c.call("scale", x, y) }

// ------------------------------------------------------------------------------------------------
// Translate adds a translation transformation by moving the canvas origin by (x, y).
func (c Context2D) Translate(x, y float64) { c.call("translate", x, y) }

// ------------------------------------------------------------------------------------------------
// Transform multiplies the current transformation matrix with the matrix described by a..f.
func (c Context2D) Transform(a, b, cVal, d, e, f float64) {
	c.call("transform", a, b, cVal, d, e, f)
}

// ------------------------------------------------------------------------------------------------
// SetTransform resets the current transform to the identity matrix and applies the matrix a..f.
func (c Context2D) SetTransform(a, b, cVal, d, e, f float64) {
	c.call("setTransform", a, b, cVal, d, e, f)
}

// ------------------------------------------------------------------------------------------------
// SetTransformMatrix sets the current transformation matrix from a DOMMatrix handle.
func (c Context2D) SetTransformMatrix(matrix Handle) {
	c.call("setTransform", js.Value(matrix))
}

// ------------------------------------------------------------------------------------------------
// GetTransform returns a copy of the current transformation matrix as a DOMMatrix handle.
func (c Context2D) GetTransform() Handle {
	return Handle(c.call("getTransform"))
}

// ------------------------------------------------------------------------------------------------
// ResetTransform resets the current transform to the identity matrix.
func (c Context2D) ResetTransform() { c.call("resetTransform") }

// ================================================================================================
// Compositing, image smoothing, and filters
// ================================================================================================

// ------------------------------------------------------------------------------------------------
// GlobalAlpha sets the alpha (transparency) value applied to shapes and images (0.0 to 1.0).
func (c Context2D) GlobalAlpha(alpha float64) { c.setProp("globalAlpha", alpha) }

// ------------------------------------------------------------------------------------------------
// GetGlobalAlpha returns the current global alpha value.
func (c Context2D) GetGlobalAlpha() float64 { return c.getFloatProp("globalAlpha") }

// ------------------------------------------------------------------------------------------------
// GlobalCompositeOperation sets the compositing operation used when drawing new shapes.
func (c Context2D) GlobalCompositeOperation(op string) { c.setProp("globalCompositeOperation", op) }

// ------------------------------------------------------------------------------------------------
// GetGlobalCompositeOperation returns the current global composite operation name.
func (c Context2D) GetGlobalCompositeOperation() string {
	return c.getStringProp("globalCompositeOperation")
}

// ------------------------------------------------------------------------------------------------
// ImageSmoothingEnabled sets whether scaled images are smoothed.
func (c Context2D) ImageSmoothingEnabled(enabled bool) { c.setProp("imageSmoothingEnabled", enabled) }

// ------------------------------------------------------------------------------------------------
// GetImageSmoothingEnabled reports whether scaled image smoothing is enabled.
func (c Context2D) GetImageSmoothingEnabled() bool {
	return c.getBoolProp("imageSmoothingEnabled")
}

// ------------------------------------------------------------------------------------------------
// ImageSmoothingQuality sets image smoothing quality ("low", "medium", or "high").
func (c Context2D) ImageSmoothingQuality(quality string) { c.setProp("imageSmoothingQuality", quality) }

// ------------------------------------------------------------------------------------------------
// GetImageSmoothingQuality returns the current image smoothing quality setting.
func (c Context2D) GetImageSmoothingQuality() string {
	return c.getStringProp("imageSmoothingQuality")
}

// ------------------------------------------------------------------------------------------------
// Filter sets the CSS filter effects string applied to drawing operations.
func (c Context2D) Filter(filter string) { c.setProp("filter", filter) }

// ------------------------------------------------------------------------------------------------
// GetFilter returns the current CSS filter effects string.
func (c Context2D) GetFilter() string { return c.getStringProp("filter") }

// ================================================================================================
// Drawing images
// ================================================================================================

// ------------------------------------------------------------------------------------------------
// DrawImage draws an image source onto the canvas at (dx, dy).
func (c Context2D) DrawImage(image Handle, dx, dy float64) {
	c.call("drawImage", js.Value(image), dx, dy)
}

// ------------------------------------------------------------------------------------------------
// DrawImageScaled draws an image source scaled to dWidth by dHeight at (dx, dy).
func (c Context2D) DrawImageScaled(image Handle, dx, dy, dWidth, dHeight float64) {
	c.call("drawImage", js.Value(image), dx, dy, dWidth, dHeight)
}

// ------------------------------------------------------------------------------------------------
// DrawImageSubrect draws a cropped subrectangle of an image source onto the destination rectangle.
func (c Context2D) DrawImageSubrect(image Handle, sx, sy, sWidth, sHeight, dx, dy, dWidth, dHeight float64) {
	c.call("drawImage", js.Value(image), sx, sy, sWidth, sHeight, dx, dy, dWidth, dHeight)
}

// ================================================================================================
// Pixel manipulation
// ================================================================================================

// ------------------------------------------------------------------------------------------------
// CreateImageData creates a blank ImageData object of the given pixel dimensions.
func (c Context2D) CreateImageData(width, height int) Handle {
	return Handle(c.call("createImageData", width, height))
}

// ------------------------------------------------------------------------------------------------
// CreateImageDataFrom creates a blank ImageData object with the same dimensions as imageData.
func (c Context2D) CreateImageDataFrom(imageData Handle) Handle {
	return Handle(c.call("createImageData", js.Value(imageData)))
}

// ------------------------------------------------------------------------------------------------
// GetImageData returns an ImageData object copying pixel data for the specified rectangle.
func (c Context2D) GetImageData(sx, sy, sw, sh int) Handle {
	return Handle(c.call("getImageData", sx, sy, sw, sh))
}

// ------------------------------------------------------------------------------------------------
// PutImageData paints pixel data from imageData onto the canvas at (dx, dy).
func (c Context2D) PutImageData(imageData Handle, dx, dy int) {
	c.call("putImageData", js.Value(imageData), dx, dy)
}

// ------------------------------------------------------------------------------------------------
// PutImageDataDirty paints only the specified dirty subrectangle of imageData onto the canvas.
func (c Context2D) PutImageDataDirty(imageData Handle, dx, dy, dirtyX, dirtyY, dirtyWidth, dirtyHeight int) {
	c.call("putImageData", js.Value(imageData), dx, dy, dirtyX, dirtyY, dirtyWidth, dirtyHeight)
}
