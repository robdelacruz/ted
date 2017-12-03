package main

import ()

type View struct {
	Area
	border Area
	*Buf
	*TextBlk
}

func NewView(x, y, w, h int, buf *Buf) *View {
	border := NewArea(x, y, w, h)
	area := NewArea(x+1, y+1, w-2, h-2)
	textBlk := NewTextBlk(area.Width, 0)

	view := &View{area, border, buf, textBlk}
	return view
}

func min(n1, n2 int) int {
	if n1 < n2 {
		return n1
	}
	return n2
}

func (v *View) Draw() {
	drawBox(v.border.X, v.border.Y, v.border.Width, v.border.Height, 0, 0)

	FillTextBlk(v.TextBlk, v.Buf)

	text := v.TextBlk.Text
	for y := 0; y < min(v.TextBlk.Height, v.Area.Height); y++ {
		for x := 0; x < min(v.TextBlk.Width, v.Area.Width); x++ {
			c := text[y][x]
			if c == 0 {
				print(string(" "), x+v.X, y+v.Y, 0, 0)
				continue
			}

			print(string(c), x+v.X, y+v.Y, 0, 0)
		}
	}
}
