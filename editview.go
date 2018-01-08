package main

import (
	"fmt"
	"log"
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

func logln(s string) {
	log.Println(s)
}

func (v *EditView) drawText(rect Rect) {
	v.bitWl.Reset()
	i := 0
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
	if v.bitWl.Seek(v.Cur) {
		contentCurPos.Y = v.bitWl.WrapLineIndex()
		contentCurPos.X = v.Cur.X - v.bitWl.Pos().X
	}

	if contentCurPos.Y < rect.H && contentCurPos.X < rect.W {
		tb.SetCursor(rect.X+contentCurPos.X, rect.Y+contentCurPos.Y)
	}
}

func (v *EditView) ResetCur() {
	v.Cur = Pos{0, 0}
	v.bitCur.Seek(v.Cur)
}

func (v *EditView) SetText(s string) {
	v.Buf.SetText(s)
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
		if v.bitCur.Pos() != v.Cur {
			v.bitCur.Seek(v.Cur)
		}
		if v.bitCur.ScanPrev() {
			v.Cur = v.bitCur.Pos()
		}
	case tb.KeyArrowRight:
		if v.bitCur.Pos() != v.Cur {
			v.bitCur.Seek(v.Cur)
		}
		if v.bitCur.ScanNext() {
			v.Cur = v.bitCur.Pos()
		}
	case tb.KeyArrowUp:
	case tb.KeyArrowDown:

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
	case tb.KeyCtrlE:

	// Scroll text
	case tb.KeyCtrlU:
		//$$ scroll up half content area
	case tb.KeyCtrlD:
		//$$ scroll down half content area

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
	case tb.KeyBackspace:
		fallthrough
	case tb.KeyBackspace2:

	// Text entry
	case tb.KeyEnter:
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
	}

	return v, WidgetEventNone
}
