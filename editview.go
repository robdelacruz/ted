package main

// Structs
// -------
// EditView
//
// Consts
// ------
// EditViewBorder
// EditViewStatusLine
//
// Functions
// ---------
//
// EditView
// --------
// NewEditView(x, y, w, h int, mode EditViewMode,
//				contentAttr, statusAttr TermAttr, buf *Buf) *EditView
// contentRect() Rect
// contentRange() (startPos, endPos Pos)
//
// Reset()
// ResetCur()
//
// Draw()
// drawText(rect Rect)
// drawStatus(rect Rect)
// drawCur(rect Rect)
//
// SetText(s string)
// Text() string
//
// HandleEvent(e *tb.Event) (Widget, WidgetEventID)
// navCurChar(chFn func() bool)
// navCurWrapline(wlFn func() bool)
// navStartWrapline()
// navEndWrapline()
// navStartWord()
// navEndWord()
// fitViewTopToCur()
// ScrollN(beforePos Pos, nWraplines int) (afterPos Pos)
//

import (
	"fmt"
	"strings"

	tb "github.com/nsf/termbox-go"
)

const _tablen = 4

type EditView struct {
	Rect
	*Buf
	Mode        EditViewMode
	ContentAttr TermAttr
	StatusAttr  TermAttr
	Cur         Pos
	bitCur      *BufIterCh
	bitWl       *BufIterWl
	ViewTop     Pos
}

type EditViewMode uint

const (
	EditViewBorder EditViewMode = 1 << iota
	EditViewStatusLine
)

func NewEditView(x, y, w, h int, mode EditViewMode, contentAttr, statusAttr TermAttr, buf *Buf) *EditView {
	v := &EditView{}
	v.Rect = NewRect(x, y, w, h)
	v.Mode = mode
	v.ContentAttr = contentAttr
	v.StatusAttr = statusAttr

	if buf == nil {
		buf = NewBuf()
		buf.AppendLine("")
	}
	v.Buf = buf
	v.bitCur = NewBufIterCh(v.Buf)
	v.bitWl = NewBufIterWl(v.Buf, v.contentRect().W)

	return v
}

func (v *EditView) contentRect() Rect {
	rect := v.Rect
	if v.Mode&EditViewBorder != 0 {
		rect.X++
		rect.Y++
		rect.W -= 2
		rect.H -= 2
	}
	if v.Mode&EditViewStatusLine != 0 {
		rect.H--
	}
	return rect
}

func (v *EditView) contentRange() (startPos, endPos Pos) {
	v.bitWl.Seek(v.ViewTop)
	startPos = v.bitWl.Pos()

	contentRect := v.contentRect()
	for i := 1; i < contentRect.H; i++ {
		if !v.bitWl.ScanNext() {
			break
		}
	}
	endPos = v.bitWl.Pos()
	endPos.X += rlen(v.bitWl.Text()) - 1

	return startPos, endPos
}

func (v *EditView) Reset() {
	v.Cur = Pos{0, 0}
	v.bitCur.Reset()
	v.bitWl.Reset()
	v.fitViewTopToCur()
}

func (v *EditView) ResetCur() {
	v.Cur = Pos{0, 0}
	v.bitCur.Seek(v.Cur)
}

func (v *EditView) Draw() {
	clearRect(v.Rect, v.ContentAttr)
	if v.Mode&EditViewBorder != 0 {
		boxAttr := v.ContentAttr
		drawBox(v.Rect.X, v.Rect.Y, v.Rect.W, v.Rect.H, boxAttr)
	}

	contentRect := v.contentRect()
	v.drawText(contentRect)

	if v.Mode&EditViewStatusLine != 0 {
		v.drawStatus(contentRect)
	}

	v.drawCur(contentRect)
}

func (v *EditView) drawText(rect Rect) {
	v.bitWl.Seek(v.ViewTop)
	i := 0

	// First wrapline.
	sline := v.bitWl.Text()
	print(sline, rect.X, rect.Y+i, v.ContentAttr)
	i++

	// Succeeding wraplines until bottommost content row.
	for v.bitWl.ScanNext() {
		if i > rect.H-1 {
			break
		}
		sline := v.bitWl.Text()
		print(sline, rect.X, rect.Y+i, v.ContentAttr)
		i++
	}
}

// Draw status line one row below content area.
func (v *EditView) drawStatus(rect Rect) {
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

	// Cur pos y,x
	sCurPos := fmt.Sprintf("%d,%d", v.Cur.Y+1, v.Cur.X+1)
	print(sCurPos, left+width-(width/3), y, v.StatusAttr)
}

func (v *EditView) drawCur(rect Rect) {
	var contentCurPos Pos

	v.bitWl.Seek(v.ViewTop)
	wliViewTop := v.bitWl.WrapLineIndex()

	v.bitWl.Seek(v.Cur)
	wliCur := v.bitWl.WrapLineIndex()

	if wliCur >= wliViewTop {
		contentCurPos.Y = wliCur - wliViewTop
		contentCurPos.X = v.Cur.X - v.bitWl.Pos().X
	}

	if contentCurPos.Y < rect.H && contentCurPos.X < rect.W {
		tb.SetCursor(rect.X+contentCurPos.X, rect.Y+contentCurPos.Y)
	}
}

func (v *EditView) SetText(s string) {
	v.Buf.SetText(s)
	v.Reset()
}

func (v *EditView) Text() string {
	return v.Buf.Text()
}

