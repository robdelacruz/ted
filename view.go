package main

import (
	tb "github.com/nsf/termbox-go"
)

type EdView struct {
	Ed                      *Editor
	CurX, CurY              int
	Fg, Bg                  tb.Attribute
	FrameLeft, FrameTop     int
	FrameWidth, FrameHeight int
	Left, Top               int
	Width, Height           int
}

func NewView(ed *Editor, fg, bg tb.Attribute, left, top, width, height int) *EdView {
	v := &EdView{}
	v.Ed = ed
	v.Fg = tb.ColorDefault
	v.Bg = tb.ColorDefault

	v.FrameLeft = left
	v.FrameTop = top
	v.FrameWidth = width
	v.FrameHeight = height

	v.Left = v.FrameLeft + 1
	v.Top = v.FrameTop + 1
	v.Width = v.FrameWidth - 2
	v.Height = v.FrameHeight - 2

	return v
}

func (v *EdView) Draw() {
	tb.Clear(v.Fg, v.Bg)

	// Border
	c := v.NewCell('┌')
	printCell(v.FrameLeft, v.FrameTop, c)
	c.Ch = '┐'
	printCell(v.FrameLeft+v.FrameWidth-1, v.FrameTop, c)

	c.Ch = '─'
	for i := v.FrameLeft + 1; i < v.FrameLeft+v.FrameWidth-1; i++ {
		printCell(i, v.FrameTop, c)
	}
	for i := v.FrameLeft + 1; i < v.FrameLeft+v.FrameWidth-1; i++ {
		printCell(i, v.FrameTop+v.FrameHeight-1, c)
	}

	c.Ch = '│'
	for j := v.FrameTop + 1; j < v.FrameTop+v.FrameHeight-1; j++ {
		printCell(v.FrameLeft, j, c)
	}
	for j := v.FrameTop + 1; j < v.FrameTop+v.FrameHeight-1; j++ {
		printCell(v.FrameLeft+v.FrameWidth-1, j, c)
	}

	c.Ch = '┘'
	printCell(v.FrameLeft+v.FrameWidth-1, v.FrameTop+v.FrameHeight-1, c)
	c.Ch = '└'
	printCell(v.FrameLeft, v.FrameTop+v.FrameHeight-1, c)

	// Content
	x, y := v.Left, v.Top
draw1:
	for _, line := range v.Ed.Lines {
		for _, cell := range line {
			if x > v.Left+v.Width-1 {
				y++
				x = v.Left

				if y > v.Top+v.Height-1 {
					break draw1
				}
			}

			printCell(x, y, cell)
			x++
		}

		y++
		x = v.Left

		if y > v.Top+v.Height-1 {
			break
		}
	}

	// Compute view cursor in relation to doc cursor
	curSetX := v.Left + (v.CurX % v.Width)
	curSetY := v.Top
	for i := 0; i < v.CurY; i++ {
		if i < len(v.Ed.Lines) {
			line := v.Ed.Lines[i]
			curSetY += len(line)/v.Width + 1
		}
	}
	if v.CurY < len(v.Ed.Lines) {
		curSetY += v.CurX / v.Width
	}
	tb.SetCursor(curSetX, curSetY)
}

// Make sure cursor stays within text bounds
func (v *EdView) BoundsCursor() {
	ed := v.Ed

	if v.CurY < 0 {
		v.CurY = 0
	}
	if v.CurY > len(ed.Lines)-1 {
		v.CurY = len(ed.Lines) - 1
	}

	if v.CurX < 0 {
		v.CurX = 0
	}
	if v.CurX > len(ed.Lines[v.CurY]) {
		v.CurX = len(ed.Lines[v.CurY])
	}
}

func (v *EdView) currentLine() EdLine {
	return v.Ed.Line(v.CurY)
}

func (v *EdView) CurLeft() {
	v.CurX--
	if v.CurX < 0 {
		if v.CurY > 0 {
			v.CurY--
			v.CurX = len(v.currentLine()) - 1
		} else {
			v.CurX++
		}
	}
	v.BoundsCursor()
}
func (v *EdView) CurRight() {
	v.CurX++
	if v.CurX > len(v.currentLine()) {
		if v.CurY < len(v.Ed.Lines)-1 {
			v.CurY++
			v.CurX = 0
		} else {
			v.CurX--
		}
	}
	v.BoundsCursor()
}
func (v *EdView) CurUp() {
	if v.CurX > v.Width-1 {
		// wrapped line
		v.CurX -= v.Width
		v.BoundsCursor()
		return
	}

	if v.CurY == 0 {
		return
	}

	// Go up one line
	v.CurY--

	// cursor is past rightmost char
	if v.CurX > len(v.currentLine())-1 {
		v.CurX = len(v.currentLine()) - 1
		v.BoundsCursor()
		return
	}

	// wrapped line adjustment
	if len(v.currentLine()) > v.Width {
		v.CurX += (len(v.currentLine()) / v.Width) * v.Width
		v.BoundsCursor()
		return
	}

	v.BoundsCursor()
}
func (v *EdView) CurDown() {
	v.CurX += v.Width

	// within wrapped line
	if v.CurX < len(v.currentLine()) {
		v.BoundsCursor()
		return
	}

	v.CurX -= v.Width

	// end of wrapped line
	if (len(v.currentLine()) > v.Width) &&
		(v.CurX < len(v.currentLine())-(len(v.currentLine())%v.Width)) {
		v.CurX = len(v.currentLine()) - 1
		v.BoundsCursor()
		return
	}

	v.CurY++
	v.BoundsCursor()
}

// Return whether cursor is in a wrapped line.
// wrapped line = line length longer than view width
func (v *EdView) IsWrapLine() bool {
	line := v.Ed.Line(v.CurY)
	if len(line) > v.Width {
		return true
	}
	return false
}

// Return whether cursor is in the trailing part of a wrapped line.
// trailing part = text that is 'wrapped' to the next line.
func (v *EdView) IsTrailLine() bool {
	if !v.IsWrapLine() {
		return false
	}
	if v.CurX > v.Width-1 {
		return true
	}
	return false
}

func (v *EdView) NewCell(c rune) *EdCell {
	return &EdCell{
		Ch: c,
		Fg: v.Fg,
		Bg: v.Bg,
	}
}
