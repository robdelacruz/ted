package main

import (
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
	view := scr.View
	ed := scr.View.Ed

	tb.SetCursor(view.CurX, view.CurY)
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
				view.CurLeft()
			case tb.KeyArrowUp:
				view.CurUp()
			case tb.KeyArrowDown:
				view.CurDown()
			case tb.KeyArrowLeft:
				view.CurLeft()
			case tb.KeyArrowRight:
				view.CurRight()
			case tb.KeyHome:
			case tb.KeyEnd:
			case tb.KeyDelete:
			case tb.KeyEnter:
				ed.InsertNewLine(view.CurY)
				view.CurDown()
			case 0:
				c = e.Ch
			}

			if c != 0 {
				cell := &EdCell{c, scr.Fg, scr.Bg}
				ed.InsertCell(view.CurX, view.CurY, cell)
				view.CurRight()
			}

			scr.Draw()

			tb.SetCursor(view.CurX, view.CurY)
			flush()
		}
	}
}

type Screen struct {
	CurX, CurY    int
	Width, Height int
	Fg, Bg        tb.Attribute
	Ed            *Editor
	View          *EdView
}

func NewScreen() *Screen {
	ed := NewEditor()
	w, h := tb.Size()
	return &Screen{
		Ed:     ed,
		View:   NewView(ed),
		Width:  w,
		Height: h,
		Fg:     tb.ColorDefault,
		Bg:     tb.ColorDefault,
	}
}

func (scr *Screen) Draw() {
	tb.Clear(scr.Fg, scr.Bg)

	x, y := 0, 0
	for _, el := range scr.Ed.Lines {
		for _, cell := range el {
			printCell(x, y, cell)
			x++
		}
		y++
		x = 0
	}
}
