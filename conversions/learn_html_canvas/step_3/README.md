# Step 3

With basic canvas rendering handled, we can start diving deeper into the actual game code.

## Time to update the index page

Let's edit our `index.html` to reflect where we are now:

```html
<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <title>GoDom Tutorial Step 3</title>
    <script src="wasm_exec.js"></script>
  </head>
  <body>
      <script>
        const go = new Go();
        WebAssembly.instantiateStreaming(fetch("app.wasm"), go.importObject).then((result) => {
          go.run(result.instance);
        }).catch(err => {
          console.error("Failed to load WASM:", err);
        });
      </script>
  </body>
</html>
```

## We're gonna need random numbers

One of the things I use a lot in my games, is random numbers. Now I know there are tons of different ways to solve this problem, but there's a particular solution I quite like: SplitMix64. You can read up on it at [RosettaCode Pseudo-random numbers](https://rosettacode.org/wiki/Pseudo-random_numbers/Splitmix64). The reason I like this method is that I can, when I need to, make it deterministic. For example, in this project, I want trees to be "random" but that randomness must be deterministic. That'll make sense when you see the game rendering, I hope.

For now, let's introduce a new Go source file called `rng.go`:

```go
package main

// For the trees, I wanted each tree to look random, but I wanted that randomness to be tied to
// the tree height. That way, I can have trees render at any xy position, but their look is
// determined by their height. So it's sort of deterministic.
//
// I used https://rosettacode.org/wiki/Pseudo-random_numbers/Splitmix64 as a reference for building
// this. It's something I've been using since my Java days!

// ------------------------------------------------------------------------------------------------
type rng struct{ s uint64 }

// ------------------------------------------------------------------------------------------------
func newRNG(seed uint64) *rng {
	return &rng{s: seed}
}

// ------------------------------------------------------------------------------------------------
// next uses the golden ration (0x9E3779B97F4A7C15) to advance the seed and generated the next
// random number. Basically, this is SplitMix64, which is pretty cool for games. I tend to use
// it quite often because it's lightweight but efficient. Don't build your crypto exchange on it
// though, unless you want to get famous for the wrong reasons!
func (r *rng) next() uint64 {
	r.s += 0x9E3779B97F4A7C15
	z := r.s
	z = (z ^ (z >> 30)) * 0xBF58476D1CE4E5B9
	z = (z ^ (z >> 27)) * 0x94D049BB133111EB
	return z ^ (z >> 31)
}

// ------------------------------------------------------------------------------------------------
func (r *rng) float64() float64 {
	return float64(r.next()>>11) / float64(1<<53)
}

// ------------------------------------------------------------------------------------------------
func (r *rng) intn(n int) int {
	return int(r.next() % uint64(n))
}
```

If you like, you can go read up more about how this works, but for now, all you need to know is that we can create a random number generator and ask it for a float64 or int value and it is deterministic.

## Getting the canvas configured

To be able to draw something, we need an HTML Canvas configured. So we could create a canvas element in our HTML page, but we can also ask Go to do that for us. Since we're trying to do as much Go as possible, let's do that.

In our `main.go`, we can create a new HTML Canvas element like this:

```go
func main() {
	// First step is to init the dom
	dom.Init()

	// now we can create elements
	// we won't need the actual container so we ignore the variable
	_ = html.Div().Class("container").Child(html.H1().Text("Hello from GoDom").Build()).AppendTo(dom.Body)

	// create our canvas and draw to it
	context2d := createAndConfigureHTMLCanvas()
	drawToTheCanvas(context2d)

	// need to block this from exiting, otherwise wasm terminates our app
	select {}
}
```

I know, those functions don't exist yet, but we'll get to them in a second... First up, let's take a look at what `createAndConfigureHTMLCanvas` does:

```go
func createAndConfigureHTMLCanvas() dom.Context2D {
	// create a new Div container for our canvas
	canvasContainer := html.Div().Class("canvas-container")

	// create the actual canvas, set the attributes and background colour
	myCanvas := html.Canvas()
	myCanvas.Attr("width", "800")
	myCanvas.Attr("height", "600")
	myCanvas.Style("background-color", "blue")

	// apped the canvas to the container, then the container to the html body
	myCanvas.AppendTo(canvasContainer.Handle)
	canvasContainer.AppendTo(dom.Body)

	// return the Context2D handle, which is what we need to draw to the canvas
	return getContext2DfromCanvasHandle(myCanvas.Handle)
}
```

That's quite a beast, but, is it, really? We've already created a div container before, so that's not new. Creating an HTML Canvas element seems pretty straight-forward, you create the element, set some attributes and styles, pretty much like you'd do in JavaScript anyway. Then we just append the canvas to the container, the container to the HTML body and... Well, that's almost it. There's a funny `getContext2DfromCanvasHandle` function:

```go
func getContext2DfromCanvasHandle(canvasHandle dom.Handle) dom.Context2D {
	if !canvasHandle.IsValid() {
		panic("godom: getContext2DfromCanvasHandle received an invalid canvas handle")
	}

	return dom.NewContext2D(canvasHandle)
}
```
Oh, ok, well, that's not too bad, is it? Basically, check if we have a valid handle, then get a Context2D from it and return that.

Finally, we have the `` function:

```go
func drawToTheCanvas(ctx dom.Context2D) {
	ctx.FillStyle("yellow")
	ctx.Arc(10.0, 10.0, 150.0, 0.0, 180.0, false)
	ctx.Fill()
}
```

Now, if you've ever done JavaScript HTML Canvas drawing, this code should look pretty familiar! And that's what I tried to do with GoDom: Make it as close to the JavaScript code I'm used to write, just with a bit of a Go flavour mixed in.

If you run this, you should get:

![Step 3 Screenshot](step3.png)
