package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type Test struct {
	Markdown string `json:"markdown"`
	HTML     string `json:"html"`
	Example  int    `json:"example"`
}

var gdoc *Document

var markdown = ``

func main() {
	example := ""

	if example == "" && len(os.Args) == 1 {
		var fp io.Reader = os.Stdin
		if markdown != "" {
			fp = strings.NewReader(markdown)
		}
		gdoc = parse(fp, -1)
		h := render(gdoc)
		fmt.Print(h)
		return
	}

	if example != "" && len(os.Args) > 1 {
		panic("dup args")
	}

	fp, err := os.Open("spec.json")
	if err != nil {
		panic(err)
	}
	defer fp.Close()
	var tests []*Test
	if err := json.NewDecoder(fp).Decode(&tests); err != nil {
		panic(err)
	}
	m := map[string]*Test{}
	for _, t := range tests {
		m[fmt.Sprint(t.Example)] = t
	}

	if example == "" && false {
		for k := range loadTestResults() {
			t := m[k]
			// fmt.Println("retest:", t.Example)
			if h := render(parse(strings.NewReader(t.Markdown), t.Example)); h != t.HTML {
				fmt.Printf("break: %d\n", t.Example)
				dumpFail(t.Markdown, t.HTML, h)
			}
		}
	}

	if example == "" {
		if len(os.Args) > 1 {
			example = os.Args[1]
		} else {
		}
	}

	t := m[example]
	gdoc = parse(strings.NewReader(t.Markdown), t.Example)
	if h := render(gdoc); h != t.HTML {
		dumpFail(t.Markdown, t.HTML, h)
		os.Exit(1)
	} else {
		saveTestResults(example)
		fmt.Println("pass")
		os.Exit(0)
	}
}

func loadTestResults() map[string]bool {
	fp, err := os.Open("result.txt")
	if err != nil {
		panic(err)
	}
	defer fp.Close()
	m := map[string]bool{}
	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		m[scanner.Text()] = true
	}
	return m
}

func saveTestResults(example string) {
	fp, err := os.OpenFile("result.txt", os.O_RDWR|os.O_APPEND, 0)
	if err != nil {
		panic(err)
	}
	defer fp.Close()
	fp.WriteString(example + "\n")
}
