package util

import (
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"strings"
	"unicode"
)

func RemoveAccents(str string) string{
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	noAccents, _, _ := transform.String(t, str)
	return noAccents
}

func Substr(str string, count int) string {
	var sb strings.Builder
	for i:=0; i < len(str) && i< count; i++ {
		sb.WriteByte(str[i])
	}
	return sb.String()
}

func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
}