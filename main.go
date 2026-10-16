package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "Usage: go run . <input.txt> <output.txt>")
		os.Exit(1)
	}

	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erreur de lecture :", err)
		os.Exit(1)
	}

	result := Process(string(data))

	if err := os.WriteFile(os.Args[2], []byte(result), 0644); err != nil {
		fmt.Fprintln(os.Stderr, "Erreur d'écriture :", err)
		os.Exit(1)
	}
}
