package main

import (
	"fmt"
	"strings"
	"unicode"

	tb "github.com/nsf/termbox-go"
)

type EditView struct {
	Rect
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
	v := &EditView{}
	v.Rect = NewRect(x, y, w, h)
	v.Mode = mode
	v.ContentAttr = contentAttr
	v.StatusAttr = statusAttr

	if mode&EditViewBorder != 0 {
		w -= 2
	}
	textBlk := NewTextBlk(w, 0)
	if buf == nil {
		buf = NewBuf()
		buf.SetText("")
	}

	v.Buf = buf
	v.TextBlk = textBlk
	v.SyncText()
	return v
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
	clearRect(v.Rect, v.ContentAttr)
	if v.Mode&EditViewBorder != 0 {
		boxAttr := v.ContentAttr
		drawBox(v.Rect.X, v.Rect.Y, v.Rect.W, v.Rect.H, boxAttr)
	}
	v.drawText()

	if v.Mode&EditViewStatusLine != 0 {
		v.drawStatus()
	}

	v.drawCursor()
}

func (v *EditView) contentRect() Rect {
	rect := v.Rect
	if v.Mode&EditViewBorder != 0 {
		rect.X++
		rect.Y++
		rect.W -= 2
		rect.H -= 2
	}
	return rect
}

func (v *EditView) drawText() {
	v.TextBlk.PrintToArea(v.contentRect(), v.ContentAttr)
}

// Draw status line one row below content area.
func (v *EditView) drawStatus() {
	rect := v.contentRect()
	left := rect.X
	width := rect.W
	y := rect.Y + rect.H

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

	// Pos in doc (%)
	var sYDist string
	nBufLines := len(v.Buf.Lines)
	if nBufLines == 0 {
		sYDist = "0%"
	}
	yDist := bufPos.Y * 100 / (nBufLines - 1)
	sYDist = fmt.Sprintf(" %d%% ", yDist)
	print(sYDist, left+width-len(sYDist), y, v.StatusAttr)
}

func (v *EditView) drawCursor() {
	rect := v.contentRect()

	x := v.Cur.X
	y := v.Cur.Y - v.TextBlk.BlkYOffset

	tb.SetCursor(rect.X+x, rect.Y+y)
}

func (v *EditView) Clear() {
	v.Buf.Clear()
	v.SyncText()
	v.ResetCur()
}
func (v *EditView) ResetCur() {
	v.Cur = Pos{0, 0}
	v.TextBlk.BlkYOffset = 0
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

func (v *EditView) HandleEvent(e *tb.Event) (Widget, WidgetEventID) {
	contentRect := v.contentRect()

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
		case tb.KeyCtrlU:
			v.TextBlk.BlkYOffset -= contentRect.H / 2
			if v.TextBlk.BlkYOffset < 0 {
				v.TextBlk.BlkYOffset = 0
			}
			v.Cur.Y -= contentRect.H / 2
			v.KeepBoundsCur()
		case tb.KeyCtrlD:
			v.TextBlk.BlkYOffset += contentRect.H / 2
			if v.TextBlk.BlkYOffset > len(v.TextBlk.Text)-1 {
				v.TextBlk.BlkYOffset = len(v.TextBlk.Text) - 1
			}
			v.Cur.Y += contentRect.H / 2
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

	return v, WidgetEventNone
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
	if v.Cur.X > v.TextBlk.W-1 {
		v.Cur.X = v.TextBlk.W - 1
	}
	if v.Cur.Y > v.TextBlk.H-1 {
		v.Cur.Y = v.TextBlk.H - 1
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
	if v.Cur.Y >= v.TextBlk.H-1 && v.Cur.X >= v.TextBlk.W-1 {
		return true
	}
	return false
}

func (v *EditView) AdjustScrollOffset() {
	if v.Cur.Y-v.TextBlk.BlkYOffset < 0 {
		v.TextBlk.BlkYOffset--
	}
	if v.Cur.Y-v.TextBlk.BlkYOffset > v.contentRect().H-1 {
		v.TextBlk.BlkYOffset++
	}
}

func (v *EditView) CurBOL() {
	v.Cur.X = 0
}
func (v *EditView) CurEOL() {
	for v.Cur.X < v.TextBlk.W && !v.IsNilCur() {
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
			for x := v.contentRect().W - 1; x >= 0; x-- {
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
