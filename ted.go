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
	Focus WhichFocus
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
	EditV = NewEditView(0, 0, termW, termH-5, true, buf)

	statusArea := NewArea(EditV.Pos().X, EditV.Pos().Y+EditV.Size().Height, EditV.Size().Width, 5)

	// Status panel
	StatusPanel = NewPanel(statusArea.X, statusArea.Y, statusArea.Width, statusArea.Height, true)

	// Prompt panel
	CmdPrompt = NewPrompt("Open file:", statusArea.X, statusArea.Y, statusArea.Width-2, 1, true)

	uiState := UIState{}
	uiState.Focus = EditFocus

	Draw(uiState)

	for {
		e := tb.PollEvent()
		if e.Type == tb.EventKey {
			if e.Key == tb.KeyCtrlQ {
				break
			}
			if e.Key == tb.KeyEsc {
				if uiState.Focus == EditFocus {
					uiState.Focus = CmdFocus
				} else {
					uiState.Focus = EditFocus
				}
			}
		}

		if uiState.Focus == EditFocus {
			EditV.HandleEvent(&e)
		} else if uiState.Focus == CmdFocus {
			tevt := CmdPrompt.HandleEvent(&e)
			if tevt == TEExit {
				_log.Printf("CmdPrompt: %s\n", CmdPrompt.Text())
				uiState.Focus = EditFocus
			}
		}

		Draw(uiState)
	}

}

func Draw(uiState UIState) {
	tb.Clear(0, 0)

	EditV.Draw()

	if uiState.Focus == EditFocus {
		bufPos := EditV.BufPos()
		StatusPanel.Clear()
		StatusPanel.WriteLine(fmt.Sprintf("x:%d y:%d", bufPos.X, bufPos.Y))
		StatusPanel.Draw()

		EditV.DrawCursor()
	}

	if uiState.Focus == CmdFocus {
		CmdPrompt.Draw()
		CmdPrompt.DrawCursor()
	}

	flush()
}
