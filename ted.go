package main

import (
	"bytes"
	"fmt"
	"log"
	"os"

	tb "github.com/nsf/termbox-go"
)

var _log *log.Logger

func main() {
	flog, err := os.Create("./log.txt")
	if err != nil {
		panic(err)
	}
	defer flog.Close()
	_log = log.New(flog, "", 0)

	err = tb.Init()
	if err != nil {
		panic(err)
	}
	defer tb.Close()

	scr := NewScreen()
	et := scr.ET

	tb.SetCursor(et.CurX, et.CurY)
	flush()

	for {
		e := tb.PollEvent()
		if e.Type == tb.EventKey {
			if e.Key == tb.KeyEsc {
				break
			}

			var c rune
			switch e.Key {
			case tb.KeySpace:
				c = ' '
			case tb.KeyBackspace:
				fallthrough
			case tb.KeyBackspace2:
				et.CurLeft()
			case tb.KeyArrowUp:
				et.CurUp()
			case tb.KeyArrowDown:
				et.CurDown()
			case tb.KeyArrowLeft:
				et.CurLeft()
			case tb.KeyArrowRight:
				et.CurRight()
			case tb.KeyHome:
			case tb.KeyEnd:
			case tb.KeyDelete:
			case tb.KeyEnter:
				et.InsertNewLineCursor()
			case 0:
				c = e.Ch
			}

			if c != 0 {
				rc := &RuneCell{c, scr.Fg, scr.Bg}
				et.InsertCell(rc)
			}

			scr.Draw()

			tb.SetCursor(et.CurX, et.CurY)
			flush()
		}
	}
}

type RuneCell struct {
	Ch     rune
	Fg, Bg tb.Attribute
}

type EdLine []*RuneCell

type EdText struct {
	CurX, CurY int
	Lines      []EdLine
}

type Screen struct {
	CurX, CurY    int
	Width, Height int
	Fg, Bg        tb.Attribute
	ET            *EdText
}

func NewScreen() *Screen {
	w, h := tb.Size()
	return &Screen{
		ET:     NewEdText(),
		Width:  w,
		Height: h,
		Fg:     tb.ColorDefault,
		Bg:     tb.ColorDefault,
	}
}

func (scr *Screen) Draw() {
	tb.Clear(scr.Fg, scr.Bg)

	x, y := 0, 0
	for _, el := range scr.ET.Lines {
		for _, rc := range el {
			printRC(x, y, rc)
			x++
		}
		y++
		x = 0
	}
}

func NewEdText() *EdText {
	et := &EdText{}

	// init with 1 empty line
	et.Lines = []EdLine{
		EdLine{},
	}

	return et
}

// Make sure cursor stays within text bounds
func (et *EdText) BoundsCursor() {
	if et.CurY < 0 {
		et.CurY = 0
	}
	if et.CurY > len(et.Lines)-1 {
		et.CurY = len(et.Lines) - 1
	}

	if et.CurX < 0 {
		et.CurX = 0
	}
	if et.CurX > len(et.Lines[et.CurY]) {
		et.CurX = len(et.Lines[et.CurY])
	}
}
func (et *EdText) CurLeft() {
	et.CurX--
	et.BoundsCursor()
}
func (et *EdText) CurRight() {
	et.CurX++
	et.BoundsCursor()
}
func (et *EdText) CurUp() {
	et.CurY--
	et.BoundsCursor()
}
func (et *EdText) CurDown() {
	et.CurY++
	et.BoundsCursor()
}

func (et *EdText) CurLine() EdLine {
	return et.Lines[et.CurY]
}
func (et *EdText) SetCurLine(el EdLine) {
	et.Lines[et.CurY] = el
}
func (et *EdText) InsertCell(rc *RuneCell) {
	// Carriage return, add new line below current line
	if rc.Ch == '\n' {
		et.InsertNewLineCursor()
		return
	}

	// Add new char
	et.SetCurLine(et.CurLine().InsertCell(et.CurX, rc))
	et.CurRight()
}

func (et *EdText) Print() {
	for _, line := range et.Lines {
		var b bytes.Buffer
		for _, rc := range line {
			b.WriteString(fmt.Sprintf("%c", rc.Ch))
		}
		_log.Println(b.String())
	}
}

func (el EdLine) InsertCells(x int, rcs []*RuneCell) EdLine {
	el = append(el, rcs...)
	copy(el[x+len(rcs):], el[x:])
	copy(el[x:], rcs)
	return el
}

func (el EdLine) InsertCell(x int, rc *RuneCell) EdLine {
	el = append(el, nil)
	copy(el[x+1:], el[x:])
	el[x] = rc
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

func (et *EdText) InsertLines(y int, els []EdLine) {
	et.Lines = append(et.Lines, els...)
	copy(et.Lines[y+len(els):], et.Lines[y:])
	copy(et.Lines[y:], els)
}

func (et *EdText) InsertLine(y int, el EdLine) {
	et.Lines = append(et.Lines, nil)
	copy(et.Lines[y+1:], et.Lines[y:])
	et.Lines[y] = el
}

// Insert new line after cursor
func (et *EdText) InsertNewLineCursor() {
	et.InsertLine(et.CurY+1, EdLine{})
	et.CurDown()
}

func flush() {
	err := tb.Flush()
	if err != nil {
		panic(err)
	}

}

func print(x, y int, fg, bg tb.Attribute, s string) {
	for _, c := range s {
		tb.SetCell(x, y, c, fg, bg)
		x++
	}
}

func strToRuneCells(s string, fg, bg tb.Attribute) []RuneCell {
	var rcs []RuneCell
	for _, c := range s {
		rcs = append(rcs, RuneCell{c, fg, bg})
	}
	return rcs
}

func printRC(x, y int, rc *RuneCell) {
	tb.SetCell(x, y, rc.Ch, rc.Fg, rc.Bg)
}

func printRCs(x, y int, rcs []*RuneCell) {
	for _, rc := range rcs {
		printRC(x, y, rc)
		x++
	}
}
