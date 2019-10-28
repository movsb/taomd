# Entities Generator

This tools generate entities that are treated as HTML entities.

> Although HTML5 does accept some entity references without a trailing semicolon (such as &copy), these are not recognized here, because it makes the grammar too ambiguous.

Entities without a trailing semicolon are not recognized by Markdown.

Also, for file size consideration, the leading `&` and trailing `;` characters are removed.

## Build

Execute the following at taomd project root:

```bash
$ go run tools/entities/entities.go
```

## Source

https://html.spec.whatwg.org/entities.json
