package main

import (
	tb "github.com/nsf/termbox-go"
)

// Panel prompt for text, interactive textual entry using textview.

type Prompt struct {
	PromptPanel *Panel
	Edit        *EditView

	Outline        Area
	Content        Area
	Mode           PromptMode
	QAttr, AnsAttr TermAttr
}

type PromptMode uint

const (
	PromptBorder PromptMode = 1 << iota
)

const (
	PromptOK WidgetEventID = iota
	PromptCancel
)

func NewPrompt(x, y, w, h int, mode PromptMode, prompt string, qAttr, ansAttr TermAttr) *Prompt {
	outline := NewArea(x, y, w, h)
	content := outline

	if mode&PromptBorder != 0 {
		content = NewArea(x+1, y+1, w-2, h-2)
	}

	nEditRows := 2
	promptPanel := NewPanel(content.X, content.Y, content.Width, content.Height-nEditRows, 0, qAttr, prompt)
	edit := NewEditView(content.X, content.Y+2, content.Width, nEditRows, 0, ansAttr, BWAttr, nil)

	pr := &Prompt{}
	pr.Outline = outline
	pr.Content = content
	pr.PromptPanel = promptPanel
	pr.Edit = edit
	pr.Mode = mode
	pr.QAttr = qAttr
	pr.AnsAttr = ansAttr
	return pr
}

func (pr *Prompt) SetEdit(s string) {
	pr.Edit.Clear()
	pr.Edit.WriteLine(s)
	pr.Edit.SyncText()
}
func (pr *Prompt) SetPrompt(prompt string) {
	pr.PromptPanel.SetText(prompt)
}
func (pr *Prompt) GetPrompt() string {
	return pr.PromptPanel.GetText()
}
func (pr *Prompt) Clear() {
	pr.SetPrompt("")
	pr.SetEdit("")
}

func (pr *Prompt) Text() string {
	return pr.Edit.Text()
}

func (pr *Prompt) Draw() {
	clearArea(pr.Outline, pr.QAttr)
	if pr.Mode&PromptBorder != 0 {
		drawBox(pr.Outline.X, pr.Outline.Y, pr.Outline.Width, pr.Outline.Height, BWAttr)
	}

	pr.PromptPanel.Draw()
	pr.Edit.Draw()
}

func (pr *Prompt) HandleEvent(e *tb.Event) (Widget, WidgetEventID) {
	if e.Type == tb.EventKey {
		if e.Key == tb.KeyEnter {
			return pr, PromptOK
		}
		if e.Key == tb.KeyEsc {
			return pr, PromptCancel
		}
	}

	return pr.Edit.HandleEvent(e)
}

func (pr *Prompt) Pos() Pos {
	return Pos{pr.Outline.X, pr.Outline.Y}
}
func (pr *Prompt) Size() Size {
	return Size{pr.Outline.Width, pr.Outline.Height}
}
func (pr *Prompt) Area() Area {
	return NewArea(pr.Outline.X, pr.Outline.Y, pr.Outline.Width, pr.Outline.Height)
}

func (pr *Prompt) SetPos(x, y int) {
	var borderWidth int
	if pr.Mode&PromptBorder != 0 {
		borderWidth = 1
	}
	paddingWidth := 0
	pr.Outline, pr.Content = adjPos(pr.Outline, pr.Content, x, y, borderWidth, paddingWidth)

	pr.PromptPanel.SetPos(pr.Content.X, pr.Content.Y)
	pr.Edit.SetPos(pr.Content.X, pr.Content.Y+2)
}
