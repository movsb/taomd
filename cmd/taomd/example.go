package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/movsb/taomd"
)

// An Example from spec.json.
type Example struct {
	Markdown string `json:"markdown"`
	HTML     string `json:"html"`
	Example  int    `json:"example"`
	Section  string `json:"section"`
}

// loadExamples load all examples from spec.json.
func loadExamples(path string) []*Example {
	fp, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer fp.Close()

	var examples []*Example
	if err := json.NewDecoder(fp).Decode(&examples); err != nil {
		panic(err)
	}

	return examples
}

func testNumbers(examples []*Example, numbers ...int) {
	testFunc(examples, func(example *Example) bool {
		for _, number := range numbers {
			if number == example.Example {
				return true
			}
		}
		return false
	})
}

func testSections(examples []*Example, sections ...string) {
	testFunc(examples, func(example *Example) bool {
		for _, section := range sections {
			if section == example.Section {
				return true
			}
		}
		return false
	})
}

func testFunc(examples []*Example, predicate func(example *Example) bool) {
	var total, passed, failed, skipped int

	for _, example := range examples {
		total++
		if !predicate(example) {
			skipped++
			continue
		}
		doc := taomd.Parse(strings.NewReader(example.Markdown))
		html := taomd.Render(doc)
		if html == example.HTML {
			passed++
			fmt.Fprintf(os.Stdout, "pass: %d\n", example.Example)
		} else {
			failed++
			fmt.Fprintf(os.Stderr, "fail: %d\n", example.Example)
			taomd.DumpFail(os.Stderr, example.Markdown, example.HTML, html)
		}
	}

	fmt.Fprintf(os.Stderr,
		"\ntotal: %d, passed: %d, failed: %d, skipped: %d\n",
		total, passed, failed, skipped,
	)
}
