package main

func render(doc *Document) string {
	var s string

	for _, block := range doc.blocks {
		switch typed := block.(type) {
		default:
			panic("unhandled block")
		case *HorizontalRule:
			_ = typed
			s += "<hr />\n"

		}
	}

	return s
}
