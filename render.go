package main

import "fmt"

func render(doc *Document) string {
	var s string

	for _, block := range doc.blocks {
		switch typed := block.(type) {
		default:
			panic("unhandled block")
		case *Paragraph:
			s += fmt.Sprintf("<p>%s</p>\n", typed.text)
		case *HorizontalRule:
			_ = typed
			s += "<hr />\n"

		}
	}

	return s
}
