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

func (buf *Buf) Text() string {
	var b bytes.Buffer
	for i, l := range buf.Lines {
		b.WriteString(l)
		if i < len(buf.Lines)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
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
	//	if y == nLines && x > 0 {
	//		return false
	//	}
	return true
}

func (buf *Buf) Clear() {
	buf.Lines = []string{}
}

func (buf *Buf) WriteLine(s string) {
	buf.Lines = append(buf.Lines, s)
}

func (buf *Buf) InsEOL(x, y int) (bufPos Pos) {
	if !buf.InWriteBounds(x, y) {
		return Pos{x, y}
	}

	nLines := len(buf.Lines)
	if y == nLines {
		buf.WriteLine("")
		return Pos{x, y + 1}
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
	return Pos{0, y + 1}
}

func (buf *Buf) InsChar(c rune, x, y int) (bufPos Pos) {
	if !buf.InWriteBounds(x, y) {
		return Pos{x, y}
	}

	// Insert new line with char.
	nLines := len(buf.Lines)
	if y == nLines {
		buf.WriteLine(string(c))
		return Pos{x + 1, y}
	}

	// Replace existing line, insert char.
	line := []rune(buf.Lines[y])
	line = append(line, 0)
	if x < len(line)-1 {
		copy(line[x+1:], line[x:])
	}
	line[x] = c
	buf.Lines[y] = string(line)

	return Pos{x + 1, y}
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
	if !buf.InWriteBounds(x, y) {
		return Pos{x, y}
	}

	line := []rune(buf.Lines[y])
	if x == len(line) || (x == 0 && len(line) == 0) {
		buf.MergeLines(y, y+1)
		n--
	}

	// Replace existing line, delete chars.
	for n > 0 && buf.InBounds(x, y) {
		line = []rune(buf.Lines[y])
		nlinechars := min(n, len(line)-x)

		copy(line[x:], line[x+nlinechars:])
		line = line[:len(line)-nlinechars]
		buf.Lines[y] = string(line)

		n -= nlinechars
		if n > 0 && (x == len(line) || len(line) == 0) {
			buf.MergeLines(y, y+1)
			n--
		}
	}

	/*
		if x+n > len(line) {
			n = len(line) - x
		}
		copy(line[x:], line[x+n:])
		line = line[:len(line)-n]
		buf.Lines[y] = string(line)
	*/

	return Pos{x, y}
}

//$$ Hack until buf.PrevPos() added, which should recognize the CR
//   as a buf position.
func (buf *Buf) DelPrevChar(x, y int) (bufPos Pos) {
	if !buf.InWriteBounds(x, y) {
		return Pos{x, y}
	}

	if x == 0 && y > 0 {
		yLineLen := len([]rune(buf.Lines[y]))
		buf.MergeLines(y-1, y)

		line := []rune(buf.Lines[y-1])
		return Pos{len(line) - yLineLen, y - 1}
	}

	return buf.DelChar(x-1, y)
}

func (buf *Buf) DelLine(y int) {
	if y < 0 || y > len(buf.Lines)-1 {
		return
	}

	copy(buf.Lines[y:], buf.Lines[y+1:])
	buf.Lines = buf.Lines[:len(buf.Lines)-1]
}

func (buf *Buf) MergeLines(y1, y2 int) {
	if y1 < 0 || y1 > len(buf.Lines)-1 ||
		y2 < 0 || y2 > len(buf.Lines)-1 {
		return
	}

	buf.Lines[y1] += buf.Lines[y2]
	buf.DelLine(y2)
}
