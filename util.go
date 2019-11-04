package taomd

import (
	"bufio"
	"bytes"
	"container/list"
	"fmt"
	"io"
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
	lines := strings.SplitAfter(s, "\n")
	// remove last empty string
	if n := len(lines); n > 0 && len(lines[n-1]) == 0 {
		lines = lines[0 : n-1]
	}

	max := 0

	converted := make([]string, len(lines))
	hexed := make([]string, len(lines))

	for i, line := range lines {
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

func DumpFail(w io.Writer, markdown string, want string, given string) {
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

	fmt.Fprintf(w, `----------Markdown----------

%s
------------Want------------

%s
------------Given-----------

%s
`, sm(max), sw(max), sg(max))
}

type LineScanner struct {
	scanner *bufio.Scanner
	buffers *list.List
	text    []rune
}

func NewLineScanner(in io.Reader) *LineScanner {
	scn := bufio.NewScanner(in)
	scn.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.IndexByte(data, '\n'); i >= 0 {
			// We have a full newline-terminated line.
			return i + 1, data[0 : i+1], nil // \n is returned
		}
		// If we're at EOF, we have a final, non-terminated line. Return it.
		if atEOF {
			return len(data), data, nil
		}
		// Request more data.
		return 0, nil, nil
	})
	return &LineScanner{
		scanner: scn,
		buffers: list.New(),
	}
}

func (ls *LineScanner) Scan() bool {
	if ls.buffers.Len() > 0 {
		ls.text = ls.buffers.Remove(ls.buffers.Front()).([]rune)
		return true
	}
	if ls.scanner.Scan() {
		ls.text = []rune(ls.scanner.Text())
		return true
	}
	return false
}

func (ls *LineScanner) Text() []rune {
	return ls.text
}

func (ls *LineScanner) PutBack(s []rune) {
	ls.buffers.PushFront(s)
}
