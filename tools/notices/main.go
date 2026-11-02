package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

var targets = [][2]string{{"darwin", "arm64"}, {"linux", "amd64"}, {"linux", "arm64"}}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "notices:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	if len(args) != 1 || (args[0] != "generate" && args[0] != "check") {
		return fmt.Errorf("usage: notices generate|check")
	}
	if _, err := os.Stat("go.mod"); err != nil {
		return fmt.Errorf("run from the repository root: %w", err)
	}
	want, err := generate()
	if err != nil {
		return err
	}
	if args[0] == "generate" {
		if err := os.WriteFile(noticesFile, []byte(want), 0o644); err != nil {
			return err
		}
		fmt.Fprintln(stdout, "wrote", noticesFile)
		return nil
	}
	have, err := os.ReadFile(noticesFile)
	if err != nil {
		return fmt.Errorf("%s is missing, run just notices: %w", noticesFile, err)
	}
	if string(have) != want {
		return fmt.Errorf("%s is out of date, run just notices", noticesFile)
	}
	fmt.Fprintln(stdout, noticesFile, "is up to date")
	return nil
}

func generate() (string, error) {
	var lists [][]moduleRef
	for _, target := range targets {
		refs, err := listModules(target[0], target[1])
		if err != nil {
			return "", err
		}
		lists = append(lists, refs)
	}
	var modules []module
	for _, ref := range union(lists...) {
		files, err := readLicenseFiles(ref)
		if err != nil {
			return "", err
		}
		modules = append(modules, module{Path: ref.Path, Version: ref.Version, Files: files})
	}
	return render(modules), nil
}

func listModules(goos, goarch string) ([]moduleRef, error) {
	command := exec.Command("go", "list", "-deps", "-f", "{{with .Module}}{{if not .Main}}{{.Path}} {{.Version}} {{.Dir}}{{end}}{{end}}", "./binary")
	command.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS="+goos, "GOARCH="+goarch)
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("list modules for %s/%s: %w", goos, goarch, err)
	}
	return parseModules(string(output))
}

func readLicenseFiles(ref moduleRef) ([]licenseFile, error) {
	entries, err := os.ReadDir(ref.Dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	var files []licenseFile
	for _, name := range licenseNames(names) {
		text, err := os.ReadFile(filepath.Join(ref.Dir, name))
		if err != nil {
			return nil, err
		}
		files = append(files, licenseFile{Name: name, Text: string(text)})
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("%s %s ships no license file", ref.Path, ref.Version)
	}
	return files, nil
}
