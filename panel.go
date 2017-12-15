package main

import ()

type Panel struct {
	Content Area
	Outline Area
	*Buf
	fOutline bool
}

func NewPanel(x, y, w, h int, fOutline bool) *Panel {
	outline := NewArea(x, y, w, h)
	content := NewArea(x+1, y+1, w-2, h-2)

	if !fOutline {
		content = outline
	}

	p := &Panel{}
	p.Outline = outline
	p.Content = content
	p.Buf = NewBuf()
	p.fOutline = fOutline

	return p
}

func (p *Panel) Pos() Pos {
	return Pos{p.Outline.X, p.Outline.Y}
}
func (p *Panel) Size() Size {
	return Size{p.Outline.Width, p.Outline.Height}
}

func (p *Panel) Draw() {
	if p.fOutline {
		drawBox(p.Outline.X, p.Outline.Y, p.Outline.Width, p.Outline.Height, 0, 0)
	}

	x, y := p.Content.X, p.Content.Y
	for i, l := range p.Lines {
		print(l, x, y, 0, 0)

		y++
		if i >= p.Content.Height-1 {
			break
		}
	}
}
