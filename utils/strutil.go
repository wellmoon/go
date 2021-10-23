package utils

import "unicode"

func IsNumber(s string) bool {
	for _, r := range s {
		if !unicode.IsNumber(r) {
			return false
		}
	}
	return true
}

func RemoveZeroWidthSpace(s string) string {
	rr := []rune(s)
	for i, r := range rr {
		if r == 8203 {
			temp := append(rr[:i], rr[i+1:]...)
			return RemoveZeroWidthSpace(string(temp))
		}
	}
	return string(rr)
}
