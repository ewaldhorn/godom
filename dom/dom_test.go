package dom

import (
	"syscall/js"
	"testing"
	"time"
)

func init() {
	// Bootstrap DOM, Console, and alert mocks for testing.
	js.Global().Get("eval").Invoke(`
		globalThis.alert = function(msg) {
			globalThis.lastAlert = msg;
		};
		globalThis.console = {
			log: function(...args) {
				globalThis.lastLog = args.join(" ");
			}
		};
		globalThis.stopRAF = false;
		globalThis.requestAnimationFrame = function(cb) {
			if (!globalThis.stopRAF) {
				setTimeout(cb, 10);
			}
		};

		let mockElements = {};

		globalThis.document = {
			body: {
				tagName: "body",
				children: [],
				appendChild: function(c) {
					this.children.push(c);
				}
			},
			head: {
				tagName: "head",
				children: [],
				appendChild: function(c) {
					this.children.push(c);
				}
			},
			createElement: function(tag) {
				let el = {
					tagName: tag,
					id: "",
					className: "",
					innerText: "",
					innerHTML: "",
					style: {
						display: ""
					},
					children: [],
					listeners: {},
					classList: {
						add: function(cls) {
							if (!el.className) {
								el.className = cls;
							} else {
								el.className += " " + cls;
							}
						},
						remove: function(cls) {
							el.className = el.className.split(" ").filter(c => c !== cls).join(" ");
						}
					},
					appendChild: function(child) {
						this.children.push(child);
					},
					replaceChildren: function() {
						this.children.length = 0;
					},
					addEventListener: function(event, cb) {
						if (!this.listeners[event]) {
							this.listeners[event] = [];
						}
						this.listeners[event].push(cb);
					},
					focus: function() {
						this.focused = true;
					}
				};
				return el;
			},
			getElementById: function(id) {
				if (!mockElements[id]) {
					let el = this.createElement("div");
					el.id = id;
					mockElements[id] = el;
				}
				return mockElements[id];
			}
		};
	`)
}

func TestInit(t *testing.T) {
	Init()

	if !Document.IsValid() {
		t.Error("expected Document to be valid after Init")
	}
	if !Body.IsValid() {
		t.Error("expected Body to be valid after Init")
	}
	if !Head.IsValid() {
		t.Error("expected Head to be valid after Init")
	}
}

func TestIsValid(t *testing.T) {
	Init()

	if Invalid.IsValid() {
		t.Error("expected Invalid handle to return false for IsValid()")
	}

	h := Handle(js.ValueOf(map[string]any{"x": 1}))
	if !h.IsValid() {
		t.Error("expected valid map handle to return true for IsValid()")
	}
}

func TestGetSet(t *testing.T) {
	el := CreateElement("div")
	el.Set("customProp", "hello")

	if val := el.Get("customProp"); val != "hello" {
		t.Errorf("expected customProp to be 'hello', got '%s'", val)
	}
}

func TestInnerTextAndHTML(t *testing.T) {
	el := CreateElement("span")
	el.SetInnerText("test-text")
	if txt := el.Get("innerText"); txt != "test-text" {
		t.Errorf("expected innerText 'test-text', got '%s'", txt)
	}

	el.SetInnerHTML("<b>test-html</b>")
	if htmlVal := el.Get("innerHTML"); htmlVal != "<b>test-html</b>" {
		t.Errorf("expected innerHTML '<b>test-html</b>', got '%s'", htmlVal)
	}
}

func TestClassList(t *testing.T) {
	el := CreateElement("div")
	el.AddClassTo("btn")
	el.AddClassTo("primary")

	if val := el.Get("className"); val != "btn primary" {
		t.Errorf("expected className 'btn primary', got '%s'", val)
	}

	el.RemoveClassFrom("btn")
	if val := el.Get("className"); val != "primary" {
		t.Errorf("expected className 'primary', got '%s'", val)
	}

	el.ReplaceClasses([]string{"a", "b", "c"})
	if val := el.Get("className"); val != "a b c" {
		t.Errorf("expected className 'a b c', got '%s'", val)
	}
}

func TestElementCreation(t *testing.T) {
	d := CreateDiv()
	if tag := d.Get("tagName"); tag != "div" {
		t.Errorf("expected tagName 'div', got '%s'", tag)
	}

	p := CreateParagraph()
	if tag := p.Get("tagName"); tag != "p" {
		t.Errorf("expected tagName 'p', got '%s'", tag)
	}

	pt := CreateParagraphWithText("ptext")
	if txt := pt.Get("innerText"); txt != "ptext" {
		t.Errorf("expected paragraph innerText 'ptext', got '%s'", txt)
	}

	btn := CreateButton("clicker")
	if typ := btn.Get("type"); typ != "button" {
		t.Errorf("expected button type to be 'button', got '%s'", typ)
	}
	if txt := btn.Get("innerText"); txt != "clicker" {
		t.Errorf("expected button text 'clicker', got '%s'", txt)
	}

	img := CreateImg("src.png")
	if src := img.Get("src"); src != "src.png" {
		t.Errorf("expected img src 'src.png', got '%s'", src)
	}
}

