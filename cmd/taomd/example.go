package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
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

func testAll(examples []*Example, compare bool, save bool) {
	var passedArray []*Example
	var failedArray []*Example

	testFunc(examples,
		func(example *Example) bool {
			return true
		},
		func(example *Example, html string, pass bool) {
			if pass {
				passedArray = append(passedArray, example)
			} else {
				failedArray = append(failedArray, example)
			}
			dumpResult(example, html, pass)
		},
	)

	if compare {
		oldPassed := func() map[int]int {
			fp, err := os.Open("pass.txt")
			if err != nil {
				return map[int]int{}
			}
			defer fp.Close()
			m := make(map[int]int)
			scn := bufio.NewScanner(fp)
			for scn.Scan() {
				n, err := strconv.Atoi(scn.Text())
				if err != nil {
					panic(err)
				}
				m[n] = 1
			}
			return m
		}()

		fmt.Fprintf(os.Stderr, "\nPassed:\n\n")
		n := 0
		for _, k := range passedArray {
			if _, ok := oldPassed[k.Example]; !ok {
				fmt.Fprintf(os.Stderr, "    %d\n", k.Example)
				n++
			}
		}
		if n == 0 {
			fmt.Fprintf(os.Stderr, "    None\n")
		}
		n = 0
		fmt.Fprintf(os.Stderr, "\nBroken:\n\n")
		for _, k := range failedArray {
			if _, ok := oldPassed[k.Example]; ok {
				fmt.Fprintf(os.Stderr, "    %d\n", k.Example)
				n++
			}
		}
		if n == 0 {
			fmt.Fprintf(os.Stderr, "    None\n")
		}
	}

	if save {
		fp, err := os.Create("pass.txt")
		if err != nil {
			panic(err)
		}
		defer fp.Close()
		for _, k := range passedArray {
			fmt.Fprintf(fp, "%d\n", k.Example)
		}
	}
}

func testNumbers(examples []*Example, numbers ...int) {
	testFunc(examples,
		func(example *Example) bool {
			for _, number := range numbers {
				if number == example.Example {
					return true
				}
			}
			return false
		},
		dumpResult,
	)
}

func testSections(examples []*Example, sections ...string) {
	testFunc(examples,
		func(example *Example) bool {
			for _, section := range sections {
				if section == example.Section {
					return true
				}
			}
			return false
		},
		dumpResult,
	)
}

func testFunc(examples []*Example, predicate func(example *Example) bool, result func(examle *Example, html string, pass bool)) {
	var total, passed, failed, skipped int

	for _, example := range examples {
		total++
		if !predicate(example) {
			skipped++
			continue
		}
		doc := taomd.Parse(strings.NewReader(example.Markdown))
		html := taomd.Render(doc)
		ok := html == example.HTML
		if ok {
			passed++
		} else {
			failed++
		}
		if result != nil {
			result(example, html, ok)
		}
	}

	if total != skipped {
		fmt.Fprintf(os.Stderr, "\n")
	}
	fmt.Fprintf(os.Stderr,
		"total: %d, passed: %d, failed: %d, skipped: %d\n",
		total, passed, failed, skipped,
	)
}

func dumpResult(example *Example, html string, pass bool) {
	if pass {
		fmt.Fprintf(os.Stdout, "pass: %d\n", example.Example)
	} else {
		fmt.Fprintf(os.Stderr, "fail: %d\n", example.Example)
		taomd.DumpFail(os.Stderr, example.Markdown, example.HTML, html)
	}
}
