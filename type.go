package main

import (
	"strings"
)

type BlankLine struct {
}

// HorizontalRule is a horizontal rule (thematic breaks).
// https://spec.commonmark.org/0.29/#thematic-breaks
type HorizontalRule struct {
}

func (hr *HorizontalRule) AddLine(s []rune) bool {
	return false
}

type Paragraph struct {
	texts []string
	Text  string
}

func (p *Paragraph) AddLine(s []rune) bool {
	if _, ok := in(s, '=', '-'); ok {
		return false
	}
	p.texts = append(p.texts, string(s))
	return true
}

func (p *Paragraph) ParseInlines() {
	c := []rune(strings.Join(p.texts, ""))
	s := []rune{}
	i := 0
	for i < len(c) {
		switch c[i] {
		case '\\':
			i++
			if i < len(c) {
				escape := escapeHTML(c[i])
				if isPunctuation(c[i]) {
					s = append(s, escape...)
				} else if c[i] == '\n' {
					// A backslash at the end of the line is a hard line break
					s = append(s, []rune("<br />\n")...)
				} else {
					s = append(s, '\\')
					s = append(s, escape...)
				}
				i++
			} else {
				s = append(s, '\\')
			}
		case '`':
			if nc, span := tryParseCodeSpan(c[i:]); span != nil {
				i = 0
				c = nc
				s = append(s, []rune(span.String())...)
				continue
			}
			for ; i < len(c) && c[i] == '`'; i++ {
				s = append(s, '`')
			}
		default:
			s = append(s, escapeHTML(c[i])...)
			i++
		}
	}
	p.Text = strings.TrimSpace(string(s))
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
