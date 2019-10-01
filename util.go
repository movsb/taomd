package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func toHexStr(s string) (h string) {
	n := len(s)
	if n == 0 {
		return ""
	}

	for i := 0; i < n-1; i++ {
		h += fmt.Sprintf("%02X ", s[i])
	}

	h += fmt.Sprintf("%02X", s[n-1])

	return h
}

func toCharStr(s string) (c string) {
	for _, r := range s {
		switch r {
		case ' ':
			c += "."
		case '\t':
			c += "---→"
		case '\n':
			c += "."
		default:
			c += string(r)
		}
	}
	return
}

func HexDump(s string) (int, func(max int) string) {
	lines := strings.Split(s, "\n")
	max := 0

	converted := make([]string, len(lines))
	hexed := make([]string, len(lines))

	for i, line := range lines {
		line += "\n"
		converted[i] = toCharStr(line)
		hexed[i] = toHexStr(line)
		if n := utf8.RuneCountInString(converted[i]); n > max {
			max = n
		}
	}

	return max, func(m int) string {
		r := ""
		for i := 0; i < len(lines); i++ {
			r += fmt.Sprintf("%2d | %-*s | %s\n", i+1, m, converted[i], hexed[i])
		}
		return r
	}
}

func dumpFail(markdown string, want string, given string) {
	nm, sm := HexDump(markdown)
	nw, sw := HexDump(want)
	ng, sg := HexDump(given)

	max := nm
	if nw > max {
		max = nw
	}
	if ng > max {
		max = ng
	}

	fmt.Printf(`----------Markdown----------

%s
------------Want------------

%s
------------Given-----------

%s
`, sm(max), sw(max), sg(max))
}
