package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sort"
	"strings"
)

type packageImports struct {
	components []string
	categories map[string][]string
}

var categoryPackages = map[string][]string{
	"filesystem": {"os", "io/fs", "io/ioutil", "path/filepath"},
	"network":    {"net", "net/http", "net/url"},
	"process":    {"os/exec", "os/signal", "syscall", "golang.org/x/sys/unix"},
	"terminal":   {"golang.org/x/term", "github.com/charmbracelet/bubbletea", "github.com/charmbracelet/lipgloss", "github.com/charmbracelet/bubbles/textinput"},
	"randomness": {"math/rand", "crypto/rand"},
}

func readGraph(root string) (map[string]packageImports, error) {
	module, err := goOutput(root, "list", "-m")
	if err != nil {
		return nil, err
	}
	module = strings.TrimSpace(module)
	listing, err := goOutput(root, "list", "-json", "./internal/...", "./binary")
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(strings.NewReader(listing))
	graph := map[string]packageImports{}
	for {
		var listed struct {
			ImportPath string
			Imports    []string
		}
		if err := decoder.Decode(&listed); err == io.EOF {
			return graph, nil
		} else if err != nil {
			return nil, err
		}
		id, ok := componentOf(module, listed.ImportPath)
		if !ok {
			continue
		}
		graph[id] = importsOf(module, listed.Imports)
	}
}

func goOutput(root string, args ...string) (string, error) {
	command := exec.Command("go", args...)
	command.Dir = root
	var stderr bytes.Buffer
	command.Stderr = &stderr
	out, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("go %s: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return string(out), nil
}

func componentOf(module string, importPath string) (string, bool) {
	if importPath == module+"/binary" {
		return "binary", true
	}
	return strings.CutPrefix(importPath, module+"/internal/")
}

func importsOf(module string, imports []string) packageImports {
	found := packageImports{categories: map[string][]string{}}
	for _, imported := range imports {
		if id, ok := componentOf(module, imported); ok {
			found.components = append(found.components, id)
			continue
		}
		for category, packages := range categoryPackages {
			for _, candidate := range packages {
				if imported == candidate {
					found.categories[category] = append(found.categories[category], imported)
				}
			}
		}
	}
	sort.Strings(found.components)
	return found
}
