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
