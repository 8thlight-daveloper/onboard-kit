package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "usage: formula CHECKSUMS_FILE TAG OWNER/REPOSITORY")
		os.Exit(2)
	}
	checksums, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "formula:", err)
		os.Exit(1)
	}
	text, err := formula(os.Args[2], os.Args[3], string(checksums))
	if err != nil {
		fmt.Fprintln(os.Stderr, "formula:", err)
		os.Exit(1)
	}
	fmt.Print(text)
}
