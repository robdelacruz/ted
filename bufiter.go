package main

// Structs
// -------
// BufChIter
//
// BufChIter - Buf char iterator
// -----------------------------
// NewBufChIter(buf *Buf) *BufChIter
// NextChar() (rune, Pos)
// PrevChar() (rune, Pos)
//

type BufChIter struct {
	buf  *Buf
	bn   *BufNode
	pos  Pos
	rstr []rune
	slen int
}

func NewBufChIter(buf *Buf) *BufChIter {
	bit := &BufChIter{}
	bit.buf = buf
	bit.bn = buf.H
	bit.pos = Pos{-1, 0}

	if bit.bn != nil {
		bit.rstr, bit.slen = runestr(bit.bn.S)
	}
	return bit
}

func (bit *BufChIter) NextChar() (rune, Pos) {
	if bit.bn == nil {
		return 0, Pos{0, 0}
	}

	bit.pos.X++
	if bit.pos.X > bit.slen-1 {
		bn := bit.bn.Next
		if bn == nil {
			bit.pos.X--
			return 0, Pos{0, 0}
		}

		bit.pos.X = 0
		bit.pos.Y++

		rstr, slen := runestr(bn.S)
		bit.bn = bn
		bit.rstr = rstr
		bit.slen = slen
	}

	if bit.slen == 0 {
		// Code should not reach here because buf lines should always
		// have at least one char '\n'.
		return bit.NextChar()
	}

	return bit.rstr[bit.pos.X], bit.pos
}

func (bit *BufChIter) PrevChar() (rune, Pos) {
	if bit.bn == nil {
		return 0, Pos{0, 0}
	}

	bit.pos.X--
	if bit.pos.X < 0 {
		bn := bit.bn.Prev
		if bn == nil {
			bit.pos.X++
			return 0, Pos{0, 0}
		}

		rstr, slen := runestr(bn.S)
		bit.bn = bn
		bit.rstr = rstr
		bit.slen = slen

		bit.pos.X = bit.slen - 1
		bit.pos.Y--
	}

	if bit.slen == 0 {
		// Code should not reach here because buf lines should always
		// have at least one char '\n'.
		return bit.PrevChar()
	}

	return bit.rstr[bit.pos.X], bit.pos
}
