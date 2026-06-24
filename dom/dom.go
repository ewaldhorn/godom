// Package dom provides a basic DOM manipulation wrapper using syscall/js.
package dom

import "syscall/js"

// ------------------------------------------------------------------------------------------------
// Handle represents a reference to a live JS object in the browser.
type Handle js.Value

// ------------------------------------------------------------------------------------------------
// Invalid represents a null/invalid handle.
var Invalid Handle

// ------------------------------------------------------------------------------------------------
// Global element references.
var (
	Document Handle
	Body     Handle
	Head     Handle
)

// ------------------------------------------------------------------------------------------------
// Init captures global JS references from the browser.
func Init() {
	Document = Handle(js.Global().Get("document"))
	Body = Handle(js.Value(Document).Get("body"))
	Head = Handle(js.Value(Document).Get("head"))
	Invalid = Handle(js.Null())
}

// ------------------------------------------------------------------------------------------------
// IsValid checks if the handle is valid (truthy and not null/undefined).
func (h Handle) IsValid() bool {
	v := js.Value(h)
	return v.Truthy() && !v.IsNull() && !v.IsUndefined()
}

// ------------------------------------------------------------------------------------------------
// Get gets a JS property on the handle as a string.
func (h Handle) Get(key string) string {
	return js.Value(h).Get(key).String()
}

// ------------------------------------------------------------------------------------------------
// Set sets a JS property on the handle to an arbitrary value.
func (h Handle) Set(key string, value any) {
	js.Value(h).Set(key, value)
}

// ------------------------------------------------------------------------------------------------
// SetInnerText sets the inner text of an element.
func (h Handle) SetInnerText(text string) {
	js.Value(h).Set("innerText", text)
}

// ------------------------------------------------------------------------------------------------
// SetInnerHTML sets the inner HTML of an element.
func (h Handle) SetInnerHTML(html string) {
	js.Value(h).Set("innerHTML", html)
}

// ------------------------------------------------------------------------------------------------
// AddClassTo adds a CSS class to the element.
func (h Handle) AddClassTo(class string) {
	js.Value(h).Get("classList").Call("add", class)
}

// ------------------------------------------------------------------------------------------------
// RemoveClassFrom removes a CSS class from the element.
func (h Handle) RemoveClassFrom(class string) {
	js.Value(h).Get("classList").Call("remove", class)
}

// ------------------------------------------------------------------------------------------------
// ReplaceClasses replaces all classes on the element with the given list.
func (h Handle) ReplaceClasses(classes []string) {
	js.Value(h).Set("className", "")
	for _, cls := range classes {
		h.AddClassTo(cls)
	}
}

// ================================================================================================
// Element CRUD
// ================================================================================================

// ------------------------------------------------------------------------------------------------
// CreateElement creates a new HTML element of the given tag name.
func CreateElement(tag string) Handle {
	return Handle(js.Global().Get("document").Call("createElement", tag))
}

// ------------------------------------------------------------------------------------------------
// CreateDiv creates a new div element.
func CreateDiv() Handle {
	return CreateElement("div")
}

// ------------------------------------------------------------------------------------------------
// CreateParagraph creates a new p element.
func CreateParagraph() Handle {
	return CreateElement("p")
}

// ------------------------------------------------------------------------------------------------
// CreateParagraphWithText creates a new p element with pre-set inner text.
func CreateParagraphWithText(text string) Handle {
	p := CreateElement("p")
	p.SetInnerText(text)
	return p
}

// ------------------------------------------------------------------------------------------------
// CreateButton creates a button element of type "button" with pre-set inner text.
func CreateButton(text string) Handle {
	b := CreateElement("button")
	b.Set("type", "button")
	b.SetInnerText(text)
	return b
}

// ------------------------------------------------------------------------------------------------
// CreateImg creates an img element with pre-set src.
func CreateImg(src string) Handle {
	img := CreateElement("img")
	img.Set("src", src)
	return img
}

// ------------------------------------------------------------------------------------------------
// AddElementTo appends a child element to a target element.
func AddElementTo(target, elem Handle) {
	js.Value(target).Call("appendChild", js.Value(elem))
}

// ------------------------------------------------------------------------------------------------
// AddToBody appends a child element to the document body.
func AddToBody(elem Handle) {
	AddElementTo(Body, elem)
}

// ------------------------------------------------------------------------------------------------
// RemoveAllChildElementsFrom removes all children from the target element.
func RemoveAllChildElementsFrom(target Handle) {
	js.Value(target).Call("replaceChildren")
}

// ------------------------------------------------------------------------------------------------
// GetElementByID returns the element handle matching the specified ID.
func GetElementByID(id string) Handle {
	el := js.Global().Get("document").Call("getElementById", id)
	if el.Truthy() && !el.IsNull() && !el.IsUndefined() {
		return Handle(el)
	}
	return Invalid
}

// ------------------------------------------------------------------------------------------------
// WrapElementWithNewDiv wraps an existing element in a new div with the given classes.
func WrapElementWithNewDiv(element Handle, classes []string) Handle {
	div := CreateDiv()
	for _, cls := range classes {
		div.AddClassTo(cls)
	}
	AddElementTo(div, element)
	return div
}

// ================================================================================================
// Visibility
// ================================================================================================

// ------------------------------------------------------------------------------------------------
// Hide sets display to "none" on the element matching the ID.
func Hide(id string) {
	elem := GetElementByID(id)
	if elem.IsValid() {
		js.Value(elem).Get("style").Set("display", "none")
	}
}

