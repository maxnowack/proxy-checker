package main

import (
	"strings"
)

func buildFillString(length int) string {
	var b strings.Builder
	b.Grow(length)
	for i := 0; i < length; i++ {
		b.WriteString("0")
	}
	return b.String()
}
