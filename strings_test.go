package main

import "testing"

func TestContainStr(t *testing.T) {
	l := []string{"test", "lol", "hello"}
	if !containsStr(l, "test") {
		t.Fatal("'test' is in the list\n")
	}
	if !containsStr(l, "lol") {
		t.Fatal("'lol' is in the list\n")
	}
	if !containsStr(l, "hello") {
		t.Fatal("'hello' is in the list\n")
	}
	if containsStr(l, "Hello") {
		t.Fatal("'Hello' is not in the list\n")
	}
}

func TestContainsStrCaseInsensitive(t *testing.T) {
	l := []string{"Hello", "lOl", "WorLD"}
	if !containsStrCaseInsensitive(l, "hello") {
		t.Fatal("'hello' is in the list\n")
	}
	if !containsStrCaseInsensitive(l, "LOL") {
		t.Fatal("'LOL' is in the list\n")
	}
	if !containsStrCaseInsensitive(l, "WorLD") {
		t.Fatal("'WorLD' is in the list\n")
	}
	if containsStrCaseInsensitive(l, "h") {
		t.Fatalf("'h' is not in the list\n")
	}
}

func TestToUTF8(t *testing.T) {
	s := "\xE2\x8A\x82"
	converted := toUTF8([]byte(s))
	if converted != "âŠ‚" {
		t.Fatalf("Should be 'âŠ,' not '%s'\n", converted)
	}
	s = "\x92"
	converted = toUTF8([]byte(s))
	if converted != "’" {
		t.Fatalf("Should be '’' not '%s'\n", converted)
	}
	s = "\x96"
	converted = toUTF8([]byte(s))
	if converted != "–" {
		t.Fatalf("Should be '–' not '%s'\n", converted)
	}
}
