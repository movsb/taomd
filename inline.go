package main

import (
	"container/list"
	"strings"
	"unicode"
)

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
	switch i.(type) {
	case *SoftLineBreak, *HardLineBreak:
		return " "
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

	runePrev rune
	runeNext rune
}

func (d *Delimiter) prevChar() rune {
	if d.runePrev == 0 {
		prev := d.textElement.Next()
		if prev != nil {
			prevText := textContent(prev.Value)
			if prevText == "" {
				prevText = "p" // dummy. prevText == "" => it is a inline element.
			}
			d.runePrev = rune(prevText[len(prevText)-1])
		} else {
			d.runePrev = ' '
		}
	}
	return d.runePrev
}

func (d *Delimiter) nextChar() rune {
	if d.runeNext == 0 {
		next := d.textElement.Prev()
		if next != nil {
			nextText := textContent(next.Value)
			if nextText == "" {
				nextText = "n" // dummy. nextText == "" => it is a inline element.
			}
			d.runeNext = rune(nextText[0])
		} else {
			d.runeNext = ' '
		}
	}
	return d.runeNext
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
	return !unicode.IsSpace(d.nextChar()) && (!isPunctuation(d.nextChar()) || (unicode.IsSpace(d.prevChar()) || isPunctuation(d.prevChar())))
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
	return !unicode.IsSpace(d.prevChar()) && (!isPunctuation(d.prevChar()) || (unicode.IsSpace(d.nextChar()) || isPunctuation(d.nextChar())))
}

func (d *Delimiter) canOpenEmphasis() bool {
	switch d.text {
	case "*", "**":
		return d.isLeftFlanking()
	case "_", "__":
		return d.isLeftFlanking() && (!d.isRightFlanking() || isPunctuation(d.prevChar()))
	}
	return false
}

func (d *Delimiter) canCloseEmphasis() bool {
	switch d.text {
	case "*", "**":
		return d.isRightFlanking()
	case "_", "__":
		return d.isRightFlanking() && (!d.isLeftFlanking() || isPunctuation(d.nextChar()))
	}
	return false
}

func (d *Delimiter) isStrong() bool {
	return d.text == "**" || d.text == "__"
}

// since delimiters in d are the same, from what end to remove is ignored.
func (d *Delimiter) consume(count int) int {
	n := len(d.text)
	if count > n {
		panic("won't happen")
	}

	d.text = d.text[0 : n-count]
	t := d.textElement.Value.(*Text)
	t.Text = t.Text[0 : n-count]

	return len(d.text)
}

// A CodeSpan begins with a backtick string and ends with a backtick string of equal length.
type CodeSpan struct {
	text string
}

func (s *CodeSpan) String() string {
	text := s.text

	// First, line endings are converted to spaces.
	if strings.HasSuffix(text, "\n") {
		text = text[:len(text)-1]
		text += " "
	}

	// If the resulting string both begins and ends with a space character,
	// but does not consist entirely of space characters,
	// a single space character is removed from the front and back.
	if n := len(text); n >= 2 {
		if text[0] == ' ' && text[n-1] == ' ' {
			allAreSpaces := true
			for j := 1; j < n-1; j++ {
				if text[j] != ' ' {
					allAreSpaces = false
					break
				}
			}
			if !allAreSpaces {
				text = text[1 : n-1]
			}
		}
	}

	return "<code>" + escapeHTMLString(text) + "</code>"
}

type Link struct {
	Inlines []*Text
	Link    string
	Title   string
	Text    string
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

// A HardLineBreak is a line break (not in a code span or HTML tag) that is preceded
// by two or more spaces and does not occur at the end of a block is parsed as a
// hard line break (rendered in HTML as a <br /> tag).
type HardLineBreak struct{}

// A SoftLineBreak is a regular line break (not in a code span or HTML tag) that is not preceded
// by two or more spaces or a backslash is parsed as a softbreak.
//
// A softbreak may be rendered in HTML either as a line ending or as a space.
// The result will be the same in browsers. In the examples, a line ending will be used.
//
// A renderer may also provide an option to render soft line breaks as hard line breaks.
type SoftLineBreak struct{}

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
