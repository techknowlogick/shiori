package utils

import (
	"unicode/utf8"
)

func FixUtf(r rune) rune {
	if r == utf8.RuneError {
		return -1
	}
	return r
}
