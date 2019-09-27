package main

type Document struct {
	blocks []interface{}
}

// HorizontalRule is a horizontal rule (thematic breaks).
// https://spec.commonmark.org/0.29/#thematic-breaks
type HorizontalRule struct {
}

type Paragraph struct {
	text string
}
