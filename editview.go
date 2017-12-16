package main

import (
	"unicode"

	tb "github.com/nsf/termbox-go"
)

type EditView struct {
	Content Area
	Outline Area
	*Buf
	*TextBlk
	fOutline bool
}

func NewEditView(x, y, w, h int, fOutline bool, buf *Buf) *EditView {
	outline := NewArea(x, y, w, h)
	content := outline
	if fOutline {
		content = NewArea(x+1, y+1, w-2, h-2)
	}
	textBlk := NewTextBlk(content.Width, 0)
	if buf == nil {
		buf = NewBuf()
		buf.WriteLine("")
	}

	v := &EditView{
		Content:  content,
		Outline:  outline,
		Buf:      buf,
		TextBlk:  textBlk,
		fOutline: fOutline,
	}
	v.SyncText()
	return v
}

func (v *EditView) Pos() Pos {
	return Pos{v.Outline.X, v.Outline.Y}
}
func (v *EditView) Size() Size {
	return Size{v.Outline.Width, v.Outline.Height}
}
func (v *EditView) Text() string {
	return v.Buf.Text()
}
func min(n1, n2 int) int {
	if n1 < n2 {
		return n1
	}
	return n2
}

func (v *EditView) Draw() {
	if v.fOutline {
		drawBox(v.Outline.X, v.Outline.Y, v.Outline.Width, v.Outline.Height, 0, 0)
	}
	v.drawText()
}

func (v *EditView) drawText() {
	text := v.TextBlk.Text
	for y := 0; y < min(v.TextBlk.Height, v.Content.Height); y++ {
		for x := 0; x < min(v.TextBlk.Width, v.Content.Width); x++ {
			c := text[y][x]
			if c == 0 {
				print(string(" "), x+v.Content.X, y+v.Content.Y, 0, 0)
				continue
			}

			print(string(c), x+v.Content.X, y+v.Content.Y, 0, 0)
		}
	}
}

func (v *EditView) DrawCursor() {
	tb.SetCursor(v.Content.X+v.Cur.X, v.Content.Y+v.Cur.Y)
}

func (v *EditView) Clear() {
	v.Buf.Clear()
	v.SyncText()
	v.Cur = Pos{0, 0}
}
func (v *EditView) SyncText() {
	v.TextBlk.FillWithBuf(v.Buf)
}
func (v *EditView) UpdateCursorBufPos(bufPos Pos) {
	v.Cur = v.BlkFromBuf[bufPos]
}

func (v *EditView) SetText(s string) {
	v.Buf.SetText(s)
	v.SyncText()
	v.Cur = Pos{0, 0}
}

func (v *EditView) GetText() string {
	return v.Buf.GetText()
}

func (v *EditView) HandleEvent(e *tb.Event) {
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
			v.SyncText()
			v.UpdateCursorBufPos(bufPos)
		case tb.KeyEnter:
			bufPos := v.BufPos()
			bufPos = v.Buf.InsEOL(bufPos.X, bufPos.Y)
			v.SyncText()
			v.UpdateCursorBufPos(bufPos)
		case tb.KeyDelete:
			bufPos := v.BufPos()
			v.Buf.DelChar(bufPos.X, bufPos.Y)
			v.SyncText()
			v.UpdateCursorBufPos(bufPos)
		case tb.KeyBackspace:
			fallthrough
		case tb.KeyBackspace2:
			bufPos := v.BufPos()
			bufPos = v.Buf.DelPrevChar(bufPos.X, bufPos.Y)
			v.SyncText()
			v.UpdateCursorBufPos(bufPos)
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
		v.SyncText()
		v.UpdateCursorBufPos(bufPos)
	}
}

func (v *EditView) InBoundsCur() bool {
	if v.Cur.Y < 0 || v.Cur.Y > len(v.TextBlk.Text)-1 {
		return false
	}
	if v.Cur.X < 0 || v.Cur.X > len(v.TextBlk.Text[v.Cur.Y])-1 {
		return false
	}
	return true
}

// Return char under the cursor or 0 if out of bounds.
func (v *EditView) CurChar() rune {
	if v.InBoundsCur() {
		return v.TextBlk.Text[v.Cur.Y][v.Cur.X]
	}
	return 0
}
func (v *EditView) IsNilCur() bool {
	return v.CurChar() == 0
}
func (v *EditView) IsNilLeftCur() bool {
	if v.Cur.X == 0 {
		return true
	}
	if v.TextBlk.Text[v.Cur.Y][v.Cur.X-1] == 0 {
		return true
	}
	return false
}
func (v *EditView) IsBOFCur() bool {
	if v.Cur.Y == 0 && v.Cur.X == 0 {
		return true
	}
	return false
}
func (v *EditView) IsEOFCur() bool {
	if v.Cur.Y >= v.TextBlk.Height-1 && v.Cur.X >= v.TextBlk.Width-1 {
		return true
	}
	return false
}

func (v *EditView) CurBOL() {
	v.Cur.X = 0
}
func (v *EditView) CurEOL() {
	for v.Cur.X < v.TextBlk.Width && !v.IsNilCur() {
		v.Cur.X++
	}
	v.Cur.X--
}
func (v *EditView) CurLeft() {
	if !v.IsBOFCur() {
		v.Cur.X--
	}

	// Past left margin, wrap to prev line if there's room.
	if v.Cur.X < 0 {
		v.Cur.X = 0
		if v.Cur.Y > 0 {
			v.Cur.Y--
			// Go to rightmost char in prev row.
			for x := v.Content.Width - 1; x >= 0; x-- {
				if v.TextBlk.Text[v.Cur.Y][x] != 0 {
					v.Cur.X = x
					break
				}
			}
		}
	}
}
func (v *EditView) CurRight() {
	v.Cur.X++

	// Past right margin, wrap to next line if there's room.
	//if v.Cur.X > v.Content.Width-1 || (v.IsNilCur() && v.IsNilLeftCur()) {
	if v.Cur.X > v.TextBlk.RowWidth-1 || (v.IsNilCur() && v.IsNilLeftCur()) {
		if v.Cur.Y < len(v.TextBlk.Text)-1 {
			v.Cur.X = 0
			v.Cur.Y++
		} else {
			v.Cur.X--
		}
	}
}
func (v *EditView) CurUp() {
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
func (v *EditView) CurDown() {
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
func (v *EditView) CurRightN(n int) {
	for i := 0; i < n; i++ {
		v.CurRight()
	}
}
func (v *EditView) CurDownN(n int) {
	for i := 0; i < n; i++ {
		v.CurDown()
	}
}
func (v *EditView) CurWordNext() {
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
func (v *EditView) CurWordBack() {
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
