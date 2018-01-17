package main

// Structs
// -------
// Buf
// BufNode
//
// Consts
// ------
// LF
//
// Functions
// ---------
// min(n1...) int
// lfStr(s string) string
// endsWithLF(s string) bool
// rlen(s string) int
// runestr(s string) ([]rune, int)
// inBounds(slen, x int) int
//
// BufNode
// -------
// NewBufNode(s string) *BufNode
// NewBufLineNode(s string) *BufNode
// InsertAfter(bnNew *BufNode)
// InsertLineAfter(s string) *BufNode
// MergeNextNode()
// InsStr(x int, s string) int
// InsChar(x int c rune) int
// InsLF(x int) *BufNode
// DelChars(x, nDel int)
//
// Buf
// ---
// NewBuf() *Buf
// Validate() error
// Clear()
// SetDirty()
// ClearDirty()
// NumNodes() int
//
// YFromNode(bnFind *BufNode) int
// NodeFromY(y int) *BufNode
// NodeFromYAutoAdd(y int) *BufNode
// Line(y int) string
// InBounds(pos Pos) bool
//
// SetText(s string)
// Text() string
// Load(file string) error
// SaveFile(file string) error
//
// AppendLine(s string)
// DelNode(bnDel *BufNode)
//
// InsStr(pos Pos s string) Pos
// InsChar(pos Pos, c rune) Pos
// InsLF(pos Pos) Pos
// InsText(pos Pos, s string) Pos
// DelChars(pos Pos, nDel int) Pos
// DelChar(pos Pos) Pos
//
// Cut(begin, end Pos) (string, int)
// Copy(begin, end Pos) (string, int)
// Paste(pos Pos, s string)
// Search(pos Pos, s string) (foundPos Pos, fFound bool)
//

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
)

const (
	LF = '\n'
)

type BufNode struct {
	S    string
	Next *BufNode
	Prev *BufNode
}

type Buf struct {
	Name  string
	Dirty bool
	H     *BufNode
}

// LF-terminated line
func lfStr(s string) string {
	if !endsWithLF(s) {
		return s + string(LF)
	}
	return s
}

func endsWithLF(s string) bool {
	rstr, slen := runestr(s)
	if slen > 0 && rstr[slen-1] == LF {
		return true
	}
	return false
}

// Remove LF from end of line if it's there
func chomp(s string) string {
	slen := len(s)
	if slen > 0 && s[slen-1] == LF {
		return s[:slen-1]
	}
	return s
}

func rlen(s string) int {
	return len([]rune(s))
}

func runestr(s string) ([]rune, int) {
	rstr := []rune(s)
	return rstr, len(rstr)
}

// Keep x always within bounds of line len.
// Last char of line is always '\n' (LF), so slen-1 points to
// location of '\n', which is last insert x col in line.
func inBounds(slen, x int) int {
	if x > slen-1 {
		x = slen - 1
	}
	return x
}

// BufNode with string
func NewBufNode(s string) *BufNode {
	n := &BufNode{}
	n.S = s
	return n
}

// BufNode with CR-terminated line string
func NewBufLineNode(s string) *BufNode {
	return NewBufNode(lfStr(s))
}

func NewBuf() *Buf {
	buf := &Buf{}
	return buf
}

// Check if buf is valid:
// - all nodes should be at least 1 char, ending in '\n'.
func (buf *Buf) Validate() error {
	bn := buf.H
	row := 0
	for bn != nil {
		rstr, slen := runestr(bn.S)
		if slen < 1 {
			return fmt.Errorf("Zero length line %d (should at least have 1 char ending in '\n'", row)
		}

		if rstr[slen-1] != '\n' {
			return fmt.Errorf("Line %d '%s' doesn't end with '\n'.", row, bn.S)
		}
		bn = bn.Next
		row++
	}

	return nil
}

func (buf *Buf) Clear() {
	buf.H = nil
}

