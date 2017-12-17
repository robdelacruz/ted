package main

import (
	"fmt"
	"log"
	"os"

	tb "github.com/nsf/termbox-go"
)

var _log *log.Logger

var CmdPrompt *Prompt
var EditV *EditView
var SplashPanel *Panel

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
	EditV = NewEditView(0, 0, termW, termH, EditViewBorder|EditViewStatusLine, editAttr, statusAttr, buf)

	// Prompt panel
	qAttr := TermAttr{tb.ColorBlack, tb.ColorYellow}
	ansAttr := TermAttr{tb.ColorGreen, tb.ColorYellow}
	CmdPrompt = NewPrompt(0, termH-5, termW, 5, 0, "", qAttr, ansAttr)

	// Splash panel
	sSplash := `Now is the time for all good men to come to the aid of the party. The quick brown fox jumps over the lazy dog. Now is the time for all good men to come to the aid of the party. The quick brown fox jumps over the lazy dog. Now is the time for all good men to come to the aid of the party. The quick brown fox jumps over the lazy dog. 

Now is the time for all good men to come to the aid of the party. The quick brown fox jumps over the lazy dog. Now is the time for all good men to come to the aid of the party. The quick brown fox jumps over the lazy dog.`

	SplashPanel = NewPanel(10, 15, 55, 18, PanelBorder, TermAttr{tb.ColorRed, tb.ColorWhite}, sSplash)

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

					fullSizeEditV()
				}
			}

			// CTRL-O: Open File prompt
			if e.Key == tb.KeyCtrlO {
				uis.Focus = CmdFocus
				CmdPrompt.SetPrompt("Open file:")
				CmdPrompt.SetEdit("")

				minSizeEditV()
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
					EditV.SyncText()
					EditV.Cur = Pos{0, 0}
					uis.Focus = EditFocus

					fullSizeEditV()
				}
			}
		}

		Draw(uis)
	}
}

func fullSizeEditV() {
	editBox := EditV.Area()
	cmdBox := CmdPrompt.Area()
	EditV.Resize(editBox.X, editBox.Y, editBox.Width, editBox.Height+cmdBox.Height-1)
}

func minSizeEditV() {
	editBox := EditV.Area()
	cmdBox := CmdPrompt.Area()
	EditV.Resize(editBox.X, editBox.Y, editBox.Width, editBox.Height-cmdBox.Height+1)
}

func Draw(uis UIState) {
	tb.Clear(0, 0)

	// Editor is always visible.
	EditV.Draw()

	if uis.Focus == EditFocus {
		EditV.DrawCursor()
	}

	if uis.Focus == CmdFocus {
		// Refresh CmdPrompt.
		CmdPrompt.Draw()
		CmdPrompt.DrawCursor()

		SplashPanel.Draw()
	}

	flush()
}
