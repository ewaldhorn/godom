package canvas

// ------------------------------------------------------------------------------------------------
// Point represents a 2D coordinate vector.
type Point struct {
	X int32
	Y int32
}

// ------------------------------------------------------------------------------------------------
// abs returns the absolute value of the 32-bit integer v.
func abs(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}
