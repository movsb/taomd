package main

import (
	"fmt"
	"html"
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
		case *CodeBlock:
			if typed.Lang == "" {
				s += fmt.Sprintf("<pre><code>%s</code></pre>\n", html.EscapeString(typed.String()))
			} else {
				lang := typed.Lang
				if p := strings.IndexAny(lang, " \t"); p != -1 {
					lang = lang[:p]
				}
				s += fmt.Sprintf(
					"<pre><code class=\"language-%s\">%s</code></pre>\n",
					lang,
					html.EscapeString(typed.String()),
				)
			}
		}
	}

	return s
}
