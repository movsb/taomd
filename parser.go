package main

import (
	"bufio"
	"bytes"
	"strings"
)

func addLine(pBlocks *[]Blocker, s []rune) bool {
	if len(s) == 0 {
		return false
	}

	blocks := *pBlocks

	defer func() {
		*pBlocks = blocks
	}()

	if len(blocks) > 0 && blocks[len(blocks)-1].AddLine(s) {
		return true
	}

	_, n := peekSpaces(s, 4)
	if n >= 0 && n < 4 {
		s = s[n:]
		if len(s) == 0 {
			return false
		}

		if r, ok := in(s, '-', '_', '*'); ok {
			rs, hr := tryParseHorizontalRule(s, r)
			if hr != nil {
				_ = rs
				if n := len(blocks); n > 0 && r == '-' {
					if p, ok := blocks[n-1].(*Paragraph); ok {
						heading := &Heading{
							Level: 2,
							text:  strings.Join(p.texts, ""),
						}
						blocks[n-1] = heading
						return true
					}
				}
				blocks = append(blocks, hr)
				return true
			}
		}

		if _, ok := in(s, '#'); ok {
			rs, heading := tryParseAtxHeading(s)
			if heading != nil {
				_ = rs
				blocks = append(blocks, heading)
				return true
			}
		}

		if r, ok := in(s, '=', '-'); ok {
			_ = r
			rs, heading := tryParseSetextHeading(s)
			if heading != nil {
				_ = rs
				if n := len(blocks); n > 0 {
					if p, ok := blocks[n-1].(*Paragraph); ok {
						heading.text = strings.Join(p.texts, "")
						blocks[n-1] = heading
						return true
					}
				}
			}
		}

		if r, ok := in(s, '`', '~'); ok {
			cb := tryParseFencedCodeBlockStart(s, r, n)
			if cb != nil {
				blocks = append(blocks, cb)
				return true
			}
		}

		if _, ok := in(s, '>'); ok {
			bq, _ := tryParseBlockQuote(s, nil)
			if bq != nil {
				blocks = append(blocks, bq)
				return true
			}
		}

		if marker, ok := in(s, '-', '+', '*'); ok {
			_ = marker
		}

		p := &Paragraph{}
		p.texts = append(p.texts, string(s))
		blocks = append(blocks, p)
		return true
	} else if n == 4 {
		var cb *CodeBlock
		if len(blocks) > 0 {
			if pcb, ok := blocks[len(blocks)-1].(*CodeBlock); ok {
				cb = pcb
			}
		}
		if cb == nil {
			cb = &CodeBlock{}
			blocks = append(blocks, cb)
		}
		s = s[4:]
		return cb.AddLine(s)
	}
	//}
	return false
}

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

func escapeHTMLString(s string) string {
	return strings.NewReplacer(
		`"`, "&quot;",
		`&`, "&amp;",
		`<`, "&lt;",
		`>`, "&gt;",
	).Replace(s)
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

	scn := bufio.NewScanner(strings.NewReader(in))
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

	//var i int

	//var lastBlock interface{}
	//var thisBlock interface{}

	for scn.Scan() {
		s := scn.Text()
		doc.AddLine([]rune(s))
	}

	return &doc

	/*
		c := []rune{}

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
	*/
}

func parseBlock(c []rune) ([]rune, interface{}) {
	// var content string

	_, n := peekSpaces(c, 4)
	if n >= 0 && n <= 3 {
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

func tryParseSetextHeading(c []rune) ([]rune, *Heading) {
	return nil, &Heading{}
}

func tryParseFencedCodeBlockStart(s []rune, start rune, indent int) *CodeBlock {
	cb := &CodeBlock{}
	cb.indent = indent
	cb.start = start

	i := 0
	for i < len(s) && s[i] == start {
		i++
	}

	// at least three consecutive backtick characters (`) or tildes (~)
	if i < 3 {
		return nil
	}

	cb.fenceLength = i

	// The line with the opening code fence may optionally
	// contain some text following the code fence;
	// this is trimmed of leading and trailing whitespace and called the info string.
	infoStart := i

	j := infoStart
	for j < len(s) && s[j] != '\n' {
		j++
	}

	// eof
	if j == len(s) {
		return cb
	}

	// j is at '\n'
	j++
	infoEnd := j

	info := strings.TrimSpace(string(s[infoStart:infoEnd]))

	// If the info string comes after a backtick fence, it may not contain any backtick characters.
	if start == '`' && strings.IndexByte(info, byte('`')) != -1 {
		return nil
	}

	cb.Info = info

	return cb
}

func tryParseCodeSpan(c []rune) ([]rune, *CodeSpan) {
	i := 0

	// a string of one or more backtick characters
	startTickCount := 0
	for i < len(c) && c[i] == '`' {
		startTickCount++
		i++
	}

	// eof, leave it unchanged
	if i == len(c) {
		return c, nil
	}

	var text string

	ap := func(s []rune) {
		text += string(s)
	}

	for i < len(c) {
		start := i

		for {
			// count to line ending
			for i < len(c) && c[i] != '`' && c[i] != '\n' {
				i++
			}

			// eof
			if i == len(c) {
				return c, nil
			}

			// eol
			if c[i] == '\n' {
				// for `First, line endings are converted to spaces.`
				ap(c[start:i])
				ap([]rune(" "))
				i++
				break
			}

			endTickStart := i

			for i < len(c) && c[i] == '`' {
				i++
			}

			if i-endTickStart == startTickCount {
				ap(c[start:endTickStart])
				goto exit
			}

			// eof
			if i == len(c) {
				return c, nil
			}
		}
	}

exit:

	return c[i:], &CodeSpan{
		text: text,
	}
}

func tryParseBlockQuote(s []rune, bq *BlockQuote) (*BlockQuote, bool) {
	// skip '>'
	s = s[1:]

	// skip ' '
	if len(s) > 0 && s[0] == ' ' {
		s = s[1:]
	}

	if bq == nil {
		bq = &BlockQuote{}
	}
	return bq, addLine(&bq.blocks, s)
}

func tryParseListItem(s []rune) {

}