func TestChildOperations(t *testing.T) {
	Init()

	parent := CreateDiv()
	child := CreateParagraph()
	AddElementTo(parent, child)

	children := js.Value(parent).Get("children")
	if children.Length() != 1 {
		t.Fatalf("expected 1 child, got %d", children.Length())
	}

	AddToBody(parent)
	bodyChildren := js.Value(Body).Get("children")
	if bodyChildren.Length() == 0 {
		t.Error("expected Body to have children after AddToBody")
	}

	RemoveAllChildElementsFrom(parent)
	if children.Length() != 0 {
		t.Errorf("expected children to be cleared, got %d left", children.Length())
	}

	wrapped := WrapElementWithNewDiv(child, []string{"wrapper-class"})
	if class := wrapped.Get("className"); class != "wrapper-class" {
		t.Errorf("expected wrapper class 'wrapper-class', got '%s'", class)
	}
	if wrappedChildren := js.Value(wrapped).Get("children"); wrappedChildren.Length() != 1 {
		t.Errorf("expected wrapped div to contain 1 child, got %d", wrappedChildren.Length())
	}
}

func TestVisibilityAndFocus(t *testing.T) {
	Hide("target-hide")
	elHide := GetElementByID("target-hide")
	if disp := js.Value(elHide).Get("style").Get("display").String(); disp != "none" {
		t.Errorf("expected target-hide display to be 'none', got '%s'", disp)
	}

	Show("target-show")
	elShow := GetElementByID("target-show")
	if disp := js.Value(elShow).Get("style").Get("display").String(); disp != "block" {
		t.Errorf("expected target-show display to be 'block', got '%s'", disp)
	}

	SetFocus("target-focus")
	elFocus := GetElementByID("target-focus")
	if foc := js.Value(elFocus).Get("focused").Bool(); !foc {
		t.Error("expected target-focus to be focused")
	}
}

func TestPropertyAccessAndClasses(t *testing.T) {
	SetValue("target-prop", "value", "123")
	if val := GetString("target-prop", "value"); val != "123" {
		t.Errorf("expected string property '123', got '%s'", val)
	}

	AddClass("target-class", "new-class")
	el := GetElementByID("target-class")
	if class := el.Get("className"); class != "new-class" {
		t.Errorf("expected class 'new-class', got '%s'", class)
	}

	RemoveClass("target-class", "new-class")
	if class := el.Get("className"); class != "" {
		t.Errorf("expected class to be empty, got '%s'", class)
	}
}

func TestEvents(t *testing.T) {
	el := CreateElement("button")
	called := false
	cb := js.FuncOf(func(this js.Value, args []js.Value) any {
		called = true
		return nil
	})
	defer cb.Release()

	AddEventListener(el, "click", cb.Value)

	listeners := js.Value(el).Get("listeners").Get("click")
	if listeners.Length() != 1 {
		t.Fatalf("expected 1 click listener, got %d", listeners.Length())
	}
	listeners.Index(0).Invoke()
	if !called {
		t.Error("expected callback to be called via event listener")
	}

	// Test by ID
	calledByID := false
	cbByID := js.FuncOf(func(this js.Value, args []js.Value) any {
		calledByID = true
		return nil
	})
	defer cbByID.Release()

	AddEventListenerByID("btn-by-id", "click", cbByID.Value)
	elByID := GetElementByID("btn-by-id")
	listenersByID := js.Value(elByID).Get("listeners").Get("click")
	listenersByID.Index(0).Invoke()
	if !calledByID {
		t.Error("expected callback to be called via event listener by ID")
	}
}

func TestHelpers(t *testing.T) {
	Log([]string{"message1", "message2"})
	lastLog := js.Global().Get("lastLog").String()
	if lastLog != "message1 message2" {
		t.Errorf("expected logged output 'message1 message2', got '%s'", lastLog)
	}

	Alert("warning message")
	lastAlert := js.Global().Get("lastAlert").String()
	if lastAlert != "warning message" {
		t.Errorf("expected alert message 'warning message', got '%s'", lastAlert)
	}
}

func TestAnimationLoop(t *testing.T) {
	called := false
	cb := js.FuncOf(func(this js.Value, args []js.Value) any {
		called = true
		return nil
	})
	defer cb.Release()

	defer js.Global().Set("stopRAF", true)

	StartAnimationLoop(cb.Value)

	// Wait up to 100ms for animation loop requestAnimationFrame timeout
	start := time.Now()
	for time.Since(start) < 100*time.Millisecond {
		if called {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if !called {
		t.Error("expected animation loop callback to be called")
	}
}
