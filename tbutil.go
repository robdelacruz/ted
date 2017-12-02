package main

import (
	tb "github.com/nsf/termbox-go"
)

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

func strToRuneCells(s string, fg, bg tb.Attribute) []*EdCell {
	var cells []*EdCell
	for _, c := range s {
		cells = append(cells, &EdCell{c, fg, bg})
	}
	return cells
}

func printCell(x, y int, cell *EdCell) {
	tb.SetCell(x, y, cell.Ch, cell.Fg, cell.Bg)
}

func printCells(x, y int, cells []*EdCell) {
	for _, cell := range cells {
		printCell(x, y, cell)
		x++
	}
}
