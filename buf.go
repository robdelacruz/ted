package main

import (
	"bufio"
	"bytes"
)

type Buf struct {
	Lines []string
}

func NewBuf() *Buf {
	buf := &Buf{}
	return buf
}

func (buf *Buf) InBounds(x, y int) bool {
	if y < 0 || x < 0 {
		return false
	}

	nLines := len(buf.Lines)
	if y > nLines-1 {
		return false
	}
	line := []rune(buf.Lines[y])
	if x > len(line)-1 {
		return false
	}
	return true
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

func (buf *Buf) InsStr(s string, x, y int) int {
	if !buf.InWriteBounds(x, y) {
		return x
	}

	// Insert new line with string.
	nLines := len(buf.Lines)
	if y == nLines {
		buf.WriteLine(s)
		return x
	}

	// Replace existing line, insert string.
	var b bytes.Buffer
	line := []rune(buf.Lines[y])
	leftPart := line[:x]
	rightPart := line[x:]

	b.WriteString(string(leftPart))
	b.WriteString(s)
	b.WriteString(string(rightPart))

	buf.Lines[y] = b.String()

	return x + runeslen(s)
}

func (buf *Buf) InsText(s string, x, y int) (bufPos Pos) {
	b := bytes.NewBufferString(s)
	scanner := bufio.NewScanner(b)

	xBuf, yBuf := x, y
	for scanner.Scan() {
		sline := scanner.Text()
		xBuf = buf.InsStr(sline, xBuf, yBuf)
		buf.InsEOL(xBuf, yBuf)

		yBuf++
		xBuf = 0
	}

	return Pos{xBuf, yBuf}
}

func (buf *Buf) DelChar(x, y int) (bufPos Pos) {
	return buf.DelChars(x, y, 1)
}

func (buf *Buf) DelChars(x, y, n int) (bufPos Pos) {
	if !buf.InBounds(x, y) {
		return Pos{x, y}
	}
	// Replace existing line, delete chars.
	line := []rune(buf.Lines[y])
	if x+n > len(line) {
		n = len(line) - x
	}
	copy(line[x:], line[x+n:])
	line = line[:len(line)-n]
	buf.Lines[y] = string(line)

	return Pos{x, y}
}
