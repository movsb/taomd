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

	for k := range loadTestResults() {
		t := m[k]
		if render(parse(t.Markdown)) != t.HTML {
			panic("break: " + k)
		}
	}

	example := "13"

	if example == "" {
		example = os.Args[1]
	}

	t := m[example]
	if render(parse(t.Markdown)) != t.HTML {
		fmt.Println(t.Markdown, "!=", t.HTML)
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
	fp, err := os.OpenFile("result.txt", os.O_RDWR, 0)
	if err != nil {
		panic(err)
	}
	defer fp.Close()
	fp.WriteString(example + "\n")
}
