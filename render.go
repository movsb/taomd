package main

import (
	"fmt"
	"net/url"
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

func escapeURL(s string) string {
	u, err := url.Parse(s)
	if err != nil {
		return escapeText(s)
	}
	return u.String()
}

func escapeText(s string) string {
	return strings.NewReplacer(
		`&`, "&amp;",
		`"`, "&quot;",
		`<`, "&lt;",
		`>`, "&gt;",
	).Replace(s)
}

func toInline(inline Inline) string {
	s := ""
	switch it := inline.(type) {
	default:
		panic("unhandled inline: " + reflect.TypeOf(it).String())
	case *Text:
		s += escapeText(it.Text)
	case *Link:
		s += fmt.Sprintf(`<a href="%s"`, escapeURL(it.Link))
		if it.Title != "" {
			s += fmt.Sprintf(` title="%s"`, escapeText(it.Title))
		}
		s += ">"
		for _, inline := range it.Inlines {
			s += toInline(inline)
		}
		s += "</a>"
	case *Image:
		s += fmt.Sprintf(`<img src="%s"`, escapeURL(it.Link))
		s += fmt.Sprintf(` alt="%s"`, escapeText(it.Alt))
		if it.Title != "" {
			s += fmt.Sprintf(` title="%s"`, escapeText(it.Title))
		}
		s += " />"
	case *Emphasis:
		switch it.Delimiter {
		default:
			panic("unknown delimiter")
		case "*", "_":
			s += "<em>"
		case "**", "__":
			s += "<strong>"
		}

		for _, i := range it.Inlines {
			s += toInline(i)
		}

		switch it.Delimiter {
		default:
			panic("unknown delimiter")
		case "*", "_":
			s += "</em>"
		case "**", "__":
			s += "</strong>"
		}
	case *HardLineBreak:
		s += "<br />\n"
	case *SoftLineBreak:
		s += "\n"
	case *CodeSpan:
		s += "<code>" + escapeText(it.TextContent()) + "</code>"
	case *HtmlTag:
		s += it.Tag
	}
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
			s += toInline(inline)
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
		s += fmt.Sprintf("<h%d>", typed.Level)
		for _, inline := range typed.Inlines {
			s += toInline(inline)
		}
		s += fmt.Sprintf("</h%d>\n", typed.Level)
	case *CodeBlock:
		if typed.Lang == "" {
			s += fmt.Sprintf("<pre><code>%s</code></pre>\n", escapeText(typed.String()))
		} else {
			lang := typed.Lang
			if p := strings.IndexAny(lang, " \t"); p != -1 {
				lang = lang[:p]
			}
			s += fmt.Sprintf(
				"<pre><code class=\"language-%s\">%s</code></pre>\n",
				lang,
				escapeText(typed.String()),
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
			if typed.Start == 1 {
				s += "<ol>\n"
			} else {
				s += fmt.Sprintf("<ol start=\"%d\">\n", typed.Start)
			}
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
			var lastParagraph *Paragraph
			for _, block := range item.(*ListItem).blocks {
				if lastParagraph != nil && lastParagraph.Tight {
					s += "\n"
				}
				if p, ok := block.(*Paragraph); ok {
					p.Tight = typed.Tight
					lastParagraph = p
				} else {
					lastParagraph = nil
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
	case *HtmlBlock:
		for _, line := range typed.Lines {
			s += string(line)
		}
	}
	return s
}