func (buf *Buf) SetDirty() {
	buf.Dirty = true
}
func (buf *Buf) ClearDirty() {
	buf.Dirty = false
}

func (buf *Buf) NumNodes() int {
	n := 0
	bn := buf.H
	for bn != nil {
		n++
		bn = bn.Next
	}
	return n
}

func (buf *Buf) YFromNode(bnFind *BufNode) int {
	bn := buf.H
	y := 0
	for bn != nil {
		if bn == bnFind {
			return y
		}
		bn = bn.Next
		y++
	}

	return 0
}

// Return row y bufnode.
func (buf *Buf) NodeFromY(y int) *BufNode {
	bn := buf.H
	for i := 0; i < y && bn != nil; i++ {
		bn = bn.Next
	}
	return bn
}

// Same as NodeFromY() but auto add any missing rows.
// Always returns a bufnode.
func (buf *Buf) NodeFromYAutoAdd(y int) *BufNode {
	bn := buf.NodeFromY(y)
	if bn == nil {
		// Create any missing lines up to y.
		nNodeLines := buf.NumNodes()
		for i := 0; i < y-(nNodeLines-1); i++ {
			buf.AppendLine("")
		}
		bn = buf.NodeFromY(y)
	}
	return bn
}

func (buf *Buf) Line(y int) string {
	bn := buf.NodeFromY(y)
	if bn == nil {
		return ""
	}
	return bn.S
}

func (buf *Buf) InBounds(pos Pos) bool {
	if pos.X < 0 {
		return false
	}

	bn := buf.NodeFromY(pos.Y)
	if bn == nil {
		return false
	}

	if pos.X > rlen(bn.S)-1 {
		return false
	}

	return true
}

func (buf *Buf) SetText(s string) {
	buf.Clear()

	var bn *BufNode

	scanner := bufio.NewScanner(bytes.NewBufferString(s))
	for scanner.Scan() {
		sline := scanner.Text()
		newBn := NewBufLineNode(sline)

		if buf.H == nil {
			buf.H = newBn
			bn = buf.H
			continue
		}

		bn.Next = newBn
		newBn.Prev = bn
		bn = newBn
	}

	// Buf should always have at least one node.
	if buf.H == nil {
		buf.AppendLine("")
	}
}

