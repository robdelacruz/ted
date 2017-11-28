package main

import ()

type EdView struct {
	Ed         *Editor
	CurX, CurY int
}

func NewView(ed *Editor) *EdView {
	v := &EdView{
		Ed: ed,
	}
	return v
}

// Make sure cursor stays within text bounds
func (v *EdView) BoundsCursor() {
	ed := v.Ed

	if v.CurY < 0 {
		v.CurY = 0
	}
	if v.CurY > len(ed.Lines)-1 {
		v.CurY = len(ed.Lines) - 1
	}

	if v.CurX < 0 {
		v.CurX = 0
	}
	if v.CurX > len(ed.Lines[v.CurY]) {
		v.CurX = len(ed.Lines[v.CurY])
	}
}

func (v *EdView) CurLeft() {
	v.CurX--
	v.BoundsCursor()
}
func (v *EdView) CurRight() {
	v.CurX++
	v.BoundsCursor()
}
func (v *EdView) CurUp() {
	v.CurY--
	v.BoundsCursor()
}
func (v *EdView) CurDown() {
	v.CurY++
	v.BoundsCursor()
}
