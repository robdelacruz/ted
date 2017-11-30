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

	ed := NewEditor()
	view := NewView(ed, tb.ColorDefault, tb.ColorDefault, 10, 10, 15, 15)

	view.Draw()
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
				ed.DeleteCell(view.CurX, view.CurY)
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
				ed.DeleteCell(view.CurX, view.CurY)
			case tb.KeyEnter:
				ed.InsertNewLine(view.CurX, view.CurY)
				view.CurY++
				view.CurX = 0
			case 0:
				c = e.Ch
			}

			if c != 0 {
				cell := view.NewCell(c)
				ed.InsertCell(view.CurX, view.CurY, cell)
				view.CurRight()
			}

			view.Draw()
			flush()
		}
	}
}

type Screen struct {
	Width, Height int
}

func NewScreen() *Screen {
	w, h := tb.Size()
	return &Screen{
		Width:  w,
		Height: h,
	}
}
