package main

import "container/list"

type _Inliner interface {
	parseInlines()
}

type Inline interface {
}

type Text struct {
	Text string
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
	nextText := next.Value.(*Text).Text
	if nextText[0] == ' ' {
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

	prevText := prev.Value.(*Text).Text
	lastChar := prevText[len(prevText)-1]
	if lastChar == ' ' || isPunctuation(rune(lastChar)) {
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
	prevText := prev.Value.(*Text).Text
	lastChar := prevText[len(prevText)-1]
	if lastChar == ' ' {
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

	nextText := prev.Value.(*Text).Text
	if nextText[0] == ' ' || isPunctuation(rune(nextText[0])) {
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
		t := te.Value.(*Text).Text
		return isPunctuation(rune(t[len(t)-1]))
	} else {
		te := d.textElement.Prev()
		if te == nil {
			return false
		}
		t := te.Value.(*Text).Text
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

type Image struct {
	Link    string
	inlines []*Text
	Alt     string
	Title   string
}
