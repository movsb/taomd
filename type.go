package main

import (
	"strings"
)

type BlankLine struct {
}

func (bl *BlankLine) AddLine(s []rune) bool {
	return false
}

// HorizontalRule is a horizontal rule (thematic breaks).
// https://spec.commonmark.org/0.29/#thematic-breaks
type HorizontalRule struct {
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
	if !addLine(&blocks, s) || len(blocks) == 0 {
		panic("won't happen")
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
	Level int
	text  string
}

func (h *Heading) AddLine(s []rune) bool {
	return false
}
