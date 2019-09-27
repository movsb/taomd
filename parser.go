package main

import (
	"strings"
)

type Parser struct {
}

func isSpace(c rune) bool {
	return c == ' '
}

func skipSpaces(r []rune, atMost int) ([]rune, int) {
	n := 0
	for n < atMost && isSpace(r[n]) {
		n++
	}
	return r[n:], n
}

func skipEnding(c []rune) ([]rune, bool) {
	if len(c) <= 0 {
		return c, true
	}

	if c[0] == '\n' {
		return c[1:], true
	}

	return c, false
}

func is(c []rune, r rune) bool {
	return len(c) > 0 && c[0] == r
}

func in(c []rune, rs ...rune) (rune, bool) {
	if len(c) <= 0 {
		return 0, false
	}

	for _, r := range rs {
		if c[0] == r {
			return r, true
		}
	}

	return 0, false
}

func parse(in string, example int) *Document {
	var doc Document
	doc.example = example

	c := []rune(in)

	var i int

	var lastBlock interface{}
	var thisBlock interface{}

	for len(c) > 0 {
		c, thisBlock = parseBlock(c[i:])
		if _, ok := thisBlock.(*BlankLine); ok {
			lastBlock = nil
			continue
		}
		block, merged := tryMerge(lastBlock, thisBlock)
		if merged {
			if lastBlock == nil {
				doc.blocks = append(doc.blocks, block)
			}
		} else {
			doc.blocks = append(doc.blocks, block)
		}
		lastBlock = block
	}

	return &doc
}

func tryMerge(b1 interface{}, b2 interface{}) (interface{}, bool) {
	switch t1 := b1.(type) {
	default:
		switch t2 := b2.(type) {
		case *Line:
			return &Paragraph{
				texts: []string{
					t2.text,
				},
			}, false
		}
	case nil:
		switch t2 := b2.(type) {
		case *Line:
			return &Paragraph{
				texts: []string{
					t2.text,
				},
			}, true
		}
	case *Paragraph:
		switch t2 := b2.(type) {
		case *Line:
			t1.texts = append(t1.texts, t2.text)
			return t1, true
		}
	}
	return b2, false
}

func parseBlock(c []rune) ([]rune, interface{}) {
	// var content string

	c, n := skipSpaces(c, 4)
	if n >= 0 && n <= 3 {
		if _, ok := in(c, ' ', '\n'); ok {
			if nc, bl := tryParseBlankLine(c); bl != nil {
				return nc, bl
			}
		}
		if r, ok := in(c, '*', '-', '_'); ok {
			if nc, hr := tryParseHorizontalRule(c, r); hr != nil {
				return nc, hr
			}
		}
		if _, ok := in(c, '#'); ok {
			if nc, heading := tryParseAtxHeading(c); heading != nil {
				return nc, heading
			}
		}
		return parseLine(c)
	}

	return c, nil
}

func tryParseBlankLine(c []rune) ([]rune, *BlankLine) {
	i := 0
	for i < len(c) && (c[i] == ' ' || c[i] == '\t') {
		i++
	}
	if i == len(c) {
		return c[i:], &BlankLine{}
	}
	if c[i] == '\n' {
		i++
		return c[i:], &BlankLine{}
	}
	return c, nil
}

func tryParseHorizontalRule(c []rune, start rune) ([]rune, *HorizontalRule) {
	i := 0
	loop := true
	n := 0

	for loop && i < len(c) {
		switch c[i] {
		case start:
			n++
			i++
		case ' ', '\t':
			i++
		default:
			loop = false
		}
	}

	if n < 3 {
		return c, nil
	}

	if nc, ok := skipEnding(c[i:]); ok {
		return nc, &HorizontalRule{}
	}

	return c, nil
}

func parseLine(c []rune) ([]rune, *Line) {
	i, n := 0, len(c)

	// skip to line ending or eof
	for i < n && c[i] != '\n' {
		i++
	}

	end := i

	// \n
	if i < n {
		i++
	}

	return c[i:], &Line{string(c[:end])}
}

func tryParseAtxHeading(c []rune) ([]rune, *Heading) {
	i, n := 0, 0
	for i < len(c) && c[i] == '#' {
		n++
		i++
	}

	// 1–6 unescaped # characters
	if n <= 0 || n >= 7 {
		return c, nil
	}

	// eof
	if i >= len(c) {
		return c[i:], &Heading{Level: n}
	}

	// end of line
	if c[i] == '\n' {
		i++
		return c[i:], &Heading{Level: n}
	}

	// not followed by a space
	if c[i] != ' ' {
		return c, nil
	}

	i++

	start := i

	lastHash := 0

	// skip to line ending
	for i < len(c) && c[i] != '\n' {
		if c[i] == '#' {
			lastHash = i
		}
		i++
	}

	cEnd := i

	// skip \n
	if i < len(c) {
		cEnd++
	}

	end := i

	// The optional closing sequence of #s ...
	if lastHash != 0 {
		// ... be followed by spaces only.
		j := end - 1
		for c[j] == ' ' {
			j--
		}

		// an optional closing sequence of any number of unescaped # characters
		for c[j] == '#' {
			j--
		}

		//  ... preceded by a space ...
		if c[j] == ' ' {
			j--
			end = j + 1
		}
	}

	// The raw contents of the heading are stripped of
	// leading and trailing spaces before being parsed as inline content
	text := string(c[start:end])
	text = strings.TrimSpace(text)

	return c[cEnd:], &Heading{
		Level: n,
		text:  text,
	}
}
