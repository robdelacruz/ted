package main

import ()

type TextSurface struct {
	W, H  int
	Lines [][]rune
}

func NewTextSurface(w, h int) *TextSurface {
	ts := &TextSurface{}
	ts.W = w
	ts.H = h

	for i := 0; i < h; i++ {
		line := make([]rune, w)
		ts.Lines = append(ts.Lines, line)
	}

	return ts
}

func (ts *TextSurface) Clear(c rune) {
	for y := 0; y < ts.H; y++ {
		tsLine := ts.Lines[y]
		for x := 0; x < ts.W; x++ {
			tsLine[x] = c
		}
	}
}

func (ts *TextSurface) WriteString(s string, x, y int) {
	if y < 0 || y > ts.H-1 {
		return
	}
	if x < 0 || x > ts.W-1 {
		return
	}

	tsLine := ts.Lines[y]
	for _, c := range s {
		tsLine[x] = c

		x++
		if x > ts.W-1 {
			break
		}
	}
}

func (ts *TextSurface) ch(x, y int) rune {
	if y < 0 || y > ts.H-1 || x < 0 || x > ts.W-1 {
		return 0
	}

	return ts.Lines[y][x]
}

func (ts *TextSurface) Char(x, y int) rune {
	ch := ts.ch(x, y)
	if ch == 0 {
		return ' '
	}
	return ch
}
