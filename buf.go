package main

import ()

type Buf struct {
	Lines []string
}

func NewBuf() *Buf {
	buf := &Buf{}
	return buf
}

func (buf *Buf) InWriteBounds(x, y int) bool {
	if y < 0 || x < 0 {
		return false
	}

	// Allow x and y to be written one char outside buffer boundaries
	// (ex. one line below bottom edge or one char past right edge)
	// This allows adding to the buffer by writing to (x,y) 1 line/char
	// beyond the boundaries.
	nLines := len(buf.Lines)
	if y > nLines {
		return false
	}
	if y < nLines && x > len(buf.Lines[y]) {
		return false
	}
	if y == nLines && x > 0 {
		return false
	}
	return true
}

func (buf *Buf) WriteLine(s string) {
	buf.Lines = append(buf.Lines, s)
}

func (buf *Buf) InsChar(c rune, x, y int) {
	if y < 0 || x < 0 {
		return
	}
	nLines := len(buf.Lines)
	if y > nLines {
		return
	}
	line := []rune(buf.Lines[y])
	if y < nLines && x > len(line) {
		return
	}

	// Insert new line with char.
	if y == nLines {
		buf.WriteLine(string(c))
		return
	}

	// Replace existing line, insert char.
	line = append(line, 0)
	if x < len(line)-1 {
		copy(line[x+1:], line[x:])
	}
	line[x] = c
	buf.Lines[y] = string(line)
}
