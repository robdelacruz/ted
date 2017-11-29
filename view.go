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
	for _, line := range v.Ed.Lines {
		for _, cell := range line {
			if x > v.Left+v.Width-1 {
				break
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

	// Show cursor if within view bounds
	if v.CurX < v.Width && v.CurY < v.Height {
		curSetX := v.Left + v.CurX
		curSetY := v.Top + v.CurY
		tb.SetCursor(curSetX, curSetY)
	}
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

func (v *EdView) CurLeft() {
	v.CurX--
	v.BoundsCursor()
}
func (v *EdView) CurRight() {
	v.CurX++
	v.BoundsCursor()
}
func (v *EdView) CurUp() {
	v.CurY--
	v.BoundsCursor()
}
func (v *EdView) CurDown() {
	v.CurY++
	v.BoundsCursor()
}

func (v *EdView) NewCell(c rune) *EdCell {
	return &EdCell{
		Ch: c,
		Fg: v.Fg,
		Bg: v.Bg,
	}
}
