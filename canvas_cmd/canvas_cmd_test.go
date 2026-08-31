package canvas_cmd

import (
	"bytes"
	"math"
	"testing"
)

// The wire format must stay identical to OdinDOM's canvas_cmd so the JS dispatcher (shared
// nearly verbatim between odindom.js and godom's docs/app.js) keeps decoding correctly.
// These tests pin the exact byte layout: opcode, then typed args (f32/u16/u8 LE, str as
// u16 length + UTF-8 bytes).

func TestSetFill(t *testing.T) {
	b := NewBuffer()
	b.SetFill("#fff")
	want := []byte{0x01, 0x04, 0x00, '#', 'f', 'f', 'f'}
	if !bytes.Equal(b.data, want) {
		t.Fatalf("SetFill: got % x, want % x", b.data, want)
	}
}

func TestFillRect(t *testing.T) {
	b := NewBuffer()
	b.FillRect(1, 2, 3, 4)
	want := []byte{0x10}
	for _, v := range []float32{1, 2, 3, 4} {
		bits := math.Float32bits(v)
		want = append(want, byte(bits), byte(bits>>8), byte(bits>>16), byte(bits>>24))
	}
	if !bytes.Equal(b.data, want) {
		t.Fatalf("FillRect: got % x, want % x", b.data, want)
	}
}

func TestArcDefaultCCW(t *testing.T) {
	b := NewBuffer()
	b.Arc(0, 0, 10, 0, 2*math.Pi)
	// opcode + 5 f32s + ccw byte
	if len(b.data) != 1+5*4+1 {
		t.Fatalf("Arc: got %d bytes, want %d", len(b.data), 1+5*4+1)
	}
	if b.data[len(b.data)-1] != 0 {
		t.Fatalf("Arc: default ccw must be 0, got %d", b.data[len(b.data)-1])
	}

	b2 := NewBuffer()
	b2.Arc(0, 0, 10, 0, 2*math.Pi, true)
	if b2.data[len(b2.data)-1] != 1 {
		t.Fatalf("Arc: explicit ccw=true must be 1, got %d", b2.data[len(b2.data)-1])
	}
}

func TestGradientOps(t *testing.T) {
	b := NewBuffer()
	b.LinearGradient(0xABCD, 0, 320, 800, 500)
	want := []byte{0x50, 0xCD, 0xAB} // opcode + u16 LE id
	if !bytes.Equal(b.data[:3], want) {
		t.Fatalf("LinearGradient header: got % x, want % x", b.data[:3], want)
	}
	if len(b.data) != 3+4*4 {
		t.Fatalf("LinearGradient: got %d bytes, want %d", len(b.data), 3+4*4)
	}
}

func TestSpriteOps(t *testing.T) {
	b := NewBuffer()
	b.BakeBegin(3, 70, 110) // exercises non-zero u16 ids
	want := []byte{0x80, 0x03, 0x00, 0x46, 0x00, 0x6E, 0x00}
	if !bytes.Equal(b.data, want) {
		t.Fatalf("BakeBegin: got % x, want % x", b.data, want)
	}

	b2 := NewBuffer()
	b2.DrawSpriteScaled(2, 1.5, -1.5, 48, 36)
	if b2.data[0] != 0x61 || len(b2.data) != 1+2+4*4 {
		t.Fatalf("DrawSpriteScaled: bad layout: % x", b2.data)
	}
}

func TestUnicodeString(t *testing.T) {
	b := NewBuffer()
	b.SetStroke("rgba(255,255,255,0.60)")
	want := append([]byte{0x02}, 0x16, 0x00) // len 22 LE
	want = append(want, []byte("rgba(255,255,255,0.60)")...)
	if !bytes.Equal(b.data, want) {
		t.Fatalf("SetStroke: got % x, want % x", b.data, want)
	}
}

func TestResetRewinds(t *testing.T) {
	b := NewBuffer()
	b.SetFill("red")
	b.Reset()
	if b.Len() != 0 {
		t.Fatalf("Reset: len = %d, want 0", b.Len())
	}
	b.FillRect(0, 0, 1, 1)
	if b.Len() != 1+4*4 {
		t.Fatalf("after re-record: len = %d, want %d", b.Len(), 1+4*4)
	}
}
