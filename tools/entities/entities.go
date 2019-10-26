package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
)

type _Entity struct {
	Codepoints []rune `json:"codepoints"`
}

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
	sort.Strings(sortedKeys)

	s1 := "var htmlEntities1 = map[string]rune {"
	s2 := "var htmlEntities2 = map[string][2]rune {"

	for _, key := range sortedKeys {
		if key[len(key)-1] != ';' {
			continue
		}
		entity := entities[key]
		quotedName := "`" + key[1:len(key)-1] + "`:"
		codepoints := entity.Codepoints
		switch len(codepoints) {
		case 1:
			s1 += fmt.Sprintf("%s%d,", quotedName, codepoints[0])
		case 2:
			s2 += fmt.Sprintf("%s{%d,%d},", quotedName, codepoints[0], codepoints[1])
		}
	}

	s1 += "}"
	s2 += "}"

	fmt.Printf("package main\n\n")
	fmt.Println(s1)
	fmt.Println(s2)
	fmt.Println()
}
