package main

import (
	"github.com/ewaldhorn/godom/canvas"
	"github.com/ewaldhorn/godom/colour"
)

// getGlyphRows returns the 3x5 bitmap rows for a supported uppercase ASCII character or digit.
func getGlyphRows(ch uint8) [5]string {
	switch ch {
	case '0':
		return [5]string{"111", "101", "101", "101", "111"}
	case '1':
		return [5]string{"010", "110", "010", "010", "111"}
	case '2':
		return [5]string{"111", "001", "111", "100", "111"}
	case '3':
		return [5]string{"111", "001", "111", "001", "111"}
	case '4':
		return [5]string{"101", "101", "111", "001", "001"}
	case '5':
		return [5]string{"111", "100", "111", "001", "111"}
	case '6':
		return [5]string{"111", "100", "111", "101", "111"}
	case '7':
		return [5]string{"111", "001", "010", "010", "010"}
	case '8':
		return [5]string{"111", "101", "111", "101", "111"}
	case '9':
		return [5]string{"111", "101", "111", "001", "111"}
	case 'A':
		return [5]string{"010", "101", "111", "101", "101"}
	case 'B':
		return [5]string{"110", "101", "110", "101", "110"}
	case 'C':
		return [5]string{"111", "100", "100", "100", "111"}
	case 'D':
		return [5]string{"110", "101", "101", "101", "110"}
	case 'E':
		return [5]string{"111", "100", "110", "100", "111"}
	case 'H':
		return [5]string{"101", "101", "111", "101", "101"}
	case 'I':
		return [5]string{"111", "010", "010", "010", "111"}
	case 'K':
		return [5]string{"101", "101", "110", "101", "101"}
	case 'L':
		return [5]string{"100", "100", "100", "100", "111"}
	case 'M':
		return [5]string{"101", "111", "111", "101", "101"}
	case 'N':
		return [5]string{"101", "111", "111", "111", "101"}
	case 'O':
		return [5]string{"111", "101", "101", "101", "111"}
	case 'P':
		return [5]string{"110", "101", "110", "100", "100"}
	case 'R':
		return [5]string{"110", "101", "110", "101", "101"}
	case 'S':
		return [5]string{"111", "100", "111", "001", "111"}
	case 'U':
		return [5]string{"101", "101", "101", "101", "111"}
	default:
		return [5]string{"000", "000", "000", "000", "000"}
	}
}

// drawBitmapText renders uppercase ASCII text onto target canvas c using scaled 3x5 blocks.
func drawBitmapText(c *canvas.Canvas, text string, x, y, scale int32, col colour.Colour) {
	cursor := x
	for i := range len(text) {
		ch := text[i]
		if ch != ' ' {
			drawBitmapGlyph(c, ch, cursor, y, scale, col)
		}
		cursor += 4 * scale
	}
}

// drawBitmapDigit renders a single decimal digit (0 to 9) onto target canvas c.
func drawBitmapDigit(c *canvas.Canvas, digit uint8, x, y, scale int32, col colour.Colour) {
	drawBitmapGlyph(c, '0'+digit, x, y, scale, col)
}

// drawBitmapGlyph renders a single 3x5 character glyph onto target canvas c.
func drawBitmapGlyph(c *canvas.Canvas, ch uint8, x, y, scale int32, col colour.Colour) {
	rows := getGlyphRows(ch)
	for rowIdx, row := range rows {
		for colIdx := range len(row) {
			if row[colIdx] == '1' {
				c.ColourFilledRectangle(
					x+int32(colIdx)*scale,
					y+int32(rowIdx)*scale,
					scale,
					scale,
					col,
				)
			}
		}
	}
}
