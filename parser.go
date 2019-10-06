package main

import (
	"bufio"
	"bytes"
	"strings"
)

func isText(s []rune) bool {
	var blocks []Blocker
	addLine(&blocks, s)
	if len(blocks) > 0 {
		if _, ok := blocks[0].(*Paragraph); ok {
			return true
		}
	}
	return false
}

func skipPrefixSpaces(s []rune, max int) []rune {
	n := 0
	for len(s) > n && s[n] == ' ' {
		n++
	}
	if n <= max {
		return s[n:]
	}
	return s
}

func skipIf(s []rune, c rune) []rune {
	if len(s) > 0 && s[0] == c {
		s = s[1:]
	}
	return s
}

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

		if len(s) == 1 && s[0] == '\n' {
			blocks = append(blocks, &BlankLine{})
			return true
		}

		if r, ok := in(s, '-', '_', '*'); ok {
			hr := tryParseHorizontalRule(s, r)
			if hr != nil {
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
			heading := tryParseAtxHeading(s)
			if heading != nil {
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
			var bq *BlockQuote
			if len(blocks) > 0 {
				if pbq, ok := blocks[len(blocks)-1].(*BlockQuote); ok {
					bq = pbq
				}
			}
			bq, _ = tryParseBlockQuote(s, bq)
			if bq != nil {
				blocks = append(blocks, bq)
				return true
			}
		}

		_, maybeListMarker := in(s, '-', '+', '*')
		maybeListStart := '0' <= s[0] && s[0] <= '9'
		if maybeListMarker || maybeListStart {
			if list, item, ok := tryParseListItem(s); ok {
				list.Items = append(list.Items, item)
				blocks = append(blocks, list)
				return true
			}
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

func tryParseHorizontalRule(c []rune, start rune) *HorizontalRule {
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
		return nil
	}

	if _, ok := skipEnding(c[i:]); ok {
		return &HorizontalRule{}
	}

	return nil
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

func tryParseAtxHeading(c []rune) *Heading {
	i, n := 0, 0
	for i < len(c) && c[i] == '#' {
		n++
		i++
	}

	// 1–6 unescaped # characters
	if n <= 0 || n >= 7 {
		return nil
	}

	// eof
	if i >= len(c) {
		return &Heading{Level: n}
	}

	// end of line
	if c[i] == '\n' {
		return &Heading{Level: n}
	}

	// not followed by a space
	if c[i] != ' ' {
		return nil
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

	return &Heading{
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
	s = skipPrefixSpaces(s, 3)

	if len(s) == 0 || s[0] != '>' {
		return bq, false
	}

	// skip '>'
	s = s[1:]

	// skip ' '
	s = skipIf(s, ' ')

	if bq == nil {
		bq = &BlockQuote{}
	}
	return bq, addLine(&bq.blocks, s)
}

func tryParseListItem(s []rune) (*List, *ListItem, bool) {
	list := &List{}

	if marker, ok := in(s, '-', '+', '*'); ok {
		list.Ordered = false
		list.Marker = byte(marker)
		list.markerWidth = 1
		s = s[1:]
	} else {
		list.Ordered = true
		start := 0
		i := 0
		for i < len(s) && '0' <= s[i] && s[i] <= '9' {
			start *= 10
			start += int(s[i]) - '0'
			i++
		}
		s = s[i:]
		if len(s) == 0 {
			return nil, nil, false
		}
		switch s[0] {
		default:
			return nil, nil, false
		case '.', ')':
			list.Delimeter = byte(s[0])
			list.Start = start
			list.markerWidth = i + 1
			s = s[1:]
		}
	}

	if len(s) == 0 {
		return nil, nil, false
	}
	_, n := peekSpaces(s, 4)
	if !(1 <= n && n <= 4) {
		return nil, nil, false
	}
	s = s[n:]
	list.spacesWidth = n

	item := &ListItem{}
	item.spaces = list.markerWidth + list.spacesWidth

	if !addLine(&item.blocks, s) {
		return nil, nil, false
	}

	return list, item, true
}
