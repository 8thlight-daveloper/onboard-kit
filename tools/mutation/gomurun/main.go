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

	gomuGraceSeconds = 10

	reportFile  = "mutation-report.json"
	historyFile = ".gomu_history.json"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	options := flag.NewFlagSet("gomurun", flag.ContinueOnError)
	workers := options.Int("workers", 1, "mutants tested at once")
	timeout := options.Int("timeout", 30, "seconds a test run may take for one mutant")
	gomu := options.String("gomu", ".direnv/bin/gomu", "path of the gomu program")
	short := options.Bool("short", false, "run the tests with -short, which skips the slow ones")
	cached := options.Bool("cached", false, "reuse the results of source files whose text and contract suite have not changed")
	if options.Parse(args) != nil || options.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: gomurun [--workers N] [--timeout S] [--short] [--cached] <noun>/<role>")
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
	defer os.Remove(filepath.Join(root, reportFile))
	defer os.Remove(filepath.Join(root, historyFile))
	var results []result
	if *cached {
		results, err = cachedResults(root, component, func(target string) error {
			return runGomu(*gomu, target, component, *workers, *timeout, *short)
		}, *short)
	} else {
		err = runGomu(*gomu, "./internal/"+component, component, *workers, *timeout, *short)
		if err == nil {
			results, err = freshResults(root)
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitProblem
	}
	return summarize(root, component, results)
}

func freshResults(root string) ([]result, error) {
	parsed, err := readReport(filepath.Join(root, reportFile))
	if err != nil {
		return nil, err
	}
	return parsed.Results, nil
}

func cachedResults(root string, component string, mutate func(target string) error, short bool) ([]result, error) {
	names, err := sourceFiles(root, component)
	if err != nil {
		return nil, err
	}
	suite, err := suiteDigest(root, component)
	if err != nil {
		return nil, err
	}
	var all []result
	for _, name := range names {
		source, err := os.ReadFile(filepath.Join(root, "internal", component, name))
		if err != nil {
			return nil, err
		}
		path := cachePath(root, component, name, cacheKey(source, suite, short))
		if results, ok := readCached(path); ok {
			fmt.Fprintf(os.Stderr, "%s: reused %d results\n", name, len(results))
			all = append(all, results...)
			continue
		}
		fmt.Fprintf(os.Stderr, "%s: mutating\n", name)
		if err := mutate("./internal/" + component + "/" + name); err != nil {
			return nil, err
		}
		results, err := freshResults(root)
		if err != nil {
			return nil, err
		}
		os.Remove(filepath.Join(root, reportFile))
		if err := writeCached(path, results); err != nil {
			return nil, err
		}
		forgetStale(root, component, name, path)
		all = append(all, results...)
	}
	return all, nil
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

func gomuArguments(gomu string, target string, component string, workers int, timeout int, short bool) []string {
	testFlags := []string{"-count=1", "-failfast", fmt.Sprintf("-timeout=%ds", timeout)}
	if short {
		testFlags = append(testFlags, "-short")
	}
	arguments := []string{
		"-n", "10", gomu, "run", target,
		"--incremental=false", "--threshold", "0", "--output", "json",
		"--workers", fmt.Sprint(workers), "--timeout", fmt.Sprint(timeout + gomuGraceSeconds),
		"--",
	}
	arguments = append(arguments, testFlags...)
	return append(arguments, contractPathFrom(component))
}

func runGomu(gomu string, target string, component string, workers int, timeout int, short bool) error {
	command := exec.Command("nice", gomuArguments(gomu, target, component, workers, timeout, short)...)
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

func summarize(root string, component string, results []result) int {
	survivors, err := survivorsOf(results, root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitProblem
	}
	accepted, err := readAccepted(filepath.Join(root, "tools", "mutation", "accepted", component+".txt"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitProblem
	}
	counts, unreviewed := tallyOf(results, survivors, accepted)
	printSurvivors(os.Stdout, unreviewed)
	printScore(os.Stdout, counts)
	return exitSuccess
}
