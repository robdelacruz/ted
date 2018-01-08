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

	buf.SetText(text)

	fmt.Println("Buf text:")
	fmt.Println(buf.Text())

	fmt.Println("BufIterCh test:")
	bit := NewBufIterCh(buf)
	for bit.ScanNext() {
		fmt.Printf("%c", bit.Ch())
	}
	fmt.Println("")

	fmt.Println("BufIterCh reverse test:")
	fmt.Printf("%c", bit.Ch())
	for bit.ScanPrev() {
		fmt.Printf("%c", bit.Ch())
	}
	fmt.Println("")

	fmt.Println("BufIterCh seek test:")
	fmt.Printf("Seek (5,2): %v\n", bit.Seek(Pos{5, 2}))

	// Nonexistent, should be ignored by bufiterch.
	fmt.Printf("Seek (8,2): %v\n", bit.Seek(Pos{8, 2}))
	fmt.Printf("Seek (0,100): %v\n", bit.Seek(Pos{0, 100}))

	fmt.Printf("%c", bit.Ch())
	for bit.ScanNext() {
		fmt.Printf("%c", bit.Ch())
	}

	fmt.Printf("Number of nodes: %d\n", buf.NumNodes())
}

func TestBufIterWl(t *testing.T) {
	buf := NewBuf()
	p := `Now is the time for all good men to come to the aid of the party. The quick brown fox jumps over the lazy dog.`
	buf.AppendLine(p)
	buf.AppendLine("")
	buf.AppendLine(p)

	err := buf.Load("sample.txt")
	if err != nil {
		panic(err)
	}

	fmt.Println("BufIterWl test:")
	bit := NewBufIterWl(buf, 40)
	for bit.ScanNext() {
		fmt.Printf("(%2d,%2d) '%s'\n", bit.Pos().X, bit.Pos().Y, bit.Text())
	}

	fmt.Println("BufIterWl reverse test:")
	fmt.Printf("(%2d,%2d) '%s'\n", bit.Pos().X, bit.Pos().Y, bit.Text())
	for bit.ScanPrev() {
		fmt.Printf("(%2d,%2d) '%s'\n", bit.Pos().X, bit.Pos().Y, bit.Text())
	}

	fmt.Println("BufIterWl forward again:")
	bit.Reset()
	fmt.Printf("(%2d,%2d) '%s'\n", bit.Pos().X, bit.Pos().Y, bit.Text())
	for bit.ScanNext() {
		fmt.Printf("(%2d,%2d) '%s'\n", bit.Pos().X, bit.Pos().Y, bit.Text())
	}

	fmt.Println("BufIterWl seek test:")
	seekPos := Pos{3, 100}
	fmt.Printf("Seek %v result: %v\n", seekPos, bit.Seek(seekPos))
	seekPos = Pos{500, 2}
	fmt.Printf("Seek %v result: %v\n", seekPos, bit.Seek(seekPos))
	seekPos = Pos{76, 2}
	fmt.Printf("Seek %v result: %v\n", seekPos, bit.Seek(seekPos))
	seekPos = Pos{45, 2}
	fmt.Printf("Seek %v result: %v\n", seekPos, bit.Seek(seekPos))
	fmt.Printf("(%2d,%2d) '%s'\n", bit.Pos().X, bit.Pos().Y, bit.Text())
	for bit.ScanNext() {
		fmt.Printf("(%2d,%2d) '%s'\n", bit.Pos().X, bit.Pos().Y, bit.Text())
	}

}
