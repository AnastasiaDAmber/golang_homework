package main

import (
	"fmt"
	"log"

	hw02unpackstring "github.com/AnastasiaDAmber/golang_homework/hw02_unpack_string"
)

func main() {
	tests := []string{
		"a4bc2d5e",
		"abcd",
		"45",
		"",
		"aaa0b",
		"🙃0",
		"aaф0b",
	}

	for _, input := range tests {
		result, err := hw02unpackstring.Unpack(input)
		if err != nil {
			log.Printf("input=%q → error: %v\n", input, err)
			continue
		}
		fmt.Printf("input=%q → %q\n", input, result)
	}
}
