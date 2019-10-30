package main

import (
	"container/list"
	"strings"
	"unicode"
)

type Blocker interface {
	AddLine(s []rune) bool
}

type Document struct {
	example int
	blocks  []Blocker
	links   map[string]*LinkReferenceDefinition
}

func (doc *Document) AddLine(s []rune) {
	addLine(&doc.blocks, s)
}

func (doc *Document) parseInlines() {
	for _, block := range doc.blocks {
		if inliner, ok := block.(_Inliner); ok {
			inliner.parseInlines()
		}
	}
}

func (doc *Document) refLink(label string, enclosed bool) (string, string, bool) {
	if !enclosed {
		label = "[" + label + "]"
	}
	label = strings.ToLower(label)
	ref, ok := doc.links[label]
	if !ok {
		return "", "", false
	}
	return ref.Destination, ref.Title, true
}

type BlankLine struct {
}

func (bl *BlankLine) AddLine(s []rune) bool {
	return false
}

// HorizontalRule is a horizontal rule (thematic breaks).
// https://spec.commonmark.org/0.29/#thematic-breaks
type HorizontalRule struct {
	Marker rune
}

func (hr *HorizontalRule) AddLine(s []rune) bool {
	return false
}

type Paragraph struct {
	texts   []string
	Tight   bool
	Inlines []Inline
}

func (p *Paragraph) AddLine(s []rune) bool {
	var blocks []Blocker
	if !addLine(&blocks, s) {
		panic("won't happen")
	}

	// A Link Reference Definition was added.
	if len(blocks) == 0 {
		return true
	}

	// Leading spaces at the beginning of the next line are ignored
	trimLeft := func(s string) string {
		return strings.TrimLeft(s, " ")
	}

	switch typed := blocks[0].(type) {
	case *Paragraph:
		// A sequence of non-blank lines that cannot be interpreted as ...
		// ... other kinds of blocks forms a paragraph.
		p.texts = append(p.texts, trimLeft(typed.texts[0]))
		return true
	case *CodeBlock:
		// An indented code block cannot interrupt a paragraph
		if !typed.isFenced() {
			// s: typed.lines[0] is trimmed 4 spaces at the beginning, don't use.
			p.texts = append(p.texts, trimLeft(string(s)))
			return true
		}
	case *List:
		// In order to solve of unwanted lists in paragraphs with hard-wrapped numerals,
		// we allow only lists starting with 1 to interrupt paragraphs.
		if typed.Ordered && typed.Start != 1 {
			p.texts = append(p.texts, trimLeft(string(s)))
			return true
		}
	case *LinkReferenceDefinition:
		// A link reference definition cannot interrupt a paragraph.
		p.texts = append(p.texts, trimLeft(string(s)))
		return true
	case *HtmlBlock:
		// HTML blocks of type 7 cannot interrupt a paragraph
		if typed.condition == 7 {
			p.texts = append(p.texts, trimLeft(string(s)))
			return true
		}
	}

	return false
}

func (p *Paragraph) parseInlines() {
	raw := strings.Join(p.texts, "")
	raw = strings.TrimSpace(raw)
	p.Inlines = parseInlines(raw)
}

type Line struct {
	text string
}

type Heading struct {
	Level   int
	Inlines []Inline
	text    string
}

func (h *Heading) AddLine(s []rune) bool {
	return false
}

func (h *Heading) parseInlines() {
	text := strings.TrimSpace(h.text)
	h.Inlines = parseInlines(text)
}

type SetextHeading struct {
	line  []rune
	level int
}

func (h *SetextHeading) AddLine(s []rune) bool {
	return false
}

// An HTML block is a group of lines that is treated as raw HTML (and will not be escaped in HTML output).
type HtmlBlock struct {
	Lines [][]rune

	condition int
	closed    bool
}

func (hb *HtmlBlock) append(c []rune) {
	hb.Lines = append(hb.Lines, c)
}

