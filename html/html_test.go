package html

import (
	"syscall/js"
	"testing"
)

// ------------------------------------------------------------------------------------------------
func init() {
	// Bootstrap a minimal DOM stub in the JS environment.
	js.Global().Get("eval").Invoke(`
		globalThis.document = {
			createElement: function(tag) {
				let el = {
					tagName: tag,
					id: "",
					className: "",
					innerText: "",
					innerHTML: "",
					children: [],
					listeners: {},
					setAttribute: function(key, val) {
						this[key] = val;
					},
					classList: {
						add: function(cls) {
							if (!el.className) {
								el.className = cls;
							} else {
								el.className += " " + cls;
							}
						}
					},
					appendChild: function(child) {
						this.children.push(child);
					},
					addEventListener: function(event, cb) {
						if (!this.listeners[event]) {
							this.listeners[event] = [];
						}
						this.listeners[event].push(cb);
					}
				};
				return el;
			}
		};
	`)
}

// ------------------------------------------------------------------------------------------------
func TestElementBuilder(t *testing.T) {
	el := Div().
		ID("test-div").
		Class("container").
		Class("active").
		Text("Hello").
		HTML("<p>Hello</p>").
		Attr("data-custom", "value")

	val := js.Value(el.Handle)

	if id := val.Get("id").String(); id != "test-div" {
		t.Errorf("expected ID 'test-div', got '%s'", id)
	}

	if class := val.Get("className").String(); class != "container active" {
		t.Errorf("expected className 'container active', got '%s'", class)
	}

	if txt := val.Get("innerText").String(); txt != "Hello" {
		t.Errorf("expected innerText 'Hello', got '%s'", txt)
	}

	if htmlVal := val.Get("innerHTML").String(); htmlVal != "<p>Hello</p>" {
		t.Errorf("expected innerHTML '<p>Hello</p>', got '%s'", htmlVal)
	}

	if custom := val.Get("data-custom").String(); custom != "value" {
		t.Errorf("expected data-custom 'value', got '%s'", custom)
	}
}

// ------------------------------------------------------------------------------------------------
func TestElementHierarchy(t *testing.T) {
	parent := Div()
	child := Span().ID("child-span")

	parent.Child(child.Build())

	valParent := js.Value(parent.Handle)
	childrenLength := valParent.Get("children").Length()
	if childrenLength != 1 {
		t.Errorf("expected parent to have 1 child, got %d", childrenLength)
	}

	firstChild := valParent.Get("children").Index(0)
	if id := firstChild.Get("id").String(); id != "child-span" {
		t.Errorf("expected first child to have ID 'child-span', got '%s'", id)
	}

	// Test AppendTo
	parent2 := Section()
	child2 := Article().ID("child-article")
	child2.AppendTo(parent2.Build())

	valParent2 := js.Value(parent2.Handle)
	if l := valParent2.Get("children").Length(); l != 1 {
		t.Errorf("expected parent2 to have 1 child, got %d", l)
	}
}

// ------------------------------------------------------------------------------------------------
func TestElementOn(t *testing.T) {
	btn := Button()
	called := false
	cb := js.FuncOf(func(this js.Value, args []js.Value) any {
		called = true
		return nil
	})
	defer cb.Release()

	btn.On("click", cb.Value)

	val := js.Value(btn.Handle)
	listeners := val.Get("listeners").Get("click")
	if listeners.Length() != 1 {
		t.Errorf("expected 1 click listener, got %d", listeners.Length())
	}

	// Invoke the callback in JS to test event binding
	listeners.Index(0).Invoke()
	if !called {
		t.Error("expected callback to be called")
	}
}
