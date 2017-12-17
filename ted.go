package main

import (
	"fmt"
	"log"
	"os"

	tb "github.com/nsf/termbox-go"
)

var _log *log.Logger

var StatusPanel *Panel
var CmdPrompt *Prompt
var EditV *EditView

type WhichFocus int

const (
	EditFocus WhichFocus = iota
	CmdFocus
)

type TedEvent int

const (
	TENone TedEvent = iota
	TEExit
)

type UIState struct {
	Focus       WhichFocus
	Prompt      string
	HotKeyTips  string
	CurrentFile string
}

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

	// Main text edit view
	buf := NewBuf()
	buf.WriteLine("Now is the time for all good men to come to the aid of the party.")
	buf.WriteLine("aaa")
	buf.WriteLine("zzz")
	buf.WriteLine(" ")
	buf.WriteLine("Word1 a the at supercalifragilisticexpialidocious, and a somewhatlongerwordisalso")
	buf.WriteLine("The quick brown fox jumps over the lazy dog.")
	buf.WriteLine("Last line!")

	termW, termH := tb.Size()

	editAttr := TermAttr{tb.ColorWhite, tb.ColorBlack}
	statusAttr := TermAttr{tb.ColorBlack, tb.ColorWhite}
	EditV = NewEditView(0, 0, termW, termH-5, EditViewBorder|EditViewStatusLine, editAttr, statusAttr, buf)

	statusArea := NewArea(EditV.Pos().X, EditV.Pos().Y+EditV.Size().Height, EditV.Size().Width, 5)

	// Status panel
	StatusPanel = NewPanel(statusArea.X, statusArea.Y, statusArea.Width, statusArea.Height, true)

	// Prompt panel
	CmdPrompt = NewPrompt("", statusArea.X, statusArea.Y, statusArea.Width-2, 1, true)

	uis := UIState{}
	uis.Focus = EditFocus
	uis.HotKeyTips = "^O: Open File  ^S: Save File  ^Q: Quit"
	uis.CurrentFile = "noname.txt"

	Draw(uis)

	for {
		e := tb.PollEvent()
		if e.Type == tb.EventKey {
			// CTRL-Q: Quit
			if e.Key == tb.KeyCtrlQ {
				break
			}

			// ESC:
			if e.Key == tb.KeyEsc {
				if uis.Focus == EditFocus {
					// ESC on editor
				} else {
					// ESC anywhere else, back to editor
					uis.Focus = EditFocus
				}
			}

			// CTRL-O: Open File prompt
			if e.Key == tb.KeyCtrlO {
				uis.Focus = CmdFocus
				CmdPrompt.SetPrompt("Open file:")
				CmdPrompt.SetEdit("")
			}

			// CTRL-S: Save File
			if e.Key == tb.KeyCtrlS {
				err := EditV.Buf.Save()
				if err != nil {
					uis.Focus = CmdFocus
					serr := fmt.Sprintf("Error writing file (%s), ESC to cancel.", err)
					CmdPrompt.SetPrompt(serr)
					CmdPrompt.SetEdit("")
				} else {
					uis.Focus = CmdFocus
					CmdPrompt.SetPrompt("File saved. Hit ESC.")
					CmdPrompt.SetEdit("")
				}
			}
		}

		if uis.Focus == EditFocus {
			// Editor receives events.
			EditV.HandleEvent(&e)
		} else if uis.Focus == CmdFocus {
			// Prompt receives events.
			tevt := CmdPrompt.HandleEvent(&e)
			if tevt == TEExit {
				file := CmdPrompt.Text()
				err := EditV.Buf.Load(file)
				if err != nil {
					serr := fmt.Sprintf("Error opening file (%s), ESC to cancel.", err)
					CmdPrompt.SetPrompt(serr)
					CmdPrompt.SetEdit("")
				} else {
					EditV.SetText(EditV.Buf.GetText())
					uis.Focus = EditFocus
				}
			}
		}

		Draw(uis)
	}
}

func UpdateStatusPanel(bufPos Pos, tips string) {
	s := fmt.Sprintf("x:%d y:%d\n%s", bufPos.X, bufPos.Y, tips)
	StatusPanel.SetText(s)
}

func Draw(uis UIState) {
	tb.Clear(0, 0)

	// Editor is always visible.
	EditV.Draw()

	if uis.Focus == EditFocus {
		// Refresh StatusPanel.
		UpdateStatusPanel(EditV.BufPos(), uis.HotKeyTips)
		StatusPanel.Draw()

		EditV.DrawCursor()
	}

	if uis.Focus == CmdFocus {
		// Refresh CmdPrompt.
		CmdPrompt.Draw()

		CmdPrompt.DrawCursor()
	}

	flush()
}
