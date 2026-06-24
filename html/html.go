// Package html provides fluent, chainable HTML tag builder constructors and utility methods.
package html

import (
	"github.com/ewaldhorn/godom/dom"
	"syscall/js"
)

// ------------------------------------------------------------------------------------------------
// Element is a chainable tag builder wrapper around a DOM element handle.
type Element struct {
	Handle dom.Handle
}

// ------------------------------------------------------------------------------------------------
// Init creates a new builder instance for the specified HTML tag.
func Init(tag string) *Element {
	return &Element{Handle: dom.CreateElement(tag)}
}

// ------------------------------------------------------------------------------------------------
// ID sets the element's DOM ID attribute.
func (e *Element) ID(val string) *Element {
	e.Handle.Set("id", val)
	return e
}

// ------------------------------------------------------------------------------------------------
// Class adds a class name to the element.
func (e *Element) Class(val string) *Element {
	e.Handle.AddClassTo(val)
	return e
}

// ------------------------------------------------------------------------------------------------
// Text sets the inner text of the element.
func (e *Element) Text(val string) *Element {
	e.Handle.SetInnerText(val)
	return e
}

// ------------------------------------------------------------------------------------------------
// HTML sets the inner HTML content of the element.
func (e *Element) HTML(val string) *Element {
	e.Handle.SetInnerHTML(val)
	return e
}

// ------------------------------------------------------------------------------------------------
// Attr sets an arbitrary string attribute.
func (e *Element) Attr(key, val string) *Element {
	e.Handle.Set(key, val)
	return e
}

// ------------------------------------------------------------------------------------------------
// Child appends another builder's element as a child.
func (e *Element) Child(childHandle dom.Handle) *Element {
	dom.AddElementTo(e.Handle, childHandle)
	return e
}

// ------------------------------------------------------------------------------------------------
// AppendTo appends this element under a parent element.
func (e *Element) AppendTo(parent dom.Handle) *Element {
	dom.AddElementTo(parent, e.Handle)
	return e
}

// ------------------------------------------------------------------------------------------------
// On registers an event listener callback on the element.
func (e *Element) On(event string, cb js.Value) *Element {
	dom.AddEventListener(e.Handle, event, cb)
	return e
}

// ------------------------------------------------------------------------------------------------
// Build returns the compiled element handle.
func (e *Element) Build() dom.Handle {
	return e.Handle
}

// ================================================================================================
// Tag constructors
// ================================================================================================

// ------------------------------------------------------------------------------------------------
// Structural tags
func Div() *Element    { return Init("div") }
func Span() *Element   { return Init("span") }
func P() *Element      { return Init("p") }
func Button() *Element { return Init("button") }
func A() *Element      { return Init("a") }

// ------------------------------------------------------------------------------------------------
// Headings
func H1() *Element { return Init("h1") }
func H2() *Element { return Init("h2") }
func H3() *Element { return Init("h3") }
func H4() *Element { return Init("h4") }
func H5() *Element { return Init("h5") }
func H6() *Element { return Init("h6") }

// ------------------------------------------------------------------------------------------------
// Semantic layout
func Article() *Element { return Init("article") }
func Aside() *Element   { return Init("aside") }
func Section() *Element { return Init("section") }
func Nav() *Element     { return Init("nav") }
func Header() *Element  { return Init("header") }
func Footer() *Element  { return Init("footer") }
func Main() *Element    { return Init("main") }

// ------------------------------------------------------------------------------------------------
// Lists
func Ul() *Element { return Init("ul") }
func Ol() *Element { return Init("ol") }
func Li() *Element { return Init("li") }
func Dl() *Element { return Init("dl") }
func Dt() *Element { return Init("dt") }
func Dd() *Element { return Init("dd") }

// ------------------------------------------------------------------------------------------------
// Inline formatting
func Strong() *Element { return Init("strong") }
func Em() *Element     { return Init("em") }
func Code() *Element   { return Init("code") }
func Pre() *Element    { return Init("pre") }
func Small() *Element  { return Init("small") }
func Mark() *Element   { return Init("mark") }
func B() *Element      { return Init("b") }
func I() *Element      { return Init("i") }

// ------------------------------------------------------------------------------------------------
// Form elements
func Form() *Element     { return Init("form") }
func Input() *Element    { return Init("input") }
func Label() *Element    { return Init("label") }
func Select() *Element   { return Init("select") }
func Option() *Element   { return Init("option") }
func Textarea() *Element { return Init("textarea") }
func Fieldset() *Element { return Init("fieldset") }
func Legend() *Element   { return Init("legend") }

// ------------------------------------------------------------------------------------------------
// Media / Void elements
func Img() *Element { return Init("img") }
func Br() *Element  { return Init("br") }
func Hr() *Element  { return Init("hr") }

// ------------------------------------------------------------------------------------------------
// Tables
func Table() *Element { return Init("table") }
func Thead() *Element { return Init("thead") }
func Tbody() *Element { return Init("tbody") }
func Tr() *Element    { return Init("tr") }
func Th() *Element    { return Init("th") }
func Td() *Element    { return Init("td") }

// ------------------------------------------------------------------------------------------------
// Miscellaneous
func Figure() *Element     { return Init("figure") }
func Figcaption() *Element { return Init("figcaption") }
func Details() *Element    { return Init("details") }
func Summary() *Element    { return Init("summary") }
func Blockquote() *Element { return Init("blockquote") }
func Cite() *Element       { return Init("cite") }
func Time() *Element       { return Init("time") }