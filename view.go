package main

import (
	tb "github.com/nsf/termbox-go"
)

type View struct {
	Area
	border Area
	*Buf
	*TextBlk
	Cur Pos
}

func NewView(x, y, w, h int, buf *Buf) *View {
	border := NewArea(x, y, w, h)
	area := NewArea(x+1, y+1, w-2, h-2)
	textBlk := NewTextBlk(area.Width, 0)

	view := &View{
		Area:    area,
		border:  border,
		Buf:     buf,
		TextBlk: textBlk,
		Cur:     Pos{0, 0},
	}
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

	v.drawText()

	tb.SetCursor(v.Area.X+v.Cur.X, v.Area.Y+v.Cur.Y)
}

func (v *View) drawText() {
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

func (v *View) HandleEvent(e *tb.Event) {
	var c rune
	if e.Type == tb.EventKey {
		switch e.Key {
		case tb.KeyArrowLeft:
			v.CurLeft()
		case tb.KeyArrowRight:
			v.CurRight()
		case tb.KeyArrowUp:
			v.CurUp()
		case tb.KeyArrowDown:
			v.CurDown()
		case 0:
			c = e.Ch
		}
	}

	// Char entered
	if c != 0 {
	}
}

func (v *View) CurXInc() bool {
	if v.Cur.X == v.Area.Width-1 && v.Cur.Y == v.Area.Height-1 {
		return false
	}

	v.Cur.X++
	if v.Cur.X > v.Area.Width-1 {
		v.Cur.Y++
		v.Cur.X = 0
	}
	return true
}
func (v *View) CurXDec() {
	if v.Cur.X == 0 && v.Cur.Y == 0 {
		return
	}

	v.Cur.X--
	if v.Cur.X < 0 {
		v.Cur.Y--
		v.Cur.X = v.Area.Width - 1
	}
}
func (v *View) CurYInc() {
	if v.Cur.Y == v.Cur.Y-1 {
		return
	}
	v.Cur.Y++
}
func (v *View) CurYDec() {
	if v.Cur.Y == 0 {
		return
	}
	v.Cur.Y--
}

// Return char under the cursor or 0 if out of bounds.
func (v *View) CurChar() rune {
	if v.InBoundsCur() {
		return v.TextBlk.Text[v.Cur.Y][v.Cur.X]
	}
	return 0
}
func (v *View) IsNilCur() bool {
	return v.CurChar() == 0
}
func (v *View) InBoundsCur() bool {
	if v.Cur.X >= 0 && v.Cur.X < v.Area.Width &&
		v.Cur.Y >= 0 && v.Cur.Y < v.Area.Height {
		return true
	}
	return false
}

func (v *View) CurLeft() {
	if !v.IsNilCur() || v.Cur.X == 0 {
		v.Cur.X--
	}

	// Past left margin, wrap to prev line if there's room.
	if v.Cur.X < 0 {
		v.Cur.X = 0
		if v.Cur.Y > 0 {
			v.Cur.Y--
			// Go to rightmost char in prev row.
			for x := v.Area.Width - 1; x >= 0; x-- {
				if v.TextBlk.Text[v.Cur.Y][x] != 0 {
					v.Cur.X = x
					break
				}
			}
		}
	}
}
func (v *View) CurRight() {
	if !v.IsNilCur() {
		v.Cur.X++
	}

	// Past right margin, wrap to next line if there's room.
	if v.Cur.X > v.Area.Width-1 || v.IsNilCur() {
		if v.Cur.Y < v.TextBlk.Height-1 {
			v.Cur.X = 0
			v.Cur.Y++
		}
	}
}
func (v *View) CurUp() {
	if v.Cur.Y > 0 {
		v.Cur.Y--
	}

	if v.IsNilCur() {
		startX := v.Cur.X
		v.Cur.X = 0
		for x := startX; x >= 0; x-- {
			if v.TextBlk.Text[v.Cur.Y][x] != 0 {
				v.Cur.X = x
				break
			}
		}
	}
}
func (v *View) CurDown() {
	if v.Cur.Y < v.Area.Height-1 {
		v.Cur.Y++
	}

	if v.IsNilCur() {
		startX := v.Cur.X
		v.Cur.X = 0
		for x := startX; x >= 0; x-- {
			if v.TextBlk.Text[v.Cur.Y][x] != 0 {
				v.Cur.X = x
				break
			}
		}
	}
}
