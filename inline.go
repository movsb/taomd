package main

import "container/list"

type _Inliner interface {
	parseInlines()
}

type Inline interface {
}

type ITextContent interface {
	TextContent() string
}

func textContent(i interface{}) string {
	if tc, ok := i.(ITextContent); ok {
		return tc.TextContent()
	}
	return ""
}

type Text struct {
	Text string
}

func (t *Text) TextContent() string {
	return t.Text
}

type Delimiter struct {
	textElement *list.Element
	active      bool
	text        string
}

// A left-flanking delimiter run is a delimiter run that is
//   (1) not followed by Unicode whitespace,
//   (2a) not followed by a punctuation character
//  or
//   (2b) followed by a punctuation character and preceded by Unicode whitespace or a punctuation character.
//
// For purposes of this definition, the beginning and the end of the line count as Unicode whitespace.
//
// 1 && (2a || 2b)
func (d *Delimiter) isLeftFlanking() bool {
	switch d.text[0] {
	case '*', '_':
		break
	default:
		return false
	}

	next := d.textElement.Prev()
	// the end of the line count as Unicode whitespace
	if next == nil {
		return false
	}

	// not followed by Unicode whitespace
	nextText := textContent(next.Value)
	if nextText[0] == ' ' || nextText[0] == '\n' {
		return false
	}

	// not followed by a punctuation character
	if !isPunctuation(rune(nextText[0])) {
		return true
	}

	prev := d.textElement.Next()
	if prev == nil {
		return true
	}

	prevText := textContent(prev.Value)
	lastChar := prevText[len(prevText)-1]
	if lastChar == ' ' || lastChar == '\n' || isPunctuation(rune(lastChar)) {
		return true
	}

	return false
}

// A right-flanking delimiter run is a delimiter run that is
//   (1) not preceded by Unicode whitespace, and either
//   (2a) not preceded by a punctuation character,
//  or
//   (2b) preceded by a punctuation character and followed by Unicode whitespace or a punctuation character.
//
// For purposes of this definition, the beginning and the end of the line count as Unicode whitespace.
//
// 1 && (2a || 2b)
func (d *Delimiter) isRightFlanking() bool {
	switch d.text[0] {
	case '*', '_':
		break
	default:
		return false
	}

	prev := d.textElement.Next()
	// the end of the line count as Unicode whitespace
	if prev == nil {
		return false
	}

	// not preceded by Unicode whitespace
	prevText := textContent(prev.Value)
	lastChar := prevText[len(prevText)-1]
	if lastChar == ' ' || lastChar == '\n' {
		return false
	}

	// not preceded by a punctuation character
	if !isPunctuation(rune(lastChar)) {
		return true
	}

	next := d.textElement.Prev()
	if next == nil {
		return true
	}

	nextText := textContent(next.Value)
	if nextText[0] == ' ' || nextText[0] == '\n' || isPunctuation(rune(nextText[0])) {
		return true
	}

	return false
}

func delimiterPrecedeOrFollowedByPunctuation(d *Delimiter, precede bool) bool {
	if precede {
		te := d.textElement.Next()
		if te == nil {
			return false
		}
		t := textContent(te.Value)
		return isPunctuation(rune(t[len(t)-1]))
	} else {
		te := d.textElement.Prev()
		if te == nil {
			return false
		}
		t := textContent(te.Value)
		return isPunctuation(rune(t[0]))
	}
}

func (d *Delimiter) canOpenEmphasis() bool {
	switch d.text {
	case "*", "**":
		return d.isLeftFlanking()
	case "_", "__":
		return d.isLeftFlanking() && (!d.isRightFlanking() || delimiterPrecedeOrFollowedByPunctuation(d, true))
	}
	return false
}

func (d *Delimiter) canCloseEmphasis() bool {
	switch d.text {
	case "*", "**":
		return d.isRightFlanking()
	case "_", "__":
		return d.isRightFlanking() && (!d.isLeftFlanking() || delimiterPrecedeOrFollowedByPunctuation(d, false))
	}
	return false
}

func (d *Delimiter) isStrong() bool {
	return d.text == "**" || d.text == "__"
}

type Link struct {
	Inlines []*Text
	Link    string
	Title   string
}

func (l *Link) TextContent() string {
	return l.Title
}

type Image struct {
	Link    string
	inlines []*Text
	Alt     string
	Title   string
}

func (i *Image) TextContent() string {
	return i.Alt
}

type Emphasis struct {
	Delimiter string
	Inlines   []Inline
}

// "<br />"
type LineBreak struct {
}

// " "
type SoftBreak struct {
}

// An HtmlTag (HTML tag) consists of an open tag, a closing tag, an HTML comment,
// a processing instruction, a declaration, or a CDATA section.
//
// Text between < and > that looks like an HTML tag is parsed as a raw HTML tag
// and will be rendered in HTML without escaping.
// Tag and attribute names are not limited to current HTML tags,
// so custom tags (and even, say, DocBook tags) may be used.
type HtmlTag struct {
	Tag string
}

/*
type HtmlTagType int

const (
	HtmlTagOpen = iota
	HtmlTagClosing
	HtmlTagComment
	HtmlTagProcessingInstruction
	HtmlTagDeclaration
	HtmlTagCDATA
)

type HtmlAttribute struct {

}
*/
