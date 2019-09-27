package main

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