func (hb *HtmlBlock) AddLine(c []rune) bool {
	if hb.closed {
		return false
	}

	switch hb.condition {
	case 1:
		hb.append(c)
		lc := strings.ToLower(string(c))
		if strings.Contains(lc, "</script>") || strings.Contains(lc, "</pre>") || strings.Contains(lc, "</style>") {
			hb.closed = true
		}
	case 2:
		hb.append(c)
		if strings.Contains(string(c), "-->") {
			hb.closed = true
		}
	case 3:
		hb.append(c)
		if strings.Index(string(c), "?>") != -1 {
			hb.closed = true
		}
	case 4:
		hb.append(c)
		if strings.Contains(string(c), ">") {
			hb.closed = true
		}
	case 5:
		hb.append(c)
		if strings.Contains(string(c), "]]>") {
			hb.closed = true
		}
	case 6:
		_, nc := skipPrefixSpaces(c, -1)
		if len(nc) <= 1 {
			hb.closed = true
		} else {
			hb.append(c)
		}
	case 7:
		_, nc := skipPrefixSpaces(c, -1)
		if len(nc) <= 1 {
			hb.closed = true
		} else {
			hb.append(c)
		}
	}
	return true
}

type BlockQuote struct {
	blocks []Blocker
}

func (bq *BlockQuote) AddLine(s []rune) bool {
	_, ok := tryParseBlockQuote(s, bq)
	tryMergeSetextHeading(&bq.blocks)
	return ok
}

func (bq *BlockQuote) parseInlines() {
	for _, block := range bq.blocks {
		if inliner, ok := block.(_Inliner); ok {
			inliner.parseInlines()
		}
	}
}

// A List is a sequence of one or more list items of the same type.
// The list items may be separated by any number of blank lines.
type List struct {
	// A list is an ordered list if its constituent list items begin with ordered list markers,
	// and a bullet list if its constituent list items begin with bullet list markers.
	Ordered bool

	// A list is loose if any of its constituent list items are separated by blank lines,
	// or if any of its constituent list items directly contain two block-level elements
	// with a blank line between them. Otherwise a list is tight.
	//
	// The difference in HTML output is that paragraphs in a loose list are wrapped in <p> tags,
	// while paragraphs in a tight list are not.)
	Tight bool

	// Two list items are of the same type if they begin with a list marker of the same type.
	// Two list markers are of the same type if either
	//     (a) they are bullet list markers using the same character (-, +, or *)
	//   or
	//     (b) they are ordered list numbers with the same delimiter (either . or )).
	MarkerChar byte

	// The start number of an ordered list is determined by the list number of its initial list item.
	// The numbers of subsequent list items are disregarded.
	Start int

	Items []Blocker
}

func (l *List) parseMarker(s []rune) (remain []rune, list *List, prefixSpaces int, markerWidth int, ok bool) {
	list = &List{}
	prefixWidth := 0

	// for example 280, subsequent lines may have indent spaces.
	prefixSpaces, s = skipPrefixSpaces(s, 3)

	if marker, ook := in(s, '-', '+', '*'); ook {
		list.Ordered = false
		list.MarkerChar = byte(marker)
		prefixWidth = 1
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
		// ordered list start numbers must be nine digits or less
		if i > 9 {
			return
		}
		s = s[i:]
		if len(s) == 0 {
			return
		}
		switch s[0] {
		default:
			return
		case '.', ')':
			list.MarkerChar = byte(s[0])
			list.Start = start
			prefixWidth = i + 1
			s = s[1:]
		}
	}

	if len(s) == 0 {
		return
	}

	// Explain for "5"
	// If the first block in the list item is an indented code block,
	// then by rule #2, the contents must be indented one space after the list marker
	_, n := peekSpaces(s, 5)
	if n < 1 {
		return
	}

	switch {
	case n == 5:
		return s[1:], list, prefixSpaces, prefixWidth + 1, true
	default:
		return s[n:], list, prefixSpaces, prefixWidth + n, true
	}
}

