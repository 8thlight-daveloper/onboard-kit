package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "architecture:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	if len(args) != 1 || args[0] != "check" {
		return fmt.Errorf("usage: architecture check")
	}
	if _, err := os.Stat("go.mod"); err != nil {
		return fmt.Errorf("run from the repository root: %w", err)
	}
	rules, err := loadRules("architecture.yaml")
	if err != nil {
		return err
	}
	graph, err := readGraph(".")
	if err != nil {
		return err
	}
	outcome := evaluate(rules, graph, contractFolderExists)
	if len(outcome.undeclared) > 0 {
		fmt.Fprintf(stdout, "components with no declared imports, held to their kind only: %s\n", strings.Join(outcome.undeclared, ", "))
	}
	if len(outcome.problems) > 0 {
		return fmt.Errorf("found problems:\n  %s", strings.Join(outcome.problems, "\n  "))
	}
	fmt.Fprintln(stdout, "architecture rules hold")
	return nil
}

func contractFolderExists(id string) bool {
	info, err := os.Stat("contracts/" + id)
	return err == nil && info.IsDir()
}
