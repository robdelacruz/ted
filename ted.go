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
	buf.WriteString("The quick brown fox jumps over the lazy dog.")

	view := NewView(10, 10, 25, 15, buf)

	view.Draw()

}
