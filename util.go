package main

import (
	"fmt"
	"strings"
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
			c += "...."
		case '\n':
			c += "."
		default:
			c += string(r)
		}
	}
	return
}

func HexDump(s string) (r string) {
	lines := strings.Split(s, "\n")
	max := 0

	converted := make([]string, len(lines))
	hexed := make([]string, len(lines))

	for i, line := range lines {
		line += "\n"
		converted[i] = toCharStr(line)
		hexed[i] = toHexStr(line)
		if n := len(converted[i]); n > max {
			max = n
		}
	}

	for i := 0; i < len(lines); i++ {
		r += fmt.Sprintf("%2d | %-*s | %s\n", i+1, max, converted[i], hexed[i])
	}

	return r
}

func dumpFail(markdown string, want string, given string) {
	fmt.Printf(`----------Markdown----------

%s
------------Want------------

%s
------------Given-----------

%s
`, markdown, HexDump(want), HexDump(given))
}
