package main

import (
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

	buf := NewBuf()
	buf.WriteLine("Now is the time for all good men to come to the aid of the party.")
	//buf.WriteLine(" ")
	//	buf.WriteLine("Word1 a the at supercalifragilisticexpialidocious, and a somewhatlongerword is also here.")
	//	buf.WriteLine("")
	//	buf.WriteLine("The quick brown fox jumps over the 123")
	//	buf.WriteLine("Last line!")

	view := NewView(5, 5, 40, 20, buf)
	//view := NewView(20, 20, 40, 10, buf)
	p := NewPanel(view.Border.X, view.Border.Y+view.Border.Height,
		view.Border.Width, 10, "x:0 y:0")

	tb.Clear(0, 0)
	view.Draw()
	view.DrawCursor()
	p.Draw()
	flush()

	for {
		e := tb.PollEvent()
		if e.Type == tb.EventKey {
			if e.Key == tb.KeyCtrlQ {
				break
			}
		}

		tb.Clear(0, 0)
		view.HandleEvent(&e)

		bufPos := view.BufPos()
		p.Text = fmt.Sprintf("x:%d y:%d", bufPos.X, bufPos.Y)
		p.Draw()
		flush()
	}

}
