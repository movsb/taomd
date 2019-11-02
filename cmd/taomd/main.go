package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

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

	i := 1

	switch os.Args[i] {
	default:
		panic("unknown command")
	case "test":
		var (
			loop    = true
			all     = false
			section = false
			number  = false
			save    = false
			compare = false
		)

		for i++; i < len(os.Args) && loop; {
			switch os.Args[i] {
			default:
				if strings.HasPrefix(os.Args[i], "-") {
					panic("unknown arguments: " + os.Args[i])
				}
				loop = false
			case "--all", "-a":
				i++
				all = true
			case "--sections", "-s":
				i++
				section = true
			case "--numbers", "-n":
				i++
				number = true
			case "--save":
				i++
				save = true
			case "--compare":
				i++
				compare = true
			}
		}

		switch {
		case all:
			testAll(examples, compare, save)
		case section:
			testSections(examples, os.Args[i:]...)
		case number:
			numbers := []int{}
			for _, arg := range os.Args[i:] {
				n, err := strconv.Atoi(arg)
				if err != nil {
					panic(err)
				}
				numbers = append(numbers, n)
			}
			testNumbers(examples, numbers...)
		}
	}
}
