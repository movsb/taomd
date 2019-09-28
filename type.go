package main

import (
	"strings"
)

type Document struct {
	example int
	blocks  []interface{}
}

type BlankLine struct {
}

// HorizontalRule is a horizontal rule (thematic breaks).
// https://spec.commonmark.org/0.29/#thematic-breaks
type HorizontalRule struct {
}

type Paragraph struct {
	texts []string
	Text  string
}

func (p *Paragraph) ParseInlines() {
	c := []rune(strings.Join(p.texts, "\n"))
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
		default:
			s = append(s, escapeHTML(c[i])...)
			i++
		}
	}
	p.Text = string(s)
}

type Line struct {
	text string
}

type Heading struct {
	Level int
	text  string
}

type CodeBlock struct {
	Lang   string
	chunks []*_CodeChunk
}

func (s *CodeBlock) String() string {
	i := 0

	for i < len(s.chunks) && s.chunks[i].text == "" {
		i++
	}
	s.chunks = s.chunks[i:]

	i = len(s.chunks) - 1
	for i > 0 && s.chunks[i].text == "" {
		i--
	}
	s.chunks = s.chunks[:i+1]

	t := ""

	for _, c := range s.chunks {
		t += c.text + "\n"
	}
	return t
}

type _CodeChunk struct {
	text string
}
