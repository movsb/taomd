package taomd

import (
	"fmt"
	"reflect"
	"strings"
)

func Render(doc *Document) string {
	var s string

	for e := doc.firstChild; e != nil; e = nd(e).next {
		s += toHTML(e)
	}

	return s
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
		s += fmt.Sprintf(`<a href="%s"`, escapeText(urlEncode(it.Link)))
		if it.Title != "" {
			s += fmt.Sprintf(` title="%s"`, escapeText(it.Title))
		}
		s += ">"
		for _, inline := range it.Inlines {
			s += toInline(inline)
		}
		s += "</a>"
	case *Image:
		s += fmt.Sprintf(`<img src="%s"`, escapeText(urlEncode(it.Link)))
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

func toHTML(node INode) string {
	s := ""
	switch typed := node.(type) {
	default:
		panic("unhandled node: " + reflect.TypeOf(typed).String())
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
		for e := typed.firstChild; e != nil; e = nd(e).next {
			s += toHTML(e)
		}
		s += "</blockquote>\n"
	case *List:
		if typed.Ordered {
			if typed.Start == 1 {
				s += "<ol>\n"
			} else {
				s += fmt.Sprintf("<ol start=\"%d\">\n", typed.Start)
			}
		} else {
			s += "<ul>\n"
		}

		for e := typed.firstChild; e != nil; e = nd(e).next {
			s += "<li>"
			for f := nd(e).firstChild; f != nil; f = nd(f).next {
				if pp, ok := f.(*Paragraph); ok {
					pp.Tight = typed.Tight
					if !typed.Tight && f == nd(e).firstChild {
						s += "\n"
					}
				} else {
					if pp, ok := nd(f).prev.(*Paragraph); ok {
						if pp.Tight {
							s += "\n"
						}
					} else if f == nd(e).firstChild {
						s += "\n"
					}
				}
				s += toHTML(f)
			}
			s += "</li>\n"
		}

		if typed.Ordered {
			s += "</ol>\n"
		} else {
			s += "</ul>\n"
		}
	case *HtmlBlock:
		s += typed.Content
	}
	return s
}
