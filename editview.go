package main

import (
	"fmt"
	"strings"
	"unicode"

	tb "github.com/nsf/termbox-go"
)

type EditView struct {
	Content Area
	Outline Area
	*Buf
	*TextBlk
	Mode        EditViewMode
	ContentAttr TermAttr
	StatusAttr  TermAttr
}

type EditViewMode uint

const (
	EditViewBorder = 1 << iota
	EditViewStatusLine
)

func NewEditView(x, y, w, h int, mode EditViewMode, contentAttr, statusAttr TermAttr, buf *Buf) *EditView {
	outline, content := getAreas(x, y, w, h, mode)
	textBlk := NewTextBlk(content.Width, 0)
	if buf == nil {
		buf = NewBuf()
		buf.WriteLine("")
	}

	v := &EditView{}
	v.Content = content
	v.Outline = outline
	v.Mode = mode
	v.ContentAttr = contentAttr
	v.StatusAttr = statusAttr
	v.Buf = buf
	v.TextBlk = textBlk

	v.SyncText()
	return v
}

func getAreas(x, y, w, h int, mode EditViewMode) (outline, content Area) {
	outline = NewArea(x, y, w, h)
	content = outline
	if mode&EditViewBorder != 0 {
		content = NewArea(x+1, y+1, w-2, h-2)
	}
	if mode&EditViewStatusLine != 0 {
		content.Height--
	}

	return outline, content
}

func (v *EditView) Resize(x, y, w, h int) {
	outline, content := getAreas(x, y, w, h, v.Mode)
	v.Content = content
	v.Outline = outline

	v.TextBlk.Resize(content.Width, content.Height)
	v.SyncText()
}

func (v *EditView) Pos() Pos {
	return Pos{v.Outline.X, v.Outline.Y}
}
func (v *EditView) Size() Size {
	return Size{v.Outline.Width, v.Outline.Height}
}
func (v *EditView) Area() Area {
	return NewArea(v.Outline.X, v.Outline.Y, v.Outline.Width, v.Outline.Height)
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
	if v.Mode&EditViewBorder != 0 {
		boxAttr := v.ContentAttr
		drawBox(v.Outline.X, v.Outline.Y, v.Outline.Width, v.Outline.Height, boxAttr)
	}
	v.drawText()

	if v.Mode&EditViewStatusLine != 0 {
		v.drawStatus()
	}
}

func (v *EditView) drawText() {
	v.TextBlk.PrintToArea(v.Content, v.ContentAttr)
}

// Draw status line one row below content area.
func (v *EditView) drawStatus() {
	left := v.Content.X
	width := v.Content.Width
	y := v.Content.Y + v.Content.Height

	// Clear status line first
	clearStr := strings.Repeat(" ", width)
	print(clearStr, left, y, v.StatusAttr)

	// Buf name
	bufName := v.Buf.Name
	if bufName == "" {
		bufName = "(new)"
	}
	if v.Buf.Dirty {
		bufName += " [+]"
	}
	print(bufName, left, y, v.StatusAttr)

	// Buf pos (x,y)
	bufPos := v.BufPos()
	sBufPos := fmt.Sprintf("%d,%d", bufPos.Y+1, bufPos.X+1)
	print(sBufPos, left+width-(width/3), y, v.StatusAttr)
}

func (v *EditView) DrawCursor() {
	//tb.SetCursor(v.Content.X+v.Cur.X, v.Content.Y+v.Cur.Y)

	x := v.Cur.X
	y := v.Cur.Y - v.TextBlk.BlkYOffset
	tb.SetCursor(v.Content.X+x, v.Content.Y+y)
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
			v.SyncText()
		case tb.KeyArrowRight:
			v.CurRight()
			v.SyncText()
		case tb.KeyArrowUp:
			v.CurUp()
			v.SyncText()
		case tb.KeyArrowDown:
			v.CurDown()
			v.SyncText()
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
		case tb.KeyCtrlU:
			v.TextBlk.BlkYOffset -= v.Content.Height / 2
			if v.TextBlk.BlkYOffset < 0 {
				v.TextBlk.BlkYOffset = 0
			}
			v.Cur.Y -= v.Content.Height / 2
			v.KeepBoundsCur()
		case tb.KeyCtrlD:
			v.TextBlk.BlkYOffset += v.Content.Height / 2
			if v.TextBlk.BlkYOffset > len(v.TextBlk.Text)-1 {
				v.TextBlk.BlkYOffset = len(v.TextBlk.Text) - 1
			}
			v.Cur.Y += v.Content.Height / 2
			v.KeepBoundsCur()
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

// Move cursor back in bounds if it's outside.
func (v *EditView) KeepBoundsCur() {
	if v.Cur.Y < 0 {
		v.Cur.Y = 0
	}
	if v.Cur.X < 0 {
		v.Cur.X = 0
	}
	if v.Cur.X > v.TextBlk.Width-1 {
		v.Cur.X = v.TextBlk.Width - 1
	}
	if v.Cur.Y > v.TextBlk.Height-1 {
		v.Cur.Y = v.TextBlk.Height - 1
	}
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

func (v *EditView) AdjustScrollOffset() {
	if v.Cur.Y-v.TextBlk.BlkYOffset < 0 {
		v.TextBlk.BlkYOffset--
	}
	if v.Cur.Y-v.TextBlk.BlkYOffset > v.Content.Height-1 {
		v.TextBlk.BlkYOffset++
	}
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

	v.AdjustScrollOffset()
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

	v.AdjustScrollOffset()
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

	v.AdjustScrollOffset()
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

	v.AdjustScrollOffset()
}
func (v *EditView) CurRightN(n int) {
	for i := 0; i < n; i++ {
		v.CurRight()
	}

	v.AdjustScrollOffset()
}
func (v *EditView) CurDownN(n int) {
	for i := 0; i < n; i++ {
		v.CurDown()
	}

	v.AdjustScrollOffset()
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

	v.AdjustScrollOffset()
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

	v.AdjustScrollOffset()
}
