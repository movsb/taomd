package main

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
