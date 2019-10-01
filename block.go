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
