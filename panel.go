package main

import (
	tb "github.com/nsf/termbox-go"
)

type Panel struct {
	Content Area
	Outline Area
	*Buf
	*TextBlk
	Mode        PanelMode
	ContentAttr TermAttr
}

type PanelMode uint

const (
	PanelBorder PanelMode = 1 << iota
)

func NewPanel(x, y, w, h int, mode PanelMode, contentAttr TermAttr, text string) *Panel {
	outline := NewArea(x, y, w, h)
	content := outline

	if mode&PanelBorder != 0 {
		content = NewArea(x+1, y+1, w-2, h-2)
	}

	p := &Panel{}
	p.Outline = outline
	p.Content = content
	p.Buf = NewBuf()
	p.TextBlk = NewTextBlk(content.Width, 0)
	p.Mode = mode
	p.ContentAttr = contentAttr

	p.Buf.SetText(text)
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
	clearArea(p.Outline, p.ContentAttr)
	if p.Mode&PanelBorder != 0 {
		drawBox(p.Outline.X, p.Outline.Y, p.Outline.Width, p.Outline.Height, p.ContentAttr)
	}

	p.drawText()
}
func (p *Panel) drawText() {
	p.TextBlk.PrintToArea(p.Content, p.ContentAttr)
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
	if p.Mode&PanelBorder != 0 {
		borderWidth = 1
	}
	paddingWidth := 0
	p.Outline, p.Content = adjPos(p.Outline, p.Content, x, y, borderWidth, paddingWidth)
}
