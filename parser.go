package main

import (
	"strings"
)

type Parser struct {
}

func isSpace(c rune) bool {
	return c == ' '
}

func escapeHTML(r rune) []rune {
	switch r {
	case '"':
		return []rune("&quot;")
	case '&':
		return []rune("&amp;")
	case '<':
		return []rune("&lt;")
	case '>':
		return []rune("&gt;")
	}
	return []rune{r}
}

var punctuations = map[rune]int{
	'!':  1,
	'"':  1,
	'#':  1,
	'$':  1,
	'%':  1,
	'&':  1,
	'\'': 1,
	'(':  1,
	')':  1,
	'*':  1,
	'+':  1,
	',':  1,
	'-':  1,
	'.':  1,
	'/':  1,
	':':  1,
	';':  1,
	'<':  1,
	'=':  1,
	'>':  1,
	'?':  1,
	'@':  1,
	'[':  1,
	'\\': 1,
	']':  1,
	'^':  1,
	'_':  1,
	'`':  1,
	'{':  1,
	'|':  1,
	'}':  1,
	'~':  1,
}

func isPunctuation(r rune) bool {
	_, ok := punctuations[r]
	return ok
}

func peekSpaces(r []rune, atMost int) ([]rune, int) {
	n := 0
	for n < atMost && n < len(r) && isSpace(r[n]) {
		n++
	}
	return r, n
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
			switch lastBlock.(type) {
			default:
				lastBlock = nil
				continue
			case *CodeBlock:
				break
			}
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

	for _, block := range doc.blocks {
		if parser, ok := block.(interface {
			ParseInlines()
		}); ok {
			parser.ParseInlines()
		}
	}

	return &doc
}

func tryMerge(b1 interface{}, b2 interface{}) (interface{}, bool) {
	switch t1 := b1.(type) {
	case *Paragraph:
		switch t2 := b2.(type) {
		case *Line:
			t1.texts = append(t1.texts, t2.text)
			return t1, true
		// An indented code block cannot interrupt a paragraph
		case *_CodeChunk:
			t1.texts = append(t1.texts, t2.text)
			return t1, true
		}
	case *CodeBlock:
		switch t2 := b2.(type) {
		case *_CodeChunk:
			t1.chunks = append(t1.chunks, t2)
			return t1, true
		case *BlankLine:
			t1.chunks = append(t1.chunks, &_CodeChunk{
				text: "",
			})
			return t1, true
		}
	}

	switch t2 := b2.(type) {
	case *Line:
		return &Paragraph{
			texts: []string{
				t2.text,
			},
		}, false
	case *_CodeChunk:
		return &CodeBlock{
			chunks: []*_CodeChunk{t2},
		}, false
	}

	return b2, false
}

func parseBlock(c []rune) ([]rune, interface{}) {
	// var content string

	_, n := peekSpaces(c, 4)
	if n >= 0 && n <= 3 {
		oc := c
		c = c[n:]
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
		if _, ok := in(c, '`', '~'); ok {
			if nc, code := tryParseFencedCodeBlock(oc, n); code != nil {
				return nc, code
			}
		}
		return parseLine(c)
	} else if n == 4 {
		return parseIndentedCodeChunk(c)
	}

	return c, nil
}

func parseIndentedCodeChunk(c []rune) ([]rune, *_CodeChunk) {
	c, line := parseLine(c[4:])
	return c, &_CodeChunk{
		text: line.text,
	}
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
		if j > start {
			if c[j] == ' ' {
				j--
				end = j + 1
			}
		} else {
			end = start
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

func tryParseFencedCodeBlock(c []rune, indent int) ([]rune, *CodeBlock) {
	old := c
	leadingSpaces := 0

	if _, n := peekSpaces(c, indent); n > 0 {
		c = c[n:]
		leadingSpaces = n
	}

	i, n := 0, 0
	sign := c[0]

	// at least three consecutive backtick characters (`) or tildes (~)
	for i < len(c) && c[i] == sign {
		n++
		i++
	}

	if n < 3 {
		return old, nil
	}

	// The line with the opening code fence may optionally
	// contain some text following the code fence;
	// this is trimmed of leading and trailing whitespace and called the info string.
	infoStart := i

	j := infoStart
	for j < len(c) && c[j] != '\n' {
		j++
	}

	// eof
	if j == len(c) {
		return c[j:], &CodeBlock{}
	}

	j++
	infoEnd := j

	infoText := strings.TrimSpace(string(c[infoStart:infoEnd]))

	// TODO If the info string comes after a backtick fence, it may not contain any backtick characters.

	// The content of the code block consists of all subsequent lines,
	var lines []string
	c = c[j:]
	for len(c) > 0 {
		var line *Line
		c, line = parseLine(c)
		sn := 0
		// If the leading code fence is indented N spaces,
		// then up to N spaces of indentation are removed
		s := strings.TrimLeftFunc(line.text, func(r rune) bool {
			if sn < leadingSpaces && r == ' ' {
				sn++
				return true
			}
			return false
		})
		// until a closing code fence of the same type as the code block
		// began with (backticks or tildes), and with at least as many backticks
		// or tildes as the opening code fence.
		sn = 0
		trimS := strings.TrimLeftFunc(s, func(r rune) bool {
			if sn < 3 && r == ' ' {
				sn++
				return true
			}
			return false
		})
		startSign := strings.Repeat(string(sign), n)
		allSign := strings.Repeat(string(sign), len(trimS))
		if strings.HasPrefix(trimS, startSign) && trimS == allSign {
			break
		}
		lines = append(lines, s)
	}

	return c, &CodeBlock{
		Lang: infoText,
		chunks: []*_CodeChunk{
			&_CodeChunk{
				text: strings.Join(lines, "\n"),
			},
		},
	}
}
