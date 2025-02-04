package main

import (
	"fmt"
	"os"

	"tfau/lib/hcl"
	"tfau/lib/module"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <terraform_file.tf>")
		return
	}

	filename := os.Args[1]
	if err := hcl.UpdateVersionInFile(filename); err != nil {
		fmt.Println("Error:", err)
	}
}
