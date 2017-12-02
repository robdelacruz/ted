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
	view := NewView(ed, tb.ColorDefault, tb.ColorDefault, 10, 10, 20, 25)

	view.InsertLine("Now is the time for all good men to come to the aid of the party. And supercalifragilisticexpialidocious, a really, really long word.")
	view.InsertLine("")
	view.InsertLine("The quick brown fox jumps over the lazy dog.")

	view.Draw()
	flush()

	for {
		e := tb.PollEvent()
		if e.Type == tb.EventKey {
			if e.Key == tb.KeyEsc {
				break
			}

			var c rune
			prevCurX, prevCurY := view.CurX, view.CurY

			switch e.Key {
			case tb.KeySpace:
				c = ' '
			case tb.KeyBackspace:
				fallthrough
			case tb.KeyBackspace2:
				view.CurLeft()
				if view.CurX != prevCurX || view.CurY != prevCurY {
					_log.Printf("view.CurX=%d, view.CurY=%d\n", view.CurX, view.CurY)
					ed.DeleteCell(view.CurX, view.CurY)
				}
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
