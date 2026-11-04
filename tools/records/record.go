package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"go.yaml.in/yaml/v3"
)

const decisionsFolder = "docs/decisions"

type frontMatter struct {
	Category           string   `yaml:"category"`
	Status             string   `yaml:"status"`
	Approval           string   `yaml:"approval"`
	Note               string   `yaml:"note"`
	Affects            []string `yaml:"affects"`
	BuildsOn           []string `yaml:"builds-on"`
	Supersedes         []string `yaml:"supersedes"`
	SupersededBy       []string `yaml:"superseded-by"`
	PartlySupersedes   []string `yaml:"partly-supersedes"`
	PartlySupersededBy []string `yaml:"partly-superseded-by"`
}

type record struct {
	number     string
	file       string
	heading    string
	front      frontMatter
	hasAffects bool
}

var recordFile = regexp.MustCompile(`^(\d{4})-.+\.md$`)

func loadRecords(root string) ([]record, []string) {
	entries, err := os.ReadDir(filepath.Join(root, decisionsFolder))
	if err != nil {
		return nil, []string{err.Error()}
	}
	var records []record
	var problems []string
	for _, entry := range entries {
		match := recordFile.FindStringSubmatch(entry.Name())
		if match == nil {
			continue
		}
		text, err := os.ReadFile(filepath.Join(root, decisionsFolder, entry.Name()))
		if err != nil {
			problems = append(problems, err.Error())
			continue
		}
		parsed, problem := parseRecord(match[1], entry.Name(), string(text))
		if problem != "" {
			problems = append(problems, problem)
			continue
		}
		records = append(records, parsed)
	}
	sort.Slice(records, func(i, j int) bool { return records[i].number < records[j].number })
	return records, problems
}

func parseRecord(number string, file string, text string) (record, string) {
	rest, ok := strings.CutPrefix(text, "---\n")
	if !ok {
		return record{}, fmt.Sprintf("%s: no front matter", file)
	}
	block, body, ok := strings.Cut(rest, "\n---\n")
	if !ok {
		return record{}, fmt.Sprintf("%s: front matter is not closed", file)
	}
	var front frontMatter
	decoder := yaml.NewDecoder(bytes.NewReader([]byte(block)))
	decoder.KnownFields(true)
	if err := decoder.Decode(&front); err != nil {
		return record{}, fmt.Sprintf("%s: front matter: %v", file, err)
	}
	var keys map[string]any
	if err := yaml.Unmarshal([]byte(block), &keys); err != nil {
		return record{}, fmt.Sprintf("%s: front matter: %v", file, err)
	}
	_, hasAffects := keys["affects"]
	return record{number: number, file: file, heading: firstHeading(body), front: front, hasAffects: hasAffects}, ""
}

func firstHeading(body string) string {
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "# ") {
			return line
		}
	}
	return ""
}
