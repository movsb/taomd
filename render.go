package main

import (
	"fmt"
	"reflect"
	"strings"
)

func render(doc *Document) string {
	var s string

	for _, block := range doc.blocks {
		switch typed := block.(type) {
		default:
			panic("unhandled block: " + fmt.Sprint(doc.example) + "," + reflect.TypeOf(typed).String())
		case *BlankLine:
			s += "\n"
		case *Paragraph:
			s += fmt.Sprintf("<p>%s</p>\n", strings.Join(typed.texts, "\n"))
		case *HorizontalRule:
			_ = typed
			s += "<hr />\n"
		case *Heading:
			s += fmt.Sprintf("<h%[1]d>%s</h%[1]d>\n", typed.Level, typed.text)
		}
	}

	return s
}
