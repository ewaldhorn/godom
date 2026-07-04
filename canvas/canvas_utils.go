package canvas

// ------------------------------------------------------------------------------------------------
// Point represents a 2D coordinate vector.
type Point struct {
	X int32
	Y int32
}

// ------------------------------------------------------------------------------------------------
// abs returns the absolute value of the 32-bit integer v.
// MinInt32 is clamped to MaxInt32 to avoid overflow on negation.
func abs(v int32) int32 {
	if v >= 0 {
		return v
	}
	if v == -1<<31 {
		return 1<<31 - 1
	}
	return -v
}
