package main

import (
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
