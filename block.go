package main

import "strings"

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

func (doc *Document) refLink(label string, link *Link, enclosed bool) bool {
	if !enclosed {
		label = "[" + label + "]"
	}
	label = strings.ToLower(label)
	ref, ok := doc.links[label]
	if !ok {
		return false
	}
	link.Link = ref.Destination
	link.Title = ref.Title
	return true
}

type BlockQuote struct {
	blocks []Blocker
}

func (bq *BlockQuote) AddLine(s []rune) bool {
	_, ok := tryParseBlockQuote(s, bq)
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

// begin tries to parse List start indicator.
// func (l *List) begin(s []rune) bool {
// 	if len(l.Items) != 0 {
// 		panic("wrong func call")
// 	}
//
// 	// TODO no spaces after marker
// 	// -
// 	// -
// 	// -
// 	// is treated as list.
// 	_, n := peekSpaces(s, 4)
// 	if !(1 <= n && n <= 4) {
// 		return false
// 	}
//
// 	s, list, markerWidth, ok := l.parseMarker(s)
//
// 	return l.AddLine(s)
// }

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

	_, n := peekSpaces(s, 4)
	if n < 1 {
		return
	}

	s = s[n:]

	return s, list, prefixSpaces, prefixWidth + n, true
}

func (l *List) AddLine(s []rune) bool {
	var lastItem *ListItem

	for i := len(l.Items) - 1; i >= 0; i-- {
		if item, ok := l.Items[i].(*ListItem); ok {
			lastItem = item
			break
		}
	}

	if lastItem != nil {
		if lastItem.AddLine(s) {
			return true
		}
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

	return addLine(&lastItem.blocks, s)
}

// A list is loose if any of its constituent list items are separated by blank lines,
// or if any of its constituent list items directly contain two block-level elements
// with a blank line between them. Otherwise a list is tight.
//
// The difference in HTML output is that paragraphs in a loose list are
// wrapped in <p> tags, while paragraphs in a tight list are not.
func (l *List) deduceIsTight() {
	var bp Blocker
	var bl *BlankLine

	for _, item := range l.Items {
		switch t := item.(type) {
		case *ListItem:
			if bl != nil {
				l.Tight = false
				return
			}
			bp = t
		case *BlankLine:
			if bp != nil {
				bl = t
			}
		}

		if pItem, ok := item.(*ListItem); ok {
			var ibp Blocker
			var ibl *BlankLine
			for _, block := range pItem.blocks {
				switch t := block.(type) {
				default:
					if ibl != nil {
						l.Tight = false
						return
					}
					ibp = t
				case *BlankLine:
					if ibp != nil {
						ibl = t
					}
				}
			}
			if ibp != nil && ibl != nil {
				l.Tight = false
				return
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
	return addLine(&li.blocks, s)
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

	// The content of the code block consists of all subsequent lines
	lines []string

	// A code fence is a sequence of at least three consecutive backtick
	// characters (`) or tildes (~). (Tildes and backticks cannot be mixed.)
	fenceStart rune

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
		if len(s) > 0 && s[0] == cb.fenceStart {
			n := 0
			for n < len(s) && s[n] == cb.fenceStart {
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

	destinationLine []rune
	titleLine       []rune

	errored         bool
	wantDestination bool
	wantTitle       bool
}

func (l *LinkReferenceDefinition) AddLine(c []rune) bool {
	if l.wantDestination {
		l.destinationLine = []rune(string(c))
		_, c = skipPrefixSpaces(c, -1)
		if len(c) == 0 {
			l.wantDestination = false
			return false
		}
		c, dest, ok := parseLinkDestination(c)
		if !ok {
			return false
		}
		l.Destination = dest
		l.wantDestination = false

		_, c = skipPrefixSpaces(c, -1)
		if len(c) == 0 {
			return true
		}

		if c[0] == '\n' {
			return true
		}

		if !l.wantTitle {
			l.errored = true
			return false
		}

		c, title, ok := parseLinkTitle(c)
		if !ok {
			l.errored = true
			return false
		}
		l.Title = title
		l.wantTitle = false

		_, c = skipPrefixSpaces(c, -1)
		if len(c) != 0 || c[0] != '\n' {
			l.errored = true
			return false
		}

		return true
	}

	if l.wantTitle {
		l.titleLine = c
		_, c = skipPrefixSpaces(c, -1)
		if len(c) == 0 {
			l.wantTitle = false
			return true
		}

		if c[0] == '\n' {
			l.wantTitle = false
			return true
		}

		c, title, ok := parseLinkTitle(c)
		if !ok {
			l.errored = true
			return false
		}

		l.Title = title
		l.wantTitle = false

		_, c = skipPrefixSpaces(c, -1)
		if len(c) != 0 || c[0] != '\n' {
			l.errored = true
			return false
		}

		return true
	}

	return false
}
