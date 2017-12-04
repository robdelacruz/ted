package main

import ()

type Panel struct {
	Area
	Text string
}

func NewPanel(x, y, w, h int, text string) *Panel {
	area := NewArea(x, y, w, h)

	panel := &Panel{
		Area: area,
		Text: text,
	}
	return panel
}

func (p *Panel) Draw() {
	print(p.Text, p.X, p.Y, 0, 0)
}
