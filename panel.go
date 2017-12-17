package main

import ()

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

func (p *Panel) Draw() {
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
func (p *Panel) SyncText() {
	p.TextBlk.FillWithBuf(p.Buf)
}
