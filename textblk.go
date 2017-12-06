package main

import (
	"unicode"
)

type TextBlk struct {
	Text   [][]rune // A block of text, with line wrapping
	PosMap [][]Pos  // Absolute line position ignoring line wrapping
	Size
}

func NewTextBlk(width, height int) *TextBlk {
	blk := &TextBlk{}
	blk.Resize(width, height)

	return blk
}

// Preserve contents while resizing textblk.
func (blk *TextBlk) Resize(width, height int) {
	if height > blk.Height {
		for y := 0; y < height-blk.Height; y++ {
			row := make([]rune, width)
			blk.Text = append(blk.Text, row)

			posRow := make([]Pos, width)
			blk.PosMap = append(blk.PosMap, posRow)
		}
	} else if height < blk.Height {
		blk.Text = blk.Text[:height]
		blk.PosMap = blk.PosMap[:height]
	}

	if width > blk.Width {
		for y := range blk.Text {
			addlRunes := make([]rune, width-blk.Width)
			blk.Text[y] = append(blk.Text[y], addlRunes...)

			addlPos := make([]Pos, width-blk.Width)
			blk.PosMap[y] = append(blk.PosMap[y], addlPos...)
		}
	} else if width < blk.Width {
		for y := range blk.Text {
			blk.Text[y] = blk.Text[y][:width]
			blk.PosMap[y] = blk.PosMap[y][:width]
		}
	}

	blk.Height = height
	blk.Width = width
}

func (blk *TextBlk) AddRows(n int) {
	blk.Resize(blk.Width, blk.Height+n)
}

func (blk *TextBlk) AddCols(n int) {
	blk.Resize(blk.Width+n, blk.Height)
}

func (blk *TextBlk) ClearRow(row, colStart, yPos, xPos int) {
	for col := colStart; col < blk.Width; col++ {
		blk.Text[row][col] = 0
		blk.PosMap[row][col] = Pos{xPos, yPos}
		xPos++
	}
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

// Write str line into startRow,
// return next row to write succeeding lines.
func (blk *TextBlk) writeLineStartRow(l string, yPos int, startRow int) (nextRow int) {
	words := parseWords(l)
	x := 0
	y := startRow

	xPos := 0

	for _, word := range words {
		// Not enough space in this line to fit word, try in next line.
		if (x + len(word) - 1) > (blk.Width - 1) {
			blk.ClearRow(y, x, yPos, xPos)
			y++
			x = 0
		}

		// Add new rows as needed.
		if y > blk.Height-1 {
			blk.AddRows(1)
		}

		// Write word in remaining space.
		for _, c := range word {
			blk.Text[y][x] = c
			blk.PosMap[y][x] = Pos{xPos, yPos}

			x++
			xPos++

			// If word is longer than entire textblk width,
			// split word into multiple lines.
			if x > blk.Width-1 {
				y++
				x = 0

				// Add new rows as needed.
				if y > blk.Height-1 {
					blk.AddRows(1)
				}
			}
		}
	}

	// Last word ended exactly at txtblk edge, so already at next row.
	if x == 0 && len(words) > 0 {
		nextRow = y
		return nextRow
	}

	blk.ClearRow(y, x, yPos, xPos)

	nextRow = y + 1
	return nextRow
}

// Write sequence of string lines into textblk, one line at a time,
// wrapping text into rows separated by word boundaries.
//
// textblk is auto-resized (rows are added when needed) to fit
// the string lines.
func (blk *TextBlk) WriteStringLines(lines []string) {
	blk.Resize(blk.Width, len(lines))

	yblk := 0
	for yPos, l := range lines {
		yblk = blk.writeLineStartRow(l, yPos, yblk)
	}
}

func FillTextBlk(blk *TextBlk, buf *Buf) {
	blk.WriteStringLines(buf.Lines)
}
