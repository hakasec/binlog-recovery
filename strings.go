package main

import "strings"

var windows1252 = map[rune]rune{
	'\x80': '\u20AC',
	'\x82': '\u201A',
	'\x83': '\u0192',
	'\x84': '\u201E',
	'\x85': '\u2026',
	'\x86': '\u2020',
	'\x87': '\u2021',
	'\x88': '\u02C6',
	'\x89': '\u2030',
	'\x8A': '\u0160',
	'\x8B': '\u2039',
	'\x8C': '\u0152',
	'\x8E': '\u017D',
	'\x91': '\u2018',
	'\x92': '\u2019',
	'\x93': '\u201C',
	'\x94': '\u201D',
	'\x95': '\u2022',
	'\x96': '\u2013',
	'\x97': '\u2014',
	'\x98': '\u02DC',
	'\x99': '\u2122',
	'\x9A': '\u0161',
	'\x9B': '\u203A',
	'\x9C': '\u0153',
	'\x9E': '\u017E',
	'\x9F': '\u0178',
}

func toUTF8(bytes []byte) string {
	var builder strings.Builder
	for _, b := range bytes {
		// try and convert from Window1252 encoding
		r := convertRuneFromWindows1252(rune(b))
		builder.WriteRune(r)
	}
	return builder.String()
}

func convertRuneFromWindows1252(r rune) rune {
	c, ok := windows1252[r]
	if !ok {
		return r
	}
	return c
}

func containsStr(l []string, val string) bool {
	for _, s := range l {
		if s == val {
			return true
		}
	}
	return false
}

func containsStrCaseInsensitive(l []string, val string) bool {
	for _, s := range l {
		if strings.ToUpper(s) == strings.ToUpper(val) {
			return true
		}
	}
	return false
}
