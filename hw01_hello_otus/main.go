package main

import (
	"fmt"

	"golang.org/x/example/hello/reverse"
)

func main() {
	s := "Hello, OTUS!"
	fmt.Println(reverse.String(s))
}
