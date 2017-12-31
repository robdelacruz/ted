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
func (ts *TextSurface) ClearLine(y int, c rune) {
	if y < 0 || y > len(ts.Lines)-1 {
		return
	}
	for x := 0; x < ts.W; x++ {
		ts.Lines[y][x] = c
	}
}

func (ts *TextSurface) ResizeLines(n int) {
	nLines := len(ts.Lines)
	if n < nLines {
		ts.Lines = ts.Lines[:n]
	}
	if n > nLines {
		for y := nLines; y < n-nLines; y++ {
			ts.Lines = append(ts.Lines, make([]rune, ts.W))
			ts.ClearLine(y, 0)
		}
	}
	ts.H = n
}

func (ts *TextSurface) WriteString(s string, x, y int) {
	if x < 0 || x > ts.W-1 {
		return
	}
	if y < 0 {
		return
	}

	if y > ts.H-1 {
		ts.Lines = append(ts.Lines, make([]rune, ts.W))
		ts.H = len(ts.Lines)
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
