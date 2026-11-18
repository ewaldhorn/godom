// Package dom provides a basic DOM manipulation wrapper using syscall/js.
package dom

import (
	"strings"
	"sync/atomic"
	"syscall/js"
)

// ------------------------------------------------------------------------------------------------
// Handle represents a reference to a live JS object in the browser.
type Handle js.Value

// ------------------------------------------------------------------------------------------------
// Invalid represents a null/invalid handle.
var Invalid = Handle(js.Null())

// ------------------------------------------------------------------------------------------------
// Global element references.
var (
	Window   Handle
	Document Handle
	Body     Handle
	Head     Handle
)

// ------------------------------------------------------------------------------------------------
// Init captures global JS references from the browser.
func Init() {
	Window = Handle(js.Global())
	Document = Handle(js.Global().Get("document"))
	Body = Handle(js.Value(Document).Get("body"))
	Head = Handle(js.Value(Document).Get("head"))
	Invalid = Handle(js.Null())
}

// ------------------------------------------------------------------------------------------------
// IsValid checks if the handle is valid (truthy in the JS sense).
func (h Handle) IsValid() bool {
	return js.Value(h).Truthy()
}

// ------------------------------------------------------------------------------------------------
// Get gets a JS property on the handle as a string.
// It returns an empty string if the property is undefined or null.
func (h Handle) Get(key string) string {
	v := js.Value(h).Get(key)
	if v.IsUndefined() || v.IsNull() {
		return ""
	}
	return v.String()
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
	js.Value(h).Set("className", strings.Join(classes, " "))
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
	el := Handle(js.Global().Get("document").Call("getElementById", id))
	if el.IsValid() {
		return el
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
// It returns an empty string if the element or the property is missing.
func GetString(elemID, key string) string {
	elem := GetElementByID(elemID)
	if !elem.IsValid() {
		return ""
	}
	return elem.Get(key)
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
// Event wraps a browser DOM Event passed to an event listener callback.
type Event struct {
	raw js.Value
}

// ------------------------------------------------------------------------------------------------
// Raw returns the underlying js.Value of the DOM event.
func (e Event) Raw() js.Value {
	return e.raw
}

// ------------------------------------------------------------------------------------------------
// PreventDefault cancels the event if it is cancelable.
func (e Event) PreventDefault() {
	if e.raw.Truthy() {
		e.raw.Call("preventDefault")
	}
}

// ------------------------------------------------------------------------------------------------
// StopPropagation prevents further propagation of the current event in the capturing and bubbling phases.
func (e Event) StopPropagation() {
	if e.raw.Truthy() {
		e.raw.Call("stopPropagation")
	}
}

// ------------------------------------------------------------------------------------------------
// Target returns the DOM element handle on which the event originally occurred.
func (e Event) Target() Handle {
	if !e.raw.Truthy() {
		return Invalid
	}
	t := Handle(e.raw.Get("target"))
	if t.IsValid() {
		return t
	}
	return Invalid
}

// ------------------------------------------------------------------------------------------------
// ClientX returns the horizontal coordinate within the application's viewport at which the event occurred.
func (e Event) ClientX() int {
	return e.intProp("clientX")
}

// ------------------------------------------------------------------------------------------------
// ClientY returns the vertical coordinate within the application's viewport at which the event occurred.
func (e Event) ClientY() int {
	return e.intProp("clientY")
}

// ------------------------------------------------------------------------------------------------
// OffsetX returns the horizontal coordinate of the pointer relative to the target node.
func (e Event) OffsetX() int {
	return e.intProp("offsetX")
}

// ------------------------------------------------------------------------------------------------
// OffsetY returns the vertical coordinate of the pointer relative to the target node.
func (e Event) OffsetY() int {
	return e.intProp("offsetY")
}

// ------------------------------------------------------------------------------------------------
// Button returns which mouse button was pressed to trigger the event (0=main, 1=auxiliary, 2=secondary).
func (e Event) Button() int {
	return e.intProp("button")
}

// ------------------------------------------------------------------------------------------------
// Key returns the value of the key pressed by the user (for example, "ArrowUp" or "a").
func (e Event) Key() string {
	return Handle(e.raw).Get("key")
}

// ------------------------------------------------------------------------------------------------
// Code returns the physical key code on the keyboard (for example, "Space" or "KeyA").
func (e Event) Code() string {
	return Handle(e.raw).Get("code")
}

func (e Event) intProp(key string) int {
	if !e.raw.Truthy() {
		return 0
	}
	v := e.raw.Get(key)
	if v.IsUndefined() || v.IsNull() {
		return 0
	}
	return v.Int()
}

// ------------------------------------------------------------------------------------------------
// On registers a plain Go callback for the given event name on the handle.
// It returns a cleanup function that removes the event listener and releases the underlying JS closure.
func (h Handle) On(event string, fn func()) (off func()) {
	if !h.IsValid() {
		return func() {}
	}
	var released atomic.Bool
	jsFn := js.FuncOf(func(this js.Value, args []js.Value) any {
		fn()
		return nil
	})
	js.Value(h).Call("addEventListener", event, jsFn)
	return func() {
		if released.CompareAndSwap(false, true) {
			js.Value(h).Call("removeEventListener", event, jsFn)
			jsFn.Release()
		}
	}
}

// ------------------------------------------------------------------------------------------------
// OnEvent registers a Go callback that receives the DOM Event wrapper for the given event name.
// It returns a cleanup function that removes the event listener and releases the underlying JS closure.
func (h Handle) OnEvent(event string, fn func(Event)) (off func()) {
	if !h.IsValid() {
		return func() {}
	}
	var released atomic.Bool
	jsFn := js.FuncOf(func(this js.Value, args []js.Value) any {
		var ev Event
		if len(args) > 0 {
			ev = Event{raw: args[0]}
		}
		fn(ev)
		return nil
	})
	js.Value(h).Call("addEventListener", event, jsFn)
	return func() {
		if released.CompareAndSwap(false, true) {
			js.Value(h).Call("removeEventListener", event, jsFn)
			jsFn.Release()
		}
	}
}

// ------------------------------------------------------------------------------------------------
// OnClick registers a plain Go callback for "click" events on the handle.
// It returns a cleanup function that removes the event listener and releases the underlying JS closure.
func (h Handle) OnClick(fn func()) (off func()) {
	return h.On("click", fn)
}

// ------------------------------------------------------------------------------------------------
// OnByID registers a plain Go callback for the given event name on the element matching the ID.
func OnByID(id, event string, fn func()) (off func()) {
	return GetElementByID(id).On(event, fn)
}

// ------------------------------------------------------------------------------------------------
// OnEventByID registers a Go callback receiving an Event on the element matching the ID.
func OnEventByID(id, event string, fn func(Event)) (off func()) {
	return GetElementByID(id).OnEvent(event, fn)
}

// ------------------------------------------------------------------------------------------------
// OnClickByID registers a plain Go callback for "click" events on the element matching the ID.
func OnClickByID(id string, fn func()) (off func()) {
	return GetElementByID(id).OnClick(fn)
}

// ------------------------------------------------------------------------------------------------
// AddEventListener registers a raw js.Value event listener on the element.
func AddEventListener(elem Handle, event string, cb js.Value) {
	js.Value(elem).Call("addEventListener", event, cb)
}

// ------------------------------------------------------------------------------------------------
// AddEventListenerByID registers a raw js.Value event listener on the element matching the ID.
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
// Log logs messages to the browser developer console.
func Log(messages ...string) {
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
	canvas.Set("width", width)
	canvas.Set("height", height)
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
// It returns a stop function; calling it halts the loop and releases the JS closure.
//
// The js.Func backing the loop is released on the first tick that observes the
// stop. It is not released directly in stop(): requestAnimationFrame callbacks
// cannot be cancelled, and js.Func.Release forbids invoking the function
// afterwards, so a pending frame would call into a released function.
func StartAnimationLoop(cb js.Value) (stop func()) {
	return StartLoop(func() {
		cb.Invoke()
	})
}

// ------------------------------------------------------------------------------------------------
// StartLoop drives a 60 FPS animation cycle via requestAnimationFrame using a plain Go callback.
// It returns a stop function; calling it halts the loop and releases the JS closure.
func StartLoop(fn func()) (stop func()) {
	var tick js.Func
	var stopped atomic.Bool
	tick = js.FuncOf(func(this js.Value, args []js.Value) any {
		if stopped.Load() {
			tick.Release()
			return nil
		}
		fn()
		js.Global().Call("requestAnimationFrame", tick)
		return nil
	})
	js.Global().Call("requestAnimationFrame", tick)
	return func() { stopped.Store(true) }
}
