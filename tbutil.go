package main

import (
	"fmt"
	"strings"

	tb "github.com/nsf/termbox-go"
)

type Pos struct{ X, Y int }
type Size struct{ Width, Height int }
type Area struct {
	Pos
	Size
}
type TermAttr struct{ Fg, Bg tb.Attribute }

var BWAttr TermAttr

func NewArea(x, y, w, h int) Area {
	return Area{
		Pos:  Pos{x, y},
		Size: Size{w, h},
	}
}

func (pos *Pos) String() string {
	return fmt.Sprintf("%d,%d", pos.X, pos.Y)
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