func (l *List) isHorizontalRule(s []rune) bool {
	if _, s = skipPrefixSpaces(s, -1); len(s) == 0 {
		return false
	}

	// very tricky
	if s[0] != rune(l.MarkerChar) {
		return false
	}

	if r, ok := in(s, '-', '_', '*'); ok {
		if hr := tryParseHorizontalRule(s, r); hr != nil {
			return true
		}
	}

	return false
}

func (l *List) AddLine(s []rune) bool {
	// When both a thematic break and a list item are possible
	// interpretations of a line, the thematic break takes precedence
	if l.isHorizontalRule(s) {
		return false
	}

	var lastItem *ListItem

	for i := len(l.Items) - 1; i >= 0; i-- {
		if item, ok := l.Items[i].(*ListItem); ok {
			lastItem = item
			break
		}
	}

	if lastItem != nil {
		if lastItem.AddLine(s) {
			if isBlankLine(s) {
				l.Items = append(l.Items, &BlankLine{})
				return true
			}
			return true
		}
	}

	if isBlankLine(s) {
		l.Items = append(l.Items, &BlankLine{})
		return true
	}

	s, list, prefixSpaces, markerWidth, ok := l.parseMarker(s)
	if !ok {
		return false
	}

	if lastItem != nil {
		// Two list markers are of the same type if
		//   (a) they are bullet list markers using the same character (-, +, or *)
		// or
		//   (b) they are ordered list numbers with the same delimiter (either . or )).
		same := (l.Ordered == list.Ordered) && (l.MarkerChar == list.MarkerChar)
		if !same {
			return false
		}
	} else {
		l.Ordered = list.Ordered
		l.MarkerChar = list.MarkerChar
		l.Start = list.Start
	}

	lastItem = &ListItem{
		prefixSpaces: prefixSpaces,
		suffixSpaces: markerWidth,
	}

	l.Items = append(l.Items, lastItem)

	if addLine(&lastItem.blocks, s) {
		return true
	}

	_, s = skipPrefixSpaces(s, -1)
	if len(s) == 0 || s[0] == '\n' {
		l.Items = append(l.Items, &BlankLine{})
		return true
	}

	return false
}

// A list is loose if any of its constituent list items are separated by blank lines,
// or if any of its constituent list items directly contain two block-level elements
// with a blank line between them. Otherwise a list is tight.
//
// The difference in HTML output is that paragraphs in a loose list are
// wrapped in <p> tags, while paragraphs in a tight list are not.
func (l *List) deduceIsTight() {
	var setListBlock Blocker
	var setListBlankLine *BlankLine

	for _, item := range l.Items {
		switch t := item.(type) {
		default:
			if setListBlankLine != nil {
				l.Tight = false
				return
			}
			if setListBlock == nil {
				setListBlock = t
			}
		case *BlankLine:
			if setListBlock != nil {
				if setListBlankLine == nil {
					setListBlankLine = t
				}
			}
		}

		if pItem, ok := item.(*ListItem); ok {
			var setItemBlock Blocker
			var setItemBlankLine *BlankLine

			for _, block := range pItem.blocks {
				switch t := block.(type) {
				default:
					if setItemBlankLine != nil {
						l.Tight = false
						return
					}
					if setItemBlock == nil {
						setItemBlock = t
					}
				case *BlankLine:
					if setItemBlock != nil {
						if setItemBlankLine == nil {
							setItemBlankLine = t
						}
					}
				}
			}
		}
	}
	l.Tight = true
}

func (l *List) parseInlines() {
	for _, item := range l.Items {
		if inliner, ok := item.(_Inliner); ok {
			inliner.parseInlines()
		}
	}
}

type ListItem struct {
	prefixSpaces int
	suffixSpaces int
	blocks       []Blocker
}

