package main

import (
	tb "github.com/nsf/termbox-go"
)

type Panel struct {
	Rect
	*Buf
	*TextBlk
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
	if opts.Mode&PanelBorder != 0 {
		w -= 2
	}
	p.TextBlk = NewTextBlk(w, 0)
	p.SyncText()

	return p
}

func (p *Panel) SetPos(x, y int) {
	p.X = x
	p.Y = y
}

func (p *Panel) Draw() {
	clearRect(p.Rect, p.Opts.Attr)
	if p.Opts.Mode&PanelBorder != 0 {
		drawBox(p.Rect.X, p.Rect.Y, p.Rect.W, p.Rect.H, p.Opts.Attr)
	}

	p.drawText()
}
func (p *Panel) drawText() {
	rect := p.Rect
	if p.Opts.Mode&PanelBorder != 0 {
		rect.X++
		rect.Y++
		rect.W -= 2
		rect.H -= 2
	}

	p.TextBlk.PrintToArea(rect, p.Opts.Attr)
}

func (p *Panel) SetText(s string) {
	p.Buf.SetText(s)
	p.SyncText()
}
func (p *Panel) GetText() string {
	return p.Buf.GetText()
}
func (p *Panel) SyncText() {
	p.TextBlk.FillWithBuf(p.Buf)
}

func (p *Panel) HandleEvent(e *tb.Event) (Widget, WidgetEventID) {
	return p, WidgetEventNone
}
