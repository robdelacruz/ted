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
	buf.WriteString("Now is the time for all good men to come to the aid of the party.")
	buf.WriteString("")
	buf.WriteString("Word1 a the at supercalifragilisticexpialidocious, and a somewhatlongerword is also here.")
	buf.WriteString("The quick brown fox jumps over the 123")
	buf.WriteString("The quick brown fox jumps over the 123")
	buf.WriteString("Last line!")

	//view := NewView(10, 10, 25, 15, buf)
	view := NewView(20, 20, 40, 10, buf)
	p := NewPanel(view.Border.X, view.Border.Y+view.Border.Height,
		view.Border.Width, 10, "x:0 y:0")

	tb.Clear(0, 0)
	view.Draw()
	p.Draw()
	flush()

	for {
		e := tb.PollEvent()
		if e.Type == tb.EventKey {
			if e.Key == tb.KeyEsc {
				break
			}
		}

		view.HandleEvent(&e)

		textPos := view.TextPos()
		p.Text = fmt.Sprintf("x:%d y:%d", textPos.X, textPos.Y)

		tb.Clear(0, 0)
		view.Draw()
		p.Draw()
		flush()
	}

}
