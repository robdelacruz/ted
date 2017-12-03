package main

import ()

type View struct {
	area Area
	buf  *Buf
}

func NewView(x, y, w, h int, buf *Buf) *View {
	area := NewArea(x, y, w, h)
	view := &View{area, buf}
	return view
}

func (v *View) Draw() {
}
