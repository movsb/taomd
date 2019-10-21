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

	fmt.Println("package main")
	fmt.Println()
	fmt.Println("var htmlEntities = map[string]rune {")

	for _, key := range sortedKeys {
		entity := entities[key]
		codepoint := entity.Codepoints[0]
		quotedName := `"` + key + `":`
		fmt.Printf("\t%-20s %d,\n", quotedName, codepoint)
	}

	fmt.Println("}")
}
