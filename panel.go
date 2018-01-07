package main

import (
	tb "github.com/nsf/termbox-go"
)

type Panel struct {
	Rect
	*Buf
	Opts PanelOptions
}

type PanelMode uint

const (
	PanelBorder PanelMode = 1 << iota
)

type PanelOptions struct {
	Text string
	Attr TermAttr
	Mode PanelMode
}

func NewPanel(x, y, w, h int, opts PanelOptions) *Panel {
	p := &Panel{}
	p.Rect = NewRect(x, y, w, h)
	p.Opts = opts
	p.Buf = NewBuf()
	p.Buf.SetText(opts.Text)

	return p
}

func (p *Panel) SetPos(x, y int) {
	p.X = x
	p.Y = y
}

func (p *Panel) contentRect() Rect {
	rect := p.Rect
	if p.Opts.Mode&PanelBorder != 0 {
		rect.X++
		rect.Y++
		rect.W -= 2
		rect.H -= 2
	}
	return rect
}

func (p *Panel) Draw() {
	clearRect(p.Rect, p.Opts.Attr)
	if p.Opts.Mode&PanelBorder != 0 {
		drawBox(p.Rect.X, p.Rect.Y, p.Rect.W, p.Rect.H, p.Opts.Attr)
	}

	p.drawText(p.contentRect())
}
func (p *Panel) drawText(rect Rect) {
	bit := NewBufIterWl(p.Buf, rect.W)
	i := 0
	for bit.ScanNext() {
		if i > rect.H-1 {
			break
		}
		sline := bit.Text()
		print(sline, rect.X, rect.Y+i, p.Opts.Attr)
		i++
	}
}

func (p *Panel) SetText(s string) {
	p.Buf.SetText(s)
}
func (p *Panel) Text() string {
	return p.Buf.Text()
}

func (p *Panel) HandleEvent(e *tb.Event) (Widget, WidgetEventID) {
	return p, WidgetEventNone
}
