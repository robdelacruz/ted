package main

// Consts
// ------
// TabSize
//
// Functions
// ---------
// nextTabStop(x int) (nextTabX int, paddingToTabX string)
// expandTabs(sline string) string
// curXFromLineCol(x int, sline string) (xli int)

import (
	"bytes"
	"strings"
)

const (
	TabSize = 5
)

func nextTabStop(x int) (nextTabX int, paddingToTabX string) {
	remX := x % TabSize
	ncolToTabStop := 5 - remX
	return x + ncolToTabStop, strings.Repeat(" ", ncolToTabStop)
}

func expandTabs(sline string) string {
	var b bytes.Buffer
	rstr, _ := runestr(sline)

	var x int
	for _, c := range rstr {
		if c == '\t' {
			nextTabX, pad := nextTabStop(x)
			b.WriteString(pad)
			x = nextTabX
			continue
		}
		b.WriteRune(c)
		x++
	}
	return b.String()
}

// Return equivalent column no with tabs expanded.
func expandTabsX(x int, sline string) (xli int) {
	rstr, _ := runestr(sline)
	for _, c := range rstr[:x] {
		if c == '\t' {
			xli, _ = nextTabStop(xli)
			continue
		}
		xli++
	}
	return xli
}

// Return equivalent column no without tabs expanded.
func unexpandTabsX(xli int, sline string) (x int) {
	if xli == 0 {
		return 0
	}

	rstr, _ := runestr(sline)
	xExpanded := 0
	for _, c := range rstr {
		if c == '\t' {
			xExpanded, _ = nextTabStop(xExpanded)
		} else {
			xExpanded++
		}
		if xExpanded > xli {
			break
		}
		x++
	}
	return x
}
