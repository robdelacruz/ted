package main

import (
	"fmt"
	"strings"

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

func min(n1, n2 int) int {
	if n1 < n2 {
		return n1
	}
	return n2
}