// ------------------------------------------------------------------------------------------------
// Show sets display to "block" on the element matching the ID.
func Show(id string) {
	elem := GetElementByID(id)
	if elem.IsValid() {
		js.Value(elem).Get("style").Set("display", "block")
	}
}

// ------------------------------------------------------------------------------------------------
// SetFocus calls focus() on the element matching the ID.
func SetFocus(id string) {
	elem := GetElementByID(id)
	if elem.IsValid() {
		js.Value(elem).Call("focus")
	}
}

// ================================================================================================
// Property access (string-based, by ID)
// ================================================================================================

// ------------------------------------------------------------------------------------------------
// GetString retrieves a string property from an element by ID.
func GetString(elemID, key string) string {
	elem := GetElementByID(elemID)
	if !elem.IsValid() {
		return ""
	}
	return js.Value(elem).Get(key).String()
}

// ------------------------------------------------------------------------------------------------
// SetValue sets a string property on an element by ID.
func SetValue(elemID, key, value string) {
	elem := GetElementByID(elemID)
	if elem.IsValid() {
		js.Value(elem).Set(key, value)
	}
}

// ================================================================================================
// Style (by element ID)
// ================================================================================================

// ------------------------------------------------------------------------------------------------
// AddClass adds a CSS class to the element looked up by ID.
func AddClass(elemID, class string) {
	elem := GetElementByID(elemID)
	if elem.IsValid() {
		elem.AddClassTo(class)
	}
}

// ------------------------------------------------------------------------------------------------
// RemoveClass removes a CSS class from the element looked up by ID.
func RemoveClass(elemID, class string) {
	elem := GetElementByID(elemID)
	if elem.IsValid() {
		elem.RemoveClassFrom(class)
	}
}

// ------------------------------------------------------------------------------------------------
// AddNewStyleElement injects a style block with raw CSS rules into document head.
func AddNewStyleElement(css string) {
	style := CreateElement("style")
	style.Set("type", "text/css")
	style.SetInnerHTML(css)
	AddElementTo(Head, style)
}

// ================================================================================================
// Events
// ================================================================================================

// ------------------------------------------------------------------------------------------------
// AddEventListener registers an event listener on the element.
func AddEventListener(elem Handle, event string, cb js.Value) {
	js.Value(elem).Call("addEventListener", event, cb)
}

// ------------------------------------------------------------------------------------------------
// AddEventListenerByID registers an event listener on the element matching the ID.
func AddEventListenerByID(id, event string, cb js.Value) {
	elem := GetElementByID(id)
	if elem.IsValid() {
		AddEventListener(elem, event, cb)
	}
}

// ================================================================================================
// Logging
// ================================================================================================

// ------------------------------------------------------------------------------------------------
// Log logs a slice of strings to the browser developer console.
func Log(messages []string) {
	args := make([]any, len(messages))
	for i, msg := range messages {
		args[i] = msg
	}
	js.Global().Get("console").Call("log", args...)
}

// ------------------------------------------------------------------------------------------------
// Alert triggers a browser modal dialog with the specified message.
func Alert(msg string) {
	js.Global().Call("alert", msg)
}

// ================================================================================================
// Canvas & Context 2D utilities
// ================================================================================================

// ------------------------------------------------------------------------------------------------
// CanvasCreate creates a canvas element and appends it to parent.
func CanvasCreate(parent Handle, width, height int) Handle {
	canvas := CreateElement("canvas")
	canvas.Set("width", js.ValueOf(width))
	canvas.Set("height", js.ValueOf(height))
	AddElementTo(parent, canvas)
	return canvas
}

// ------------------------------------------------------------------------------------------------
// CanvasGetContext retrieves the 2D rendering context of the canvas.
func CanvasGetContext(canvas Handle) Handle {
	return Handle(js.Value(canvas).Call("getContext", "2d"))
}

// ------------------------------------------------------------------------------------------------
// StartAnimationLoop drives a 60 FPS animation cycle via requestAnimationFrame.
func StartAnimationLoop(cb js.Value) {
	var tick js.Func
	tick = js.FuncOf(func(this js.Value, args []js.Value) any {
		cb.Invoke()
		js.Global().Call("requestAnimationFrame", tick)
		return nil
	})
	js.Global().Call("requestAnimationFrame", tick)
}

// ------------------------------------------------------------------------------------------------
// Context2D wraps the canvas 2D rendering context.
type Context2D struct {
	Ctx Handle
}

// ------------------------------------------------------------------------------------------------
// BeginPath begins a new path in the rendering context.
func (c Context2D) BeginPath() {
	js.Value(c.Ctx).Call("beginPath")
}

// ------------------------------------------------------------------------------------------------
// Fill fills the current path with the current fill style.
func (c Context2D) Fill() {
	js.Value(c.Ctx).Call("fill")
}

// ------------------------------------------------------------------------------------------------
// Arc draws a circular arc on the rendering context.
func (c Context2D) Arc(x, y, radius, startAngle, endAngle float64, ccw bool) {
	js.Value(c.Ctx).Call("arc", x, y, radius, startAngle, endAngle, ccw)
}

// ------------------------------------------------------------------------------------------------
// FillStyle sets the fill style/color of the rendering context.
func (c Context2D) FillStyle(style string) {
	js.Value(c.Ctx).Set("fillStyle", style)
}
