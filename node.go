package taomd

import (
	"container/list"
	"strings"
	"unicode"
)

type Blocker interface {
	AddLine(p *Parser, s []rune) bool
}

type Document struct {
	blocks []Blocker
	links  map[string]*LinkReferenceDefinition
}

func (doc *Document) AddLine(p *Parser, s []rune) bool {
	p.reset(s)
	p.tip = doc
	return addLine(&doc.blocks, s)
}

func (doc *Document) parseInlines() {
	for _, block := range doc.blocks {
		if inliner, ok := block.(_Inliner); ok {
			inliner.parseInlines()
		}
	}
}

func (doc *Document) parseDefinitions() {
	for _, block := range doc.blocks {
		switch t := block.(type) {
		case *Paragraph:
			t.parseDefinitions()
		case *List:
			t.parseDefinitions()
		case *BlockQuote:
			t.parseDefinitions()
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

func (bl *BlankLine) AddLine(p *Parser, s []rune) bool {
	return false
}

// HorizontalRule is a horizontal rule (thematic breaks).
// https://spec.commonmark.org/0.29/#thematic-breaks
type HorizontalRule struct {
	Marker rune

	// temp for parse as setext heading
	s []rune
}

func (hr *HorizontalRule) AddLine(p *Parser, s []rune) bool {
	return false
}

type Paragraph struct {
	texts   []string
	Tight   bool
	Inlines []Inline
	closed  bool
	lazying bool
}

func (pp *Paragraph) AddLine(p *Parser, s []rune) bool {
	if pp.closed {
		return false
	}

	var blocks []Blocker
	p.tip = pp
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
		pp.texts = append(pp.texts, trimLeft(typed.texts[0]))
		return true
	case *CodeBlock:
		// An indented code block cannot interrupt a paragraph
		if !typed.isFenced() {
			// s: typed.lines[0] is trimmed 4 spaces at the beginning, don't use.
			pp.texts = append(pp.texts, trimLeft(string(s)))
			return true
		}
	case *List:
		// In order to solve of unwanted lists in paragraphs with hard-wrapped numerals,
		// we allow only lists starting with 1 to interrupt paragraphs.
		if typed.Ordered && typed.Start != 1 {
			pp.texts = append(pp.texts, trimLeft(string(s)))
			return true
		}
		// However, an empty list item cannot interrupt a paragraph
		if len(typed.Items) == 1 {
			if subItem, ok := typed.Items[0].(*ListItem); ok {
				if len(subItem.blocks) == 0 {
					pp.texts = append(pp.texts, trimLeft(string(s)))
					return true
				}
			}
		}
	case *LinkReferenceDefinition:
		// A link reference definition cannot interrupt a paragraph.
		pp.texts = append(pp.texts, trimLeft(string(s)))
		return true
	case *HtmlBlock:
		// HTML blocks of type 7 cannot interrupt a paragraph
		if typed.condition == 7 {
			pp.texts = append(pp.texts, trimLeft(string(s)))
			return true
		}
	}

	pp.closed = true
	//pp.parseDefinitions(p)
	return false
}

func (pp *Paragraph) parseDefinitions() bool {
	raw := []rune(strings.Join(pp.texts, ""))
	for {
		remain, link := tryParseLinkReferenceDefinition(raw)
		if link == nil {
			break
		}

		lower := strings.ToLower(link.Label)
		if _, ok := p.doc.links[lower]; !ok {
			p.doc.links[lower] = link
		}

		raw = remain
	}
	pp.texts = []string{string(raw)}
	return len(raw) == 0
}

func (pp *Paragraph) addLaziness(p *Parser, s []rune) bool {
	_, s = skipPrefixSpaces(s, -1)
	pp.lazying = true
	defer func() { pp.lazying = false }()
	return pp.AddLine(p, s)
}

func (pp *Paragraph) parseInlines() {
	raw := strings.Join(pp.texts, "")
	raw = strings.TrimSpace(raw)
	pp.Inlines = parseInlines(raw)
}

type Line struct {
	text string
}

type Heading struct {
	Level   int
	Inlines []Inline
	text    string
}

func (h *Heading) AddLine(p *Parser, s []rune) bool {
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

func (h *SetextHeading) AddLine(p *Parser, s []rune) bool {
	return false
}

// An HTML block is a group of lines that is treated as raw HTML (and will not be escaped in HTML output).
type HtmlBlock struct {
	Lines [][]rune

	condition int
	closed    bool
}

// NewHtmlBlock news a HtmlBlock.
func NewHtmlBlock(p *Parser, condition int, c []rune) *HtmlBlock {
	h := HtmlBlock{}
	h.condition = condition
	h.AddLine(p, c)
	return &h
}

func (hb *HtmlBlock) append(c []rune) {
	hb.Lines = append(hb.Lines, c)
}

func (hb *HtmlBlock) AddLine(p *Parser, c []rune) bool {
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

func (bq *BlockQuote) parseDefinitions() {
	for _, block := range bq.blocks {
		switch it := block.(type) {
		case *Paragraph:
			it.parseDefinitions()
		case *List:
			it.parseDefinitions()
		}
	}
}

func (bq *BlockQuote) AddLine(p *Parser, s []rune) bool {
	_, ok := tryParseBlockQuote(s, bq)
	if ok {
		tryMergeSetextHeading(&bq.blocks)
		return true
	}
	return bq.addLaziness(p, s)
}

func (bq *BlockQuote) parseInlines() {
	for _, block := range bq.blocks {
		if inliner, ok := block.(_Inliner); ok {
			inliner.parseInlines()
		}
	}
}

func (bq *BlockQuote) addLaziness(p *Parser, s []rune) bool {
	if len(bq.blocks) == 0 {
		return false
	}
	switch typed := bq.blocks[len(bq.blocks)-1].(type) {
	case *List:
		return typed.addLaziness(p, s)
	case *BlockQuote:
		return typed.addLaziness(p, s)
	case *Paragraph:
		return typed.addLaziness(p, s)
	}
	return false
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

	closed bool
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

	// When the list item starts with a blank line, the number of spaces
	// following the list marker doesn’t change the required indentation
	if isBlankLine(s) {
		return s[len(s):], list, prefixSpaces, prefixWidth + 1, true
	}

	// Explain for "5"
	// If the first block in the list item is an indented code block,
	// then by rule #2, the contents must be indented one space after the list marker
	_, n := peekSpaces(s, 5)

	if n < 1 {
		if len(s) == 0 || len(s) == 1 && s[0] == '\n' {
			// A list may start or end with an empty list item
			return s, list, prefixSpaces, prefixWidth + 1, true
		}
		return s, nil, 0, 0, false
	}

	switch {
	case n == 5:
		return s[1:], list, prefixSpaces, prefixWidth + 1, true
	default:
		return s[n:], list, prefixSpaces, prefixWidth + n, true
	}
}

func (l *List) parseDefinitions() {
	for _, item := range l.Items {
		switch t := item.(type) {
		case *ListItem:
			for _, block := range t.blocks {
				switch it := block.(type) {
				case *Paragraph:
					it.parseDefinitions()
				case *List:
					it.parseDefinitions()
				case *BlockQuote:
					it.parseDefinitions()
				}
			}
		}
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

func (l *List) AddLine(p *Parser, s []rune) bool {
	if l.closed {
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
		if lastItem.AddLine(p, s) {
			if isBlankLine(s) {
				l.Items = append(l.Items, &BlankLine{})
				return true
			}
			return true
		}

		if isBlankLine(s) {
			// lastItem.closed = true
			l.Items = append(l.Items, &BlankLine{})
			return true
		}

		//if lastItem.addLaziness(s) {
		//	return true
		//}
	}

	if isBlankLine(s) {
		l.Items = append(l.Items, &BlankLine{})
		return true
	}

	// Notice: returning s may be empty while ok == true.
	os := s
	s, list, prefixSpaces, markerWidth, ok := l.parseMarker(s)
	if !ok {
		// return false
		if lastItem != nil {
			return lastItem.addLaziness(p, os)
		}
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

	// If any line is a thematic break then that line is not a list item.
	//
	// When both a thematic break and a list item are possible
	// interpretations of a line, the thematic break takes precedence
	if l.isHorizontalRule(os) {
		return false
	}

	lastItem = &ListItem{
		prefixSpaces: prefixSpaces,
		suffixSpaces: markerWidth,
	}

	l.Items = append(l.Items, lastItem)

	// trick: A list may start or end with an empty list item
	if len(s) == 0 || len(s) == 1 && s[0] == '\n' {
		// lastItem.blocks = append(lastItem.blocks, &BlankLine{})
		return true
	}

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

func (l *List) addLaziness(p *Parser, s []rune) bool {
	if len(l.Items) > 0 {
		if lastItem, ok := l.Items[len(l.Items)-1].(*ListItem); ok {
			return lastItem.addLaziness(p, s)
		}
	}
	return false
}

type ListItem struct {
	prefixSpaces int
	suffixSpaces int
	blocks       []Blocker
	closed       bool
}

func (li *ListItem) addLaziness(p *Parser, s []rune) bool {
	if len(li.blocks) > 0 {
		switch typed := li.blocks[len(li.blocks)-1].(type) {
		case *List:
			return typed.addLaziness(p, s)
		case *Paragraph:
			return typed.addLaziness(p, s)
		case *BlockQuote:
			return typed.addLaziness(p, s)
		}
	}
	return false
}

func (li *ListItem) AddLine(p *Parser, s []rune) bool {
	if li.closed {
		return false
	}

	if len(s) == 1 && s[0] == '\n' {
		if len(li.blocks) > 0 && li.blocks[len(li.blocks)-1].AddLine(p, s) {
			return true
		}

		// A list item can begin with at most one blank line.
		// A blank line has been added while creating this ListItem.
		if len(li.blocks) == 0 {
			li.closed = true
			return false
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

func (cb *CodeBlock) isClosingFence(s []rune) bool {
	i := 0

	// Closing fences may be indented by 0-3 spaces, and their
	// indentation need not match that of the opening fence.
	for i < 3 && isSpace(s[i]) {
		i++
	}

	// until a closing code fence of the same type as the code block
	// began with (backticks or tildes), and with at least as many backticks
	// or tildes as the opening code fence.
	n := i
	for n < len(s) && s[n] == cb.fenceMarker {
		n++
	}
	if n-i >= cb.fenceLength && isBlankLine(s[n:]) {
		return true
	}

	return false
}

func (cb *CodeBlock) AddLine(p *Parser, s []rune) bool {
	if cb.closed {
		return false
	}

	if cb.isFenced() {
		return cb.addLineFenced(s)
	}

	return cb.addLineIndented(s)
}

func (cb *CodeBlock) addLineFenced(s []rune) bool {
	if cb.isClosingFence(s) {
		cb.closed = true
		return true
	}

	// If the leading code fence is indented N spaces,
	// then up to N spaces of indentation are removed
	n := 0
	for n < cb.fenceIndent && n < len(s) && s[n] == ' ' {
		n++
	}
	s = s[n:]

	cb.lines = append(cb.lines, string(s))
	return true
}

func (cb *CodeBlock) addLineIndented(s []rune) bool {
	if p.indented {
		advanceOffset(4, true)
		line := string(s[p.offset:])
		cb.lines = append(cb.lines, line)
		return true
	}
	if isBlankLine(s) {
		cb.lines = append(cb.lines, "\n")
		return true
	}
	return false
}

func (cb *CodeBlock) String() string {
	if cb.isFenced() {
		return strings.Join(cb.lines, "")
	}

	// indented code block starts from a blank line.
	// So, there must be at least one non-blank line.

	// Blank lines preceding or following an indented code block are not included in it
	// TODO optimize blank line checking
	i, j := 0, len(cb.lines)-1
	for isBlankLine([]rune(cb.lines[i])) {
		i++
	}
	for isBlankLine([]rune(cb.lines[j])) {
		j--
	}
	cb.lines = cb.lines[i : j+1]

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

func (l *LinkReferenceDefinition) AddLine(p *Parser, c []rune) bool {
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
