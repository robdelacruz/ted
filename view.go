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

	buf [][]*EdCell
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

	for i := 0; i < v.Height; i++ {
		bufLine := make([]*EdCell, v.Width)
		v.buf = append(v.buf, bufLine)
	}

	return v
}

func (v *EdView) InsertLine(s string) {
	l := EdLine(strToRuneCells(s, v.Fg, v.Bg))
	v.Ed.InsertLine(len(v.Ed.Lines), l)
}

func (v *EdView) drawBox(x, y, width, height int) {
	c := v.NewCell('┌')
	printCell(x, y, c)
	c.Ch = '┐'
	printCell(x+width-1, y, c)

	c.Ch = '─'
	for i := x + 1; i < x+width-1; i++ {
		printCell(i, y, c)
	}
	for i := x + 1; i < x+width-1; i++ {
		printCell(i, y+height-1, c)
	}

	c.Ch = '│'
	for j := y + 1; j < y+height-1; j++ {
		printCell(x, j, c)
	}
	for j := y + 1; j < y+height-1; j++ {
		printCell(x+width-1, j, c)
	}

	c.Ch = '┘'
	printCell(x+width-1, y+height-1, c)
	c.Ch = '└'
	printCell(x, y+height-1, c)
}

func (v *EdView) drawBuf() {
	blankCell := v.NewCell(' ')

	for y := 0; y < v.Height; y++ {
		bufLine := v.buf[y]
		for x := 0; x < v.Width; x++ {
			if bufLine[x] == nil {
				printCell(x, y, blankCell)
				continue
			}
			printCell(x, y, bufLine[x])
		}
	}
}

func clearRestOfLine(bufLine []*EdCell, x int) {
	for x < len(bufLine) {
		bufLine[x] = nil
		x++
	}
}

func (v *EdView) copyEdLineToBuf(l EdLine, y int) int {
	x := 0
	for _, c := range l {
		if x > v.Width-1 {
			y++
			x = 0
		}

		if y > len(v.buf)-1 {
			return y
		}

		v.buf[y][x] = c
		x++
	}

	clearRestOfLine(v.buf[y], x)
	return y + 1
}

// Copy editor contents to view buffer
func (v *EdView) copyEdToBuf() {
	y := 0
	for _, line := range v.Ed.Lines {
		y = v.copyEdLineToBuf(line, y)
	}

	for y < len(v.buf) {
		clearRestOfLine(v.buf[y], 0)
		y++
	}
}

func (v *EdView) Draw() {
	tb.Clear(v.Fg, v.Bg)

	v.drawBox(v.FrameLeft, v.FrameTop, v.FrameWidth, v.FrameHeight)

	v.copyEdToBuf()

	for y := 0; y < v.Height; y++ {
		for x := 0; x < v.Width; x++ {
			cell := v.buf[y][x]
			if cell != nil {
				printCell(v.Left+x, v.Top+y, cell)
			}
		}
	}

	tb.SetCursor(v.Left+v.CurX, v.Top+v.CurY)
}

// Make sure cursor stays within text bounds
func (v *EdView) BoundsCursor() {
	if v.CurY < 0 {
		v.CurY = 0
	}
	if v.CurY > v.Height-1 {
		v.CurY = v.Height - 1
	}

	if v.CurX < 0 {
		v.CurX = 0
	}
	if v.CurX > v.Width-1 {
		v.CurX = v.Width - 1
	}
}

func (v *EdView) CurLeft() {
	if v.CurX == 0 && v.CurY == 0 {
		return
	}
	v.CurX--
	if v.CurX < 0 {
		v.CurY--
		v.CurX = v.Width - 1
	}
}
func (v *EdView) CurRight() {
	if v.CurX == v.Width-1 && v.CurY == v.Height-1 {
		return
	}
	v.CurX++
	if v.CurX > v.Width-1 {
		v.CurY++
		v.CurX = 0
	}
}
func (v *EdView) CurUp() {
	if v.CurY == 0 {
		return
	}
	v.CurY--
}
func (v *EdView) CurDown() {
	if v.CurY == v.Height-1 {
		return
	}
	v.CurY++
}

func (v *EdView) NewCell(c rune) *EdCell {
	return &EdCell{
		Ch: c,
		Fg: v.Fg,
		Bg: v.Bg,
	}
}
