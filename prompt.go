package main

import (
	tb "github.com/nsf/termbox-go"
)

// Panel prompt for text, interactive textual entry using textview.

type Prompt struct {
	QPanel      *Panel
	Edit        *EditView
	HintPanel   *Panel
	StatusPanel *Panel

	Outline Area
	Content Area
	Mode    PromptMode
	Opts    *PromptOptions
}

type PromptMode uint

const (
	PromptBorder PromptMode = 1 << iota
)

const (
	PromptOK WidgetEventID = iota
	PromptCancel
)

type PromptOptions struct {
	ContentPadding int
	QText          string
	QAttr          TermAttr
	QHeight        int
	AnsText        string
	AnsAttr        TermAttr
	AnsHeight      int
	HintText       string
	HintAttr       TermAttr
	HintHeight     int
	StatusText     string
	StatusAttr     TermAttr
	StatusHeight   int
}

func NewPrompt(x, y, w int, mode PromptMode, opts *PromptOptions) *Prompt {
	outline := NewArea(x, y, w, 0)
	var borderWidth int
	if mode&PromptBorder != 0 {
		borderWidth = 1
	}
	content := NewArea(x+borderWidth+opts.ContentPadding, y+borderWidth+opts.ContentPadding, w-borderWidth*2-opts.ContentPadding*2, 0)

	if opts.QHeight == 0 {
		opts.QHeight = 1
	}
	if opts.AnsHeight == 0 {
		opts.AnsHeight = 1
	}

	x = content.X
	y = content.Y
	w = content.Width

	qOpts := PanelOptions{opts.QText, opts.QAttr, 0}
	qPanel := NewPanel(x, y, w, opts.QHeight, qOpts)

	y += qPanel.Area().Height
	edit := NewEditView(x, y, w, opts.AnsHeight, 0, opts.AnsAttr, BWAttr, nil)

	y += edit.Area().Height

	var hintPanel *Panel
	if opts.HintHeight > 0 {
		hintOpts := PanelOptions{opts.HintText, opts.HintAttr, 0}
		hintPanel = NewPanel(x, y, w, opts.HintHeight, hintOpts)

		y += hintPanel.Area().Height
	}

	var statusPanel *Panel
	if opts.StatusHeight > 0 {
		y++
		statusOpts := PanelOptions{opts.StatusText, opts.StatusAttr, 0}
		statusPanel = NewPanel(x, y, w, opts.StatusHeight, statusOpts)

		y += statusPanel.Area().Height
	}

	content.Height = y - content.Y
	outline.Height = content.Height + borderWidth*2 + opts.ContentPadding*2

	pr := &Prompt{}
	pr.Outline = outline
	pr.Content = content
	pr.Mode = mode
	pr.Opts = opts
	pr.QPanel = qPanel
	pr.Edit = edit
	pr.HintPanel = hintPanel
	pr.StatusPanel = statusPanel
	return pr
}

func (pr *Prompt) SetPos(x, y int) {
	var borderWidth int
	if pr.Mode&PromptBorder != 0 {
		borderWidth = 1
	}
	pr.Outline, pr.Content = adjPos(pr.Outline, pr.Content, x, y, borderWidth, pr.Opts.ContentPadding)

	x = pr.Content.X
	y = pr.Content.Y
	pr.QPanel.SetPos(x, y)
	y += pr.QPanel.Area().Height
	pr.Edit.SetPos(x, y)
	y += pr.Edit.Area().Height

	if pr.HintPanel != nil {
		pr.HintPanel.SetPos(x, y)
		y += pr.HintPanel.Area().Height
	}
	if pr.StatusPanel != nil {
		y++
		pr.StatusPanel.SetPos(x, y)
	}
}

func (pr *Prompt) SetPrompt(s string) {
	pr.QPanel.SetText(s)
}
func (pr *Prompt) GetPrompt() string {
	return pr.QPanel.GetText()
}
func (pr *Prompt) SetEdit(s string) {
	pr.Edit.SetText(s)
}
func (pr *Prompt) GetEditText() string {
	return pr.Edit.Text()
}
func (pr *Prompt) SetHint(s string) {
	if pr.HintPanel != nil {
		pr.HintPanel.SetText(s)
	}
}
func (pr *Prompt) GetHint(s string) string {
	if pr.HintPanel != nil {
		return pr.HintPanel.GetText()
	}
	return ""
}
func (pr *Prompt) SetStatus(s string) {
	if pr.StatusPanel != nil {
		pr.StatusPanel.SetText(s)
	}
}
func (pr *Prompt) GetStatus(s string) string {
	if pr.StatusPanel != nil {
		return pr.StatusPanel.GetText()
	}
	return ""
}
func (pr *Prompt) Clear() {
	pr.SetPrompt("")
	pr.SetEdit("")
	pr.SetHint("")
	pr.SetStatus("")
}

func (pr *Prompt) Draw() {
	clearArea(pr.Outline, pr.Opts.QAttr)
	if pr.Mode&PromptBorder != 0 {
		drawBox(pr.Outline.X, pr.Outline.Y, pr.Outline.Width, pr.Outline.Height, BWAttr)
	}

	pr.QPanel.Draw()
	pr.Edit.Draw()
	if pr.HintPanel != nil {
		pr.HintPanel.Draw()
	}
	if pr.StatusPanel != nil {
		pr.StatusPanel.Draw()
	}
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