func (buf *Buf) Text() string {
	bn := buf.H

	var b bytes.Buffer
	for bn != nil {
		// Remove any trailing LF in last line.
		if bn.Next == nil {
			b.WriteString(chomp(bn.S))
			break
		}
		b.WriteString(bn.S)
		bn = bn.Next
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

	bs := []byte(buf.Text())
	err := ioutil.WriteFile(file, bs, 0644)
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	buf.Name = file
	buf.ClearDirty()
	return nil
}

// Append node to buf end
// H ... bnLast -> bnNext
func (buf *Buf) appendNode(bn *BufNode) {
	if buf.H == nil {
		buf.H = bn
		bn.Prev = nil
		return
	}

	n := buf.H
	for n.Next != nil {
		n = n.Next
	}
	n.Next = bn
	bn.Prev = n
}

// Append line node to buf end
func (buf *Buf) AppendLine(s string) {
	bn := NewBufLineNode(s)
	buf.appendNode(bn)
}

// bn1 -> (bnNew) -> bn2...
func (bn *BufNode) InsertAfter(bnNew *BufNode) {
	bn1 := bn
	bn2 := bn.Next

	bn1.Next = bnNew
	bnNew.Prev = bn1
	bnNew.Next = bn2
	if bn2 != nil {
		bn2.Prev = bnNew
	}
}

func (bn *BufNode) InsertLineAfter(s string) *BufNode {
	bnNew := NewBufLineNode(s)
	bn.InsertAfter(bnNew)
	return bnNew
}

// bn -> bn2 -> bn3
// returns
// bn (bn.S + bn2.S) -> bn3
func (bn *BufNode) MergeNextNode() {
	bn2 := bn.Next
	if bn2 == nil {
		return
	}
	bn3 := bn2.Next

	bn.S = chomp(bn.S) + bn2.S
	bn.Next = bn3

	if bn3 != nil {
		bn3.Prev = bn
	}
}

func (buf *Buf) DelNode(bnDel *BufNode) *BufNode {
	if buf.H == nil {
		return nil
	}

	bnDelNext := bnDel.Next

	// Del head node
	if buf.H == bnDel {
		buf.H = bnDelNext
		if buf.H != nil {
			buf.H.Prev = nil
		}
		return bnDelNext
	}

	// Del middle node
	bnPrev := buf.H
	bn := buf.H.Next
	for bn != nil {
		if bn == bnDel {
			bnPrev.Next = bnDelNext
			if bnDelNext != nil {
				bnDelNext.Prev = bnPrev
			}
			return bnDelNext
		}
		bnPrev = bn
		bn = bn.Next
	}

	return nil
}

func (bn *BufNode) InsStr(x int, s string) int {
	rstr, slen := runestr(bn.S)
	if slen == 0 {
		// Initialize to LF-terminated line.
		bn.S = string(LF)
		rstr, slen = runestr(bn.S)
	}
	// If x out of bounds, fill in missing chars to make in bounds.
	if x > slen-1 {
		bn.S = strings.Repeat(" ", x-(slen-1)) + bn.S
		rstr, slen = runestr(bn.S)
	}

	var b bytes.Buffer
	b.WriteString(string(rstr[:x]))
	b.WriteString(s)
	b.WriteString(string(rstr[x:]))

	bn.S = b.String()

	return x + rlen(s)
}

func (bn *BufNode) InsChar(x int, c rune) int {
	return bn.InsStr(x, string(c))
}

func (bn *BufNode) InsLF(x int) *BufNode {
	rstr, slen := runestr(bn.S)
	if slen == 0 {
		// Initialize to LF-terminated line.
		bn.S = string(LF)
		rstr, slen = runestr(bn.S)
	}
	// If x out of bounds, fill in missing chars to make in bounds.
	if x > slen-1 {
		bn.S = strings.Repeat(" ", x-(slen-1)) + bn.S
		rstr, slen = runestr(bn.S)
	}

	var b bytes.Buffer
	b.WriteString(string(rstr[:x]))
	b.WriteString("\n")
	bn.S = b.String()

	b.Reset()
	b.WriteString(string(rstr[x:]))
	bnRight := NewBufNode(b.String())
	bn.InsertAfter(bnRight)

	return bnRight
}

func (buf *Buf) InsStr(pos Pos, s string) Pos {
	bn := buf.NodeFromYAutoAdd(pos.Y)

	buf.SetDirty()

	x := bn.InsStr(pos.X, s)
	return Pos{x, pos.Y}
}

func (buf *Buf) InsChar(pos Pos, c rune) Pos {
	return buf.InsStr(pos, string(c))
}

func (buf *Buf) InsLF(pos Pos) Pos {
	bn := buf.NodeFromYAutoAdd(pos.Y)

	buf.SetDirty()

	bnNextLine := bn.InsLF(pos.X)
	return Pos{0, buf.YFromNode(bnNextLine)}
}

func (buf *Buf) InsText(pos Pos, s string) Pos {
	var lastChar rune
	rstr, rstrLen := runestr(s)
	if rstrLen > 0 {
		lastChar = rstr[rstrLen-1]
	}

	var slines []string
	scanner := bufio.NewScanner(bytes.NewBufferString(s))
	for scanner.Scan() {
		slines = append(slines, scanner.Text())
	}

	nslines := len(slines)
	for i, sline := range slines {
		pos = buf.InsStr(pos, sline)

		// Newline for first and middle lines
		// and if last line ends with LF.
		if i < nslines-1 || lastChar == '\n' {
			pos = buf.InsLF(pos)
		}
	}

	buf.SetDirty()

	return pos
}

func (bn *BufNode) DelChars(x, nDel int) {
	if x > rlen(bn.S)-1 {
		return
	}

	for nDel > 0 {
		rstr, slen := runestr(bn.S)
		nLineDel := min(slen-x, nDel)

		var b bytes.Buffer
		b.WriteString(string(rstr[:x]))
		b.WriteString(string(rstr[x+nLineDel:]))
		bn.S = b.String()

		if !endsWithLF(bn.S) {
			bn.MergeNextNode()
		}

		nDel -= nLineDel
		if nLineDel == 0 {
			break
		}
	}

	// Make valid node if all chars gone.
	if !endsWithLF(bn.S) {
		bn.S += "\n"
	}
}

func (buf *Buf) DelChars(pos Pos, nDel int) Pos {
	bn := buf.NodeFromY(pos.Y)
	if bn == nil {
		return pos
	}

	// Out of range in first line, so start on next line.
	if pos.X > rlen(bn.S)-1 {
		orgPos := pos

		nDel -= pos.X - rlen(bn.S)
		pos.Y++
		pos.X = 0

		bn = buf.NodeFromY(pos.Y)
		if bn == nil {
			return orgPos
		}
	}

	bn.DelChars(pos.X, nDel)
	buf.SetDirty()

	return pos
}

func (buf *Buf) DelChar(pos Pos) Pos {
	return buf.DelChars(pos, 1)
}

func (buf *Buf) Cut(begin, end Pos) (string, int) {
	clip, slen := buf.Copy(begin, end)
	buf.DelChars(begin, slen)
	return clip, slen
}

func (buf *Buf) Copy(begin, end Pos) (string, int) {
	var b bytes.Buffer

	// One line only
	if begin.Y == end.Y {
		bn := buf.NodeFromY(begin.Y)
		if bn == nil {
			return "", 0
		}
		startX := begin.X
		endX := end.X
		rstr, slen := runestr(bn.S)
		if startX > slen {
			return "", 0
		}
		if endX > slen-1 {
			endX = slen - 1
		}

		rclip := rstr[startX : endX+1]
		return string(rclip), len(rclip)
	}

	// First line
	y := begin.Y
	bn := buf.NodeFromY(y)
	if bn == nil {
		return "", 0
	}
	startX := begin.X
	rstr, slen := runestr(bn.S)
	if startX < slen {
		b.WriteString(string(rstr[startX:]))
	}
	y++

	// Middle lines
	for y < end.Y {
		bn = buf.NodeFromY(y)
		if bn == nil {
			clip := b.String()
			return clip, rlen(clip)
		}
		b.WriteString(bn.S)
		y++
	}

	// Last line
	if y == end.Y {
		bn = buf.NodeFromY(y)
		if bn == nil {
			clip := b.String()
			return clip, rlen(clip)
		}
		endX := end.X
		rstr, slen = runestr(bn.S)
		if endX > slen-1 {
			endX = slen - 1
		}
		b.WriteString(string(rstr[:endX+1]))
	}

	clip := b.String()
	return clip, rlen(clip)
}

func (buf *Buf) Paste(pos Pos, s string) {
	buf.InsText(pos, s)
}

func (buf *Buf) Search(pos Pos, s string) (foundPos Pos, fFound bool) {
	bn := buf.NodeFromY(pos.Y)
	if bn == nil {
		return Pos{0, 0}, false
	}

	// Search first node starting from pos.X
	if pos.X < rlen(bn.S) {
		iFound := strings.Index(bn.S[pos.X:], s)
		if iFound >= 0 {
			return Pos{pos.X + iFound, pos.Y}, true
		}
	}

	// Search succeeding nodes
	bn = bn.Next
	foundPos = Pos{0, pos.Y + 1}

	for bn != nil {
		iFound := strings.Index(bn.S, s)
		if iFound >= 0 {
			return Pos{foundPos.X + iFound, foundPos.Y}, true
		}

		bn = bn.Next
		foundPos.Y++
	}

	return Pos{0, 0}, false
}
