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
	fmt.Printf("(%2d,%2d) '%s'\n", bit.Pos().X, bit.Pos().Y, bit.Text())
	for bit.ScanNext() {
		fmt.Printf("(%2d,%2d) '%s'\n", bit.Pos().X, bit.Pos().Y, bit.Text())
	}

	fmt.Println("BufIterWl seek test:")
	seekPos := Pos{3, 100}
	fmt.Printf("Seek %v result: %v\n", seekPos, bit.Seek(seekPos))
	seekPos = Pos{500, 5}
	fmt.Printf("Seek %v result: %v\n", seekPos, bit.Seek(seekPos))
	seekPos = Pos{76, 6}
	fmt.Printf("Seek %v result: %v\n", seekPos, bit.Seek(seekPos))
	seekPos = Pos{320, 8}
	fmt.Printf("Seek %v result: %v\n", seekPos, bit.Seek(seekPos))
	fmt.Printf("(%2d,%2d) '%s'\n", bit.Pos().X, bit.Pos().Y, bit.Text())
	for bit.ScanNext() {
		fmt.Printf("(%2d,%2d) '%s'\n", bit.Pos().X, bit.Pos().Y, bit.Text())
	}

}

func TestBufCopyPaste(t *testing.T) {
	buf := NewBuf()

	text := `Line 1.
Line 2.
Line 3.

Now is the time
for all good men
to come to the aid
of the party.`

	var begin, end Pos
	var clip string
	var slen int

	buf.SetText(text)
	begin = Pos{1, 0}
	end = Pos{5, 0}
	clip, slen = buf.Cut(begin, end)
	fmt.Printf("Cut %v - %v:\nClip:\n'%s'\nLen: %d\nAfter:\n%s\n", begin, end, clip, slen, buf.Text())

	buf.SetText(text)
	begin = Pos{2, 1}
	end = Pos{3, 4}
	clip, slen = buf.Cut(begin, end)
	fmt.Printf("Cut %v - %v:\nClip:\n'%s'\nLen: %d\nAfter:\n%s\n", begin, end, clip, slen, buf.Text())

	buf.SetText(text)
	begin = Pos{0, 0}
	end = Pos{0, 2}
	clip, slen = buf.Copy(begin, end)
	fmt.Printf("Copy: %v - %v:\nClip:\n'%s'\nLen: %d\nAfter:\n%s\n", begin, end, clip, slen, buf.Text())

	buf.SetText(text)
	begin = Pos{2, 1}
	end = Pos{4, 1}
	clip, slen = buf.Copy(begin, end)
	fmt.Printf("Copy: %v - %v:\nClip:\n'%s'\nLen: %d\nAfter:\n%s\n", begin, end, clip, slen, buf.Text())

	buf.SetText(text)
	begin = Pos{0, 1}
	pasteText := "Line1a.\nabc"
	buf.Paste(begin, pasteText)
	fmt.Printf("Paste %v:\nPaste text:\n'%s'\nAfter:\n%s\n", begin, pasteText, buf.Text())
}

func TestBufSearch(t *testing.T) {
	buf := NewBuf()

	text := `Line 1.
Line 2.
Line 3.

Now is the time
for all good men
to come to the aid
of the party.`

	buf.SetText(text)

	startPos := Pos{0, 0}
	s := "is the"
	foundPos, found := buf.Search(startPos, s)
	fmt.Printf("BufSearch(%v, '%s') returned %v, %v\n", startPos, s, foundPos, found)

	startPos = Pos{2, 5}
	s = "good men"
	foundPos, found = buf.Search(startPos, s)
	fmt.Printf("BufSearch(%v, '%s') returned %v, %v\n", startPos, s, foundPos, found)

	startPos = Pos{9, 5}
	s = "good men"
	foundPos, found = buf.Search(startPos, s)
	fmt.Printf("BufSearch(%v, '%s') returned %v, %v\n", startPos, s, foundPos, found)

	startPos = Pos{1, 1}
	s = "12345"
	foundPos, found = buf.Search(startPos, s)
	fmt.Printf("BufSearch(%v, '%s') returned %v, %v\n", startPos, s, foundPos, found)
}
