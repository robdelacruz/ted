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

type TraverseBufOp uint

const (
	UpdateTS TraverseBufOp = 1 << iota
	UpdateTSPos
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
	v.TraverseBuf(UpdateTS)

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

func (v *EditView) updateTsPos(yBuf, xBufWordStart, xBufWordEnd, xTs, yTs int) bool {
	if v.BufPos.Y == yBuf &&
		v.BufPos.X >= xBufWordStart && v.BufPos.X < xBufWordEnd {
		// Update tsPos corresponding to bufPos.
		v.TsPos.Y = yTs
		v.TsPos.X = xTs - (xBufWordEnd - v.BufPos.X)
		return true
	}

	return false
}

func (v *EditView) layoutLine(op TraverseBufOp, yBuf int, ts *TextSurface, yTs int) int {
	var fSet bool
	xTs := 0
	xBuf := 0

	bufLine := v.Buf.Lines[yBuf]
	words := parseWords(bufLine)

	for _, word := range words {
		lenWord := len([]rune(word))
		// word can't fit in remaining line, add to next line.
		if xTs+lenWord > ts.W {
			yTs++
			xTs = 0
			if op&UpdateTS != 0 {
				ts.WriteString(word, xTs, yTs)
			}

			xTs = lenWord
			xBuf += lenWord

			if op&UpdateTSPos != 0 && !fSet {
				fSet = v.updateTsPos(yBuf, xBuf-lenWord, xBuf, xTs, yTs)
			}
			continue
		}

		// add word to remaining line.
		if op&UpdateTS != 0 {
			ts.WriteString(word, xTs, yTs)
		}

		xTs += lenWord
		xBuf += lenWord

		if op&UpdateTSPos != 0 && !fSet {
			fSet = v.updateTsPos(yBuf, xBuf-lenWord, xBuf, xTs, yTs)
		}
	}

	// bufPos falls outside line bounds
	if op&UpdateTSPos != 0 && !fSet && v.BufPos.Y == yBuf {
		v.TsPos.Y = yTs

		if v.BufPos.X == 0 {
			v.TsPos.X = 0
		} else {
			v.TsPos.X = xTs
		}
		fSet = true
	}

	return yTs + 1
}

func (v *EditView) TraverseBuf(op TraverseBufOp) {
	if op&UpdateTS != 0 {
		v.Ts.Clear(0)
	}

	yBuf := v.YBufOffset
	yTs := 0

	for yBuf < len(v.Buf.Lines) {
		yTs = v.layoutLine(op, yBuf, v.Ts, yTs)
		yBuf++
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
	v.TraverseBuf(UpdateTSPos)

	rect := v.contentRect()
	tb.SetCursor(rect.X+v.TsPos.X, rect.Y+v.TsPos.Y)
}

func (v *EditView) Clear() {
	v.Buf.Clear()
	v.TraverseBuf(UpdateTS)
	v.ResetCur()
}
func (v *EditView) ResetCur() {
	v.BufPos = Pos{0, 0}
	v.YBufOffset = 0
}

func (v *EditView) SetText(s string) {
	v.Buf.SetText(s)
	v.TraverseBuf(UpdateTS)
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
		case tb.KeyArrowDown:
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
