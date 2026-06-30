package canvas

// ------------------------------------------------------------------------------------------------
// Triangle draws a wireframe triangle through the three Points using the active color.
func (c *Canvas) Triangle(p1, p2, p3 Point) {
	c.LinePoint(p1, p2)
	c.LinePoint(p2, p3)
	c.LinePoint(p1, p3)
}
