package main

import (
	"fmt"
	"testing"
)

func TestBuf(t *testing.T) {
	buf := NewBuf()

	text := `Line 1.
Line 2.
Line 3.

Now is the time
for all good men
to come to the aid
of the party.`

	//	buf.SetText(text)
	buf.InsText(Pos{5, 2}, text)
	buf.DelChars(Pos{0, 2}, 5)

	fmt.Println("Buf text:")
	fmt.Println(buf.Text())

	fmt.Println("BufChIter test:")
	bit := NewBufChIter(buf)
	ch, _ := bit.NextChar()
	for ch != 0 {
		fmt.Printf("%c", ch)
		ch, _ = bit.NextChar()
	}
	fmt.Println("")

	fmt.Println("BufChIter reverse test:")
	ch, _ = bit.PrevChar()
	for ch != 0 {
		fmt.Printf("%c", ch)
		ch, _ = bit.PrevChar()
	}
	fmt.Println("")

	fmt.Printf("Number of nodes: %d\n", buf.NumNodes())
}
