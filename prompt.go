package main

import (
	tb "github.com/nsf/termbox-go"
)

// Panel prompt for text, interactive textual entry using textview.

type Prompt struct {
	Rect
	QPanel      *Panel
	Edit        *EditView
	HintPanel   *Panel
	StatusPanel *Panel

	Mode PromptMode
	Opts *PromptOptions
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
	outlineRect := NewRect(x, y, w, 0)

	if opts.QHeight == 0 {
		opts.QHeight = 1
	}
	if opts.AnsHeight == 0 {
		opts.AnsHeight = 1
	}

	if mode&PromptBorder != 0 {
		w -= 2
	}
	w -= opts.ContentPadding * 2

	offsX := 0
	offsY := 0

	qOpts := PanelOptions{opts.QText, opts.QAttr, 0}
	qPanel := NewPanel(offsX, offsY, w, opts.QHeight, qOpts)
	offsY += qPanel.Rect.H

	edit := NewEditView(offsX, offsY, w, opts.AnsHeight, 0, opts.AnsAttr, BWAttr, nil)
	offsY += edit.Rect.H

	var hintPanel *Panel
	if opts.HintHeight > 0 {
		hintOpts := PanelOptions{opts.HintText, opts.HintAttr, 0}
		hintPanel = NewPanel(offsX, offsY, w, opts.HintHeight, hintOpts)

		offsY += hintPanel.Rect.H
	}

	var statusPanel *Panel
	if opts.StatusHeight > 0 {
		offsY++
		statusOpts := PanelOptions{opts.StatusText, opts.StatusAttr, 0}
		statusPanel = NewPanel(offsX, offsY, w, opts.StatusHeight, statusOpts)

		offsY += statusPanel.Rect.H
	}

	pr := &Prompt{}
	pr.Rect = outlineRect
	pr.Mode = mode
	pr.Opts = opts
	pr.QPanel = qPanel
	pr.Edit = edit
	pr.HintPanel = hintPanel
	pr.StatusPanel = statusPanel
	return pr
}

func (pr *Prompt) SetPrompt(s string) {
	pr.QPanel.SetText(s)
}
func (pr *Prompt) GetPrompt() string {
	return pr.QPanel.Text()
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
func (pr *Prompt) GetHint() string {
	if pr.HintPanel != nil {
		return pr.HintPanel.Text()
	}
	return ""
}
func (pr *Prompt) SetStatus(s string) {
	if pr.StatusPanel != nil {
		pr.StatusPanel.SetText(s)
	}
}
func (pr *Prompt) GetStatus() string {
	if pr.StatusPanel != nil {
		return pr.StatusPanel.Text()
	}
	return ""
}
func (pr *Prompt) Clear() {
	pr.SetPrompt("")
	pr.SetEdit("")
	pr.SetHint("")
	pr.SetStatus("")
}

func (pr *Prompt) contentRect() Rect {
	rect := pr.Rect

	rect.H = pr.QPanel.Rect.H + pr.Edit.Rect.H
	if pr.HintPanel != nil && pr.GetHint() != "" {
		rect.H += pr.HintPanel.Rect.H
	}
	if pr.StatusPanel != nil && pr.GetStatus() != "" {
		rect.H += pr.StatusPanel.Rect.H
	}

	if pr.Mode&PromptBorder != 0 {
		rect.X++
		rect.Y++
		rect.W -= 2
	}

	rect.X += pr.Opts.ContentPadding
	rect.Y += pr.Opts.ContentPadding
	rect.W -= pr.Opts.ContentPadding * 2

	return rect
}

func (pr *Prompt) outlineRect() Rect {
	rect := pr.Rect

	rect.H = pr.contentRect().H
	if pr.Mode&PromptBorder != 0 {
		rect.H += 2
	}
	rect.H += pr.Opts.ContentPadding * 2

	return rect
}

func (pr *Prompt) Height() int {
	return pr.outlineRect().H
}

func (pr *Prompt) Draw() {
	clearRect(pr.outlineRect(), pr.Opts.QAttr)

	rect := pr.contentRect()

	// Save offset positions of child widgets,
	// to be restored after drawing.
	xQPanel, yQPanel := pr.QPanel.X, pr.QPanel.Y
	xEdit, yEdit := pr.Edit.X, pr.Edit.Y
	xHintPanel, yHintPanel := pr.HintPanel.X, pr.HintPanel.Y
	xStatusPanel, yStatusPanel := pr.StatusPanel.X, pr.StatusPanel.Y

	pr.QPanel.X += rect.X
	pr.QPanel.Y += rect.Y
	pr.QPanel.Draw()

	pr.Edit.X += rect.X
	pr.Edit.Y += rect.Y
	pr.Edit.Draw()

	if pr.HintPanel != nil && pr.GetHint() != "" {
		pr.HintPanel.X += rect.X
		pr.HintPanel.Y += rect.Y
		pr.HintPanel.Draw()
	}
	if pr.StatusPanel != nil && pr.GetStatus() != "" {
		pr.StatusPanel.X += rect.X
		pr.StatusPanel.Y += rect.Y
		pr.StatusPanel.Draw()
	}

	// Restore offset positions of child widgets.
	pr.QPanel.X, pr.QPanel.Y = xQPanel, yQPanel
	pr.Edit.X, pr.Edit.Y = xEdit, yEdit
	pr.HintPanel.X, pr.HintPanel.Y = xHintPanel, yHintPanel
	pr.StatusPanel.X, pr.StatusPanel.Y = xStatusPanel, yStatusPanel

	if pr.Mode&PromptBorder != 0 {
		outlineRect := pr.outlineRect()
		drawBox(outlineRect.X, outlineRect.Y, outlineRect.W, outlineRect.H, BWAttr)
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
