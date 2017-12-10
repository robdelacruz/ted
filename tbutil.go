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

func print(s string, x, y int, fg, bg tb.Attribute) {
	for _, c := range s {
		tb.SetCell(x, y, c, fg, bg)
		x++
	}
}

func drawBox(x, y, width, height int, fg, bg tb.Attribute) {
	print("┌", x, y, fg, bg)
	print("┐", x+width-1, y, fg, bg)

	hline := strings.Repeat("─", width-2)
	print(hline, x+1, y, fg, bg)
	print(hline, x+1, y+height-1, fg, bg)

	vchar := "│"
	for j := y + 1; j < y+height-1; j++ {
		print(vchar, x, j, fg, bg)
	}
	for j := y + 1; j < y+height-1; j++ {
		print(vchar, x+width-1, j, fg, bg)
	}

	print("┘", x+width-1, y+height-1, fg, bg)
	print("└", x, y+height-1, fg, bg)
}

func runeslen(s string) int {
	return len([]rune(s))
}
