package main

import "strings"

type Blocker interface {
	AddLine(s []rune) bool
}

type Document struct {
	example int
	blocks  []Blocker
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

type List struct {
	Ordered bool
	Tight   bool

	Marker byte

	Start     int
	Delimeter byte

	Items []Blocker

	markerWidth int
	spacesWidth int
}

func (l *List) AddLine(s []rune) bool {
	if len(l.Items) == 0 {
		panic("list items == 0")
	}
	return l.Items[len(l.Items)-1].AddLine(s)
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
	spaces int
	blocks []Blocker
}

func (li *ListItem) AddLine(s []rune) bool {
	if len(s) == 1 && s[0] == '\n' {
		li.blocks = append(li.blocks, &BlankLine{})
		return true
	}
	_, nSkipped := peekSpaces(s, li.spaces)
	if nSkipped != li.spaces {
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
