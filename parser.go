package main

import (
	"bufio"
	"bytes"
	"container/list"
	"io"
	"strings"
)

func skipPrefixSpaces(s []rune, max int) (int, []rune) {
	n := 0
	for len(s) > n && s[n] == ' ' {
		n++
	}
	if max == -1 || n <= max {
		return n, s[n:]
	}
	return n, s
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

	os := s

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
			rs, heading := tryParseSetextHeadingUnderline(s)
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
			list := &List{}
			if list.AddLine(os) {
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

func parse(in io.Reader, example int) *Document {
	var doc Document
	doc.example = example

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

	//var i int

	//var lastBlock interface{}
	//var thisBlock interface{}

	for scn.Scan() {
		s := scn.Text()
		doc.AddLine([]rune(s))
	}

	doc.parseInlines()

	return &doc
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

func tryParseSetextHeadingUnderline(c []rune) ([]rune, *Heading) {
	start := c[0]
	i := 0
	for c[i] == start {
		i++
	}
	for c[i] == ' ' {
		i++
	}
	if i == len(c) || c[i] == '\n' {
		level := 1
		if start == '-' {
			level = 2
		}
		return []rune{}, &Heading{
			Level: level,
		}
	}
	return c, nil
}

func tryParseFencedCodeBlockStart(s []rune, start rune, indent int) *CodeBlock {
	cb := &CodeBlock{}
	cb.fenceIndent = indent
	cb.fenceStart = start

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
	_, s = skipPrefixSpaces(s, 3)

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

func parseInlinesToDeimiters(raw string) (*list.List, *list.List) {
	c := []rune(raw)
	texts := list.New()
	text := []rune{}
	delimiters := list.New()

	appendDelimiter := func(s string) {
		if len(text) > 0 {
			texts.PushFront(&Text{
				Text: string(text),
			})
			text = text[:0]
		}
		if s != "" {
			t := &Text{
				Text: s,
			}
			te := texts.PushFront(t)
			d := &Delimiter{
				textElement: te,
				active:      true,
				text:        s,
			}
			delimiters.PushFront(d)
		}
	}

	i := 0
	for i < len(c) {
		switch ch := c[i]; ch {
		case '*', '_':
			start := i
			end := i
			for end < len(c) && c[end] == ch {
				end++
			}
			appendDelimiter(string(c[start:end]))
			i = end
		case '!':
			i++
			if i < len(c) && c[i] == '[' {
				appendDelimiter("![")
				i++
			} else {
				text = append(text, '!')
				i++
			}
		case '[':
			appendDelimiter("[")
			i++
		case ']':
			appendDelimiter("")
			i++
			var opener *Delimiter
			var openerElement *list.Element
			for e := delimiters.Front(); e != nil; e = e.Next() {
				d := e.Value.(*Delimiter)
				if d.text == "[" || d.text == "![" {
					openerElement = e
					opener = d
					break
				}
			}
			if opener == nil {
				appendDelimiter("")
				texts.PushFront(&Text{
					Text: "]",
				})
				continue
			}
			if !opener.active {
				delimiters.Remove(openerElement)
				appendDelimiter("")
				texts.PushFront(&Text{
					Text: "]",
				})
				continue
			}

			var link Link
			var image Image
			var nc []rune
			var ok bool

			if opener.text == "[" {
				nc, ok = parseLink(c[i:], &link)
			} else {
				nc, ok = parseImage(c[i:], &image)
			}

			if !ok {
				delimiters.Remove(openerElement)
				appendDelimiter("")
				texts.PushFront(&Text{
					Text: "]",
				})
				continue
			}

			c = nc
			i = 0

			if opener.text == "[" {
				for e := opener.textElement.Prev(); e != nil; e = e.Prev() {
					link.Inlines = append(link.Inlines, e.Value.(*Text))
					texts.Remove(e)
				}

			} else {
				for e := opener.textElement.Prev(); e != nil; e = e.Prev() {
					image.inlines = append(image.inlines, e.Value.(*Text))
					texts.Remove(e)
				}
				image.Alt = textOnlyFromInlines(image.inlines)
			}

			texts.Remove(opener.textElement)

			// TODO process inlines for [ ]

			if opener.text == "[" {
				for e := openerElement.Next(); e != nil; e = e.Next() {
					d := e.Value.(*Delimiter)
					if d.text == "[" {
						d.active = false
					}
				}
			}

			delimiters.Remove(openerElement)

			if opener.text == "[" {
				texts.PushFront(&link)
			} else {
				texts.PushFront(&image)
			}
		case '\\':
			if j := i + 1; j < len(c) && isPunctuation(c[j]) {
				text = append(text, c[j])
				i++
				i++
				continue
			}
			text = append(text, '\\')
			i++
		case '<':
			if nc, link, ok := parseAutoLink(c); ok {
				c = nc
				i = 0
				texts.PushFront(link)
				continue
			}
			text = append(text, '<')
			i++
		default:
			text = append(text, ch)
			i++
		}
	}

	appendDelimiter("")

	return texts, delimiters
}

func parseInlines(raw string) []Inline {
	texts, delimiters := parseInlinesToDeimiters(raw)
	_ = delimiters
	inlines := []Inline{}
	for t := texts.Back(); t != nil; t = t.Prev() {
		inlines = append(inlines, t.Value)
	}
	return inlines
}

func parseEmphases(texts *list.List, delimiters *list.List, bottom *Delimiter) []Inline {
	return nil
}

func textOnlyFromInlines(inlines []*Text) string {
	s := ""
	for _, inline := range inlines {
		s += inline.Text
	}
	return s
}

func parseEscape(c []rune) (rune, bool) {
	i := 0
	if j := i + 1; j < len(c) && isPunctuation(c[j]) {
		return c[j], true
	}
	return 0, false
}

func parseLink(c []rune, link *Link) ([]rune, bool) {
	if len(c) == 0 || c[0] != '(' {
		return c, false
	}
	c = c[1:]
	i := 0
	_, c = skipPrefixSpaces(c, -1)
	if len(c) == 0 {
		return nil, false
	}
	angle := c[0] == '<'
	if angle {
		c = c[1:]
		dest := []rune{}
		i := 0
		for i < len(c) && c[i] != '>' && c[i] != '\n' {
			if c[i] == '\\' {
				if j := i + 1; j < len(c) && c[j] == '>' {
					dest = append(dest, '>')
					i++ // skip \
					i++ // skip >
					continue
				}
			}
			dest = append(dest, c[i])
			i++
		}
		if i == len(c) || c[i] != '>' {
			return nil, false
		}
		link.Link = string(dest)
		i++ // skip '>'
		c = c[i:]
	} else {
		dest := []rune{}
		i := 0
		nParen := 0
		for i < len(c) {
			if c[i] <= ' ' || (c[i] == ')' && nParen == 0) {
				break
			}
			switch c[i] {
			default:
				dest = append(dest, c[i])
				i++
			case '(':
				nParen++
				dest = append(dest, '(')
				i++
			case ')':
				nParen--
				dest = append(dest, ')')
				i++
			case '\\':
				// Parentheses inside the link destination may be escaped:
				// Parentheses and other symbols can also be escaped, as usual in Markdown:
				if r, ok := parseEscape(c[i:]); ok {
					dest = append(dest, r)
					i += 2
					continue
				}
				dest = append(dest, '\\')
				i++
			}
		}
		link.Link = string(dest)
		c = c[i:]
	}

	_, c = skipPrefixSpaces(c, -1)

	if len(c) == 0 {
		return nil, false
	}

	// The title may be omitted
	if c[0] == ')' {
		return c[1:], true
	}

	// parse title
	marker, ok := in(c, '\'', '"', '(')
	if !ok {
		return nil, false
	}

	start := 1
	i = 1
	for i < len(c) && (c[i] != marker && !(marker == '(' && c[i] == ')')) {
		i++
	}

	if i == len(c) {
		return nil, false
	}

	link.Title = string(c[start:i])

	i++ // skip marker

	c = c[i:]
	i = 0
	_, c = skipPrefixSpaces(c, -1)
	if len(c) == 0 || c[0] != ')' {
		return nil, false
	}

	i++
	return c[i:], true
}

func parseImage(c []rune, image *Image) ([]rune, bool) {
	if len(c) == 0 || c[0] != '(' {
		return c, false
	}
	c = c[1:]
	i := 0
	_, c = skipPrefixSpaces(c, -1)
	if len(c) == 0 {
		return nil, false
	}
	angle := c[0] == '<'
	if angle {
		c = c[1:]
		dest := []rune{}
		i := 0
		for i < len(c) && c[i] != '>' && c[i] != '\n' {
			if c[i] == '\\' {
				if j := i + 1; j < len(c) && c[j] == '>' {
					dest = append(dest, '>')
					i++ // skip \
					i++ // skip >
					continue
				}
			}
			dest = append(dest, c[i])
			i++
		}
		if i == len(c) || c[i] != '>' {
			return nil, false
		}
		image.Link = string(dest)
		i++ // skip '>'
		c = c[i:]
	} else {
		dest := []rune{}
		i := 0
		nParen := 0
		for i < len(c) {
			if c[i] <= ' ' || (c[i] == ')' && nParen == 0) {
				break
			}
			switch c[i] {
			default:
				dest = append(dest, c[i])
				i++
			case '(':
				nParen++
				dest = append(dest, '(')
				i++
			case ')':
				nParen--
				dest = append(dest, ')')
				i++
			case '\\':
				// Parentheses inside the link destination may be escaped:
				// Parentheses and other symbols can also be escaped, as usual in Markdown:
				if r, ok := parseEscape(c[i:]); ok {
					dest = append(dest, r)
					i += 2
					continue
				}
				dest = append(dest, '\\')
				i++
			}
		}
		image.Link = string(dest)
		c = c[i:]
	}

	_, c = skipPrefixSpaces(c, -1)

	if len(c) == 0 {
		return nil, false
	}

	// The title may be omitted
	if c[0] == ')' {
		return c[1:], true
	}

	// parse title
	marker, ok := in(c, '\'', '"', '(')
	if !ok {
		return nil, false
	}

	start := 1
	i = 1
	for i < len(c) && (c[i] != marker && !(marker == '(' && c[i] == ')')) {
		i++
	}

	if i == len(c) {
		return nil, false
	}

	image.Title = string(c[start:i])

	i++ // skip marker

	c = c[i:]
	i = 0
	_, c = skipPrefixSpaces(c, -1)
	if len(c) == 0 || c[0] != ')' {
		return nil, false
	}

	i++
	return c[i:], true
}

func parseAutoLink(c []rune) ([]rune, *Link, bool) {
	if len(c) == 0 {
		return nil, nil, false
	}

	i := 0
	if c[i] != '<' {
		return nil, nil, false
	}

	i++ // skip '<'

	if i == len(c) {
		return nil, nil, false
	}

	switch {
	case 'a' <= c[i] && c[i] <= 'z':
		break
	case 'A' <= c[i] && c[i] <= 'Z':
		break
	default:
		return nil, nil, false
	}

	isSchemeChar := func(ch rune) bool {
		switch {
		case 'a' <= ch && ch <= 'z':
			return true
		case 'A' <= ch && ch <= 'Z':
			return true
		case '0' <= ch && ch <= '9':
			return true
		default:
			switch ch {
			case '+', '.', '-':
				return true
			}
			return false
		}
	}

	schemeStart := i
	for i < len(c) && isSchemeChar(c[i]) {
		i++
	}

	// 2–32 characters
	if n := i - schemeStart; n < 2 || n > 32 {
		return nil, nil, false
	}

	if i == len(c) {
		return nil, nil, false
	}

	if c[i] != ':' {
		return nil, nil, false
	}

	i++

	// parse opaque
	for i < len(c) {
		if c[i] <= ' ' {
			break
		}
		if c[i] == '<' || c[i] == '>' {
			break
		}
		i++
	}
	if i == len(c) {
		return nil, nil, false
	}
	if c[i] != '>' {
		return nil, nil, false
	}

	full := string(c[schemeStart:i])
	link := &Link{
		Link: full,
		Inlines: []*Text{
			&Text{
				Text: full,
			},
		},
	}
	return c[i+1:], link, true
}
