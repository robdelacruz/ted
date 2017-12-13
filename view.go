package main

import (
	"unicode"

	tb "github.com/nsf/termbox-go"
)

type View struct {
	Area
	Border Area
	*Buf
	*TextBlk
}

func NewView(x, y, w, h int, buf *Buf) *View {
	border := NewArea(x, y, w, h)
	area := NewArea(x+1, y+1, w-2, h-2)
	textBlk := NewTextBlk(area.Width, 0)

	view := &View{
		Area:    area,
		Border:  border,
		Buf:     buf,
		TextBlk: textBlk,
	}
	return view
}

func min(n1, n2 int) int {
	if n1 < n2 {
		return n1
	}
	return n2
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

func (v *View) Draw() {
	drawBox(v.Border.X, v.Border.Y, v.Border.Width, v.Border.Height, 0, 0)
	v.TextBlk.FillWithBuf(v.Buf)
	v.drawText()
}

func (v *View) DrawCursor() {
	tb.SetCursor(v.Area.X+v.Cur.X, v.Area.Y+v.Cur.Y)
}

func (v *View) DrawCursorBufPos(bufPos Pos) {
	v.Cur = v.BlkFromBuf[bufPos]
	v.DrawCursor()
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
		case tb.KeyCtrlN:
			fallthrough
		case tb.KeyCtrlF:
			v.CurWordNext()
		case tb.KeyCtrlP:
			fallthrough
		case tb.KeyCtrlB:
			v.CurWordBack()
		case tb.KeyCtrlA:
			v.CurBOL()
		case tb.KeyCtrlE:
			v.CurEOL()
		case tb.KeyCtrlV:
			s := "12345\n678\n90\n"
			bufPos := v.BufPos()
			bufPos = v.Buf.InsText(s, bufPos.X, bufPos.Y)
			v.Draw()
			v.DrawCursorBufPos(bufPos)
			return
		case tb.KeyEnter:
			bufPos := v.BufPos()
			bufPos = v.Buf.InsEOL(bufPos.X, bufPos.Y)
			v.Draw()
			v.DrawCursorBufPos(bufPos)
			return
		case tb.KeyDelete:
			bufPos := v.BufPos()
			v.Buf.DelChar(bufPos.X, bufPos.Y)
			v.Draw()
			v.DrawCursorBufPos(bufPos)
			return
		case tb.KeyBackspace:
			fallthrough
		case tb.KeyBackspace2:
			prevbufPos := v.BufPos()
			v.CurLeft()
			bufPos := v.BufPos()
			if bufPos.X != prevbufPos.X || bufPos.Y != prevbufPos.Y {
				v.Buf.DelChar(bufPos.X, bufPos.Y)
				v.Draw()
				v.DrawCursorBufPos(bufPos)
				return
			}
		case tb.KeySpace:
			c = ' '
		case 0:
			c = e.Ch
		}
	}

	// Char entered
	if c != 0 {
		bufPos := v.BufPos()
		bufPos = v.Buf.InsChar(c, bufPos.X, bufPos.Y)
		v.Draw()
		v.DrawCursorBufPos(bufPos)
		return
	}

	v.Draw()
	v.DrawCursor()
}

func (v *View) InBoundsCur() bool {
	if v.Cur.Y < 0 || v.Cur.Y > len(v.TextBlk.Text)-1 {
		return false
	}
	if v.Cur.X < 0 || v.Cur.X > len(v.TextBlk.Text[v.Cur.Y])-1 {
		return false
	}
	return true
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
func (v *View) IsNilLeftCur() bool {
	if v.Cur.X == 0 {
		return true
	}
	if v.TextBlk.Text[v.Cur.Y][v.Cur.X-1] == 0 {
		return true
	}
	return false
}
func (v *View) IsBOFCur() bool {
	if v.Cur.Y == 0 && v.Cur.X == 0 {
		return true
	}
	return false
}
func (v *View) IsEOFCur() bool {
	if v.Cur.Y >= v.TextBlk.Height-1 && v.Cur.X >= v.TextBlk.Width-1 {
		return true
	}
	return false
}

func (v *View) CurBOL() {
	v.Cur.X = 0
}
func (v *View) CurEOL() {
	for v.Cur.X < v.TextBlk.Width && !v.IsNilCur() {
		v.Cur.X++
	}
	v.Cur.X--
}
func (v *View) CurLeft() {
	if !v.IsBOFCur() {
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
	v.Cur.X++

	// Past right margin, wrap to next line if there's room.
	if v.Cur.X > v.Area.Width-1 || (v.IsNilCur() && v.IsNilLeftCur()) {
		if v.Cur.Y < len(v.TextBlk.Text)-1 {
			v.Cur.X = 0
			v.Cur.Y++
		} else {
			v.Cur.X--
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
	if v.Cur.Y < len(v.TextBlk.Text)-1 {
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
func (v *View) CurRightN(n int) {
	for i := 0; i < n; i++ {
		v.CurRight()
	}
}
func (v *View) CurDownN(n int) {
	for i := 0; i < n; i++ {
		v.CurDown()
	}
}
func (v *View) CurWordNext() {
	if v.IsNilCur() {
		v.CurRight()
	}

	// Skip to first space.
	for !unicode.IsSpace(v.CurChar()) && !v.IsNilCur() && !v.IsEOFCur() {
		v.CurRight()
	}

	// Skip spaces to first letter.
	for unicode.IsSpace(v.CurChar()) && !v.IsNilCur() && !v.IsEOFCur() {
		v.CurRight()
	}
}
func (v *View) CurWordBack() {
	if v.IsNilCur() {
		v.CurLeft()
	}

	for !unicode.IsSpace(v.CurChar()) && !v.IsNilCur() && !v.IsBOFCur() {
		v.CurLeft()
	}

	for unicode.IsSpace(v.CurChar()) && !v.IsNilCur() && !v.IsBOFCur() {
		v.CurLeft()
	}
}
