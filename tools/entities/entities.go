package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

type _Entity struct {
	Codepoints []rune `json:"codepoints"`
}

// LengthString : shorter strings first.
type LengthString []string

func (s LengthString) Len() int { return len(s) }
func (s LengthString) Less(i, j int) bool {
	n := len(s[i]) - len(s[j])
	switch {
	case n < 0:
		return true
	case n > 0:
		return false
	default:
		// less efficiency
		return strings.ToLower(s[i]) < strings.ToLower(s[j])
	}
}
func (s LengthString) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func main() {
	resp, err := http.Get("https://html.spec.whatwg.org/entities.json")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var entities map[string]_Entity
	if err := json.NewDecoder(resp.Body).Decode(&entities); err != nil {
		panic(err)
	}

	sortedKeys := make([]string, len(entities))
	i := 0
	for k := range entities {
		sortedKeys[i] = k
		i++
	}

	// sort by length
	sort.Stable(LengthString(sortedKeys))

	s1 := "var htmlEntities1 = map[string]rune {"
	s2 := "var htmlEntities2 = map[string][2]rune {"

	lastKeyLen1 := 0
	lastKeyLen2 := 0

	for _, key := range sortedKeys {
		if key[len(key)-1] != ';' {
			continue
		}

		entity := entities[key]
		codepoints := entity.Codepoints

		key = key[1 : len(key)-1]

		switch len(codepoints) {
		case 1:
			if len(key) != lastKeyLen1 {
				s1 += "\n"
			}
			s1 += fmt.Sprintf("\t`%s`: %d,\n", key, codepoints[0])
			lastKeyLen1 = len(key)
		case 2:
			if len(key) != lastKeyLen2 {
				s2 += "\n"
			}
			s2 += fmt.Sprintf("\t`%s`: {%d, %d, },\n", key, codepoints[0], codepoints[1])
			lastKeyLen2 = len(key)
		}

	}

	s1 += "}"
	s2 += "}"

	fmt.Println(s1)
	fmt.Println()
	fmt.Println(s2)
	fmt.Println()
}
