package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	tb "github.com/nsf/termbox-go"
)

var _log *log.Logger

func main() {
	flog, err := os.Create("./log.txt")
	if err != nil {
		panic(err)
	}
	defer flog.Close()
	_log = log.New(flog, "", 0)

	err = tb.Init()
	if err != nil {
		panic(err)
	}
	defer tb.Close()

	termW, termH := tb.Size()

	// Main text edit view
	editBuf := NewBuf()
	editBuf.SetText(`



ted - A terminal text editor

`)
	editAttr := TermAttr{tb.ColorWhite, tb.ColorBlack}
	statusAttr := TermAttr{tb.ColorBlack, tb.ColorWhite}
	editW := NewEditView(0, 0, termW, termH, EditViewBorder|EditViewStatusLine, editAttr, statusAttr, editBuf)
	editLI := NewLayoutItem(editW, true)

	// Prompt panel
	qAttr := TermAttr{tb.ColorWhite, tb.ColorBlack}
	ansAttr := TermAttr{tb.ColorBlack, tb.ColorYellow}
	hintAttr := TermAttr{tb.ColorCyan, tb.ColorBlack}
	statusPromptAttr := TermAttr{tb.ColorRed, tb.ColorBlack}
	promptOpts := PromptOptions{
		ContentPadding: 1,
		QAttr:          qAttr,
		QHeight:        1,
		AnsAttr:        ansAttr,
		AnsHeight:      1,
		HintAttr:       hintAttr,
		HintHeight:     1,
		StatusAttr:     statusPromptAttr,
		StatusHeight:   2,
	}
	promptWWidth := termW / 2
	promptW := NewPrompt(0, 0, promptWWidth, PromptBorder, &promptOpts)
	promptW.SetPos(termW/2-promptWWidth/2, termH/2-promptW.Area().Height)
	promptLI := NewLayoutItem(promptW, false)

	// About panel
	sAbout := `ted - A terminal text editor
    by Rob de la Cruz

    Thanks to termbox-go library`

	aboutOpts := PanelOptions{sAbout, TermAttr{tb.ColorRed, tb.ColorWhite}, PanelBorder}
	aboutW := NewPanel(10, 15, 55, 18, aboutOpts)
	aboutLI := NewLayoutItem(aboutW, false)

	layout := NewLayout()
	layout.AddItem(editLI)
	layout.AddItem(promptLI)
	layout.AddItem(aboutLI)
	layout.SetFocusItem(editLI)

	tb.Clear(0, 0)
	layout.Draw()
	flush()

	for {
		e := tb.PollEvent()
		if e.Type == tb.EventKey {
			// CTRL-Q: Quit
			if e.Key == tb.KeyCtrlQ {
				break
			}

			// CTRL-O: Open File
			if e.Key == tb.KeyCtrlO {
				promptW.SetPrompt("Open file:")
				promptW.SetHint("<ENTER> to Open, <ESC> to Cancel")
				promptW.SetStatus("")

				promptLI.Visible = true
				layout.SetFocusItem(promptLI)
			}

			// CTRL-S: Save File
			if e.Key == tb.KeyCtrlS {
				promptW.SetPrompt("Save file:")
				promptW.SetHint("<ENTER> to Open, <ESC> to Cancel")
				promptW.SetStatus("")

				file := editBuf.Name
				promptW.SetEdit(file)

				promptLI.Visible = true
				layout.SetFocusItem(promptLI)
			}

			w, evid := layout.HandleEvent(&e)
			switch w := w.(type) {
			case *Prompt:
				promptW := w

				if evid == PromptCancel {
					promptW.SetEdit("")
					layout.SetFocusItem(editLI)
					promptLI.Visible = false
				} else if evid == PromptOK {
					prompt := strings.TrimSpace(promptW.GetPrompt())
					_log.Printf("prompt = '%s'\n", prompt)
					if prompt == "Open file:" {
						file := promptW.GetEditText()
						err := editBuf.Load(file)
						if err != nil {
							serr := fmt.Sprintf("Error (%s).", err)
							promptW.SetStatus(serr)
							promptW.SetEdit("")
						} else {
							promptW.Clear()
							layout.SetFocusItem(editLI)
							editW.ResetCur()
							editW.SyncText()
							promptLI.Visible = false
						}
					}

					if prompt == "Save file:" {
						file := promptW.GetEditText()
						editBuf.Name = file
						err := editBuf.Save(file)
						if err != nil {
							serr := fmt.Sprintf("Error (%s).", err)
							promptW.SetStatus(serr)
							promptW.SetEdit("")
						} else {
							promptW.Clear()
							layout.SetFocusItem(editLI)
							promptLI.Visible = false
						}
					}
				}
			}
		}

		tb.Clear(0, 0)
		layout.Draw()
		flush()
	}
}
