package main

import (
	"fmt"
	"strings"
	"unicode"

	tb "github.com/nsf/termbox-go"
)

type Pos struct{ X, Y int }
type Size struct{ W, H int }
type Rect struct{ X, Y, W, H int }

type TermAttr struct{ Fg, Bg tb.Attribute }

var BWAttr TermAttr

func NewRect(x, y, w, h int) Rect {
	return Rect{x, y, w, h}
}
func (rect Rect) String() string {
	return fmt.Sprintf("x: %d, y: %d, w: %d, h %d", rect.X, rect.Y, rect.W, rect.H)
}

func (pos *Pos) String() string {
	return fmt.Sprintf("%d,%d", pos.X, pos.Y)
}

func reverseAttr(attr TermAttr) TermAttr {
	return TermAttr{attr.Bg, attr.Fg}
}

func flush() {
	err := tb.Flush()
	if err != nil {
		panic(err)
	}

}

func print(s string, x, y int, attr TermAttr) {
	for _, c := range s {
		tb.SetCell(x, y, c, attr.Fg, attr.Bg)
		x++
	}
}

func printCh(c rune, x, y int, attr TermAttr) {
	tb.SetCell(x, y, c, attr.Fg, attr.Bg)
}

func clearRect(rect Rect, attr TermAttr) {
	srow := strings.Repeat(" ", rect.W)
	for y := rect.Y; y < rect.Y+rect.H; y++ {
		print(srow, rect.X, y, attr)
	}
}

func drawBox(x, y, width, height int, attr TermAttr) {
	print("┌", x, y, attr)
	print("┐", x+width-1, y, attr)

	hline := strings.Repeat("─", width-2)
	print(hline, x+1, y, attr)
	print(hline, x+1, y+height-1, attr)

	vchar := "│"
	for j := y + 1; j < y+height-1; j++ {
		print(vchar, x, j, attr)
	}
	for j := y + 1; j < y+height-1; j++ {
		print(vchar, x+width-1, j, attr)
	}

	print("┘", x+width-1, y+height-1, attr)
	print("└", x, y+height-1, attr)
}

func runeslen(s string) int {
	return len([]rune(s))
}

func adjPos(outline, content Rect, x, y, borderWidth, paddingWidth int) (retOutline, retContent Rect) {
	retOutline = outline
	retContent = content

	retOutline.X = x
	retOutline.Y = y

	retContent = NewRect(x+borderWidth+paddingWidth, y+borderWidth+paddingWidth, retOutline.W-borderWidth*2-paddingWidth*2, retOutline.H-borderWidth*2-paddingWidth*2)

	return retOutline, retContent
}

func min(ns ...int) int {
	lenNs := len(ns)
	if lenNs == 0 {
		return 0
	}

	smallest := ns[0]
	for _, n := range ns {
		if n < smallest {
			smallest = n
		}
	}
	return smallest
}

func WaitKBEvent() tb.Event {
	for {
		e := tb.PollEvent()
		if e.Type != tb.EventKey {
			continue
		}

		return e
	}

	return tb.Event{}
}

// Parse line to get sequence of words.
// Each whitespace char is considered a single word.
// Ex. "One two  three" => ["One", " ", "two", " ", " ", "three"]
func parseWords(s string) []string {
	var currentWord string
	var words []string

	for _, c := range s {
		if unicode.IsSpace(c) {
			// Add pending word
			words = append(words, currentWord)

			// Add single space word
			words = append(words, string(c))

			currentWord = ""
			continue
		}

		// Add char to pending word
		currentWord += string(c)
	}

	if len(currentWord) > 0 {
		words = append(words, currentWord)
	}

	return words
}

// -1 pos is before posStart
// +1 pos is after posEnd
//  0 pos is within posStart, posEnd
func cmpPosRange(pos, posStart, posEnd Pos) int {
	if pos.Y < posStart.Y {
		return -1
	}
	if pos.Y > posEnd.Y {
		return 1
	}
	if pos.Y == posStart.Y && pos.X < posStart.X {
		return -1
	}
	if pos.Y == posEnd.Y && pos.X > posEnd.X {
		return 1
	}
	return 0
}