func (li *ListItem) AddLine(s []rune) bool {
	if len(s) == 1 && s[0] == '\n' {
		if len(li.blocks) > 0 && li.blocks[len(li.blocks)-1].AddLine(s) {
			li.blocks = append(li.blocks, &BlankLine{})
			return true
		}
		li.blocks = append(li.blocks, &BlankLine{})
		return true
	}
	_, nSkipped := peekSpaces(s, li.prefixSpaces)
	s = s[nSkipped:]
	_, nSkipped = peekSpaces(s, li.suffixSpaces)
	if nSkipped != li.suffixSpaces {
		return false
	}
	s = s[nSkipped:]
	if addLine(&li.blocks, s) {
		tryMergeSetextHeading(&li.blocks)
		li.tryMergeCodeBlock()
		return true
	}
	return false
}

// A list item that contains an indented code block will preserve
// empty lines within the code block verbatim.
func (li *ListItem) tryMergeCodeBlock() {
	if lastCode, ok := li.blocks[len(li.blocks)-1].(*CodeBlock); ok {
		j := len(li.blocks) - 2
		n := 0
		for j = len(li.blocks) - 2; j >= 0; j-- {
			if _, ok := li.blocks[j].(*BlankLine); !ok {
				break
			}
			n++
		}
		if j > 0 && n > 0 {
			if prevCode, ok := li.blocks[j].(*CodeBlock); ok {
				for n > 0 {
					prevCode.lines = append(prevCode.lines, "\n")
					n--
				}
				prevCode.lines = append(prevCode.lines, lastCode.lines[0])
				li.blocks = li.blocks[:j+1]
			}
		}
	}
}

func (li *ListItem) parseInlines() {
	for _, block := range li.blocks {
		if inliner, ok := block.(_Inliner); ok {
			inliner.parseInlines()
		}
	}
}

// CodeBlock is either a fenced code block or an indented code block.
type CodeBlock struct {
	// The line with the opening code fence may optionally contain some
	// text following the code fence; this is trimmed of leading and
	// trailing whitespace and called the info string.
	Info string

	//  The first word of the info string is typically used to specify the
	// language of the code sample, and rendered in the class attribute of the code tag.
	Lang string

	// Custom: Info = Lang + Args
	Args string

	// The content of the code block consists of all subsequent lines
	lines []string

	// A code fence is a sequence of at least three consecutive backtick
	// characters (`) or tildes (~). (Tildes and backticks cannot be mixed.)
	fenceMarker rune

	// A fenced code block begins with a code fence, indented no more than three spaces.
	// If the leading code fence is indented N spaces, then up to N spaces of indentation
	// are removed from each line of the content (if present).
	fenceIndent int

	// The content of the code block consists of all subsequent lines,
	// until a closing code fence of the same type as the code block
	// began with (backticks or tildes), and with at least as many backticks
	// or tildes as the opening code fence.
	fenceLength int

	closed bool
}

func (cb *CodeBlock) isFenced() bool {
	return cb.fenceLength > 0
}

func (cb *CodeBlock) AddLine(s []rune) bool {
	if cb.closed {
		return false
	}

	if cb.isFenced() {
		// If the leading code fence is indented N spaces,
		// then up to N spaces of indentation are removed
		n := 0
		for n < cb.fenceIndent && n < len(s) && s[n] == ' ' {
			n++
		}
		s = s[n:]

		// until a closing code fence of the same type as the code block
		// began with (backticks or tildes), and with at least as many backticks
		// or tildes as the opening code fence.
		if len(s) > 0 && s[0] == cb.fenceMarker {
			n := 0
			for n < len(s) && s[n] == cb.fenceMarker {
				n++
			}
			if (n == len(s) || s[n] == '\n') && n >= cb.fenceLength {
				cb.closed = true
				return true
			}
		}
	} else {
		isIndented := len(s) >= 4 && s[0] == ' ' && s[1] == ' ' && s[2] == ' ' && s[3] == ' '
		if !isIndented {
			cb.closed = true
			return false
		}
		s = s[4:]
	}

	cb.lines = append(cb.lines, string(s))
	return true
}

