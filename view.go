package main

import (
	"unicode"

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

	// view buffer of {Height} lines of {Width} cells
	// 1:1 map of termbox cells to be draw to the view
	buf [][]*EdCell

	// mapping between each cell in buf[][]
	// to the x,y position of cell in Ed
	edCurPos [][]CurPos
}

type CurPos struct{ X, Y int }

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

		posLine := make([]CurPos, v.Width)
		v.edCurPos = append(v.edCurPos, posLine)
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

// Parse an edline into sequence of words.
// Each space char is represented as a single word.
// Ex. "abc def  ghi" => ["abc", " ", "def", " ", " ", "ghi"
func parseEdLineWords(l EdLine) (words [][]*EdCell) {
	currentWord := []*EdCell{}

	for _, c := range l {
		if unicode.IsSpace(c.Ch) {
			// Add pending word
			words = append(words, currentWord)

			// Add single space word
			words = append(words, []*EdCell{c})

			currentWord = []*EdCell{}
			continue
		}

		// Add char to pending word
		currentWord = append(currentWord, c)
	}

	if len(currentWord) > 0 {
		words = append(words, currentWord)
	}

	return words
}

func writeLineToBuf(l EdLine, yLine int, buf [][]*EdCell, edCurPos [][]CurPos, y int) int {
	words := parseEdLineWords(l)
	x := 0

	for _, word := range words {
		// Not enough space in this line to fit word, try in next line
		if (x + len(word) - 1) > (len(buf[y]) - 1) {
			y++
			x = 0
		}

		// Past bottom view line
		if y > len(buf)-1 {
			return y
		}

		// Write word in remaining space
		// Also remember editor cursor position for each char pos
		for xLine, c := range word {
			buf[y][x] = c
			edCurPos[y][x] = CurPos{xLine, yLine}
			x++

			// Word is longer than entire buf width, so split it into
			// multiple lines
			if x > len(buf[y])-1 {
				y++
				x = 0

				if y > len(buf)-1 {
					return y
				}
			}
		}
	}

	clearRestOfLine(buf[y], x)
	return y + 1
}

func (v *EdView) writeEdToBuf() {
	y := 0
	for yLine, l := range v.Ed.Lines {
		y = writeLineToBuf(l, yLine, v.buf, v.edCurPos, y)
		if y > len(v.buf)-1 {
			return
		}
	}

	for y < len(v.buf) {
		clearRestOfLine(v.buf[y], 0)
		y++
	}
}

func (v *EdView) Draw() {
	tb.Clear(v.Fg, v.Bg)

	v.drawBox(v.FrameLeft, v.FrameTop, v.FrameWidth, v.FrameHeight)

	v.writeEdToBuf()

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
