package main

import (
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

func flush() {
	err := tb.Flush()
	if err != nil {
		panic(err)
	}

}

func print(x, y int, fg, bg tb.Attribute, s string) {
	for _, c := range s {
		tb.SetCell(x, y, c, fg, bg)
		x++
	}
}
