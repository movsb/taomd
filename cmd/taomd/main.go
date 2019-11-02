package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/movsb/taomd"
)

func main() {
	if len(os.Args) == 1 {
		doc := taomd.Parse(os.Stdin)
		html := taomd.Render(doc)
		fmt.Print(html)
		os.Exit(0)
	}

	examples := loadExamples("spec.json")

	switch os.Args[1] {
	default:
		panic("unknown arguments provided.")
	case "--test-all", "-a":
		testFunc(examples, func(*Example) bool { return true })
	case "--test-sections", "-s":
		testSections(examples, os.Args[2:]...)
	case "--test-numbers", "-n":
		numbers := []int{}
		for _, arg := range os.Args[2:] {
			n, err := strconv.Atoi(arg)
			if err != nil {
				panic(err)
			}
			numbers = append(numbers, n)
		}
		testNumbers(examples, numbers...)
	}
}
