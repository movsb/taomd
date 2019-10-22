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

		if s[0] == '[' {
			if link := tryParseLinkReferenceDefinition(s); link != nil {
				blocks = append(blocks, link)
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

func any(c rune, rs ...rune) bool {
	_, ok := in([]rune{c}, rs...)
	return ok
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

	appendText := func(inline Inline) {
		if len(text) > 0 {
			texts.PushFront(&Text{
				Text: string(text),
			})
			text = text[:0]
		}
		if inline != nil {
			texts.PushFront(inline)
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
			if j := i + 1; j < len(c) {
				if isPunctuation(c[j]) {
					text = append(text, c[j])
					i++
					i++
					continue
				} else if c[j] == '\n' {
					// A backslash at the end of the line is a hard line break
					appendText(&HardLineBreak{})
					i++
					i++
					continue
				}
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
			if nc, tag := tryParseHtmlTag(c[i:]); tag != nil {
				i = 0
				c = nc
				appendText(tag)
				continue
			}
			text = append(text, '<')
			i++
		case '`':
			if nc, span := tryParseCodeSpan(c[i:]); span != nil {
				i = 0
				c = nc
				appendText(span)
				continue
			}
			j := i
			for ; j < len(c) && c[j] == '`'; j++ {
				j++
			}
			appendText(&Text{
				Text: strings.Repeat("`", j-i+1),
			})
			i = j
		case '&':
			if nc, entity, ok := tryParseHtmlEntity(c[i:]); ok {
				i = 0
				c = nc
				if entity == 0 {
					entity = 0xFFFD
				}
				appendText(&Text{
					Text: string(entity),
				})
				continue
			}
			appendText(&Text{
				Text: "&",
			})
			i++
		case '\n':
			appendText(&Text{
				Text: "\n",
			})
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
	parseLineBreaks(texts)
	return parseEmphases(texts, delimiters, nil)
}

func parseLineBreaks(texts *list.List) {
	if texts.Len() < 3 {
		return
	}

	hasTwoSpaces := func(s string) bool {
		n := len(s)
		return n >= 2 && s[n-2] == ' ' && s[n-1] == ' '
	}

	current := texts.Back().Prev()
	last := texts.Front()

	for current != last {
		prevText, ok2 := current.Next().Value.(*Text)
		currText, ok1 := current.Value.(*Text)
		if ok1 && ok2 {
			if currText.Text == "\n" {
				if hasTwoSpaces(prevText.Text) {
					hard := &HardLineBreak{}
					texts.InsertAfter(hard, current)
				} else {
					soft := &SoftLineBreak{}
					texts.InsertAfter(soft, current)
				}
				// Spaces at the end of the line and beginning of the next line are removed
				// beginnings are remove while adding line to paragraph.
				prevText.Text = strings.TrimRight(prevText.Text, " ")
				next := current.Prev()
				texts.Remove(current)
				current = next
				continue
			}
		}
		current = current.Prev()
	}
}

func parseEmphases(texts *list.List, delimiters *list.List, bottom *list.Element) []Inline {
	if bottom == nil {
		bottom = delimiters.Back()
	}
	openersBottom := map[string]*list.Element{
		"*":  bottom,
		"_":  bottom,
		"**": bottom,
		"__": bottom,
	}

	var closer *list.Element

	closer = bottom

	for {
		var cd *Delimiter

		for ; closer != nil; closer = closer.Prev() {
			cd = closer.Value.(*Delimiter)
			if cd.canCloseEmphasis() {
				break
			}
		}

		if closer == nil {
			break
		}

		var opener *list.Element
		var od *Delimiter

		for opener = closer.Next(); opener != nil; opener = opener.Next() {
			od = opener.Value.(*Delimiter)
			if od.canOpenEmphasis() && od.text == closer.Value.(*Delimiter).text {
				break
			}
		}

		if opener == nil {
			openersBottom[cd.text] = closer
			next := closer.Prev()
			if !cd.canOpenEmphasis() {
				delimiters.Remove(closer)
			}
			closer = next
			continue
		}

		if !od.canOpenEmphasis() {
			e := closer
			closer = closer.Prev()
			delimiters.Remove(e)
			continue
		}

		e := &Emphasis{}
		e.Delimiter = od.text
		for t := od.textElement.Prev(); t != cd.textElement; {
			e.Inlines = append(e.Inlines, t.Value)
			t = t.Prev()
		}
		texts.InsertAfter(e, od.textElement)

		for t := od.textElement; t != nil; {
			next := t.Prev()
			texts.Remove(t)
			if t == closer.Value.(*Delimiter).textElement {
				break
			}
			t = next
		}

		delimiters.Remove(opener)
		save := closer.Prev()
		delimiters.Remove(closer)
		closer = save
	}

	var inlines []Inline
	for t := texts.Back(); t != nil; t = t.Prev() {
		inlines = append(inlines, t.Value)
	}
	return inlines
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

	// after this, c is reset
	c, dest, ok := parseLinkDestination(c)
	if !ok {
		return nil, false
	}

	link.Link = dest
	i = 0

	// The title may be omitted
	if c[0] == ')' {
		return c[1:], true
	}

	c, title, ok := parseLinkTitle(c)
	if !ok {
		return nil, false
	}

	link.Title = title
	i = 0

	_, c = skipPrefixSpaces(c, -1)
	if len(c) == 0 || c[0] != ')' {
		return nil, false
	}

	i++
	return c[i:], true
}

func parseLinkDestination(c []rune) ([]rune, string, bool) {
	if len(c) == 0 {
		return c, "", false
	}

	i := 0
	dest := []rune{}

	angle := c[0] == '<'
	if angle {
		c = c[1:]
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
			return nil, "", false
		}
		i++ // skip '>'
	} else {
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
	}

	return c[i:], string(dest), true
}

func parseLinkTitle(c []rune) ([]rune, string, bool) {
	_, c = skipPrefixSpaces(c, -1)

	if len(c) == 0 {
		return nil, "", false
	}

	// parse title
	marker, ok := in(c, '\'', '"', '(')
	if !ok {
		return nil, "", false
	}

	start := 1
	i := 1

	for i < len(c) && (c[i] != marker && !(marker == '(' && c[i] == ')')) {
		i++
	}

	if i == len(c) {
		return nil, "", false
	}

	title := string(c[start:i])

	i++ // skip marker

	return c[i:], title, true
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

func isHexDigit(r rune) bool {
	return '0' <= r && r <= '9' ||
		'a' <= r && r <= 'f' ||
		'A' <= r && r <= 'F'
}

func isAlNum(r rune) bool {
	return '0' <= r && r <= '9' ||
		'a' <= r && r <= 'z' ||
		'A' <= r && r <= 'Z'
}

func isAlpha(r rune) bool {
	return 'a' <= r && r <= 'z' ||
		'A' <= r && r <= 'Z'
}

func isNum(r rune) bool {
	return '0' <= r && r <= '9'
}

func isWahitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n'
}

func tryParseHtmlEntity(c []rune) ([]rune, rune, bool) {
	i := 0
	if len(c) < 1 || c[i] != '&' {
		return c, 0, false
	}

	i++
	if i == len(c) {
		return c, 0, false
	}

	switch c[i] {
	default:
		j := i
		for j < len(c) && isAlNum(c[j]) {
			j++
		}
		if j < len(c) && c[j] == ';' {
			j++ // after ';'
		}
		if codepoint, ok := htmlEntities[string(c[0:j])]; ok {
			return c[j:], codepoint, true
		}
		return c, 0, false
	case '#':
		i++
		if i == len(c) {
			return c, 0, false
		}
		switch c[i] {
		default:
			j := i
			n := 0
			for j < len(c) && isNum(c[j]) {
				n *= 10
				n += int(c[j]) - '0'
				j++
			}
			if j == i || j-i > 7 {
				return c, 0, false
			}
			if j == len(c) || c[j] != ';' {
				return c, 0, false
			}
			j++
			return c[j:], rune(n), true
		case 'x', 'X':
			i++
			j := i
			n := 0
			for isHexDigit(c[j]) {
				n *= 16
				switch {
				case '0' <= c[j] && c[j] <= '9':
					n += int(c[j]) - '0'
				case 'a' <= c[j] && c[j] <= 'f':
					n += int(c[j]) - 'a'
				case 'A' <= c[j] && c[j] <= 'F':
					n += int(c[j]) - 'A'
				}
				j++
			}
			if j == i || j-i > 6 {
				return c, 0, false
			}
			if j == len(c) || c[j] != ';' {
				return c, 0, false
			}
			j++
			return c[j:], rune(n), true
		}
	}
}

func tryParseHtmlTag(c []rune) ([]rune, *HtmlTag) {
	i := 0
	if i == len(c) || c[i] != '<' {
		return c, nil
	}

	i++ // skip '<'

	if i == len(c) {
		return c, nil
	}

	tryParseTagName := func() int {
		j := i

		// A tag name consists of an ASCII letter
		if j == len(c) || !isAlpha(c[j]) {
			return j
		}

		j++

		// followed by zero or more ASCII letters, digits, or hyphens (-).
		for j < len(c) && (isAlNum(c[j]) || c[j] == '-') {
			j++
		}

		return j
	}

	tryParseAttributeName := func() int {
		j := i

		// An attribute name consists of an ASCII letter, _, or :,
		if j == len(c) || !(isAlpha(c[j]) || c[j] == '_' || c[j] == ':') {
			return j
		}

		j++

		// followed by zero or more ASCII letters, digits, _, ., :, or -.
		for j < len(c) && (isAlNum(c[j]) || any(c[j], '_', '.', ':', '-')) {
			j++
		}

		return j
	}

	tryParseAttributeValue := func() (end int, quoted bool) {
		j := i

		if j == len(c) {
			return j, false
		}

		switch c[j] {
		// A single-quoted attribute value consists of ', zero or more characters not including ', and a final '.
		// A double-quoted attribute value consists of ", zero or more characters not including ", and a final ".
		case '\'', '"':
			q := c[j]
			j++
			for j < len(c) && c[j] != q {
				j++
			}
			if j == len(c) || c[j] != q {
				return i, false
			}
			return j + 1, true
		// An unquoted attribute value is a nonempty string of characters not including whitespace, ", ', =, <, >, or `.
		default:
			for j < len(c) && !(isWahitespace(c[j]) || any(c[j], '"', '\'', '=', '<', '>', '`')) {
				j++
			}
			return j, false
		}
	}

	skipWhitespaces := func(atLeast int) bool {
		j := i
		for j < len(c) && isWahitespace(c[j]) {
			j++
		}
		if j-i >= atLeast {
			i = j
			return true
		}
		return false
	}

	switch c[i] {
	// A closing tag consists of the string </, a tag name, optional whitespace, and the character >.
	case '/':
		i++

		j := tryParseTagName()
		// A tag name is non-empty.
		if j == i {
			return c, nil
		}
		i = j

		skipWhitespaces(0)

		if i == len(c) || c[i] != '>' {
			return c, nil
		}

		i++

		return c[i:], &HtmlTag{
			Tag: string(c[0:i]),
		}
	// A processing instruction consists of the string <?,
	// a string of characters not including the string ?>, and the string ?>.
	case '?':
		i++

		for i < len(c) {
			for i < len(c) && c[i] != '>' {
				i++
			}

			if i == len(c) {
				return c, nil
			}

			// ends with '?>' and '<?>' is illegal
			if c[i-1] == '?' && len(c) > 3 {
				i++
				return c[i:], &HtmlTag{
					Tag: string(c[0:i]),
				}
			}

			i++ // skip '>'
		}

		return c, nil
	case '!':
		i++

		if i == len(c) {
			return c, nil
		}

		switch c[i] {
		// A CDATA section consists of the string <![CDATA[, a string of
		// characters not including the string ]]>, and the string ]]>.
		case '[':
			i++
			if j := i; !(j+5 < len(c) && c[j+0] == 'C' && c[j+1] == 'D' && c[j+2] == 'A' && c[j+3] == 'T' && c[j+4] == 'A' && c[j+5] == '[') {
				return c, nil
			}
			i += 6 // "CDATA["

			if i == len(c) {
				return c, nil
			}

			for i < len(c) {
				for i < len(c) && c[i] != '>' {
					i++
				}

				if i == len(c) {
					return c, nil
				}

				i++ // skip '>'

				// ends with ']]>'
				if c[i-2] == ']' && c[i-3] == ']' {
					return c[i:], &HtmlTag{
						Tag: string(c[0:i]),
					}
				}
			}

			return c, nil
		case '-':
			i++
			if i == len(c) || c[i] != '-' {
				return c, nil
			}
			i++
			if i == len(c) {
				return c, nil
			}

			//  text does not start with > or ->
			if c[i] == '>' && (i+1 < len(c) && c[i+0] == '-' && c[i+1] == '>') {
				return c, nil
			}

			for i < len(c) {
				for i < len(c) && c[i] != '>' {
					i++
				}

				if i == len(c) {
					return c, nil
				}

				i++ // skip '>'

				// ends with '-->'
				// NOT COMPATIBLE: does not end with -, and does not contain --
				if len(c) >= 7 && c[i-3] == '-' && c[i-2] == '-' && c[i-1] == '>' {
					return c[i:], &HtmlTag{
						Tag: string(c[0:i]),
					}
				}
			}

			return c, nil

		// A declaration consists of the string <!, a name consisting of one or more uppercase ASCII letters,
		// whitespace, a string of characters not including the character >, and the character >.
		default:
			j := i
			for j < len(c) && ('A' <= c[j] && c[j] <= 'Z') {
				j++
			}
			if j == i {
				return c, nil
			}
			i = j

			if !skipWhitespaces(1) {
				return c, nil
			}

			for i < len(c) && c[i] != '>' {
				i++
			}

			if i == len(c) {
				return c, nil
			}

			i++

			return c[i:], &HtmlTag{
				Tag: string(c[0:i]),
			}
		}
	// An open tag consists of a < character, a tag name, zero or more attributes,
	// optional whitespace, an optional / character, and a > character.
	default:
		j := tryParseTagName()
		// A tag name is non-empty.
		if j == i {
			return c, nil
		}
		i = j

		skipped := false

		for i < len(c) {
			if !skipped && !skipWhitespaces(1) {
				break
			}
			skipped = false
			if i == len(c) {
				break
			}

			j := tryParseAttributeName()
			if j == i {
				break
			}
			i = j

			skipWhitespaces(0)
			if i == len(c) {
				return c, nil
			}

			if c[i] != '=' {
				skipped = true
				continue
			}

			i++ // skip '='

			skipWhitespaces(0)
			if i == len(c) {
				break
			}

			j, quoted := tryParseAttributeValue()
			if j == i {
				break
			}
			_ = quoted
			i = j
		}

		if i == len(c) {
			return c, nil
		}

		if c[i] == '/' {
			i++
			if i == len(c) || c[i] != '>' {
				return c, nil
			}
			i++
			return c[i:], &HtmlTag{
				Tag: string(c[0:i]),
			}
		}

		if c[i] == '>' {
			i++
			return c[i:], &HtmlTag{
				Tag: string(c[0:i]),
			}
		}

		return c, nil
	}
}
func tryParseLinkReferenceDefinition(c []rune) *LinkReferenceDefinition {
	i := 0
	if i == len(c) || c[i] != '[' {
		return nil
	}

	i++ // skip '['

	j := i
	for j < len(c) && c[j] != ']' {
		j++
	}

	if j == len(c) {
		return nil
	}

	label := strings.TrimSpace(string(c[i:j]))

	j++ // skip ']'

	if n := len(label); n <= 0 || n > 999 {
		return nil
	}

	label = "[" + label + "]"

	if j == len(c) || c[j] != ':' {
		return nil
	}

	j++

	i = j

	for i < len(c) && isWahitespace(c[i]) && c[i] != '\n' {
		i++
	}

	if i == len(c) {
		return nil
	}

	if c[i] == '\n' {
		return &LinkReferenceDefinition{
			Label:           label,
			wantDestination: true,
		}
	}

	c, dest, ok := parseLinkDestination(c[i:])
	if !ok {
		return nil
	}

	i = 0

	for i < len(c) && isWahitespace(c[i]) && c[i] != '\n' {
		i++
	}

	if i == len(c) {
		return &LinkReferenceDefinition{
			Label:       label,
			Destination: dest,
		}
	}

	if c[i] == '\n' {
		return &LinkReferenceDefinition{
			Label:       label,
			Destination: dest,
			wantTitle:   true,
		}
	}

	c, title, ok := parseLinkTitle(c[i:])
	if !ok {
		return nil
	}

	i = 0

	for i < len(c) && isWahitespace(c[i]) && c[i] != '\n' {
		i++
	}

	if i == len(c) || c[i] == '\n' {
		return &LinkReferenceDefinition{
			Label:       label,
			Destination: dest,
			Title:       title,
		}
	}

	return nil
}
