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

	buf := NewBuf()
	buf.WriteString("Now is the time for all good men to come to the aid of the party.")
	buf.WriteString("")
	buf.WriteString("Word1 a the at supercalifragilisticexpialidocious, and a somewhatlongerword is also here.")
	buf.WriteString("The quick brown fox jumps over the lazy dog.")
	buf.WriteString("Last line!")

	view := NewView(10, 10, 25, 15, buf)
	view2 := NewView(20, 20, 40, 10, buf)

	tb.Clear(0, 0)
	view.Draw()
	view2.Draw()
	flush()

	for {
		e := tb.PollEvent()
		if e.Type == tb.EventKey {
			if e.Key == tb.KeyEsc {
				break
			}
		}
	}

}
