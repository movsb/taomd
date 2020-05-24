package taomd

import (
	"container/list"
	"regexp"
	"strings"
	"unicode"
)

// INode is implemented by all blocks.
type INode interface {
	getNodeData() *NodeData
	appendText(s string)
	acceptsLines() bool
	finalize(p *Parser)
	shouldFollow(p *Parser) int
}

// NodeData are private data for manipulating document parse tree.
type NodeData struct {
	node            INode
	parent          INode
	prev            INode
	next            INode
	firstChild      INode
	lastChild       INode
	lastLineBlank   bool
	lastLineChecked bool
	closed          bool
	Pos
}

func nd(node INode) *NodeData {
	return node.getNodeData()
}

func (n *NodeData) unlink() {
	if n.prev != nil {
		nd(n.prev).next = n.next
	} else if n.parent != nil {
		nd(n.parent).firstChild = n.next
	}

	if n.next != nil {
		nd(n.next).prev = n.prev
	} else if n.parent != nil {
		nd(n.parent).lastChild = n.prev
	}

	n.parent = nil
	n.next = nil
	n.prev = nil
}

func (n *NodeData) appendChild(child INode) {
	nd(child).unlink()
	nd(child).parent = n.node
	if n.lastChild != nil {
		nd(n.lastChild).next = child
		nd(child).prev = n.lastChild
		n.lastChild = child
	} else {
		n.firstChild = child
		n.lastChild = child
	}
}

func (n *NodeData) prependChild(child INode) {
	nd(child).unlink()
	nd(child).parent = n.node
	if n.firstChild != nil {
		nd(n.firstChild).prev = child
		nd(child).next = n.firstChild
		n.firstChild = child
	} else {
		n.firstChild = child
		n.lastChild = child
	}
}

func (n *NodeData) insertAfter(node INode) {
	nd(node).unlink()
	nd(node).next = n.next
	if n.next != nil {
		nd(n.next).prev = node
	}
	nd(node).prev = n.node
	n.next = node
	nd(node).parent = n.parent
	if nd(node).next == nil {
		nd(nd(node).parent).lastChild = node
	}
}

func (n *NodeData) insertBefore(node INode) {
	nd(node).unlink()
	nd(node).prev = n.prev
	if nd(node).prev != nil {
		nd(nd(node).prev).next = node
	}
	nd(node).next = n.node
	n.prev = node
	nd(node).parent = n.parent
	if nd(node).prev == nil {
		nd(nd(node).parent).firstChild = node
	}
}

func (n *NodeData) getNodeData() *NodeData {
	return n
}

type EmptyFinalizer struct{}

func (EmptyFinalizer) finalize(*Parser) {}

type NotAcceptLines struct{}

func (NotAcceptLines) acceptsLines() bool { return false }

func (NotAcceptLines) appendText(string) {}

type LineAcceptor struct {
	Content string
}

func (la *LineAcceptor) acceptsLines() bool { return true }
func (la *LineAcceptor) appendText(s string) {
	la.Content += s
}

type Document struct {
	NodeData
	NotAcceptLines
	EmptyFinalizer

	links map[string]*LinkReferenceDefinition
}

func (doc *Document) shouldFollow(p *Parser) int {
	return 0
}

