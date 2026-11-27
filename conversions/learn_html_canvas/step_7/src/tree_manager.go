package main

import (
	"math/rand/v2"

	"github.com/ewaldhorn/godom/dom"
)

// ------------------------------------------------------------------------------------------------
const (
	TREE_WIDTH    = 120
	TREE_GAP      = 20
	TREE_SPACING  = TREE_WIDTH + TREE_GAP
	NUM_TREES     = 2 + canvasWidth/TREE_WIDTH
	MIN_TREE_SIZE = 80
	MAX_TREE_SIZE = 120
)

// ------------------------------------------------------------------------------------------------
// randomTreeSize returns a random tree height between MIN_TREE_SIZE and MAX_TREE_SIZE inclusive.
func randomTreeSize() int {
	return MIN_TREE_SIZE + rand.IntN(MAX_TREE_SIZE-MIN_TREE_SIZE+1)
}

// ------------------------------------------------------------------------------------------------
type TreePosition struct {
	X, Y float64
	Size int
}

// ------------------------------------------------------------------------------------------------
func (t *TreePosition) Update() {
	t.X -= 1

	if t.X < -TREE_WIDTH {
		t.X += float64(NUM_TREES * TREE_SPACING)
		t.Size = randomTreeSize()
	}
}

// ------------------------------------------------------------------------------------------------
type TreeManager struct {
	trees []TreePosition
}

// ------------------------------------------------------------------------------------------------
func NewTreeManager() *TreeManager {
	mngr := &TreeManager{
		trees: make([]TreePosition, NUM_TREES),
	}
	mngr.initTrees()

	return mngr
}

// ------------------------------------------------------------------------------------------------
// initTrees initialises each tree in the list
func (t *TreeManager) initTrees() {
	for pos := range t.trees {
		t.trees[pos].X = float64(TREE_SPACING * pos)
		t.trees[pos].Size = randomTreeSize()
	}
}

// ------------------------------------------------------------------------------------------------
func (t *TreeManager) Update() {
	for pos := range t.trees {
		t.trees[pos].Update()
	}
}

// ------------------------------------------------------------------------------------------------
func (t *TreeManager) Render(ctx dom.Context2D) {
	for pos := range t.trees {
		drawTree(t.trees[pos].X, canvasHeight, float64(t.trees[pos].Size), ctx)
	}
}
