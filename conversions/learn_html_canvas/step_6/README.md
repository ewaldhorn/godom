# Step 6

Up and down movement for the balloon works! Well, sort of, but it's a big step forward. I think it's time we saw the forest for the trees, or something like that...

## Update the index page

Let's edit our `index.html`, as we always do:

```html
.
.
  <head>
    <meta charset="utf-8" />
    <title>GoDom Tutorial Step 6</title>
    <script src="wasm_exec.js"></script>
  </head>
  .
  .
```

## Tweaks

First thing I'm going to do is make the balloon a bit smaller, so in `balloon.go`:

```go
const (
	BASKET_WIDTH      = 80
	BASKET_HALF_WIDTH = BASKET_WIDTH / 2
	BASKET_HEIGHT     = 40
)

func drawBalloon(pos BalloonStatus, ctx dom.Context2D) {
	// first we render the basket
	ctx.FillStyle("brown")
	ctx.FillRect(pos.X-BASKET_HALF_WIDTH, pos.Y, BASKET_WIDTH, BASKET_HEIGHT)

	// now the balloon
	ctx.FillStyle("#28DD28")
	ctx.BeginPath()
	ctx.MoveTo(pos.X, pos.Y)                                       // middle of the balloon
	ctx.QuadraticCurveTo(pos.X-60, pos.Y-50, pos.X-50, pos.Y-100)  // bottom-left
	ctx.QuadraticCurveTo(pos.X-25, pos.Y-170, pos.X, pos.Y-170)    // top-left to apex
	ctx.QuadraticCurveTo(pos.X+25, pos.Y-170, pos.X+50, pos.Y-100) // apex to right
	ctx.QuadraticCurveTo(pos.X+60, pos.Y-50, pos.X, pos.Y)         // right down to middle
	ctx.ClosePath()
	ctx.Fill()
}
```

Ok now that the balloon is a bit smaller, we can also change the start position in `main.go`:

```go
var balloonPos = BalloonStatus{X: 100, Y: 200}
```

Puts out balloon more top-left than before. This is now our starting position in the game. When you run with these changes, you will see that the balloon still goes up and down, so nothing changed there. Just the size and starting position.

## Trees are good, forests are better

With that done, let's revisit our tree. Right now we have a simple function or two that renders a tree, but we don't actually keep state of a trees position so that we can animate it. Given how much needs to happen, I'm going to create a `tree_manager.go` file and put the tree management related code in there. Don't worry, in Step 7 we'll do a lot more code cleanups, but right now, we just want to get stuff working!

In `tree_manager.go`, we'll need some constants.

```go
const (
	TREE_WIDTH    = 120
	TREE_GAP      = 20
	TREE_SPACING  = TREE_WIDTH + TREE_GAP
	NUM_TREES     = 2 + canvasWidth/TREE_WIDTH
	MIN_TREE_SIZE = 80
	MAX_TREE_SIZE = 120
)
```

Each tree takes up 120 pixels, i want about a 20 pixel gap between each. This isn't EXACT, but it's good enough for the game. Some trees will overlap, some won't, but that adds a bit of spice! I want more trees than what fit on the screen, so that we can scroll them and trees can be partially on the screen. Everytime a tree goes out of view, we move it to the far right of the screen, randomise its height and that way it looks like an endless forest of varying trees.

The simplest way to get a random tree size is with this function:

```go
func randomTreeSize() int {
	return MIN_TREE_SIZE + rand.IntN(MAX_TREE_SIZE-MIN_TREE_SIZE+1)
}
```

Now keeping track of the tree position requires a few variables, so I think putting it into a struct makes sense, like so:

```go
type TreePosition struct {
	X, Y float64
	Size int
}
```

For now, we have an XY position and a size. I am defaulting the Y to the canvas height, but who knows what we do in the future... Every tree needs to move to the left on each game tick, so we create a function for that:

```go
func (t *TreePosition) Update() {
	t.X -= 1

	if t.X < -TREE_WIDTH {
		t.X += float64(NUM_TREES * TREE_SPACING)
		t.Size = randomTreeSize()
	}
}
```

Here we decrease the X position of the tree and check if it's entirely off the screen.  If it is, put it on the far right hand side by changing the X position and then give it a random size. And with that we create the illusion of more trees appearing, when, in fact, we are just recycling the same variables over and over.



You should get this in your browser:

![Step 5 Screenshot](step5.png)

If you run into issues, check your code against the repo, there might be something missing.

## Step 5 is done, Step 6 is next

Let's go to Step 6 when you are ready:  [Step 6](../step_6/)
