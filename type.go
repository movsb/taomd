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