func (doc *Document) parseInlines() {
	for e := nd(doc).firstChild; e != nil; e = nd(e).next {
		if inliner, ok := e.(_Inliner); ok {
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

// HorizontalRule is a horizontal rule (thematic breaks).
// https://spec.commonmark.org/0.29/#thematic-breaks
type HorizontalRule struct {
	Marker rune

	NodeData
	EmptyFinalizer
	NotAcceptLines

	// temp for parse as setext heading
	s []rune
}

func (hr *HorizontalRule) shouldFollow(p *Parser) int {
	return 1
}

type Paragraph struct {
	LineAcceptor
	texts   []string
	Tight   bool
	Inlines []Inline
	NodeData
}

func (pp *Paragraph) shouldFollow(p *Parser) int {
	if p.blank {
		return 1
	}
	return 0
}

func (pp *Paragraph) finalize(p *Parser) {
	pp.parseDefinitions()
	if isBlankLine([]rune(pp.Content)) {
		pp.unlink()
	}
}

func (pp *Paragraph) parseDefinitions() {
	raw := []rune(pp.Content)
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

	pp.Content = string(raw)
}

func (pp *Paragraph) parseInlines() {
	raw := strings.TrimSpace(pp.Content)
	pp.Inlines = parseInlines(raw)
}

type Heading struct {
	Level   int
	Inlines []Inline
	text    string
	NodeData
	NotAcceptLines
	EmptyFinalizer
}

func (h *Heading) shouldFollow(*Parser) int {
	return 1
}

func (h *Heading) parseInlines() {
	text := strings.TrimSpace(h.text)
	h.Inlines = parseInlines(text)
}

// An HTML block is a group of lines that is treated as raw HTML (and will not be escaped in HTML output).
type HtmlBlock struct {
	Lines [][]rune

	condition int
	closed    bool

	NodeData
	LineAcceptor
	EmptyFinalizer
}

func (hb *HtmlBlock) shouldFollow(p *Parser) int {
	if p.blank && (hb.condition == 6 || hb.condition == 7) {
		return 1
	}
	return 0
}

// NewHtmlBlock news a HtmlBlock.
func NewHtmlBlock(p *Parser, condition int, c []rune) *HtmlBlock {
	h := HtmlBlock{}
	h.condition = condition
	return &h
}

type BlockQuote struct {
	NodeData
	NotAcceptLines
	EmptyFinalizer
}

func (bq *BlockQuote) shouldFollow(p *Parser) int {
	if !p.indented && p.at(p.nextNonspace) == '>' {
		advanceNextNonspace()
		advanceOffset(1, false)
		if isSpaceOrTab(byte(p.at(p.offset))) {
			advanceOffset(1, true)
		}
		return 0
	}
	return 1
}

func (bq *BlockQuote) parseInlines() {
	for e := bq.firstChild; e != nil; e = nd(e).next {
		if inliner, ok := e.(_Inliner); ok {
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

	NodeData

	NotAcceptLines

	closed bool
}

func (l *List) match(r *List) bool {
	return l.Ordered == l.Ordered && l.MarkerChar == l.MarkerChar
}

func (l *List) shouldFollow(p *Parser) int {
	return 0
}

// A list is loose if any of its constituent list items are separated by blank lines,
// or if any of its constituent list items directly contain two block-level elements
// with a blank line between them. Otherwise a list is tight.
//
// The difference in HTML output is that paragraphs in a loose list are
// wrapped in <p> tags, while paragraphs in a tight list are not.
func (l *List) deduceIsTight() {
	l.Tight = true

	for item := nd(l).firstChild; item != nil; {
		if p.endsWithBlank(item) && nd(item).next != nil {
			l.Tight = false
			break
		}

		for sub := nd(item).firstChild; sub != nil; {
			if p.endsWithBlank(sub) && (nd(item).next != nil || nd(sub).next != nil) {
				l.Tight = false
				break
			}
			sub = nd(sub).next
		}

		item = nd(item).next
	}
}

func (l *List) finalize(p *Parser) {
	l.deduceIsTight()
}

func (l *List) parseInlines() {
	for e := l.firstChild; e != nil; e = nd(e).next {
		if inliner, ok := e.(_Inliner); ok {
			inliner.parseInlines()
		}
	}
}

type ListMarkerData struct {
	Ordered      bool
	MarkerChar   byte
	Start        int
	Padding      int
	MarkerOffset int
}

type ListItem struct {
	markerOffset int
	padding      int
	NodeData
	NotAcceptLines
	EmptyFinalizer
}

func (li *ListItem) shouldFollow(p *Parser) int {
	if p.blank {
		if li.firstChild == nil {
			return 1
		}
		advanceNextNonspace()
		return 0
	} else if p.indentation >= li.markerOffset+li.padding {
		advanceOffset(li.markerOffset+li.padding, true)
		return 0
	} else {
		return 1
	}
}

func (li *ListItem) parseInlines() {
	for e := li.firstChild; e != nil; e = nd(e).next {
		if inliner, ok := e.(_Inliner); ok {
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

	// A code fence is a sequence of at least three consecutive backtick
	// characters (`) or tildes (~). (Tildes and backticks cannot be mixed.)
	fenceMarker byte

	// A fenced code block begins with a code fence, indented no more than three spaces.
	// If the leading code fence is indented N spaces, then up to N spaces of indentation
	// are removed from each line of the content (if present).
	fenceIndent int

	// The content of the code block consists of all subsequent lines,
	// until a closing code fence of the same type as the code block
	// began with (backticks or tildes), and with at least as many backticks
	// or tildes as the opening code fence.
	fenceLength int

	NodeData
	LineAcceptor
}

func (cb *CodeBlock) shouldFollow(p *Parser) int {
	if cb.isFenced() {
		if !p.indented && cb.isClosingFence(p.line[p.nextNonspace:]) {
			p.finalize(cb, p.ln)
			return 2
		}
		// TODO skip optional spaces of fence offset
		for i := cb.fenceIndent; i > 0 && isSpaceOrTab(p.at(p.offset)); i-- {
			advanceOffset(1, true)
		}
		return 0
	} else {
		if p.indented {
			advanceOffset(4, true)
			return 0
		} else if p.blank {
			advanceNextNonspace()
			return 0
		} else {
			return 1
		}
	}
}

func (cb *CodeBlock) finalize(p *Parser) {
	if cb.isFenced() {
		pos := strings.IndexByte(cb.Content, '\n')
		info := unescapeString(strings.TrimSpace(cb.Content[0:pos]))
		cb.Content = cb.Content[pos+1:]
		if pos := strings.IndexFunc(info, unicode.IsSpace); pos != -1 {
			cb.Lang = info[:pos]
			cb.Args = info[pos:]
		} else {
			cb.Lang = info
		}
	} else {
		cb.Content = regexp.MustCompile(`(\n *)+$`).ReplaceAllLiteralString(cb.Content, "\n")
	}
}

func (cb *CodeBlock) isFenced() bool {
	return cb.fenceLength > 0
}

func (cb *CodeBlock) isClosingFence(s []rune) bool {
	i := 0

	// Closing fences may be indented by 0-3 spaces, and their
	// indentation need not match that of the opening fence.
	for i < 3 && i < len(s) && isSpace(s[i]) {
		i++
	}

	// until a closing code fence of the same type as the code block
	// began with (backticks or tildes), and with at least as many backticks
	// or tildes as the opening code fence.
	n := i
	for n < len(s) && byte(s[n]) == cb.fenceMarker {
		n++
	}
	if n-i >= cb.fenceLength && isBlankLine(s[n:]) {
		return true
	}

	return false
}

func (cb *CodeBlock) String() string {
	return cb.Content
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
