package main

import ()

type Buf struct {
	Lines []string
}

func NewBuf() *Buf {
	// Initialize with one empty line.
	buf := &Buf{}

	return buf
}

func (buf *Buf) WriteString(s string) {
	buf.Lines = append(buf.Lines, s)
}
