package main

import (
	"testing"

	tb "github.com/nsf/termbox-go"
)

func TestWidgetPanel(t *testing.T) {
	err := tb.Init()
	if err != nil {
		panic(err)
	}
	defer tb.Close()

	text := `Now is the time for all good men to come to the aid of the party. The quick brown fox jumps over the lazy dog. Now is the time for all good men to come to the aid of the party.`

	text += "\n"
	text += text

	attr := TermAttr{tb.ColorWhite, tb.ColorBlack}

	opts := PanelOptions{
		Text: text,
		Attr: attr,
		Mode: PanelBorder,
	}
	p := NewPanel(0, 0, 20, 25, opts)
	p.Draw()

	p2 := p
	p2.SetPos(10, 10)
	p2.SetText("SetPos()'d" + p2.Text())
	p2.Draw()

	opts = PanelOptions{
		Text: text,
		Attr: attr,
		Mode: 0,
	}
	p3 := NewPanel(30, 30, 50, 15, opts)
	p3.Draw()

	tb.Flush()
	WaitKBEvent()
}
