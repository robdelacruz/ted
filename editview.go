package main

import (
	"fmt"
	"strings"

	tb "github.com/nsf/termbox-go"
)

type EditView struct {
	Rect
	*Buf
	Ts          *TextSurface
	Mode        EditViewMode
	ContentAttr TermAttr
	StatusAttr  TermAttr
	BufPos      Pos
	ScrollPos   Pos

	SelMode  bool
	SelBegin Pos
	SelEnd   Pos
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

	contentRect := v.contentRect()
	v.Ts = NewTextSurface(contentRect.W, contentRect.H)

	if buf == nil {
		buf = NewBuf()
		buf.SetText("")
	}
	v.Buf = buf
	v.SyncBufText()

	return v
}

func (v *EditView) SyncBufText() {
	v.syncWithBuf(v.Ts)
}
func (v *EditView) convBufPos(bufPos ...Pos) []*Pos {
	return v.syncWithBuf(nil, bufPos...)
}
func (v *EditView) tsPosFromBufPos(bufPos Pos) Pos {
	retTsPos := v.convBufPos(bufPos)
	pTsPos := retTsPos[0]
	if pTsPos == nil {
		return Pos{0, 0}
	}
	return *pTsPos
}

// Remove any trailing '\n'
func chomp(line string) string {
	nline := len(line)
	if nline > 0 && line[nline-1] == '\n' {
		return line[:nline-1]
	}
	return line
}

func (v *EditView) syncWithBuf(pTs *TextSurface, bufPosItems ...Pos) []*Pos {
	yTs := 0
	xBuf, yBuf := 0, 0

	maxlenWrapline := v.Ts.W
	if pTs != nil {
		maxlenWrapline = pTs.W
	}

	retTsPos := make([]*Pos, len(bufPosItems))

	if pTs != nil {
		pTs.Clear(0)
	}

	cbWord := func(w string) {
	}

	cbWrapLine := func(wrapline string) {
		// Write new wrapline to display.
		if pTs != nil {
			pTs.WriteString(chomp(expandTabs(wrapline, _tablen)), 0, yTs)
		}

		lenWrapline := len([]rune(wrapline))

		for i, bufPos := range bufPosItems {
			if retTsPos[i] != nil {
				continue
			}

			// Update ts pos if bufpos in this wrapline.
			if bufPos.Y == yBuf {
				//if bufPos.X >= xBuf && bufPos.X <= (xBuf+lenWrapline) {
				if bufPos.X >= xBuf && bufPos.X < (xBuf+lenWrapline) {
					x := v.Buf.Distance(yBuf, xBuf, bufPos.X)
					y := yTs
					retTsPos[i] = &Pos{x, y}
				}

				if bufPos.X == 0 && xBuf == 0 && lenWrapline == 0 {
					retTsPos[i] = &Pos{0, yTs}
				}
			}
		}

		yTs++
		xBuf += lenWrapline
	}

	for yBuf < len(v.Buf.Lines) {
		bufLine := v.Buf.Lines[yBuf]

		processLine(bufLine, maxlenWrapline, cbWord, cbWrapLine)

		yBuf++
		xBuf = 0
	}
	v.Ts.ResizeLines(yTs)

	return retTsPos
}

func (v *EditView) UpPos(bufPos Pos) Pos {
	_, nline := v.Buf.PosLine(bufPos.Y)
	if nline == 0 {
		if bufPos.Y > 0 {
			bufPos.Y--
			_, nline := v.Buf.PosLine(bufPos.Y)
			if nline == 0 {
				bufPos.X = 0
			} else if bufPos.X > nline-1 {
				bufPos.X = nline - 1
			}
		}
		return bufPos
	}

	// Get Ts pos of bufPos
	// Ts pos will be used to nav up
	retTsPos := v.syncWithBuf(v.Ts, bufPos)
	pTsPos := retTsPos[0]
	if pTsPos == nil {
		return bufPos
	}
	tsPos := *pTsPos

	// If no Ts row above
	if tsPos.Y == 0 {
		return bufPos
	}

	tsUpPos := Pos{tsPos.X, tsPos.Y - 1}

	tsChs := v.Ts.RangeChars(tsUpPos, tsPos)
	for range tsChs {
		bufPos = v.Buf.PrevPos(bufPos)
	}

	return bufPos
}

