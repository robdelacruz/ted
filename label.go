package main

import (
	tb "github.com/nsf/termbox-go"
)

type Label struct {
	Content Rect
	Mode    LabelMode
	Text    string
	Attr    TermAttr
}

type LabelMode uint

const (
	LabelAutoSize LabelMode = 1 << iota
)

func NewLabel(s string, x, y, w, h int, attr TermAttr, mode LabelMode) *Label {
	l := &Label{}
	l.Content = NewRect(x, y, w, h)
	l.Mode = mode
	l.Text = s
	l.Attr = attr

	return l
}

func (l *Label) SetText(s string) {
	l.Text = s
}

func (l *Label) Draw() {
	x, y := l.Content.X, l.Content.Y

	// AutoSize set, so ignore label width.
	if l.Mode&LabelAutoSize != 0 {
		print(l.Text, x, y, l.Attr)
		return
	}

	// Don't print beyond label width.
	for i, c := range l.Text {
		if i > l.Content.W-1 {
			break
		}
		tb.SetCell(x, y, c, l.Attr.Fg, l.Attr.Bg)
		x++
	}
}
