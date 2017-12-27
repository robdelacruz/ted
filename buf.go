package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"unicode"
)

type Buf struct {
	Name  string
	Dirty bool
	Lines []string
}

const _tablen = 4

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

func (buf *Buf) InBounds(pos Pos) bool {
	if pos.Y < 0 || pos.X < 0 {
		return false
	}

	if pos.Y > buf.NumLines()-1 {
		return false
	}
	_, nline := buf.PosLine(pos.Y)
	if pos.X > nline-1 {
		return false
	}
	return true
}

func (buf *Buf) InWriteBounds(pos Pos) bool {
	if pos.Y > buf.NumLines()-1 {
		return false
	}

	_, nline := buf.PosLine(pos.Y)
	if pos.X <= nline {
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
	if !buf.InWriteBounds(pos) {
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
	if !buf.InWriteBounds(pos) {
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

func (buf *Buf) InsEOL(pos Pos) Pos {
	if !buf.InWriteBounds(pos) {
		return pos
	}

	buf.SetDirty()

	x, y := pos.X, pos.Y

	if y == buf.NumLines() {
		buf.WriteLine("")
		return Pos{x, y + 1}
	}

	line, nline := buf.PosLine(y)
	var leftPart, rightPart []rune
	leftPart = line[:x]
	if x <= nline-1 {
		rightPart = line[x:]
	}

	buf.Lines = append(buf.Lines, "")
	if y < buf.NumLines()-2 {
		copy(buf.Lines[y+2:], buf.Lines[y+1:])
	}

	buf.Lines[y] = string(leftPart)
	buf.Lines[y+1] = string(rightPart)
	return Pos{0, y + 1}
}

func (buf *Buf) InsChar(pos Pos, c rune) Pos {
	if !buf.InWriteBounds(pos) {
		return pos
	}

	buf.SetDirty()

	x, y := pos.X, pos.Y

	// Insert new line with char.
	if y == buf.NumLines() {
		buf.WriteLine(string(c))
		return Pos{x + 1, y}
	}

	// Replace existing line, insert char.
	line, nline := buf.PosLine(y)
	line = append(line, 0)
	if x < nline-1 {
		copy(line[x+1:], line[x:])
	}
	line[x] = c
	buf.Lines[y] = string(line)

	return Pos{x + 1, y}
}

func (buf *Buf) InsStr(pos Pos, s string) Pos {
	if !buf.InWriteBounds(pos) {
		return pos
	}

	buf.SetDirty()

	x, y := pos.X, pos.Y

	// Insert new line with string.
	if y == buf.NumLines() {
		buf.WriteLine(s)
		return pos
	}

	// Replace existing line, insert string.
	var b bytes.Buffer
	line, _ := buf.PosLine(y)
	leftPart := line[:x]
	rightPart := line[x:]

	b.WriteString(string(leftPart))
	b.WriteString(s)
	b.WriteString(string(rightPart))

	buf.Lines[y] = b.String()

	return Pos{x + runeslen(s), y}
}

func (buf *Buf) InsText(pos Pos, s string) Pos {
	b := bytes.NewBufferString(s)
	scanner := bufio.NewScanner(b)

	bufPos := pos
	for scanner.Scan() {
		sline := scanner.Text()
		bufPos = buf.InsStr(bufPos, sline)
		buf.InsEOL(bufPos)

		bufPos.Y++
		bufPos.X = 0
	}

	buf.SetDirty()

	return bufPos
}

func (buf *Buf) DelChar(pos Pos) Pos {
	return buf.DelChars(pos, 1)
}
func (buf *Buf) DelChars(pos Pos, n int) Pos {
	if !buf.InWriteBounds(pos) {
		return pos
	}

	buf.SetDirty()

	x, y := pos.X, pos.Y

	line, nline := buf.PosLine(y)
	if x == nline || (x == 0 && nline == 0) {
		buf.MergeLines(y, y+1)
		n--
	}

	// Replace existing line, delete chars.
	for n > 0 && buf.InBounds(Pos{x, y}) {
		line, nline = buf.PosLine(y)
		nlinechars := min(n, nline-x)

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
func (buf *Buf) DelPrevChar(pos Pos) Pos {
	if !buf.InWriteBounds(pos) {
		return pos
	}

	buf.SetDirty()

	x, y := pos.X, pos.Y

	if x == 0 && y > 0 {
		_, yLineLen := buf.PosLine(y)
		buf.MergeLines(y-1, y)

		_, nline := buf.PosLine(y - 1)
		return Pos{nline - yLineLen, y - 1}
	}

	if x == 0 && y == 0 {
		return pos
	}

	return buf.DelChar(Pos{x - 1, y})
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

// Parse line to get sequence of words.
// Each whitespace char is considered a single word.
// Ex. "One two  three" => ["One", " ", "two", " ", " ", "three"]
func parseWords(s string) []string {
	var currentWord string
	var words []string

	for _, c := range s {
		if unicode.IsSpace(c) {
			// Add pending word
			words = append(words, currentWord)

			// Add single space word
			words = append(words, string(c))

			currentWord = ""
			continue
		}

		// Add char to pending word
		currentWord += string(c)
	}

	if len(currentWord) > 0 {
		words = append(words, currentWord)
	}

	return words
}

func nextTabStop(x, tablen int) int {
	x++
	for x%tablen != 0 {
		x++
	}
	return x
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

		lenW := len([]rune(w))

		// Expand tabs with enough spaces to reach next tab stop.
		if w == "\t" {
			lenW = nextTabStop(xWL, _tablen) - xWL
		}

		// word can't fit in remaining wrapline, add to next wrapline.
		if xWL+lenW > maxlenWrapLine {
			cbWrapLine(bWL.String())

			if w == "\t" {
				lenW = _tablen
			}

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

// Return char distance between xStart to xEnd in line y.
// Take into account tab expansion when computing distance.
func (buf *Buf) Distance(y, xStart, xEnd int) int {
	if !buf.InBounds(Pos{xStart, y}) || !buf.InWriteBounds(Pos{xEnd, y}) {
		return 0
	}

	line, _ := buf.PosLine(y)
	line = line[xStart:xEnd]
	dist := 0
	x := 0

	for _, c := range line {
		if c == '\t' {
			dist += nextTabStop(x, _tablen) - x
			x = dist
			continue
		}
		dist++
		x++
	}

	return dist
}

func expandTabs(s string, tablen int) string {
	var b bytes.Buffer
	x := 0
	for _, c := range s {
		if c == '\t' {
			tabSpaces := strings.Repeat(" ", nextTabStop(x, tablen)-x)
			b.WriteString(tabSpaces)
			x += len([]rune(tabSpaces))
			continue
		}

		b.WriteRune(c)
		x++
	}
	return b.String()
}
