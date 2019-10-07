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
		s += toHTML(block)
	}

	return s
}

func escapeAttr(s string) string {
	s = strings.ReplaceAll(s, " ", "%20")
	s = strings.ReplaceAll(s, "\\", "%5C")
	s = strings.ReplaceAll(s, "\"", "%22")
	s = strings.ReplaceAll(s, "\xc2\xa0", "%C2%A0")
	return s
}

func toHTML(block Blocker) string {
	s := ""
	switch typed := block.(type) {
	default:
		panic("unhandled block: " + reflect.TypeOf(typed).String())
	case *Paragraph:
		if !typed.Tight {
			s += "<p>"
		}
		for _, inline := range typed.Inlines {
			switch it := inline.(type) {
			default:
				panic("unhandled inline: " + reflect.TypeOf(it).String())
			case *Text:
				s += html.EscapeString(it.Text)
			case *Link:
				s += fmt.Sprintf(`<a href="%s"`, escapeAttr(it.Link))
				if it.Title != "" {
					s += fmt.Sprintf(` title="%s"`, escapeAttr(it.Title))
				}
				s += ">"
				for _, inline := range it.Inlines {
					s += html.EscapeString(inline.Text)
				}
				s += "</a>"
			case *Image:
				panic("no image")
			}
		}

		if !typed.Tight {
			s += "</p>\n"
		}
	case *BlankLine:
		break
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
	case *BlockQuote:
		s += "<blockquote>\n"
		for _, b := range typed.blocks {
			s += toHTML(b)
		}
		s += "</blockquote>\n"
	case *List:
		typed.deduceIsTight()
		if typed.Ordered {
			s += "<ol>\n"
		} else {
			s += "<ul>\n"
		}

		for _, item := range typed.Items {
			if _, ok := item.(*BlankLine); ok {
				continue
			}
			if typed.Tight {
				s += "<li>"
			} else {
				s += "<li>\n"
			}
			for _, block := range item.(*ListItem).blocks {
				if p, ok := block.(*Paragraph); ok {
					p.Tight = typed.Tight
				}
				s += toHTML(block)
			}
			s += "</li>\n"
		}

		if typed.Ordered {
			s += "</ol>\n"
		} else {
			s += "</ul>\n"
		}
	}
	return s
}
