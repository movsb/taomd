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
	if len(s) == 0 || len(s) == 1 && s[0] == '\n' {
		return false
	}
	if _, ok := in(s, '=', '-'); ok {
		return false
	}
	p.texts = append(p.texts, string(s))
	return true
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

type CodeBlock struct {
	Lang        string
	lines       []string
	start       rune
	indent      int
	fenceLength int
	Info        string
	closed      bool
}

func (cb *CodeBlock) AddLine(s []rune) bool {
	if cb.closed {
		return false
	}

	if !isText(s) {
		return false
	}

	// If the leading code fence is indented N spaces,
	// then up to N spaces of indentation are removed
	n := 0
	for n < cb.indent && n < len(s) && s[n] == ' ' {
		n++
	}
	s = s[n:]

	// until a closing code fence of the same type as the code block
	// began with (backticks or tildes), and with at least as many backticks
	// or tildes as the opening code fence.
	if len(s) > 0 && s[0] == cb.start {
		n := 0
		for n < len(s) && s[n] == cb.start {
			n++
		}
		if (n == len(s) || s[n] == '\n') && n >= cb.fenceLength {
			cb.closed = true
			return true
		}
	}

	cb.lines = append(cb.lines, string(s))
	return true
}

func (cb *CodeBlock) String() string {
	return strings.Join(cb.lines, "")
}

type _CodeChunk struct {
	text string
}

type CodeSpan struct {
	text string
}

func (s *CodeSpan) String() string {
	text := s.text

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

	return "<code>" + escapeHTMLString(text) + "</code>"
}
