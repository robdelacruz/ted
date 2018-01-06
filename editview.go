package main

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
	BufPos      Pos
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

	v.drawText()

	if v.Mode&EditViewStatusLine != 0 {
		v.drawStatus()
	}
}

func (v *EditView) drawText() {
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
}

func (v *EditView) Clear() {
	v.Buf.Clear()
	v.ResetCur()
}
func (v *EditView) ResetCur() {
	v.BufPos = Pos{0, 0}
}

func (v *EditView) SetText(s string) {
}

func (v *EditView) GetText() string {
	return v.Buf.Text()
}

func (v *EditView) HandleEvent(e *tb.Event) (Widget, WidgetEventID) {
	var bufChanged bool
	var c rune

	switch e.Key {
	case tb.KeyEsc:

	// Nav single char
	case tb.KeyArrowLeft:
	case tb.KeyArrowRight:
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
		v.BufPos = v.Buf.InsChar(v.BufPos, c)
		bufChanged = true
	}

	if bufChanged {
	}

	return v, WidgetEventNone
}
