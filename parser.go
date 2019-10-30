package main

import (
	"container/list"
	"io"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
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
			if hr := tryParseHorizontalRule(s, r); hr != nil {
				blocks = append(blocks, hr)
				return true
			}
		}

		if _, ok := in(s, '=', '-'); ok {
			if heading := tryParseSetextHeadingUnderline(s); heading != nil {
				blocks = append(blocks, heading)
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

		if _, ok := in(s, '<'); ok {
			if hb := tryParseHtmlBlock(os); hb != nil {
				blocks = append(blocks, hb)
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
				label := strings.ToLower(link.Label)
				if _, ok := gdoc.links[label]; !ok {
					gdoc.links[label] = link
				}
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

func isPunctuation(r rune) bool {
	_, ok := punctuation[r]
	return ok
}

var reBlankLine = regexp.MustCompile(`^\s*$`)

func isBlankLine(r []rune) bool {
	return reBlankLine.MatchString(string(r))
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

var ls *LineScanner

func parse(in io.Reader, example int) *Document {
	var doc Document
	gdoc = &doc

	doc.example = example
	doc.links = make(map[string]*LinkReferenceDefinition)

	ls = NewLineScanner(in)

	for ls.Scan() {
		doc.AddLine(ls.Text())
		tryMergeSetextHeading(&doc.blocks)
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
		return &HorizontalRule{Marker: start}
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

func tryParseSetextHeadingUnderline(c []rune) *SetextHeading {
	oc := c
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
		return &SetextHeading{
			line:  oc,
			level: level,
		}
	}
	return nil
}

func tryParseFencedCodeBlockStart(c []rune, marker rune, indent int) *CodeBlock {
	cb := &CodeBlock{}
	cb.fenceIndent = indent
	cb.fenceMarker = marker

	i := 0

	// count markers
	for i < len(c) && c[i] == marker {
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
	s := []rune{}
	for i < len(c) {
		switch c[i] {
		case '\\':
			if r, ok := parseEscape(c[i:]); ok {
				s = append(s, r)
				i += 2
			} else {
				s = append(s, '\\')
				i++
			}
		case '&':
			if nc, cp1, cp2, ok := tryParseHtmlEntity(c[i:]); ok {
				i = 0
				c = nc
				if cp1 == 0 {
					cp1 = utf8.RuneError
				}
				r := []rune{cp1}
				if cp2 != 0 {
					r = append(r, cp2)
				}
				s = append(s, r...)
			} else {
				s = append(s, '&')
				i++
			}
		default:
			s = append(s, c[i])
			i++
		}
	}
	info := strings.TrimSpace(string(s))

	// If the info string comes after a backtick fence, it may not contain any backtick characters.
	// The reason for this restriction is that otherwise some inline code would be
	// incorrectly interpreted as the beginning of a fenced code block.
	if marker == '`' && strings.IndexByte(info, byte('`')) != -1 {
		return nil
	}

	cb.Info = info

	if p := strings.IndexFunc(info, unicode.IsSpace); p != -1 {
		cb.Lang = info[:p]
		cb.Args = info[p:]
	} else {
		cb.Lang = info
	}

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
			appendDelimiter("]")
			i++
			if nc, ok := parseRightBracket(texts, delimiters, c[i:]); ok {
				c = nc
				i = 0
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
			if nc, link, ok := parseUriAutoLink(c[i:]); ok {
				c = nc
				i = 0
				appendText(link)
				continue
			}
			if nc, tag := tryParseHtmlTag(c[i:]); tag != nil {
				i = 0
				c = nc
				appendText(tag)
				continue
			}
			if nc, link := parseEmailAutoLink(c[i:]); link != nil {
				c = nc
				i = 0
				appendText(link)
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
			for j < len(c) && c[j] == '`' {
				j++
			}
			appendText(&Text{
				Text: strings.Repeat("`", j-i),
			})
			i = j
		case '&':
			if nc, cp1, cp2, ok := tryParseHtmlEntity(c[i:]); ok {
				i = 0
				c = nc
				if cp1 == 0 {
					cp1 = utf8.RuneError
				}
				r := []rune{cp1}
				if cp2 != 0 {
					r = append(r, cp2)
				}
				appendText(&Text{
					Text: string(r),
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

func parseInlines(raw string) (inlines []Inline) {
	texts, delimiters := parseInlinesToDeimiters(raw)
	parseLineBreaks(texts)
	parseEmphases(texts, delimiters, nil)
	for e := texts.Back(); e != nil; e = e.Prev() {
		inlines = append(inlines, e.Value)
	}
	return
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

func parseRightBracket(texts *list.List, delimiters *list.List, c []rune) ([]rune, bool) {
	toDelimiter := func(v *list.Element) *Delimiter {
		return v.Value.(*Delimiter)
	}

	var (
		i             int
		opener        *Delimiter
		openerElement *list.Element
	)

	for openerElement = delimiters.Front().Next(); openerElement != nil; {
		d := toDelimiter(openerElement)
		if d.text == "[" || d.text == "![" {
			opener = d
			break
		}
		openerElement = openerElement.Next()
	}

	if opener == nil {
		return nil, false
	}

	if !opener.active {
		delimiters.Remove(openerElement)
		return nil, false
	}

	var link Link
	var image Image
	var nc []rune
	var ok bool

	nc, destination, title, ref, hasRef, ok := parseLink(c[i:])

	if !ok {
		delimiters.Remove(openerElement)
		return nil, false
	}

	// hack, to get texts betwee "[" and "]", for ref.
	// TODO remember positions of "[" and "]".
	if hasRef {
		if ref == "" || ref == "[]" {
			ref = ""
			for e := opener.textElement; e != nil; {
				tc, ok := e.Value.(ITextContent)
				if !ok {
					panic("implement ITextContnet")
				}
				ref += tc.TextContent()
				e = e.Prev()
			}
		}
		ref = strings.ToLower(ref)
		if ref[0] == '!' {
			ref = ref[1:]
		}
		dest, tt, ok := gdoc.refLink(ref, true)
		if !ok {
			delimiters.Remove(openerElement)
			return nil, false
		}

		destination = dest
		title = tt
	}

	if opener.text == "[" {
		link.Link = destination
		link.Title = title
	} else {
		image.Link = destination
		image.Title = title
	}

	c = nc
	i = 0

	// remove "]" before processing emphases
	texts.Remove(texts.Front())
	delimiters.Remove(delimiters.Front())

	if opener.text == "[" {
		parseEmphases(texts, delimiters, openerElement)
		for e := opener.textElement.Prev(); e != nil; {
			link.Inlines = append(link.Inlines, e.Value)
			e = e.Prev()
		}
		// opener is about to be removed, hence we insert it after (at left) opener.
		texts.InsertAfter(&link, opener.textElement)
	} else {
		parseEmphases(texts, delimiters, openerElement)
		for e := opener.textElement.Prev(); e != nil; {
			image.inlines = append(image.inlines, e.Value)
			if tc, ok := e.Value.(ITextContent); ok {
				image.Alt += tc.TextContent()
			}
			e = e.Prev()
		}
		texts.InsertAfter(&image, opener.textElement)
	}

	// If we have a link (and not an image), we also set all [ delimitersbefore the opening delimiter
	// to inactive. (This will prevent us from getting links within links.)
	if opener.text == "[" {
		for e := openerElement.Next(); e != nil; {
			d := toDelimiter(e)
			if d.text == "[" {
				d.active = false
			}
			e = e.Next()
		}
	}

	// remove from "["
	for e := opener.textElement; e != nil; {
		next := e.Prev()
		texts.Remove(e)
		e = next
	}
	for e := openerElement; e != nil; {
		next := e.Prev()
		delimiters.Remove(e)
		e = next
	}

	return c, true
}

func parseEmphases(texts *list.List, delimiters *list.List, bottom *list.Element) {
	toDelimiter := func(v *list.Element) *Delimiter {
		return v.Value.(*Delimiter)
	}

	var (
		opener        *Delimiter
		closer        *Delimiter
		openerElement *list.Element
		closerElement *list.Element
	)

	// first set closerElement above bottom
	if bottom == nil {
		closerElement = delimiters.Back() // may be nil, if empty.
	} else {
		closerElement = bottom.Prev() // may be nil, if at top.
	}

	for closerElement != nil {
		closer = toDelimiter(closerElement)
		if !closer.canCloseEmphasis() {
			// Move current_position forward in the delimiter stack (if needed)
			// until we find the first potential closer with delimiter * or _.
			closerElement = closerElement.Prev()
			continue
		}

		// Now, look back in the stack (staying above stack_bottom and the openers_bottom
		// for this delimiter type) for the first matching potential opener (“matching” means same delimiter).
		for openerElement = closerElement.Next(); openerElement != bottom; openerElement = openerElement.Next() {
			opener = toDelimiter(openerElement)
			if opener.canOpenEmphasis() && opener.match(closer) && !opener.oddMatch(closer) {
				break
			}
		}

		// If none is found
		if openerElement == bottom {
			next := closerElement.Prev()

			// If the closer at current_position is not a potential opener,
			// remove it from the delimiter stack (since we know it can’t be a closer either).
			if !closer.canOpenEmphasis() {
				delimiters.Remove(closerElement)
			}

			// Advance current_position to the next element in the stack.
			closerElement = next
			continue
		}

		// If one is found

		// Figure out whether we have emphasis or strong emphasis:
		// if both closer and opener spans have length >= 2, we have strong, otherwise regular.
		n := 1 // not strong by default
		if opener.canBeStrong() && closer.canBeStrong() {
			n = 2
		}

		// Insert an emph or strong emph node accordingly, after the text node corresponding to the opener.
		emphasis := &Emphasis{
			Delimiter: string(opener.text[0:n]),
		}

		// texts between opener and closer are contents of emphasis.
		for e := opener.textElement.Prev(); e != nil && e != closer.textElement; {
			emphasis.Inlines = append(emphasis.Inlines, e.Value)
			next := e.Prev()
			texts.Remove(e)
			e = next
		}
		texts.InsertBefore(emphasis, opener.textElement)

		// Remove any delimiters between the opener and closer from the delimiter stack.
		for e := openerElement.Prev(); e != nil && e != closerElement; {
			next := e.Prev()
			delimiters.Remove(e)
			e = next
		}

		// Remove 1 (for regular emph) or 2 (for strong emph) delimiters from the opening and closing text nodes.
		// If they become empty as a result, remove them and remove the corresponding element of the delimiter stack.
		// If the closing node is removed, reset current_position to the next element in the stack.
		openerEmpty := opener.consume(n) == 0
		closerEmpty := closer.consume(n) == 0
		if openerEmpty {
			delimiters.Remove(openerElement)
			texts.Remove(opener.textElement)
		}
		if closerEmpty {
			next := closerElement.Prev()
			delimiters.Remove(closerElement)
			texts.Remove(closer.textElement)
			closerElement = next
		}
	}
}

func parseEscape(c []rune) (rune, bool) {
	i := 0
	if j := i + 1; j < len(c) && isPunctuation(c[j]) {
		return c[j], true
	}
	return 0, false
}

func parseLink(c []rune) (remain []rune, dest string, title string, ref string, hasRef bool, oook bool) {
	i := 0
	oc := c

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

	// shortcut reference link
	if i == len(c) {
		remain = c[i:]
		hasRef = true
		oook = true
		return
	}

	switch c[i] {
	case '[':
		nc, label, ok := parseLinkLabel(c[i:])
		if !ok {
			remain = c
			hasRef = true
			oook = true
			return
		}
		ref = label
		hasRef = true
		remain = nc
		oook = true
		return
	}

	if c[i] != '(' {
		// shortcut link ref
		remain = c[i:]
		hasRef = true
		oook = true
		return
	}
	i++ // skip '('

	skipWhitespaces(0)
	if i == len(c) || c[i] == ')' {
		i++
		remain = c[i:]
		oook = true
		return
	}

	var ok bool

	c, dest, ok = parseLinkDestination(c[i:])
	if !ok {
		return
	}

	i = 0

	skipWhitespaces(0)
	if i == len(c) || c[i] == ')' {
		i++
		remain = c[i:]
		oook = true
		return
	}

	c, title, ok = parseLinkTitle(c[i:])
	if !ok {
		remain = oc
		hasRef = true
		dest = ""
		oook = true
		return
	}

	i = 0

	skipWhitespaces(0)

	if i == len(c) || c[i] != ')' {
		return
	}

	i++

	remain = c[i:]
	oook = true
	return
}

func parseLinkLabel(c []rune) ([]rune, string, bool) {
	i := 0
	if i == len(c) || c[i] != '[' {
		return nil, "", false
	}

	i++ // skip '['

	j := i // remember start
	dest := []rune{}
	for j < len(c) && c[j] != ']' && c[j] != '[' {
		switch c[j] {
		case '\\':
			if r, ok := parseEscape(c[j:]); ok {
				dest = append(dest, r)
				j += 2
			} else {
				dest = append(dest, '\\')
				j++
			}
		default:
			dest = append(dest, c[j])
			j++
		}
	}

	if j == len(c) {
		return nil, "", false
	}

	// Link labels cannot contain brackets, unless they are backslash-escaped:
	if c[j] == '[' {
		return nil, "", false
	}

	label := strings.TrimSpace(string(dest))

	j++ // skip ']'

	if n := len(label); n <= 0 || n > 999 {
		// return nil, "", false
	}

	label = "[" + label + "]"

	i = j

	return c[i:], label, true
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
		i = 0
		for i < len(c) && c[i] != '>' && c[i] != '\n' {
			switch c[i] {
			case '\\':
				if r, ok := parseEscape(c[i:]); ok {
					dest = append(dest, r)
					i += 2
				} else {
					dest = append(dest, '\\')
					i++
				}
			case '&':
				if nc, cp1, cp2, ok := tryParseHtmlEntity(c[i:]); ok {
					i = 0
					c = nc
					if cp1 == 0 {
						cp1 = utf8.RuneError
					}
					r := []rune{cp1}
					if cp2 != 0 {
						r = append(r, cp2)
					}
					dest = append(dest, r...)
				} else {
					dest = append(dest, '&')
					i++
				}
			default:
				dest = append(dest, c[i])
				i++
			}
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
			case '&':
				if nc, cp1, cp2, ok := tryParseHtmlEntity(c[i:]); ok {
					i = 0
					c = nc
					if cp1 == 0 {
						cp1 = utf8.RuneError
					}
					r := []rune{cp1}
					if cp2 != 0 {
						r = append(r, cp2)
					}
					dest = append(dest, r...)
				} else {
					dest = append(dest, '&')
					i++
				}
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

	i := 1

	dest := []rune{}

	for i < len(c) && (c[i] != marker && !(marker == '(' && c[i] == ')')) {
		switch c[i] {
		case '\\':
			if r, ok := parseEscape(c[i:]); ok {
				dest = append(dest, r)
				i += 2
			} else {
				dest = append(dest, '\\')
				i++
			}
		case '&':
			if nc, cp1, cp2, ok := tryParseHtmlEntity(c[i:]); ok {
				i = 0
				c = nc
				if cp1 == 0 {
					cp1 = utf8.RuneError
				}
				r := []rune{cp1}
				if cp2 != 0 {
					r = append(r, cp2)
				}
				dest = append(dest, r...)
			} else {
				dest = append(dest, '&')
				i++
			}
		default:
			dest = append(dest, c[i])
			i++
		}
	}

	if i == len(c) {
		return nil, "", false
	}

	title := string(dest)

	i++ // skip marker

	return c[i:], title, true
}

func parseUriAutoLink(c []rune) ([]rune, *Link, bool) {
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
		Inlines: []Inline{
			&Text{
				Text: full,
			},
		},
	}
	return c[i+1:], link, true
}

// https://spec.commonmark.org/0.29/#email-address
// not fully compatible
var reEmail = regexp.MustCompile(`^<[[:alnum:].+-]+@[[:alnum:]](?:[[:alnum:]-]{0,61}[[:alnum:]])?(?:\.[[:alnum:]](?:[[:alnum:]-]{0,61}[[:alnum:]])?)*>`)

func parseEmailAutoLink(c []rune) ([]rune, *Link) {
	bys := []byte(string(c))
	loc := reEmail.FindIndex(bys)
	if loc == nil {
		return nil, nil
	}
	email := string(bys[loc[0]+1 : loc[1]-1])
	link := Link{
		Inlines: []Inline{&Text{Text: email}},
		Link:    "mailto:" + email,
	}
	remain := []rune(string(bys[loc[1]:]))
	return remain, &link
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

// Don't use html.UnescapeString since it will use ReplacementChar.
func tryParseHtmlEntity(c []rune) (remains []rune, first rune, second rune, oook bool) {
	i := 0
	if len(c) < 1 || c[i] != '&' {
		return
	}

	i++
	if i == len(c) {
		return
	}

	switch c[i] {
	default:
		j := i
		for j < len(c) && isAlNum(c[j]) {
			j++
		}
		if j == len(c) || c[j] != ';' {
			return
		}

		j++ // after ';'

		name := string(c[1 : j-1])

		if codepoint, ok := htmlEntities1[name]; ok {
			return c[j:], codepoint, 0, true
		}
		if codepoints, ok := htmlEntities2[name]; ok {
			return c[j:], codepoints[0], codepoints[1], true
		}

		return
	case '#':
		i++
		if i == len(c) {
			return
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
				return
			}
			if j == len(c) || c[j] != ';' {
				return
			}
			j++
			return c[j:], rune(n), 0, true
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
					n += int(c[j]) - 'a' + 10
				case 'A' <= c[j] && c[j] <= 'F':
					n += int(c[j]) - 'A' + 10
				}
				j++
			}
			if j == i || j-i > 6 {
				return
			}
			if j == len(c) || c[j] != ';' {
				return
			}
			j++

			buf := make([]byte, 4)
			w := utf8.EncodeRune(buf, rune(n))
			rs := []rune(string(buf[:w]))

			first = rs[0]
			second = 0
			if len(rs) > 1 {
				second = rs[1]
			}

			return c[j:], first, second, true
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
				if len(c[:i]) >= 7 && c[i-3] == '-' && c[i-2] == '-' && c[i-1] == '>' {
					tag := string(c[0:i])
					// An HTML comment consists of <!-- + text + -->, where text does not start with > or ->,
					// does not end with -, and does not contain --. (See the HTML5 spec.)
					text := string(tag[3 : len(tag)-3])
					if strings.HasPrefix(text, ">") || strings.HasPrefix(text, "->") || strings.HasSuffix(text, "-") || strings.Contains(text, "--") {
						return nil, nil
					}
					return c[i:], &HtmlTag{
						Tag: tag,
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

func tryParseHtmlBlock(c []rune) *HtmlBlock {
	h := &HtmlBlock{}

	i := 0
	for i < len(c) && isWahitespace(c[i]) {
		i++
	}

	if i == len(c) || c[i] != '<' {
		return nil
	}

	i++ // skip '<'
	if i == len(c) {
		return nil
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

	_ = tryParseTagName

	switch c[i] {
	// Start condition: 3
	case '?':
		h.Lines = append(h.Lines, c)
		h.condition = 3
		return h
	// Start condition: 6
	case '/':
		i++

		j := tryParseTagName()
		// A tag name is non-empty.
		if j == i {
			return nil
		}
		tagName := string(c[i:j])

		i = j

		if _, ok := htmlBlockStartCondition6TagNames[strings.ToLower(tagName)]; !ok {
			// 7 complete closing tag
			_, nc := skipPrefixSpaces(c[i:], -1)
			if len(nc) > 1 && nc[0] == '>' && (nc[1] == '\n' || unicode.IsSpace(nc[1])) {
				h.condition = 7
				h.append(c)
				return h
			}

			return nil
		}

		skipWhitespaces(0)

		if i == len(c) {
			h.condition = 6
			h.append(c)
			return h
		}

		if i+0 < len(c) && c[i+0] == '>' || i+1 < len(c) && c[i+0] == '/' && c[i+1] == '>' {
			h.append(c)
			h.condition = 6
			return h
		}

		return nil
	case '!':
		i++

		if i == len(c) {
			return nil
		}

		switch c[i] {
		case '[':
			i++
			if j := i; j+5 < len(c) && c[j+0] == 'C' && c[j+1] == 'D' && c[j+2] == 'A' && c[j+3] == 'T' && c[j+4] == 'A' && c[j+5] == '[' {
				h.append(c)
				h.condition = 5
				return h
			}
			return nil
		case '-':
			i++
			if i == len(c) || c[i] != '-' {
				return nil
			}
			h.append(c)
			h.condition = 2
			return h
		default:
			if 'A' <= c[i] && c[i] <= 'Z' {
				h.append(c)
				h.condition = 4
				return h
			}
			return nil
		}
	default:
		j := tryParseTagName()
		if j == i {
			return nil
		}
		tagName := string(c[i:j])
		i = j
		if tagName == "script" || tagName == "pre" || tagName == "style" {
			if i == len(c) || isWahitespace(c[i]) || c[i] == '>' {
				h.append(c)
				h.condition = 1
				return h
			}
		}
		if _, ok := htmlBlockStartCondition6TagNames[strings.ToLower(tagName)]; ok {
			if i == len(c) || isWahitespace(c[i]) || i+0 < len(c) && c[i+0] == '>' || i+1 < len(c) && c[i+0] == '/' && c[i+1] == '>' {
				h.append(c)
				h.condition = 6
				return h
			}
		}
		// condition 7

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
				return nil
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
			return nil
		}

		if c[i] == '/' {
			i++
			if i == len(c) || c[i] != '>' {
				return nil
			}
			i++
			// followed only by whitespace or the end of the line
			if i < len(c) && !unicode.IsSpace(c[i]) {
				return nil
			}
			h.condition = 7
			h.append(c)
			return h
		}

		if c[i] == '>' {
			i++
			_, nc := skipPrefixSpaces(c[i:], -1)
			if len(nc) == 0 || nc[0] == '\n' {
				h.condition = 7
				h.append(c)
				return h
			}
		}

		return nil

	}
	return nil
}

func tryParseLinkReferenceDefinition(c []rune) *LinkReferenceDefinition {
	i := 0
	l := LinkReferenceDefinition{}

	nc, label, ok := parseLinkLabel(c)
	if !ok || label == "[]" {
		return nil
	}
	l.Label = label
	c = nc
	i = 0

	if i == len(c) || c[i] != ':' {
		return nil
	}

	i++ // skip ':'

	for i < len(c) && isWahitespace(c[i]) {
		i++
	}

	if i == len(c) {
		return nil
	}

	nc, dest, ok := parseLinkDestination(c[i:])
	if !ok {
		return nil
	}

	l.Destination = dest
	c = nc
	i = 0

	for i < len(c) && isWahitespace(c[i]) {
		i++
	}

	if i == len(c) {
		return &l
	}

	nc, title, ok := parseLinkTitle(c[i:])
	if !ok {
		return nil
	}

	l.Title = title
	c = nc
	i = 0

	for i < len(c) && isWahitespace(c[i]) {
		i++
	}

	if i != len(c) {
		return nil
	}

	return &l
}

func tryMergeSetextHeading(pbs *[]Blocker) {
	n := len(*pbs)
	if n < 1 {
		return
	}

	blocks := *pbs
	defer func() { *pbs = blocks }()

	switch typed := blocks[n-1].(type) {
	case *SetextHeading:
		if n == 1 {
			blocks[n-1] = &Paragraph{
				texts: []string{string(typed.line)},
			}
			return
		}
		if p, ok := blocks[n-2].(*Paragraph); ok {
			heading := Heading{
				Level: typed.level,
				text:  strings.Join(p.texts, ""),
			}
			blocks[n-2] = &heading
			blocks = blocks[:n-1]
		} else {
			blocks[n-1] = &Paragraph{
				texts: []string{string(typed.line)},
			}
		}
	case *HorizontalRule:
		if typed.Marker == '-' && n >= 2 {
			if p, ok := blocks[n-2].(*Paragraph); ok {
				heading := Heading{
					Level: 2,
					text:  strings.Join(p.texts, ""),
				}
				blocks[n-2] = &heading
				blocks = blocks[:n-1]
			}
		}
	}
}