func (v *EditView) DownPos(bufPos Pos) Pos {
	_, nline := v.Buf.PosLine(bufPos.Y)
	if nline <= v.Ts.W {
		if bufPos.Y < v.Buf.NumLines()-1 {
			bufPos.Y++
			_, nline := v.Buf.PosLine(bufPos.Y)
			if nline == 0 {
				bufPos.X = 0
			} else if bufPos.X > nline-1 {
				bufPos.X = nline - 1
			}
		}
		return bufPos
	}

	// Get Ts pos of bufPos
	// Ts pos will be used to nav down
	retTsPos := v.syncWithBuf(v.Ts, bufPos)
	pTsPos := retTsPos[0]
	if pTsPos == nil {
		return bufPos
	}
	tsPos := *pTsPos

	// If no Ts row below
	if tsPos.Y+1 > v.Ts.H-1 {
		return bufPos
	}

	tsDownPos := Pos{tsPos.X, tsPos.Y + 1}

	tsChs := v.Ts.RangeChars(tsPos, tsDownPos)
	for range tsChs {
		bufPos = v.Buf.NextPosBounds(bufPos)
	}

	return bufPos
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

func (v *EditView) Draw() {
	clearRect(v.Rect, v.ContentAttr)
	if v.Mode&EditViewBorder != 0 {
		boxAttr := v.ContentAttr
		drawBox(v.Rect.X, v.Rect.Y, v.Rect.W, v.Rect.H, boxAttr)
	}

	bufSelBegin, bufSelEnd := v.OrderedSelPos(v.SelBegin, v.SelEnd)
	retTsPos := v.convBufPos(v.BufPos, bufSelBegin, bufSelEnd)
	pTsPos := retTsPos[0]
	pSelTsBeginPos := retTsPos[1]
	pSelTsEndPos := retTsPos[2]

	if pSelTsBeginPos == nil {
		pSelTsBeginPos = &Pos{0, 0}
	}
	if pSelTsEndPos == nil {
		pSelTsEndPos = &Pos{v.Ts.H - 1, v.Ts.W - 1}
	}

	v.drawText(*pSelTsBeginPos, *pSelTsEndPos)

	if v.Mode&EditViewStatusLine != 0 {
		v.drawStatus()
	}

	if pTsPos != nil {
		v.drawCursor(*pTsPos)
	}
}

func (v *EditView) drawText(selTsBeginPos, selTsEndPos Pos) {
	rect := v.contentRect()

	for yTs := v.ScrollPos.Y; yTs < rect.H && yTs < v.Ts.H; yTs++ {
		for xTs := 0; xTs < rect.W; xTs++ {
			c := v.Ts.Char(xTs, yTs)
			printCh(c, rect.X+xTs, rect.Y+yTs-v.ScrollPos.Y, v.ContentAttr)
		}
	}

	if v.SelMode {
		v.drawSelText(selTsBeginPos, selTsEndPos)
	}
}

func (v *EditView) drawTsLine(xTsStart, xTsEnd, yTs int, attr TermAttr, contentRect Rect) {
	for xTs := xTsStart; xTs <= xTsEnd; xTs++ {
		c := v.Ts.Ch(xTs, yTs)
		if c != 0 {
			printCh(c, contentRect.X+xTs, contentRect.Y+yTs-v.ScrollPos.Y, attr)
		}
	}
}

func (v *EditView) drawSelText(tsBegin, tsEnd Pos) {
	rect := v.contentRect()
	selAttr := reverseAttr(v.ContentAttr)

	// One line only
	if tsBegin.Y == tsEnd.Y {
		v.drawTsLine(tsBegin.X, tsEnd.X, tsBegin.Y, selAttr, rect)
		return
	}

	// Topmost line
	v.drawTsLine(tsBegin.X, v.Ts.W-1, tsBegin.Y, selAttr, rect)

	// Middle lines
	for yTs := tsBegin.Y + 1; yTs < tsEnd.Y; yTs++ {
		v.drawTsLine(0, v.Ts.W-1, yTs, selAttr, rect)
	}

	// Bottommost line
	v.drawTsLine(0, tsEnd.X, tsEnd.Y, selAttr, rect)
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

	// Buf pos y,x
	sBufPos := fmt.Sprintf("%d,%d", v.BufPos.Y+1, v.BufPos.X+1)
	print(sBufPos, left+width-(width/3), y, v.StatusAttr)

	//$$ Ts pos y,x
	tsPos := v.tsPosFromBufPos(v.BufPos)
	sTsPos := fmt.Sprintf("Ts:(%d,%d)", tsPos.Y, tsPos.X)
	print(sTsPos, left+width-(width*2/3), y, v.StatusAttr)

	// Sel range y,x - y,x
	if v.SelMode {
		selBegin, selEnd := v.OrderedSelPos(v.SelBegin, v.SelEnd)
		sSelRange := fmt.Sprintf("%d,%d - %d,%d", selBegin.Y+1, selBegin.X+1, selEnd.Y+1, selEnd.X+1)
		print(sSelRange, left+width-(width*2/3), y, v.StatusAttr)
	}

	// Pos in doc (%)
	var sYDist string
	nBufLines := len(v.Buf.Lines)
	if nBufLines == 0 {
		sYDist = "0%"
	}
	yDist := v.BufPos.Y * 100 / (nBufLines - 1)
	sYDist = fmt.Sprintf(" %d%% ", yDist)
	print(sYDist, left+width-len(sYDist), y, v.StatusAttr)
}

func (v *EditView) drawCursor(tsPos Pos) {
	rect := v.contentRect()
	tsPos.Y -= v.ScrollPos.Y

	if tsPos.X < rect.W && tsPos.Y < rect.H {
		tb.SetCursor(rect.X+tsPos.X, rect.Y+tsPos.Y)
	}
}

func (v *EditView) Clear() {
	v.Buf.Clear()
	v.SyncBufText()
	v.ResetCur()
}
func (v *EditView) ResetCur() {
	v.BufPos = Pos{0, 0}
}

func (v *EditView) SetText(s string) {
	v.Buf.SetText(s)
	v.SyncBufText()
	v.ResetCur()
}

func (v *EditView) GetText() string {
	return v.Buf.GetText()
}

func (v *EditView) StartSelMode() {
	v.SelMode = true
	v.SelBegin = v.BufPos
	v.SelEnd = v.BufPos
}
func (v *EditView) EndSelMode() {
	v.SelMode = false
	v.SelBegin = v.BufPos
	v.SelEnd = v.BufPos
}
func (v *EditView) UpdateSelPos() {
	if v.SelMode {
		v.SelEnd = v.BufPos
	}
}

// Return pos in first, last order..
func (v *EditView) OrderedSelPos(pos1, pos2 Pos) (Pos, Pos) {
	if (pos2.Y > pos1.Y) ||
		(pos2.Y == pos1.Y && pos2.X > pos1.X) {
		return pos1, pos2
	}
	return pos2, pos1
}

func (v *EditView) HandleEvent(e *tb.Event) (Widget, WidgetEventID) {
	var bufChanged bool
	var c rune

	switch e.Key {
	case tb.KeyEsc:
		v.EndSelMode()

	// Nav single char
	case tb.KeyArrowLeft:
		v.BufPos = v.Buf.PrevPos(v.BufPos)
		v.UpdateSelPos()
	case tb.KeyArrowRight:
		v.BufPos = v.Buf.NextPos(v.BufPos)
		v.UpdateSelPos()
	case tb.KeyArrowUp:
		v.BufPos = v.Buf.UpPos(v.BufPos)
		//v.BufPos = v.UpPos(v.BufPos)
		v.UpdateSelPos()
	case tb.KeyArrowDown:
		v.BufPos = v.Buf.DownPos(v.BufPos)
		//v.BufPos = v.DownPos(v.BufPos)
		v.UpdateSelPos()

	// Nav word/line
	case tb.KeyCtrlP:
		fallthrough
	case tb.KeyCtrlB:
		//$$ go to start of previous word
	case tb.KeyCtrlN:
		fallthrough
	case tb.KeyCtrlF:
		//$$ go to start of next word
	case tb.KeyCtrlA:
		v.BufPos = v.Buf.BOLPos(v.BufPos)
		v.UpdateSelPos()
	case tb.KeyCtrlE:
		v.BufPos = v.Buf.EOLPos(v.BufPos)
		v.UpdateSelPos()

	// Scroll text
	case tb.KeyCtrlU:
		//$$ scroll up half content area
	case tb.KeyCtrlD:
		//$$ scroll down half content area

	// Select/copy/paste text
	case tb.KeyCtrlK:
		v.SelMode = !v.SelMode
		if v.SelMode {
			v.StartSelMode()
		} else {
			v.EndSelMode()
		}
	case tb.KeyCtrlC:
		//$$ copy selected text
	case tb.KeyCtrlV:
		//$$ paste into
	case tb.KeyCtrlX:
		//$$ cut selected text

	// Delete text
	case tb.KeyDelete:
		v.BufPos = v.Buf.DelChar(v.BufPos)
		bufChanged = true
	case tb.KeyBackspace:
		fallthrough
	case tb.KeyBackspace2:
		v.BufPos = v.Buf.DelPrevChar(v.BufPos)
		bufChanged = true

	// Text entry
	case tb.KeyEnter:
		v.BufPos = v.Buf.InsEOL(v.BufPos)
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
		v.BufPos = v.Buf.InsChar(v.BufPos, c)
		bufChanged = true
	}

	if bufChanged {
		v.SyncBufText()
	}

	return v, WidgetEventNone
}
