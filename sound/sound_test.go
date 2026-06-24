package sound

import (
	"math"
	"testing"
)

func TestFillClick(t *testing.T) {
	// 50ms click buffer at 44100Hz = 2205 samples
	buf := make([]float32, 2205)
	FillClick(buf)

	var maxVal float32
	var hasNonZero bool

	for idx, val := range buf {
		if math.IsNaN(float64(val)) {
			t.Fatalf("found NaN in buffer at index %d", idx)
		}
		absVal := val
		if val < 0 {
			absVal = -val
		}
		if absVal > maxVal {
			maxVal = absVal
		}
		if absVal > 0 {
			hasNonZero = true
		}
	}

	if !hasNonZero {
		t.Error("expected click buffer to contain non-zero sample values")
	}

	// Click max amplitude is scaled to 0.25.
	// Allow a small floating-point threshold margin.
	if maxVal > 0.2501 {
		t.Errorf("expected maximum amplitude to be <= 0.25, got %f", maxVal)
	}

	// Verify exponential decay by comparing early peak amplitude to final amplitude
	firstHalfMax := float32(0.0)
	for i := 0; i < len(buf)/2; i++ {
		absVal := buf[i]
		if absVal < 0 {
			absVal = -absVal
		}
		if absVal > firstHalfMax {
			firstHalfMax = absVal
		}
	}

	lastQuarterMax := float32(0.0)
	for i := len(buf) - len(buf)/4; i < len(buf); i++ {
		absVal := buf[i]
		if absVal < 0 {
			absVal = -absVal
		}
		if absVal > lastQuarterMax {
			lastQuarterMax = absVal
		}
	}

	if lastQuarterMax >= firstHalfMax {
		t.Errorf("expected sound to decay, but early max %f is not greater than late max %f", firstHalfMax, lastQuarterMax)
	}
}
