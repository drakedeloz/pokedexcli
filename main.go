package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println("Hello, World!")
}

func cleanInput(text string) []string {
	parsedText := strings.TrimSpace(text)
	parsedText = strings.ToLower(parsedText)
	splitText := strings.Fields(parsedText)

	return splitText
}
