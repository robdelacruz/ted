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
	line := []rune(buf.Lines[y])
	if y < nLines && x > len(line) {
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

func (buf *Buf) InsEOL(x, y int) {
	if !buf.InWriteBounds(x, y) {
		return
	}

	nLines := len(buf.Lines)
	if y == nLines {
		buf.WriteLine("")
		return
	}

	line := []rune(buf.Lines[y])
	var leftPart, rightPart []rune
	leftPart = line[:x]
	if x <= len(line)-1 {
		rightPart = line[x:]
	}

	buf.Lines = append(buf.Lines, "")
	if y < len(buf.Lines)-2 {
		copy(buf.Lines[y+2:], buf.Lines[y+1:])
	}

	buf.Lines[y] = string(leftPart)
	buf.Lines[y+1] = string(rightPart)
}

func (buf *Buf) InsChar(c rune, x, y int) {
	if !buf.InWriteBounds(x, y) {
		return
	}

	// Insert new line with char.
	nLines := len(buf.Lines)
	if y == nLines {
		buf.WriteLine(string(c))
		return
	}

	// Replace existing line, insert char.
	line := []rune(buf.Lines[y])
	line = append(line, 0)
	if x < len(line)-1 {
		copy(line[x+1:], line[x:])
	}
	line[x] = c
	buf.Lines[y] = string(line)
}