func (cb *CodeBlock) String() string {
	return strings.Join(cb.lines, "")
}

// A LinkReferenceDefinition defines a label which can be used in reference links and reference-style images elsewhere in the document.
// It does not correspond to a structural element of a document. Instead, it can come either before or after the links that use them.
type LinkReferenceDefinition struct {
	// It consists of a link label, indented up to three spaces, followed by a colon (:)
	Label string

	// optional whitespace (including up to one line ending), a link destination,
	// optional whitespace (including up to one line ending),
	Destination string

	// and an optional link title, which if it is present must be separated from the link destination by whitespace.
	Title string
}

func (l *LinkReferenceDefinition) AddLine(c []rune) bool {
	return false
}

/* INLINES BELOW */

type _Inliner interface {
	parseInlines()
}

type Inline interface {
}

type ITextContent interface {
	TextContent() string
}

func textContent(i interface{}) string {
	if tc, ok := i.(ITextContent); ok {
		return tc.TextContent()
	}
	switch i.(type) {
	case *SoftLineBreak, *HardLineBreak:
		return " "
	}
	return ""
}

type Text struct {
	Text string
}

func (t *Text) TextContent() string {
	return t.Text
}

type Delimiter struct {
	textElement *list.Element
	active      bool
	text        string

	runePrev rune
	runeNext rune
}

func (d *Delimiter) prevChar() rune {
	if d.runePrev == 0 {
		prev := d.textElement.Next()
		if prev != nil {
			prevText := textContent(prev.Value)
			if prevText == "" {
				prevText = "p" // dummy. prevText == "" => it is a inline element.
			}
			d.runePrev = rune(prevText[len(prevText)-1])
		} else {
			d.runePrev = ' '
		}
	}
	return d.runePrev
}

func (d *Delimiter) nextChar() rune {
	if d.runeNext == 0 {
		next := d.textElement.Prev()
		if next != nil {
			nextText := textContent(next.Value)
			if nextText == "" {
				nextText = "n" // dummy. nextText == "" => it is a inline element.
			}
			d.runeNext = rune(nextText[0])
		} else {
			d.runeNext = ' '
		}
	}
	return d.runeNext
}

// A left-flanking delimiter run is a delimiter run that is
//   (1) not followed by Unicode whitespace,
//   (2a) not followed by a punctuation character
//  or
//   (2b) followed by a punctuation character and preceded by Unicode whitespace or a punctuation character.
//
// For purposes of this definition, the beginning and the end of the line count as Unicode whitespace.
//
// 1 && (2a || 2b)
func (d *Delimiter) isLeftFlanking() bool {
	return !unicode.IsSpace(d.nextChar()) && (!isPunctuation(d.nextChar()) || (unicode.IsSpace(d.prevChar()) || isPunctuation(d.prevChar())))
}

// A right-flanking delimiter run is a delimiter run that is
//   (1) not preceded by Unicode whitespace, and either
//   (2a) not preceded by a punctuation character,
//  or
//   (2b) preceded by a punctuation character and followed by Unicode whitespace or a punctuation character.
//
// For purposes of this definition, the beginning and the end of the line count as Unicode whitespace.
//
// 1 && (2a || 2b)
func (d *Delimiter) isRightFlanking() bool {
	return !unicode.IsSpace(d.prevChar()) && (!isPunctuation(d.prevChar()) || (unicode.IsSpace(d.nextChar()) || isPunctuation(d.nextChar())))
}

func (d *Delimiter) canOpenEmphasis() bool {
	switch d.text[0] {
	case '*':
		return d.isLeftFlanking()
	case '_':
		return d.isLeftFlanking() && (!d.isRightFlanking() || isPunctuation(d.prevChar()))
	}
	return false
}

func (d *Delimiter) canCloseEmphasis() bool {
	switch d.text[0] {
	case '*':
		return d.isRightFlanking()
	case '_':
		return d.isRightFlanking() && (!d.isLeftFlanking() || isPunctuation(d.nextChar()))
	}
	return false
}

