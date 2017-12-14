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
	buf.WriteLine("aaa")
	buf.WriteLine("zzz")
	buf.WriteLine(" ")
	buf.WriteLine("Word1 a the at supercalifragilisticexpialidocious, and a somewhatlongerword is also here.")
	buf.WriteLine("")
	buf.WriteLine("The quick brown fox jumps over the lazy dog.")
	buf.WriteLine("Last line!")

	termW, termH := tb.Size()
	view := NewView(0, 0, termW, termH-5, buf)

	statusP := NewPanel(view.Border.X, view.Border.Y+view.Border.Height,
		view.Border.Width, 5, "x:0 y:0")

	tb.Clear(0, 0)
	view.Draw()
	view.DrawCursor()
	statusP.Draw()
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
		statusP.Text = fmt.Sprintf("x:%d y:%d", bufPos.X, bufPos.Y)
		statusP.Draw()
		flush()
	}

}
