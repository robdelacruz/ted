package main

import (
	tb "github.com/nsf/termbox-go"
)

type Panel struct {
	Content Area
	Outline Area
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
	outline := NewArea(x, y, w, h)
	content := outline

	if opts.Mode&PanelBorder != 0 {
		content = NewArea(x+1, y+1, w-2, h-2)
	}

	p := &Panel{}
	p.Outline = outline
	p.Content = content
	p.Buf = NewBuf()
	p.TextBlk = NewTextBlk(content.Width, 0)
	p.Opts = opts

	p.Buf.SetText(opts.Text)
	p.SyncText()

	return p
}

func (p *Panel) Pos() Pos {
	return Pos{p.Outline.X, p.Outline.Y}
}
func (p *Panel) Size() Size {
	return Size{p.Outline.Width, p.Outline.Height}
}
func (p *Panel) Area() Area {
	return NewArea(p.Outline.X, p.Outline.Y, p.Outline.Width, p.Outline.Height)
}

func (p *Panel) Draw() {
	clearArea(p.Outline, p.Opts.Attr)
	if p.Opts.Mode&PanelBorder != 0 {
		drawBox(p.Outline.X, p.Outline.Y, p.Outline.Width, p.Outline.Height, p.Opts.Attr)
	}

	p.drawText()
}
func (p *Panel) drawText() {
	p.TextBlk.PrintToArea(p.Content, p.Opts.Attr)
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

func (p *Panel) SetPos(x, y int) {
	var borderWidth int
	if p.Opts.Mode&PanelBorder != 0 {
		borderWidth = 1
	}
	paddingWidth := 0
	p.Outline, p.Content = adjPos(p.Outline, p.Content, x, y, borderWidth, paddingWidth)
}
