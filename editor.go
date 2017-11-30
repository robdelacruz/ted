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
	Lines []EdLine
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

func (ed *Editor) Line(y int) EdLine {
	if y < len(ed.Lines) {
		return ed.Lines[y]
	}
	return EdLine{}
}
func (ed *Editor) ReplaceLine(y int, el EdLine) {
	if y < len(ed.Lines) {
		ed.Lines[y] = el
	}
}
func (ed *Editor) InsertCell(x, y int, cell *EdCell) {
	// Carriage return, add new line below current line
	if cell.Ch == '\n' {
		ed.InsertNewLine(x, y)
		return
	}

	// Add new char
	ed.ReplaceLine(y, ed.Line(y).InsertCell(x, cell))
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

func (ed *Editor) InsertNewLine(x, y int) {
	curLine := ed.Lines[y]
	newLine := curLine[x:]
	ed.Lines[y] = curLine[:x]

	ed.InsertLine(y+1, newLine)
}

func (ed *Editor) DeleteCell(x, y int) {
	curLine := ed.Lines[y]
	ed.Lines[y] = curLine.DeleteCell(x)
}

// EdLine methods
//
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
	el = el.DeleteCells(x, 1)
	return el
}
