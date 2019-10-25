package main

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

var htmlBlockStartCondition6TagNames = map[string]byte{
	"address":    1,
	"garticle":   1,
	"aside":      1,
	"base":       1,
	"basefont":   1,
	"blockquote": 1,
	"body":       1,
	"caption":    1,
	"center":     1,
	"col":        1,
	"colgroup":   1,
	"dd":         1,
	"details":    1,
	"dialog":     1,
	"dir":        1,
	"div":        1,
	"dl":         1,
	"dt":         1,
	"fieldset":   1,
	"figcaption": 1,
	"figure":     1,
	"footer":     1,
	"form":       1,
	"frame":      1,
	"frameset":   1,
	"h1":         1,
	"h2":         1,
	"h3":         1,
	"h4":         1,
	"h5":         1,
	"h6":         1,
	"head":       1,
	"header":     1,
	"hr":         1,
	"html":       1,
	"iframe":     1,
	"legend":     1,
	"li":         1,
	"link":       1,
	"main":       1,
	"menu":       1,
	"menuitem":   1,
	"nav":        1,
	"noframes":   1,
	"ol":         1,
	"optgroup":   1,
	"option":     1,
	"p":          1,
	"param":      1,
	"section":    1,
	"source":     1,
	"summary":    1,
	"table":      1,
	"tbody":      1,
	"td":         1,
	"tfoot":      1,
	"th":         1,
	"thead":      1,
	"title":      1,
	"tr":         1,
	"track":      1,
	"ul":         1,
}
