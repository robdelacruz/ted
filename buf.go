package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
)

type Buf struct {
	Name  string
	Dirty bool
	Lines []string
}

func NewBuf() *Buf {
	buf := &Buf{}
	buf.Dirty = false
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
	if buf.InBounds(x, y) {
		return true
	}

	line := []rune(buf.Lines[y])
	if x <= len(line) {
		return true
	}

	return false
}

func (buf *Buf) PrevChar(bufPos Pos) Pos {
	if !buf.InWriteBounds(bufPos.X, bufPos.Y) {
		return bufPos
	}

	if bufPos.X > 0 {
		bufPos.X--
		return bufPos
	}

	if bufPos.Y > 0 {
		bufPos.Y--
		line := []rune(buf.Lines[bufPos.Y])
		bufPos.X = len(line)
		if bufPos.X < 0 {
			bufPos.X = 0
		}
		return bufPos
	}

	return bufPos
}
func (buf *Buf) NextChar(bufPos Pos) Pos {
	if !buf.InWriteBounds(bufPos.X, bufPos.Y) {
		return bufPos
	}

	line := []rune(buf.Lines[bufPos.Y])
	nline := len(line)
	if bufPos.X < nline {
		bufPos.X++
		return bufPos
	}

	if bufPos.Y < len(buf.Lines) {
		bufPos.Y++
		bufPos.X = 0
		return bufPos
	}

	return bufPos
}

func (buf *Buf) Clear() {
	buf.Lines = []string{}
}

func (buf *Buf) WriteLine(s string) {
	buf.Lines = append(buf.Lines, s)
}

func (buf *Buf) SetText(s string) {
	buf.Clear()

	b := bytes.NewBufferString(s)
	scanner := bufio.NewScanner(b)
	for scanner.Scan() {
		buf.WriteLine(scanner.Text())
	}

	// Always have at least one line.
	if len(buf.Lines) == 0 {
		buf.WriteLine("")
	}
}

func (buf *Buf) GetText() string {
	var b bytes.Buffer

	for i, l := range buf.Lines {
		b.WriteString(l)
		if i < len(buf.Lines)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (buf *Buf) Load(file string) error {
	bs, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	buf.ClearDirty()
	buf.Name = file
	buf.SetText(string(bs))
	return nil
}

// Writes contents to filename as indicated in buf.Name.
func (buf *Buf) Save(file string) error {
	if file == "" {
		return errors.New("No filename given")
	}

	bs := []byte(buf.GetText())
	err := ioutil.WriteFile(file, bs, 0644)
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	buf.Name = file
	buf.ClearDirty()
	return nil
}

func (buf *Buf) SetDirty() {
	buf.Dirty = true
}
func (buf *Buf) ClearDirty() {
	buf.Dirty = false
}

func (buf *Buf) InsEOL(x, y int) (bufPos Pos) {
	if !buf.InWriteBounds(x, y) {
		return Pos{x, y}
	}

	buf.SetDirty()

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

	buf.SetDirty()

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

	buf.SetDirty()

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

	buf.SetDirty()

	return Pos{xBuf, yBuf}
}

func (buf *Buf) DelChar(x, y int) (bufPos Pos) {
	return buf.DelChars(x, y, 1)
}
func (buf *Buf) DelChars(x, y, n int) (bufPos Pos) {
	if !buf.InWriteBounds(x, y) {
		return Pos{x, y}
	}

	buf.SetDirty()

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

	return Pos{x, y}
}

//$$ Hack until buf.PrevPos() added, which should recognize the CR
//   as a buf position.
func (buf *Buf) DelPrevChar(x, y int) (bufPos Pos) {
	if !buf.InWriteBounds(x, y) {
		return Pos{x, y}
	}

	buf.SetDirty()

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

	buf.SetDirty()

	copy(buf.Lines[y:], buf.Lines[y+1:])
	buf.Lines = buf.Lines[:len(buf.Lines)-1]
}

func (buf *Buf) MergeLines(y1, y2 int) {
	if y1 < 0 || y1 > len(buf.Lines)-1 ||
		y2 < 0 || y2 > len(buf.Lines)-1 {
		return
	}

	buf.SetDirty()

	buf.Lines[y1] += buf.Lines[y2]
	buf.DelLine(y2)
}
