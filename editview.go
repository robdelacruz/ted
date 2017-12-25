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
	Ts          *TextSurface
	Mode        EditViewMode
	ContentAttr TermAttr
	StatusAttr  TermAttr
	BufPos      Pos
	TsPos       Pos
	YBufOffset  int
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
	v.SyncText(v.Ts, nil)

	return v
}

// Parse line to get sequence of words.
// Each whitespace char is considered a single word.
// Ex. "One two  three" => ["One", " ", "two", " ", " ", "three"]
func parseWords(s string) []string {
	var currentWord string
	var words []string

	for _, c := range s {
		if unicode.IsSpace(c) {
			// Add pending word
			words = append(words, currentWord)

			// Add single space word
			words = append(words, expandWhitespaceChar(c))

			currentWord = ""
			continue
		}

		// Add char to pending word
		currentWord += string(c)
	}

	if len(currentWord) > 0 {
		words = append(words, currentWord)
	}

	return words
}

func (v *EditView) SyncText(pTs *TextSurface, pTsPos *Pos) {
	yTs := 0
	xBuf, yBuf := 0, v.YBufOffset

	maxlenWrapline := v.Ts.W
	if pTs != nil {
		maxlenWrapline = pTs.W
	}

	var fTsSet bool

	if pTs != nil {
		pTs.Clear(0)
	}

	cbWord := func(w string) {
	}

	cbWrapLine := func(wrapline string) {
		// Write new wrapline to display.
		if pTs != nil {
			pTs.WriteString(wrapline, 0, yTs)
		}

		lenWrapline := len([]rune(wrapline))

		// Update ts pos if bufpos in this wrapline.
		if pTsPos != nil && !fTsSet && v.BufPos.Y == yBuf {
			if v.BufPos.X >= xBuf && v.BufPos.X <= (xBuf+lenWrapline) {
				pTsPos.X = v.BufPos.X - xBuf
				pTsPos.Y = yTs
				fTsSet = true
			} else if v.BufPos.X == 0 {
				pTsPos.X = 0
				pTsPos.Y = yTs
				fTsSet = true
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

	v.drawCursor()
}

func (v *EditView) drawText() {
	rect := v.contentRect()

	for yTs := 0; yTs < v.Ts.H; yTs++ {
		for xTs := 0; xTs < v.Ts.W; xTs++ {
			c := v.Ts.Char(xTs, yTs)
			if c == 0 {
				c = ' '
			}
			printCh(c, rect.X+xTs, rect.Y+yTs, v.ContentAttr)
		}
	}
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
	sBufPos := fmt.Sprintf("%d,%d", v.BufPos.Y+1, v.BufPos.X+1)
	print(sBufPos, left+width-(width/3), y, v.StatusAttr)

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

func (v *EditView) drawCursor() {
	v.SyncText(nil, &v.TsPos)
	rect := v.contentRect()
	tb.SetCursor(rect.X+v.TsPos.X, rect.Y+v.TsPos.Y)
}

func (v *EditView) Clear() {
	v.Buf.Clear()
	v.SyncText(v.Ts, nil)
	v.ResetCur()
}
func (v *EditView) ResetCur() {
	v.BufPos = Pos{0, 0}
	v.YBufOffset = 0
}

func (v *EditView) SetText(s string) {
	v.Buf.SetText(s)
	v.SyncText(v.Ts, nil)
	v.ResetCur()
}

func (v *EditView) GetText() string {
	return v.Buf.GetText()
}

func (v *EditView) HandleEvent(e *tb.Event) (Widget, WidgetEventID) {
	var c rune
	if e.Type == tb.EventKey {
		switch e.Key {
		case tb.KeyArrowLeft:
			v.BufPos = v.Buf.PrevPos(v.BufPos)
		case tb.KeyArrowRight:
			v.BufPos = v.Buf.NextPos(v.BufPos)
		case tb.KeyArrowUp:
			v.BufPos = v.Buf.UpPos(v.BufPos)
		case tb.KeyArrowDown:
			v.BufPos = v.Buf.DownPos(v.BufPos)
		case tb.KeyCtrlN:
			fallthrough
		case tb.KeyCtrlF:
		case tb.KeyCtrlP:
			fallthrough
		case tb.KeyCtrlB:
		case tb.KeyCtrlA:
			v.BufPos = v.Buf.BOLPos(v.BufPos)
		case tb.KeyCtrlE:
			v.BufPos = v.Buf.EOLPos(v.BufPos)
		case tb.KeyCtrlU:
		case tb.KeyCtrlD:
		case tb.KeyCtrlV:
		case tb.KeyEnter:
		case tb.KeyDelete:
		case tb.KeyBackspace:
			fallthrough
		case tb.KeyBackspace2:
		case tb.KeySpace:
			c = ' '
		case 0:
			c = e.Ch
		}
	}

	// Char entered
	if c != 0 {
	}

	return v, WidgetEventNone
}