func (v *EditView) HandleEvent(e *tb.Event) (Widget, WidgetEventID) {
	var bufChanged bool
	var c rune

	switch e.Key {
	case tb.KeyEsc:

	// Nav single char
	case tb.KeyArrowLeft:
		v.navCurChar(v.bitCur.ScanPrev)
	case tb.KeyArrowRight:
		v.navCurChar(v.bitCur.ScanNext)
	case tb.KeyArrowUp:
		v.navCurWrapline(v.bitWl.ScanPrev)
	case tb.KeyArrowDown:
		v.navCurWrapline(v.bitWl.ScanNext)

	// Nav word/line
	case tb.KeyCtrlP:
		fallthrough
	case tb.KeyCtrlB:
		v.navStartWord()
	case tb.KeyCtrlN:
		fallthrough
	case tb.KeyCtrlF:
		v.navEndWord()
	case tb.KeyCtrlA:
		v.navStartWrapline()
	case tb.KeyCtrlE:
		v.navEndWrapline()

	// Scroll text
	case tb.KeyCtrlU:
		v.ViewTop = v.ScrollN(v.ViewTop, 0-v.contentRect().H/2)
		v.Cur = v.ScrollN(v.Cur, 0-v.contentRect().H/2)
	case tb.KeyCtrlD:
		v.ViewTop = v.ScrollN(v.ViewTop, v.contentRect().H/2)
		v.Cur = v.ScrollN(v.Cur, v.contentRect().H/2)

	// Select/copy/paste text
	case tb.KeyCtrlK:
	case tb.KeyCtrlC:
		//$$ copy selected text
	case tb.KeyCtrlV:
		//$$ paste into
	case tb.KeyCtrlX:
		//$$ cut selected text

	// Delete text
	case tb.KeyDelete:
		v.Cur = v.Buf.DelChar(v.Cur)
		bufChanged = true
	case tb.KeyBackspace:
		fallthrough
	case tb.KeyBackspace2:
		if v.bitCur.Pos() != v.Cur {
			v.bitCur.Seek(v.Cur)
		}
		if v.bitCur.ScanPrev() {
			v.Cur = v.bitCur.Pos()
			v.Cur = v.Buf.DelChar(v.Cur)
		}
		bufChanged = true

	// Text entry
	case tb.KeyEnter:
		v.Cur = v.Buf.InsLF(v.Cur)
		bufChanged = true
	case tb.KeyTab:
		c = '\t'
	case tb.KeySpace:
		c = ' '
	case 0:
		c = e.Ch
	}

	// Char entered
	if c != 0 {
		v.Cur = v.Buf.InsChar(v.Cur, c)
		bufChanged = true
	}

	if bufChanged {
		v.bitCur.Reset()
		v.bitWl.Reset()
		v.fitViewTopToCur()
	}

	return v, WidgetEventNone
}

func (v *EditView) navCurChar(chFn func() bool) {
	if v.bitCur.Pos() != v.Cur {
		v.bitCur.Seek(v.Cur)
	}
	if chFn() {
		v.Cur = v.bitCur.Pos()

		v.fitViewTopToCur()
	}
}

func (v *EditView) navCurWrapline(wlFn func() bool) {
	if v.bitWl.Seek(v.Cur) {
		curWraplineCol := v.Cur.X - v.bitWl.Pos().X
		if wlFn() {
			v.Cur.Y = v.bitWl.Pos().Y
			v.Cur.X = v.bitWl.Pos().X + curWraplineCol
			eolX := v.bitWl.Pos().X + rlen(v.bitWl.Text()) - 1
			if v.Cur.X > eolX {
				v.Cur.X = eolX
			}

			v.fitViewTopToCur()
		}
	}
}

func (v *EditView) navStartWrapline() {
	if v.bitWl.Seek(v.Cur) {
		v.Cur = v.bitWl.Pos()
	}
}
func (v *EditView) navEndWrapline() {
	if v.bitWl.Seek(v.Cur) {
		endWlPos := v.bitWl.Pos()
		endWlPos.X += rlen(v.bitWl.Text()) - 1
		v.Cur = endWlPos
	}
}
func (v *EditView) navStartWord() {
	if v.bitCur.Pos() != v.Cur {
		v.bitCur.Seek(v.Cur)
	}
}
func (v *EditView) navEndWord() {
	if v.bitCur.Pos() != v.Cur {
		v.bitCur.Seek(v.Cur)
	}
}

func (v *EditView) fitViewTopToCur() {
	contentStartPos, contentEndPos := v.contentRange()
	cmpCurRange := cmpPosRange(v.Cur, contentStartPos, contentEndPos)
	if cmpCurRange < 0 {
		v.ViewTop = v.ScrollN(v.ViewTop, -1)
	} else if cmpCurRange > 0 {
		v.ViewTop = v.ScrollN(v.ViewTop, 1)
	}
}

func (v *EditView) ScrollN(beforePos Pos, nWraplines int) (afterPos Pos) {
	if nWraplines == 0 {
		return
	}

	// negative nWraplines means ScanPrev()
	// positive nWrapLines means ScanNext()
	scanfn := v.bitWl.ScanNext
	if nWraplines < 0 {
		scanfn = v.bitWl.ScanPrev
		nWraplines = -nWraplines
	}

	v.bitWl.Seek(beforePos)
	for i := 0; i < nWraplines; i++ {
		scanfn()
	}
	return v.bitWl.Pos()
}
