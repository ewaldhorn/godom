package colour

import "testing"

func TestIsEmpty(t *testing.T) {
	if !Empty.IsEmpty() {
		t.Error("expected Empty color to be empty")
	}

	if White.IsEmpty() {
		t.Error("expected White color to not be empty")
	}

	if Black.IsEmpty() {
		t.Error("expected Black color to not be empty")
	}

	c := Colour{R: 0, G: 0, B: 0, A: 1}
	if c.IsEmpty() {
		t.Error("expected Colour with A: 1 to not be empty")
	}
}

func TestConvertToGrayscale(t *testing.T) {
	tests := []struct {
		input    Colour
		expected uint8
	}{
		{Colour{R: 255, G: 255, B: 255, A: 255}, 255}, // White -> 255
		{Colour{R: 0, G: 0, B: 0, A: 255}, 0},         // Black -> 0
		{Colour{R: 255, G: 0, B: 0, A: 255}, 76},      // Red: 0.299*255 = 76
		{Colour{R: 0, G: 255, B: 0, A: 255}, 149},     // Green: 0.587*255 = 149
		{Colour{R: 0, G: 0, B: 255, A: 255}, 29},      // Blue: 0.114*255 = 29
	}

	for _, tc := range tests {
		c := tc.input
		c.ConvertToGrayscale()
		if c.R != tc.expected || c.G != tc.expected || c.B != tc.expected {
			t.Errorf("expected grayscale shade %d for color %+v, got R=%d G=%d B=%d", tc.expected, tc.input, c.R, c.G, c.B)
		}
	}
}

func TestSeedAndRandomColour(t *testing.T) {
	// Re-seed with a known value
	Seed(42)
	c1 := RandomColour()
	c2 := RandomColour()

	// Reset seed and verify identical sequence (reproducibility)
	Seed(42)
	c1_second := RandomColour()
	c2_second := RandomColour()

	if c1 != c1_second || c2 != c2_second {
		t.Error("expected identical random colors after re-seeding with the same value")
	}

	// Verify A component is always 255 (fully opaque)
	if c1.A != 255 || c2.A != 255 {
		t.Errorf("expected alpha to be 255, got A1=%d A2=%d", c1.A, c2.A)
	}

	// Verify that a seed of 0 is ignored
	Seed(0) // should not change rngState
	c3 := RandomColour()
	
	Seed(42)
	_ = RandomColour()
	_ = RandomColour()
	c3_expected := RandomColour()

	if c3 != c3_expected {
		t.Error("expected seed of 0 to be ignored, but rngState sequence changed")
	}
}
