// Package html provides fluent, chainable HTML tag builder constructors and utility methods.
package html

import (
	"godom/dom"
	"syscall/js"
)

// ------------------------------------------------------------------------------------------------
// Elm is a chainable tag builder wrapper around a DOM element handle.
type Elm struct {
	Handle dom.Handle
}

// ------------------------------------------------------------------------------------------------
// Init creates a new builder instance for the specified HTML tag.
func Init(tag string) *Elm {
	return &Elm{Handle: dom.CreateElement(tag)}
}

// ------------------------------------------------------------------------------------------------
// ID sets the element's DOM ID attribute.
func (e *Elm) ID(val string) *Elm {
	e.Handle.Set("id", val)
	return e
}

// ------------------------------------------------------------------------------------------------
// Class adds a class name to the element.
func (e *Elm) Class(val string) *Elm {
	e.Handle.AddClassTo(val)
	return e
}

// ------------------------------------------------------------------------------------------------
// Text sets the inner text of the element.
func (e *Elm) Text(val string) *Elm {
	e.Handle.SetInnerText(val)
	return e
}

// ------------------------------------------------------------------------------------------------
// HTML sets the inner HTML content of the element.
func (e *Elm) HTML(val string) *Elm {
	e.Handle.SetInnerHTML(val)
	return e
}

// ------------------------------------------------------------------------------------------------
// Attr sets an arbitrary string attribute.
func (e *Elm) Attr(key, val string) *Elm {
	e.Handle.Set(key, val)
	return e
}

// ------------------------------------------------------------------------------------------------
// Child appends another builder's element as a child.
func (e *Elm) Child(childHandle dom.Handle) *Elm {
	dom.AddElementTo(e.Handle, childHandle)
	return e
}

// ------------------------------------------------------------------------------------------------
// AppendTo appends this element under a parent element.
func (e *Elm) AppendTo(parent dom.Handle) *Elm {
	dom.AddElementTo(parent, e.Handle)
	return e
}

// ------------------------------------------------------------------------------------------------
// On registers an event listener callback on the element.
func (e *Elm) On(event string, cb js.Value) *Elm {
	dom.AddEventListener(e.Handle, event, cb)
	return e
}

// ------------------------------------------------------------------------------------------------
// Build returns the compiled element handle.
func (e *Elm) Build() dom.Handle {
	return e.Handle
}

// ================================================================================================
// Tag constructors
// ================================================================================================

// ------------------------------------------------------------------------------------------------
// Structural tags
func Div() *Elm    { return Init("div") }
func Span() *Elm   { return Init("span") }
func P() *Elm      { return Init("p") }
func Button() *Elm { return Init("button") }
func A() *Elm      { return Init("a") }

// ------------------------------------------------------------------------------------------------
// Headings
func H1() *Elm { return Init("h1") }
func H2() *Elm { return Init("h2") }
func H3() *Elm { return Init("h3") }
func H4() *Elm { return Init("h4") }
func H5() *Elm { return Init("h5") }
func H6() *Elm { return Init("h6") }

// ------------------------------------------------------------------------------------------------
// Semantic layout
func Article() *Elm { return Init("article") }
func Aside() *Elm   { return Init("aside") }
func Section() *Elm { return Init("section") }
func Nav() *Elm     { return Init("nav") }
func Header() *Elm  { return Init("header") }
func Footer() *Elm  { return Init("footer") }
func Main() *Elm    { return Init("main") }

// ------------------------------------------------------------------------------------------------
// Lists
func Ul() *Elm { return Init("ul") }
func Ol() *Elm { return Init("ol") }
func Li() *Elm { return Init("li") }
func Dl() *Elm { return Init("dl") }
func Dt() *Elm { return Init("dt") }
func Dd() *Elm { return Init("dd") }

// ------------------------------------------------------------------------------------------------
// Inline formatting
func Strong() *Elm { return Init("strong") }
func Em() *Elm     { return Init("em") }
func Code() *Elm   { return Init("code") }
func Pre() *Elm    { return Init("pre") }
func Small() *Elm  { return Init("small") }
func Mark() *Elm   { return Init("mark") }
func B() *Elm      { return Init("b") }
func I() *Elm      { return Init("i") }

// ------------------------------------------------------------------------------------------------
// Form elements
func Form() *Elm     { return Init("form") }
func Input() *Elm    { return Init("input") }
func Label() *Elm    { return Init("label") }
func Select() *Elm   { return Init("select") }
func Option() *Elm   { return Init("option") }
func Textarea() *Elm { return Init("textarea") }
func Fieldset() *Elm { return Init("fieldset") }
func Legend() *Elm   { return Init("legend") }

// ------------------------------------------------------------------------------------------------
// Media / Void elements
func Img() *Elm { return Init("img") }
func Br() *Elm  { return Init("br") }
func Hr() *Elm  { return Init("hr") }

// ------------------------------------------------------------------------------------------------
// Tables
func Table() *Elm { return Init("table") }
func Thead() *Elm { return Init("thead") }
func Tbody() *Elm { return Init("tbody") }
func Tr() *Elm    { return Init("tr") }
func Th() *Elm    { return Init("th") }
func Td() *Elm    { return Init("td") }

// ------------------------------------------------------------------------------------------------
// Miscellaneous
func Figure() *Elm     { return Init("figure") }
func Figcaption() *Elm { return Init("figcaption") }
func Details() *Elm    { return Init("details") }
func Summary() *Elm    { return Init("summary") }
func Blockquote() *Elm { return Init("blockquote") }
func Cite() *Elm       { return Init("cite") }
func Time() *Elm       { return Init("time") }
