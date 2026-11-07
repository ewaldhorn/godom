// Package sound provides audio synthesis utilities for the browser.
package sound

import "math"

// ------------------------------------------------------------------------------------------------
// SampleRate defines the audio sample rate (44.1kHz).
const SampleRate = 44100.0

// ------------------------------------------------------------------------------------------------
// FillClick pre-renders a 50ms triangle wave click sound with exponential decay.
func FillClick(buf []float32) {
	if len(buf) == 0 {
		return
	}
	const (
		freqStart = float32(2000.0)
		freqEnd   = float32(600.0)
		decayRate = float32(20.0)
	)

	var phase float32
	bufLen := float32(len(buf))

	for i := range buf {
		t := float32(i) / SampleRate
		progress := float32(i) / bufLen
		freq := freqStart + (freqEnd-freqStart)*progress

		phase += freq / SampleRate
		for phase >= 1.0 {
			phase -= 1.0
		}

		var osc float32
		if phase < 0.5 {
			osc = -1.0 + 4.0*phase
		} else {
			osc = 3.0 - 4.0*phase
		}

		decay := float32(math.Exp(float64(-decayRate * t)))
		buf[i] = osc * decay * 0.25
	}
}
