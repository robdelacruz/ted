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
// posInRange(pos Pos, area PosRange) bool
// lineRange(bit *BufIterWl) PosRange
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
// trySetCurPos(newCurPos Pos)
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
	bitWl2      *BufIterWl
	ViewTop     Pos
	SelMode     bool
	SelRange    PosRange
	ClipText    string
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
	v.bitWl2 = NewBufIterWl(v.Buf, v.contentRect().W)

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
	v.SelMode = false
	v.SelRange = PosRange{Pos{0, 0}, Pos{0, 0}}
	v.ViewTop = Pos{0, 0}

	v.bitCur.Reset()
	v.bitWl.Reset()
	v.bitWl2.Reset()
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

	var inSelRange bool
	selRange := v.SelRange.Sorted()
	if v.SelMode {
		viewTop := v.bitWl.Pos()
		if v.SelRange.Begin.Y < viewTop.Y && v.SelRange.End.Y >= viewTop.Y {
			inSelRange = true
		}
	}

	// Draw wraplines from viewTop until bottommost content row.
	for i := 0; i < rect.H; i++ {
		sline := v.bitWl.Text()
		print(expandTabs(sline), rect.X, rect.Y+i, v.ContentAttr)

		if v.SelMode {
			inSelRange = drawSelLine(rect, selRange, v.bitWl, i, reverseAttr(v.ContentAttr), inSelRange)
		}

		if !v.bitWl.ScanNext() {
			break
		}
	}
}

func drawSelLine(rect Rect, selRange PosRange, bit *BufIterWl, yContent int, selAttr TermAttr, inSelRange bool) bool {
	slineRange := lineRange(bit)
	sline := bit.Text()

	selStartX := 0
	selEndX := rlen(sline)

	var lastSelLine bool
	if posInRange(selRange.Begin, slineRange) {
		selStartX = selRange.Begin.X - slineRange.Begin.X
		inSelRange = true
	}
	if posInRange(selRange.End, slineRange) {
		selEndX = rlen(sline) - (slineRange.End.X - selRange.End.X)
		lastSelLine = true
		inSelRange = false
	}
	if inSelRange || lastSelLine {
		sline = expandTabs(sline[selStartX:selEndX])
		selStartX = expandTabsX(selStartX, bit.Text())
		print(sline, rect.X+selStartX, rect.Y+yContent, selAttr)
	}

	return inSelRange
}

func posInRange(pos Pos, area PosRange) bool {
	begin := area.Begin
	end := area.End

	if pos.Y >= begin.Y && pos.Y <= end.Y &&
		pos.X >= begin.X && pos.X <= end.X {
		return true
	}
	return false
}

func lineRange(bit *BufIterWl) PosRange {
	begin := bit.Pos()
	end := begin
	end.X += rlen(bit.Text()) - 1

	return PosRange{begin, end}
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

	// Sel range [y,x - y,x]
	if v.SelMode {
		selRange := v.SelRange.Sorted()
		sSelRange := fmt.Sprintf("[%d,%d - %d,%d]", selRange.Begin.Y+1, selRange.Begin.X+1, selRange.End.Y+1, selRange.End.X+1)
		print(sSelRange, left+width-(width/2), y, v.StatusAttr)
	}

	// Cur pos y,x
	sCurPos := fmt.Sprintf("%d,%d", v.Cur.Y+1, v.Cur.X+1)
	print(sCurPos, left+width-(width/4), y, v.StatusAttr)

	// Scroll pos %
	sScrollPos := fmt.Sprintf("%d%%", v.Cur.Y*100/(v.Buf.NumNodes()-1))
	print(sScrollPos, left+width-4, y, v.StatusAttr)
}

