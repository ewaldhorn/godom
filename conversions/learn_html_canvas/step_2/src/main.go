package main

import (
	"github.com/ewaldhorn/godom/dom"
	"github.com/ewaldhorn/godom/html"
)

func main() {
	// First step is to init the dom
	dom.Init()

	// now we can create elements
	// we won't need the actual container so we ignore the variable
	_ = html.Div().Class("container").Child(html.H1().Text("Hello from GoDom").Build()).AppendTo(dom.Body)

	// need to block this from exiting, otherwise wasm terminates our app
	select {}
}
