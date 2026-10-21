package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	exitSuccess = 0
	exitProblem = 1
	exitUsage   = 2

	reportFile  = "mutation-report.json"
	historyFile = ".gomu_history.json"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	options := flag.NewFlagSet("gomurun", flag.ContinueOnError)
	workers := options.Int("workers", 1, "mutants tested at once")
	timeout := options.Int("timeout", 120, "seconds a test run may take for one mutant")
	gomu := options.String("gomu", ".direnv/bin/gomu", "path of the gomu program")
	if options.Parse(args) != nil || options.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: gomurun [--workers N] [--timeout S] <noun>/<role>")
		return exitUsage
	}
	component := options.Arg(0)
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitProblem
	}
	if problem := checkComponent(root, component); problem != nil {
		fmt.Fprintln(os.Stderr, problem)
		return exitUsage
	}
	if err := runGomu(*gomu, component, *workers, *timeout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitProblem
	}
	defer os.Remove(filepath.Join(root, reportFile))
	defer os.Remove(filepath.Join(root, historyFile))
	return summarize(root, component)
}

func checkComponent(root string, component string) error {
	for _, folder := range []string{"internal", "contracts"} {
		if info, err := os.Stat(filepath.Join(root, folder, component)); err != nil || !info.IsDir() {
			return fmt.Errorf("%s/%s is not a folder", folder, component)
		}
	}
	return nil
}

func contractPathFrom(component string) string {
	depth := len(strings.Split(component, "/")) + 1
	return strings.Repeat("../", depth) + "contracts/" + component + "/"
}

func runGomu(gomu string, component string, workers int, timeout int) error {
	arguments := []string{
		"-n", "10", gomu, "run", "./internal/" + component,
		"--incremental=false", "--threshold", "0", "--output", "json",
		"--workers", fmt.Sprint(workers), "--timeout", fmt.Sprint(timeout),
		"--", "-count=1", contractPathFrom(component),
	}
	command := exec.Command("nice", arguments...)
	command.Env = append(os.Environ(), "GOFLAGS="+limitedGoFlags(os.Getenv("GOFLAGS")))
	command.Stdout = os.Stderr
	command.Stderr = os.Stderr
	return command.Run()
}

func limitedGoFlags(existing string) string {
	if strings.Contains(existing, "-p=") {
		return existing
	}
	return strings.TrimSpace(existing + " -p=2")
}

func summarize(root string, component string) int {
	parsed, err := readReport(filepath.Join(root, reportFile))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitProblem
	}
	survivors, err := survivorsOf(parsed.Results, root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitProblem
	}
	accepted, err := readAccepted(filepath.Join(root, "tools", "mutation", "accepted", component+".txt"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitProblem
	}
	counts, unreviewed := tallyOf(parsed.Results, survivors, accepted)
	printSurvivors(os.Stdout, unreviewed)
	printScore(os.Stdout, counts)
	return exitSuccess
}
