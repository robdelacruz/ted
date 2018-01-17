package main

// Globals
// -------
// _log
//
// Functions
// ---------
// main()
//

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

	// Last search string
	var searchS string

	// Main text edit view
	editBuf := NewBuf()
	/*	editBuf.SetText(`



		ted - A terminal text editor

		`)*/

	err = editBuf.Load("sample.txt")
	if err != nil {
		_log.Printf("Error loading buf (%s)\n", err)
	}

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
	promptW.X = termW/2 - promptWWidth/2
	promptW.Y = termH/2 - promptW.Height()
	promptLI := NewLayoutItem(promptW, false)

	// About panel
	aboutOpts := PanelOptions{"", TermAttr{tb.ColorRed, tb.ColorWhite}, PanelBorder}
	aboutW := NewPanel(13, 12, 55, 20, aboutOpts)
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
		e := WaitKBEvent()

		// CTRL-Q: Quit
		if e.Key == tb.KeyCtrlQ {
			break
		}

		// CTRL-H: About and Help
		if e.Key == tb.KeyCtrlH {
			sHelp := `  TED - a console text editor
  by Rob de la Cruz
  https://robdelacruz.github.io/ted.html
  Available under MIT License.

  Commands:
    CTRL-Q: Quit
    CTRL-H: About Ted and Help
    CTRL-O: Open file
    CTRL-S: Save file
    CTRL-F: Search text
    CTRL-G: Repeat last search
    CTRL-K: Select text
    CTRL-C: Copy selected text
    CTRL-X: Cut selected text
    CTRL-V: Paste text

  ESC to return
`
			aboutW.SetText(sHelp)
			aboutLI.Visible = true
			layout.SetFocusItem(aboutLI)
		}

		// CTRL-O: Open File
		if e.Key == tb.KeyCtrlO {
			promptW.SetPrompt("Open file:")
			promptW.SetHint("<ENTER> to Open, <ESC> to Cancel")
			promptW.SetStatus("")

			promptW.X = termW/2 - promptWWidth/2
			promptW.Y = termH/2 - promptW.Height()

			promptLI.Visible = true
			layout.SetFocusItem(promptLI)
		}

		// CTRL-S: Save File
		if e.Key == tb.KeyCtrlS {
			promptW.SetPrompt("Save file:")
			promptW.SetHint("<ENTER> to Save, <ESC> to Cancel")
			promptW.SetStatus("")

			promptW.X = termW/2 - promptWWidth/2
			promptW.Y = termH/2 - promptW.Height()

			file := editBuf.Name
			promptW.SetEdit(file)

			promptLI.Visible = true
			layout.SetFocusItem(promptLI)
		}

		// CTRL-F: Search text
		if e.Key == tb.KeyCtrlF {
			promptW.SetPrompt("Search:")
			promptW.SetHint("<ENTER> to Search, <ESC> to Cancel")
			promptW.SetStatus("")

			promptW.X = termW/2 - promptWWidth/2
			promptW.Y = termH/2 - promptW.Height()

			promptLI.Visible = true
			layout.SetFocusItem(promptLI)
		}

		// CTRL-G: Repeat last search
		if e.Key == tb.KeyCtrlG && searchS != "" {
			editW.SearchForward(searchS)
		}

		w, evid := layout.HandleEvent(&e)
		switch w := w.(type) {
		case *Panel:
			if evid == PanelClose {
				layout.SetFocusItem(editLI)
				aboutLI.Visible = false
			}

		case *Prompt:
			promptW := w

			if evid == PromptCancel {
				promptW.SetEdit("")
				layout.SetFocusItem(editLI)
				promptLI.Visible = false
			} else if evid == PromptOK {
				prompt := strings.TrimSpace(promptW.GetPrompt())
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
						editW.Reset()
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

				if prompt == "Search:" {
					searchS = promptW.GetEditText()
					editW.SearchForward(searchS)

					promptW.Clear()
					layout.SetFocusItem(editLI)
					promptLI.Visible = false
				}
			}
		}

		tb.Clear(0, 0)
		layout.Draw()
		flush()
	}
}
