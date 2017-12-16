package main

import (
	tb "github.com/nsf/termbox-go"
)

// Panel prompt for text, interactive textual entry using textview.

type Prompt struct {
	PromptPanel *Panel
	Edit        *EditView

	Outline  Area
	fOutline bool
}

func NewPrompt(prompt string, x, y, wEdit, hEdit int, fOutline bool) *Prompt {
	var borderW int
	if fOutline {
		borderW = 1
	}
	promptPanel := NewPanel(x+borderW, y+borderW, wEdit, 1, false)
	promptPanel.WriteLine(prompt)
	ppPos, ppSize := promptPanel.Pos(), promptPanel.Size()

	edit := NewEditView(ppPos.X, ppPos.Y+1, wEdit, hEdit, 0, nil)
	editSize := edit.Size()

	outline := NewArea(ppPos.X-borderW, ppPos.Y-borderW, ppSize.Width+2*borderW, ppSize.Height+editSize.Height+2*borderW)

	pr := &Prompt{}
	pr.PromptPanel = promptPanel
	pr.Edit = edit
	pr.Outline = outline
	pr.fOutline = fOutline
	return pr
}

func (pr *Prompt) SetEdit(s string) {
	pr.Edit.Clear()
	pr.Edit.WriteLine(s)
	pr.Edit.SyncText()
}
func (pr *Prompt) SetPrompt(prompt string) {
	pr.PromptPanel.Clear()
	pr.PromptPanel.WriteLine(prompt)
}

func (pr *Prompt) Text() string {
	return pr.Edit.Text()
}

func (pr *Prompt) Draw() {
	if pr.fOutline {
		drawBox(pr.Outline.X, pr.Outline.Y, pr.Outline.Width, pr.Outline.Height, BWAttr)
	}

	pr.PromptPanel.Draw()
	pr.Edit.Draw()
}

func (pr *Prompt) DrawCursor() {
	pr.Edit.DrawCursor()
}

func (pr *Prompt) HandleEvent(e *tb.Event) TedEvent {
	if e.Type == tb.EventKey {
		if e.Key == tb.KeyEnter {
			return TEExit
		}
	}

	pr.Edit.HandleEvent(e)
	return TENone
}
