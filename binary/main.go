package main

import (
	"os"

	"onboard-kit/internal/command/root"
)

func main() {
	os.Exit(root.Run(os.Args[1:], os.Stdout, os.Stderr))
}