func (d *Delimiter) canBeStrong() bool {
	return len(d.text) >= 2
}

func (d *Delimiter) match(closer *Delimiter) bool {
	return d.text[0] == closer.text[0]
}

// very hard, taken from commonmark.js
func (d *Delimiter) oddMatch(closer *Delimiter) bool {
	opener := d
	return (closer.canOpenEmphasis() || opener.canCloseEmphasis()) &&
		len(closer.text)%3 > 0 &&
		(len(opener.text)+len(closer.text))%3 == 0
}

// since delimiters in d are the same, from what end to remove is ignored.
func (d *Delimiter) consume(count int) int {
	n := len(d.text)
	if count > n {
		panic("won't happen")
	}

	d.text = d.text[0 : n-count]
	t := d.textElement.Value.(*Text)
	t.Text = t.Text[0 : n-count]

	return len(d.text)
}

// A CodeSpan begins with a backtick string and ends with a backtick string of equal length.
type CodeSpan struct {
	text string
}

func (cs *CodeSpan) TextContent() string {
	text := cs.text

	// First, line endings are converted to spaces.
	if strings.HasSuffix(text, "\n") {
		text = text[:len(text)-1]
		text += " "
	}

	// If the resulting string both begins and ends with a space character,
	// but does not consist entirely of space characters,
	// a single space character is removed from the front and back.
	if n := len(text); n >= 2 {
		if text[0] == ' ' && text[n-1] == ' ' {
			allAreSpaces := true
			for j := 1; j < n-1; j++ {
				if text[j] != ' ' {
					allAreSpaces = false
					break
				}
			}
			if !allAreSpaces {
				text = text[1 : n-1]
			}
		}
	}

	return text
}

type Link struct {
	Inlines []Inline

	Link  string
	Title string
}

func (l *Link) TextContent() string {
	s := ""
	for _, inline := range l.Inlines {
		if tc, ok := inline.(ITextContent); ok {
			s += tc.TextContent()
		}
	}
	return s
}

type Image struct {
	Link  string
	Alt   string
	Title string

	inlines []Inline
}

func (i *Image) TextContent() string {
	return i.Alt
}

type Emphasis struct {
	Delimiter string
	Inlines   []Inline
}

func (e *Emphasis) TextContent() (s string) {
	for _, i := range e.Inlines {
		tc, ok := i.(ITextContent)
		if !ok {
			panic("implement ITextContent")
		}
		s += tc.TextContent()
	}
	return
}

// A HardLineBreak is a line break (not in a code span or HTML tag) that is preceded
// by two or more spaces and does not occur at the end of a block is parsed as a
// hard line break (rendered in HTML as a <br /> tag).
type HardLineBreak struct{}

// A SoftLineBreak is a regular line break (not in a code span or HTML tag) that is not preceded
// by two or more spaces or a backslash is parsed as a softbreak.
//
// A softbreak may be rendered in HTML either as a line ending or as a space.
// The result will be the same in browsers. In the examples, a line ending will be used.
//
// A renderer may also provide an option to render soft line breaks as hard line breaks.
type SoftLineBreak struct{}

// An HtmlTag (HTML tag) consists of an open tag, a closing tag, an HTML comment,
// a processing instruction, a declaration, or a CDATA section.
//
// Text between < and > that looks like an HTML tag is parsed as a raw HTML tag
// and will be rendered in HTML without escaping.
// Tag and attribute names are not limited to current HTML tags,
// so custom tags (and even, say, DocBook tags) may be used.
type HtmlTag struct {
	Tag string
}

/*
type HtmlTagType int

const (
	HtmlTagOpen = iota
	HtmlTagClosing
	HtmlTagComment
	HtmlTagProcessingInstruction
	HtmlTagDeclaration
	HtmlTagCDATA
)

type HtmlAttribute struct {

}
*/
