package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "records:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	verbose := len(args) == 2 && args[1] == "-v"
	if len(args) == 0 || args[0] != "check" || len(args) > 2 || (len(args) == 2 && !verbose) {
		return fmt.Errorf("usage: records check [-v]")
	}
	if _, err := os.Stat("go.mod"); err != nil {
		return fmt.Errorf("run from the repository root: %w", err)
	}
	result, err := check(".")
	if err != nil {
		return err
	}
	if verbose {
		for _, note := range result.notes {
			fmt.Fprintln(stdout, "note:", note)
		}
	} else if len(result.notes) > 0 {
		fmt.Fprintf(stdout, "records that list a component whose contract does not cite them: %d, see records check -v\n", len(result.notes))
	}
	if len(result.problems) > 0 {
		return fmt.Errorf("found problems:\n  %s", strings.Join(result.problems, "\n  "))
	}
	fmt.Fprintln(stdout, "decision records are consistent")
	return nil
}
