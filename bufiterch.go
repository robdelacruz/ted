package main

import ()

// Structs
// -------
// BufIterCh
//
// BufIterCh - Buf char iterator
// -----------------------------
// BufIterCh(buf *Buf) *BufIterCh
// Reset()
// Ch() rune
// Pos() Pos
// ScanNext() bool
// ScanPrev() bool
// ScanDown() bool
// ScanUp() bool
// Seek(pos Pos) bool
//

type BufIterCh struct {
	buf     *Buf
	bn      *BufNode
	pos     Pos
	rstr    []rune
	rstrLen int
}

func NewBufIterCh(buf *Buf) *BufIterCh {
	bit := &BufIterCh{}
	bit.buf = buf
	bit.Reset()

	return bit
}

func (bit *BufIterCh) Reset() {
	bit.pos = Pos{-1, 0}
	bit.bn = bit.buf.H
	if bit.bn != nil {
		bit.rstr, bit.rstrLen = runestr(bit.bn.S)
	}
}

func (bit *BufIterCh) Ch() rune {
	if bit.pos.X > bit.rstrLen-1 {
		return 0
	}
	return bit.rstr[bit.pos.X]
}
func (bit *BufIterCh) Pos() Pos {
	return bit.pos
}

func (bit *BufIterCh) ScanNext() bool {
	if bit.bn == nil {
		return false
	}

	bit.pos.X++
	if bit.pos.X > bit.rstrLen-1 {
		bn := bit.bn.Next
		if bn == nil {
			bit.pos.X--
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

	return true
}

func (bit *BufIterCh) ScanPrev() bool {
	if bit.bn == nil {
		return false
	}

	bit.pos.X--
	if bit.pos.X < 0 {
		bn := bit.bn.Prev
		if bn == nil {
			bit.pos.X++
			return false
		}

		rstr, rstrLen := runestr(bn.S)
		bit.bn = bn
		bit.rstr = rstr
		bit.rstrLen = rstrLen

		bit.pos.X = bit.rstrLen - 1
		bit.pos.Y--
	}

	if bit.rstrLen == 0 {
		// Code should not reach here because buf lines should always
		// have at least one char '\n'.
		return bit.ScanPrev()
	}

	return true
}

func (bit *BufIterCh) ScanDown() bool {
	if bit.bn == nil {
		return false
	}

	// Get next line.
	bn := bit.bn.Next
	if bn == nil {
		return false
	}
	rstr, rstrLen := runestr(bn.S)
	if bit.pos.X > rstrLen-1 {
		bit.pos.X = rstrLen - 1
	}
	bit.pos.Y++

	bit.bn = bn
	bit.rstr = rstr
	bit.rstrLen = rstrLen

	return true
}

func (bit *BufIterCh) ScanUp() bool {
	if bit.bn == nil {
		return false
	}

	// Get prev line.
	bn := bit.bn.Prev
	if bn == nil {
		return false
	}
	rstr, rstrLen := runestr(bn.S)
	if bit.pos.X > rstrLen-1 {
		bit.pos.X = rstrLen - 1
	}
	bit.pos.Y--

	bit.bn = bn
	bit.rstr = rstr
	bit.rstrLen = rstrLen

	return true
}

func (bit *BufIterCh) Seek(pos Pos) bool {
	bn := bit.buf.NodeFromY(pos.Y)
	if bn == nil {
		return false
	}

	rstr, rstrLen := runestr(bn.S)
	if pos.X > rstrLen-1 {
		return false
	}

	bit.bn = bn
	bit.rstr = rstr
	bit.rstrLen = rstrLen
	bit.pos = pos

	return true
}