func (v *EditView) drawCur(rect Rect) {
	var contentCurPos Pos

	v.bitWl.Seek(v.ViewTop)
	wliViewTop := v.bitWl.WrapLineIndex()

	v.bitWl.Seek(v.Cur)
	wliCur := v.bitWl.WrapLineIndex()

	if wliCur >= wliViewTop {
		contentCurPos.Y = wliCur - wliViewTop
		wliX := v.Cur.X - v.bitWl.Pos().X
		contentCurPos.X = expandTabsX(wliX, v.bitWl.Text())
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
	var bufChanged, navChanged bool
	var c rune

	switch e.Key {
	case tb.KeyEsc:
		v.endSel()

	// Nav single char
	case tb.KeyArrowLeft:
		v.navCurChar(v.bitCur.ScanPrev)
		navChanged = true
	case tb.KeyArrowRight:
		v.navCurChar(v.bitCur.ScanNext)
		navChanged = true
	case tb.KeyArrowUp:
		v.navCurWrapline(v.bitWl.ScanPrev)
		navChanged = true
	case tb.KeyArrowDown:
		v.navCurWrapline(v.bitWl.ScanNext)
		navChanged = true

	// Nav word/line
	case tb.KeyCtrlP:
		v.navStartWord()
		navChanged = true
	case tb.KeyCtrlN:
		v.navEndWord()
		navChanged = true
	case tb.KeyCtrlA:
		v.navStartWrapline()
		navChanged = true
	case tb.KeyCtrlE:
		v.navEndWrapline()
		navChanged = true

	// Scroll text
	case tb.KeyCtrlU:
		v.ViewTop = v.ScrollN(v.ViewTop, 0-v.contentRect().H/2)
		v.Cur = v.ScrollN(v.Cur, 0-v.contentRect().H/2)
		navChanged = true
	case tb.KeyCtrlD:
		v.ViewTop = v.ScrollN(v.ViewTop, v.contentRect().H/2)
		v.Cur = v.ScrollN(v.Cur, v.contentRect().H/2)
		navChanged = true

	// Select/copy/paste text
	case tb.KeyCtrlK:
		if v.SelMode {
			v.endSel()
		} else {
			v.startSel()
		}
	case tb.KeyCtrlC:
		if v.SelMode {
			selRange := v.SelRange.Sorted()
			v.ClipText, _ = v.Buf.Copy(selRange.Begin, selRange.End)
			v.endSel()
		}
	case tb.KeyCtrlV:
		if len(v.ClipText) > 0 {
			if v.SelMode {
				selRange := v.SelRange.Sorted()
				v.Buf.Cut(selRange.Begin, selRange.End)
				v.trySetCurPos(selRange.Begin)
			}
			v.Buf.Paste(v.Cur, v.ClipText)
			bufChanged = true
		}
		v.endSel()
	case tb.KeyCtrlX:
		if v.SelMode {
			selRange := v.SelRange.Sorted()
			v.ClipText, _ = v.Buf.Cut(selRange.Begin, selRange.End)
			v.endSel()

			v.trySetCurPos(selRange.Begin)
			bufChanged = true
		}

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

	if navChanged {
		v.updateSel()
	}

	if bufChanged {
		v.bitCur.Reset()
		v.bitWl.Reset()
		v.fitViewTopToCur()
	}

	return v, WidgetEventNone
}

func (v *EditView) startSel() {
	v.SelMode = true
	v.SelRange = PosRange{v.Cur, v.Cur}
}
func (v *EditView) endSel() {
	v.SelMode = false
	v.SelRange = PosRange{Pos{0, 0}, Pos{0, 0}}
}
func (v *EditView) updateSel() {
	if v.SelMode {
		v.SelRange.End = v.Cur
	}
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

/*
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
*/
func (v *EditView) navCurWrapline(wlFn func() bool) {
	if v.bitWl.Seek(v.Cur) {
		curWraplineCol := v.Cur.X - v.bitWl.Pos().X
		curWraplineCol = expandTabsX(curWraplineCol, v.bitWl.Text())
		if wlFn() {
			v.Cur.Y = v.bitWl.Pos().Y
			v.Cur.X = v.bitWl.Pos().X + curWraplineCol

			maxWraplineCol := expandTabsX(rlen(v.bitWl.Text())-1, v.bitWl.Text())
			if curWraplineCol > maxWraplineCol {
				curWraplineCol = maxWraplineCol
			}
			curWraplineCol = unexpandTabsX(curWraplineCol, v.bitWl.Text())
			v.Cur.X = v.bitWl.Pos().X + curWraplineCol

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

func (v *EditView) trySetCurPos(newCurPos Pos) {
	v.Cur = newCurPos
	if !v.Buf.InBounds(v.Cur) {
		// If cur out of bounds after cut, set cursor to nearest
		// position before the out of bounds position.
		v.bitCur.Seek(v.Cur)
		v.Cur = v.bitCur.Pos()
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

func (v *EditView) SearchForward(s string) {
	if v.bitCur.Seek(v.Cur) && v.bitCur.ScanNext() {
		curNextPos := v.bitCur.Pos()
		foundPos, found := v.Buf.Search(curNextPos, s)
		if found {
			v.Cur = foundPos
		}
	}
}
