package main

import (
	"bytes"
	"fmt"

	tb "github.com/nsf/termbox-go"
)

type EdCell struct {
	Ch     rune
	Fg, Bg tb.Attribute
}

type EdLine []*EdCell

type Editor struct {
	CurX, CurY int
	Lines      []EdLine
}

func NewEditor() *Editor {
	ed := &Editor{}
	ed.Lines = append(ed.Lines, EdLine{})

	return ed
}

func (ed *Editor) Print() {
	for _, line := range ed.Lines {
		var b bytes.Buffer
		for _, cell := range line {
			b.WriteString(fmt.Sprintf("%c", cell.Ch))
		}
		_log.Println(b.String())
	}
}

// Make sure cursor stays within text bounds
func (ed *Editor) BoundsCursor() {
	if ed.CurY < 0 {
		ed.CurY = 0
	}
	if ed.CurY > len(ed.Lines)-1 {
		ed.CurY = len(ed.Lines) - 1
	}

	if ed.CurX < 0 {
		ed.CurX = 0
	}
	if ed.CurX > len(ed.Lines[ed.CurY]) {
		ed.CurX = len(ed.Lines[ed.CurY])
	}
}
func (ed *Editor) CurLeft() {
	ed.CurX--
	ed.BoundsCursor()
}
func (ed *Editor) CurRight() {
	ed.CurX++
	ed.BoundsCursor()
}
func (ed *Editor) CurUp() {
	ed.CurY--
	ed.BoundsCursor()
}
func (ed *Editor) CurDown() {
	ed.CurY++
	ed.BoundsCursor()
}

func (ed *Editor) curLine() EdLine {
	return ed.Lines[ed.CurY]
}
func (ed *Editor) setCurLine(el EdLine) {
	ed.Lines[ed.CurY] = el
}
func (ed *Editor) InsertCell(cell *EdCell) {
	// Carriage return, add new line below current line
	if cell.Ch == '\n' {
		ed.InsertNewLine()
		return
	}

	// Add new char
	ed.setCurLine(ed.curLine().InsertCell(ed.CurX, cell))
	ed.CurRight()
}

func (ed *Editor) InsertLines(y int, els []EdLine) {
	ed.Lines = append(ed.Lines, els...)
	copy(ed.Lines[y+len(els):], ed.Lines[y:])
	copy(ed.Lines[y:], els)
}

func (ed *Editor) InsertLine(y int, el EdLine) {
	ed.Lines = append(ed.Lines, nil)
	copy(ed.Lines[y+1:], ed.Lines[y:])
	ed.Lines[y] = el
}

// Insert new line after cursor
func (ed *Editor) InsertNewLine() {
	ed.InsertLine(ed.CurY+1, EdLine{})
	ed.CurDown()
}

func (el EdLine) InsertCells(x int, rcs []*EdCell) EdLine {
	el = append(el, rcs...)
	copy(el[x+len(rcs):], el[x:])
	copy(el[x:], rcs)
	return el
}

func (el EdLine) InsertCell(x int, cell *EdCell) EdLine {
	el = append(el, nil)
	copy(el[x+1:], el[x:])
	el[x] = cell
	return el
}

func (el EdLine) DeleteCells(x, n int) EdLine {
	el = append(el[:x], el[x+n:]...)
	return el
}

func (el EdLine) DeleteCell(x int) EdLine {
	el.DeleteCells(x, 1)
	return el
}
