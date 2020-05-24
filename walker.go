package taomd

type Walker struct {
	root     INode
	current  INode
	entering bool
}

func NewWalker(node INode) {}

func (w *Walker) isContainer(element interface{}) bool {
	switch element.(type) {
	case *Document, *BlockQuote, *List, *ListItem, *Paragraph, *Heading:
		return true
	case *Emphasis, *Link, *Image:
		return true
	default:
		return false
	}
}

func (w *Walker) Next() (INode, bool) {
	current := w.current
	entering := w.entering

	if current == nil {
		return nil, false
	}

	container := w.isContainer(current)

	if entering && container {
		if first := nd(current).firstChild; first != nil {
			w.current = first
			w.entering = true
		} else {
			w.entering = false
		}
	} else if current == w.root {
		w.current = nil
	} else if nd(current).next == nil {
		w.current = nd(current).parent
		w.entering = false
	} else {
		w.current = nd(current).next
		w.entering = true
	}

	return current, entering
}
