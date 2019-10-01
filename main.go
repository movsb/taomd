package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

type Test struct {
	Markdown string `json:"markdown"`
	HTML     string `json:"html"`
	Example  int    `json:"example"`
}

func main() {
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

	example := "15"

	if example == "" {
		for k := range loadTestResults() {
			t := m[k]
			// fmt.Println("retest:", t.Example)
			if h := render(parse(t.Markdown, t.Example)); h != t.HTML {
				fmt.Printf("break: %d\n", t.Example)
				dumpFail(t.Markdown, t.HTML, h)
			}
		}
	}

	if example == "" {
		example = os.Args[1]
	}

	t := m[example]
	doc := parse(t.Markdown, t.Example)
	if h := render(doc); h != t.HTML {
		dumpFail(t.Markdown, t.HTML, h)
	} else {
		saveTestResults(example)
		fmt.Println("pass")
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
