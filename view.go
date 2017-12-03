package main

import ()

type View struct {
	Area
	border Area
	*Buf
	TextBlk [][]rune
}

func NewView(x, y, w, h int, buf *Buf) *View {
	border := NewArea(x, y, w, h)
	area := NewArea(x+1, y+1, w-2, h-2)
	textBlk := allocTextBlk(area)

	view := &View{area, border, buf, textBlk}
	return view
}

// Allocate space for textblk
func allocTextBlk(area Area) [][]rune {
	textBlk := make([][]rune, 0)
	for y := 0; y < area.Height; y++ {
		blkline := make([]rune, area.Width)
		textBlk = append(textBlk, blkline)
	}
	return textBlk
}

func (v *View) Draw() {
	drawBox(v.border.X, v.border.Y, v.border.Width, v.border.Height, 0, 0)

	v.Buf.CopyToBlk(v.TextBlk)
	for y := 0; y < len(v.TextBlk); y++ {
		for x := 0; x < len(v.TextBlk[y]); x++ {
			c := v.TextBlk[y][x]
			if c == 0 {
				print(string(" "), x+v.X, y+v.Y, 0, 0)
				continue
			}

			print(string(c), x+v.X, y+v.Y, 0, 0)
		}
	}
}
