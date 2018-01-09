package main

// Structs
// -------
// BufIterWl

// Functions
// ---------
// readNextWord(rstr []rune, rstrLen, startX int) (word string, nextX int)
//
// BufIterWl - Buf wraplines iterator
// ----------------------------------
// NewBufIterWl(buf *Buf, wlMaxLen int) *BufWlIter
// Reset()
// Text() string
// Pos() Pos
// ScanNext() bool
// ScanPrev() bool
// Seek(pos Pos) bool
// WrapLineIndex() int
//

import (
	"bytes"
	"log"
	"unicode"
)

type BufIterWl struct {
	buf      *Buf
	wlMaxLen int
	bn       *BufNode
	rstr     []rune
	rstrLen  int
	pos      Pos
	wlnode   *wlNode
}

type wlNode struct {
	S    string
	Pos  Pos
	bn   *BufNode
	Next *wlNode
	Prev *wlNode
}

func NewBufIterWl(buf *Buf, wlMaxLen int) *BufIterWl {
	bit := &BufIterWl{}
	bit.buf = buf
	bit.wlMaxLen = wlMaxLen
	bit.Reset()
	return bit
}

func (bit *BufIterWl) Reset() {
	bit.pos = Pos{-1, 0}
	bit.bn = bit.buf.H
	if bit.bn != nil {
		bit.rstr, bit.rstrLen = runestr(bit.bn.S)
	}
	bit.wlnode = nil
}

func (bit *BufIterWl) Text() string {
	if bit.wlnode == nil {
		return ""
	}
	return bit.wlnode.S
}
func (bit *BufIterWl) Pos() Pos {
	if bit.wlnode == nil {
		return Pos{-1, 0}
	}
	return bit.wlnode.Pos
}

func (bit *BufIterWl) ScanPrev() bool {
	if bit.wlnode == nil || bit.wlnode.Prev == nil {
		return false
	}
	bit.wlnode = bit.wlnode.Prev
	return true
}

func (bit *BufIterWl) ScanNext() bool {
	// If next wrapline was iterated on previously.
	if bit.wlnode != nil && bit.wlnode.Next != nil {
		bit.wlnode = bit.wlnode.Next
		bit.bn = bit.wlnode.bn
		bit.rstr, bit.rstrLen = runestr(bit.bn.S)
		return true
	}

	if bit.bn == nil {
		return false
	}

	bit.pos.X++
	if bit.pos.X > bit.rstrLen-1 {
		bn := bit.bn.Next
		if bn == nil {
			bit.pos.X--
			bit.bn = bn
			return false
		}

		bit.pos.X = 0
		bit.pos.Y++

		rstr, rstrLen := runestr(bn.S)
		bit.bn = bn
		bit.rstr = rstr
		bit.rstrLen = rstrLen
	}

	if bit.rstrLen == 0 {
		// Code should not reach here because buf lines should always
		// have at least one char '\n'.
		return bit.ScanNext()
	}

	// Read next wrapline of max length wlMaxLen
	// starting at bit.pos.X of rstr.
	// Next wrapline is stored in a wlNode struct and appended to bit.wlnode.
	var b bytes.Buffer
	wlNumChars := 0
	wlStartPos := bit.pos
	wlX := bit.pos.X
	for {
		w, endwX := readNextWord(bit.rstr, bit.rstrLen, wlX)
		wlen := rlen(w)

		if wlen == 0 || wlNumChars+wlen > bit.wlMaxLen {
			break
		}

		b.WriteString(w)
		wlNumChars += wlen
		wlX = endwX + 1
		bit.pos.X = endwX
	}

	newWlNode := &wlNode{
		S:    b.String(),
		Pos:  wlStartPos,
		bn:   bit.bn,
		Next: nil,
		Prev: bit.wlnode,
	}

	if bit.wlnode == nil {
		bit.wlnode = newWlNode
	} else {
		bit.wlnode.Next = newWlNode
		bit.wlnode = newWlNode
	}

	return true
}

// Read next word from rstr and rightmost x index of next word.
// Blank word returned indicates end of line reached.
// Each individual whitespace char (space, '\t', etc.) is returned as a
// separate word.
// Ex. "word1  word2" returns word sequence of "word1", " ", " ", "word2".
func readNextWord(rstr []rune, rstrLen, startX int) (word string, endwX int) {
	if startX > rstrLen-1 {
		return "", startX
	}

	if unicode.IsSpace(rstr[startX]) {
		return string(rstr[startX]), startX
	}

	var b bytes.Buffer
	x := startX
	for x < rstrLen {
		c := rstr[x]
		if unicode.IsSpace(c) {
			break
		}
		b.WriteRune(c)
		x++
	}

	return b.String(), x - 1
}

func (bit *BufIterWl) seekFirstLine() {
	for bit.ScanPrev() {
		// Keep going back until we hit the start.
	}
}

func (bit *BufIterWl) logTextPos() {
	log.Printf("(%2d,%2d) '%s'\n", bit.Pos().X, bit.Pos().Y, bit.Text())
}

func (bit *BufIterWl) Seek(pos Pos) bool {
	bit.seekFirstLine()

	// Seek to wrapline row.
	for bit.Pos().Y < pos.Y {
		if !bit.ScanNext() {
			break
		}
	}
	if pos.Y != bit.Pos().Y {
		return false
	}

	// Seek to col within wrapline
	for pos.Y == bit.Pos().Y && (bit.Pos().X+rlen(bit.Text())-1) < pos.X {
		if !bit.ScanNext() {
			break
		}
	}
	bitPos := bit.Pos()
	wlEndX := bitPos.X + rlen(bit.Text()) - 1
	if pos.Y == bitPos.Y && pos.X >= bitPos.X && pos.X <= wlEndX {
		// pos is within wrapline row.
		return true
	}

	return false
}

// Return zero-based index to current wrapline.
// Ex. -1 = BOF, 0 = first wrapline, 1 = second wrapline, etc.
func (bit *BufIterWl) WrapLineIndex() int {
	wlnode := bit.wlnode
	i := -1
	for wlnode != nil {
		wlnode = wlnode.Prev
		i++
	}
	return i
}
