package main

import (
	"unicode"
)

type Buf struct {
	Lines []string
}

func NewBuf() *Buf {
	// Initialize with one empty line.
	buf := &Buf{
		Lines: []string{""},
	}

	return buf
}

func (buf *Buf) WriteString(s string) {
	buf.Lines = append(buf.Lines, s)
}

// Copy buffer contents to textblk.
// Buffer lines will wrap to fit dimensions of textblk.
func (buf *Buf) CopyToBlk(textBlk [][]rune) {
	yBlk := 0
	for yBufLine, l := range buf.Lines {
		yBlk = writeBufLineToBlk(l, yBufLine, textBlk, yBlk)
		if yBlk > len(textBlk)-1 {
			return
		}
	}

	// Clear remaining textblk lines
	for yBlk < len(textBlk) {
		clearRestOfLine(textBlk[yBlk], 0)
		yBlk++
	}
}

func writeBufLineToBlk(l string, yBufLine int, blk [][]rune, yBlk int) int {
	words := parseWords(l)
	x := 0

	for _, word := range words {
		// Not enough space in this line to fit word, try in next line
		if (x + len(word) - 1) > (len(blk[yBlk]) - 1) {
			yBlk++
			x = 0
		}

		// Past bottom view line
		if yBlk > len(blk)-1 {
			return yBlk
		}

		// Write word in remaining space
		for _, c := range word {
			blk[yBlk][x] = c
			x++

			// Word is longer than entire buf width, so split it into
			// multiple lines
			if x > len(blk[yBlk])-1 {
				yBlk++
				x = 0

				if yBlk > len(blk)-1 {
					return yBlk
				}
			}
		}
	}

	clearRestOfLine(blk[yBlk], x)
	return yBlk + 1
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

func clearRestOfLine(blkLine []rune, x int) {
	for x < len(blkLine) {
		blkLine[x] = 0
		x++
	}
}
