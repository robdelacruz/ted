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

	if y > buf.NumLines()-1 {
		return false
	}
	_, nline := buf.PosLine(y)
	if x > nline-1 {
		return false
	}
	return true
}

func (buf *Buf) InWriteBounds(x, y int) bool {
	if y > buf.NumLines()-1 {
		return false
	}

	_, nline := buf.PosLine(y)
	if x <= nline {
		return true
	}

	return false
}

func (buf *Buf) NumLines() int {
	return len(buf.Lines)
}

func (buf *Buf) PosLine(y int) ([]rune, int) {
	nLines := len(buf.Lines)
	if y >= 0 && y < nLines {
		line := []rune(buf.Lines[y])
		return line, len(line)
	}
	return []rune{}, 0
}

func (buf *Buf) PrevPos(pos Pos) Pos {
	if !buf.InWriteBounds(pos.X, pos.Y) {
		return pos
	}

	if pos.X > 0 {
		pos.X--
		return pos
	}

	if pos.Y > 0 {
		pos.Y--
		_, nline := buf.PosLine(pos.Y)
		pos.X = nline
		if pos.X < 0 {
			pos.X = 0
		}
		return pos
	}

	return pos
}
func (buf *Buf) NextPos(pos Pos) Pos {
	if !buf.InWriteBounds(pos.X, pos.Y) {
		return pos
	}

	_, nline := buf.PosLine(pos.Y)
	if pos.X < nline {
		pos.X++
		return pos
	}

	if pos.Y < buf.NumLines()-1 {
		pos.Y++
		pos.X = 0
		return pos
	}

	return pos
}

func (buf *Buf) BOLPos(pos Pos) Pos {
	pos.X = 0
	return pos
}

func (buf *Buf) EOLPos(pos Pos) Pos {
	_, nline := buf.PosLine(pos.Y)

	pos.X = nline
	return pos
}

func (buf *Buf) UpPos(pos Pos) Pos {
	if pos.Y <= 0 {
		return pos
	}

	pos.Y--
	_, nline := buf.PosLine(pos.Y)
	if nline == 0 {
		pos.X = 0
	} else if pos.X > nline-1 {
		pos.X = nline - 1
	}

	return pos
}

func (buf *Buf) DownPos(pos Pos) Pos {
	if pos.Y >= buf.NumLines()-1 {
		return pos
	}

	pos.Y++
	_, nline := buf.PosLine(pos.Y)
	if nline == 0 {
		pos.X = 0
	} else if pos.X > nline-1 {
		pos.X = nline - 1
	}

	return pos
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

func processLine(line string, maxlenWrapLine int, cbWord func(word string), cbWrapLine func(wrapline string)) {
	if len(line) == 0 {
		cbWrapLine(line)
		return
	}

	xWL := 0 // x in wrapline

	var bWL bytes.Buffer // current wrapline

	words := parseWords(line)
	for _, w := range words {
		cbWord(w)

		// word can't fit in remaining wrapline, add to next wrapline.
		lenW := len([]rune(w))
		if xWL+lenW > maxlenWrapLine {
			cbWrapLine(bWL.String())

			// Start new wrapline
			xWL = 0
			bWL.Reset()
			bWL.WriteString(w)
			xWL += lenW

			continue
		}

		// add word to remaining wrapline.
		bWL.WriteString(w)
		xWL += lenW
	}

	// Process any leftover wrapline.
	remWL := bWL.String()
	if len(remWL) > 0 {
		cbWrapLine(remWL)
	}
}
