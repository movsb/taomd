package main

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

func parse(in string) *Document {
	var doc Document

	c := []rune(in)

	var block interface{}
	var i int

	for len(c) > 0 {
		c, block = parseBlock(c[i:])
		doc.blocks = append(doc.blocks, block)
	}

	return &doc
}

func parseBlock(c []rune) ([]rune, interface{}) {
	// var content string

	c, n := skipSpaces(c, 4)
	if n >= 0 && n <= 3 {
		if r, ok := in(c, '*', '-', '_'); ok {
			if nc, hr := tryParseHorizontalRule(c, r); hr != nil {
				return nc, hr
			}
		}
		return parseParagraph(c)
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

func parseParagraph(c []rune) ([]rune, *Paragraph) {
	i, n := 0, len(c)
	end := 0
	for {
		// skip to line ending or eof
		for i < n && c[i] != '\n' {
			i++
		}
		// eof
		if i >= n {
			end = i
			break
		}
		// \n
		i++
		if i >= n {
			end = i - 1
			break
		}
		if i < n && c[i] == '\n' {
			end = i - 1
			i++
			break
		}
	}
	return c[i:], &Paragraph{string(c[:end])}
}
