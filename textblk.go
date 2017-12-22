package main

import (
	"strings"
	"unicode"
)

type TextBlk struct {
	Text       [][]rune // A block of text, with line wrapping
	RowWidth   int
	BufFromBlk map[Pos]Pos // Buf pos corresponding to blk pos
	BlkFromBuf map[Pos]Pos // Blk pos corresponding to buf pos
	Size
	Cur        Pos
	BlkYOffset int
}

func NewTextBlk(width, height int) *TextBlk {
	blk := &TextBlk{}
	blk.BufFromBlk = map[Pos]Pos{}
	blk.BlkFromBuf = map[Pos]Pos{}

	blk.Resize(width, height)

	return blk
}

// Preserve contents while resizing textblk.
func (blk *TextBlk) Resize(width, height int) {
	if height > blk.H {
		for y := 0; y < height-blk.H; y++ {
			row := make([]rune, width)
			blk.Text = append(blk.Text, row)
		}
	} else if height < blk.H {
		blk.Text = blk.Text[:height]
	}

	if width > blk.W {
		for y := range blk.Text {
			addlRunes := make([]rune, width-blk.W)
			blk.Text[y] = append(blk.Text[y], addlRunes...)
		}
	} else if width < blk.W {
		for y := range blk.Text {
			blk.Text[y] = blk.Text[y][:width]
		}
	}

	blk.H = height
	blk.W = width
	blk.RowWidth = width - 1 // Leave 1 char room for CR/LF.
}

func (blk *TextBlk) AddRows(n int) {
	blk.Resize(blk.W, blk.H+n)
}

func (blk *TextBlk) ClearRow(yBlk, xBlkStart, yBuf, xBuf int) {
	for xBlk := xBlkStart; xBlk < blk.W; xBlk++ {
		blk.Text[yBlk][xBlk] = 0

		blk.BufFromBlk[Pos{xBlk, yBlk}] = Pos{xBuf, yBuf}
		blk.BlkFromBuf[Pos{xBuf, yBuf}] = Pos{xBlk, yBlk}

		xBuf++
	}
}

// Tab chars '\t' are expanded into n spaces
func expandWhitespaceChar(c rune) string {
	if c == '\t' {
		//$$ Hardcode expand to 4 for now.
		return strings.Repeat(" ", 4)
	}
	return string(c)
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
			words = append(words, expandWhitespaceChar(c))

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

// Write buf line (buf[yBuf]) into blk.
// Return next available blk row.
func (blk *TextBlk) writeBufLine(buf *Buf, yBuf int, yBlk int) (nextYBlk int) {
	words := parseWords(buf.Lines[yBuf])
	xBlk := 0

	// Add new rows as needed.
	if yBlk > blk.H-1 {
		blk.AddRows(1)
	}

	xBuf := 0

	for _, word := range words {
		// Not enough space in this line to fit word, try in next line.
		if (xBlk + len(word) - 1) > (blk.RowWidth - 1) {
			blk.ClearRow(yBlk, xBlk, yBuf, xBuf)
			yBlk++
			xBlk = 0
		}

		// Add new rows as needed.
		if yBlk > blk.H-1 {
			blk.AddRows(1)
		}

		// Write word in remaining space.
		for _, c := range word {
			blk.Text[yBlk][xBlk] = c

			blkPos := Pos{xBlk, yBlk}
			bufPos := Pos{xBuf, yBuf}
			blk.BufFromBlk[blkPos] = bufPos
			blk.BlkFromBuf[bufPos] = blkPos

			xBlk++
			xBuf++

			// If word is longer than entire row width,
			// split word into multiple lines.
			if xBlk > blk.RowWidth-1 {
				yBlk++
				xBlk = 0

				// Add new rows as needed.
				if yBlk > blk.H-1 {
					blk.AddRows(1)
				}
			}
		}
	}

	blk.ClearRow(yBlk, xBlk, yBuf, xBuf)

	return yBlk + 1
}

// Write buf lines to textblk, line wrapping text on word boundaries
// as necessary to fit the textblk width.
//
// textblk is auto-resized as needed to fit the number of buf lines.
func (blk *TextBlk) FillWithBuf(buf *Buf) {
	bufPos := blk.BufPos()

	yBlk := 0
	for yBuf := range buf.Lines {
		yBlk = blk.writeBufLine(buf, yBuf, yBlk)
	}

	// Remove any extra rows leftover from previous draw.
	blk.Resize(blk.W, yBlk)

	blk.Cur = blk.BlkFromBuf[bufPos]
}

func (blk *TextBlk) BlkPos() Pos {
	return blk.Cur
}
func (blk *TextBlk) BufPos() Pos {
	return blk.BufFromBlk[Pos{blk.Cur.X, blk.Cur.Y}]
}

// Print textblk contents to a rectangular area.
// Textblk contents that don't fit are cropped out.
// Empty space is painted to the background color (attribute).
func (blk *TextBlk) PrintToArea(content Rect, contentAttr TermAttr) {
	text := blk.Text
	yBlk := blk.BlkYOffset
	yBlkEdge := min(blk.H, yBlk+content.H)
	y := 0

	for yBlk < yBlkEdge {
		for xBlk := 0; xBlk < min(blk.W, content.W); xBlk++ {
			c := text[yBlk][xBlk]
			if c == 0 {
				c = ' '
			}
			print(string(c), xBlk+content.X, y+content.Y, contentAttr)
		}

		yBlk++
		y++
	}

	// Paint the rest of the content area.
	s := strings.Repeat(" ", content.W)
	for y < content.H {
		print(s, content.X, y+content.Y, contentAttr)
		y++
	}
}
