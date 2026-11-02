package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

const noticesFile = "THIRD_PARTY_NOTICES.md"

var licenseFileName = regexp.MustCompile(`(?i)^(licen[sc]e|copying|notice)`)

type moduleRef struct {
	Path    string
	Version string
	Dir     string
}

type licenseFile struct {
	Name string
	Text string
}

type module struct {
	Path    string
	Version string
	Files   []licenseFile
}

func parseModules(listing string) ([]moduleRef, error) {
	var refs []moduleRef
	for _, line := range strings.Split(strings.TrimSpace(listing), "\n") {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 3 {
			return nil, fmt.Errorf("unexpected module line %q", line)
		}
		refs = append(refs, moduleRef{Path: fields[0], Version: fields[1], Dir: fields[2]})
	}
	return refs, nil
}

func union(lists ...[]moduleRef) []moduleRef {
	seen := map[string]moduleRef{}
	for _, list := range lists {
		for _, ref := range list {
			seen[ref.Path+" "+ref.Version] = ref
		}
	}
	merged := make([]moduleRef, 0, len(seen))
	for _, ref := range seen {
		merged = append(merged, ref)
	}
	sort.Slice(merged, func(i, j int) bool {
		if merged[i].Path != merged[j].Path {
			return merged[i].Path < merged[j].Path
		}
		return merged[i].Version < merged[j].Version
	})
	return merged
}

func licenseNames(entries []string) []string {
	var names []string
	for _, entry := range entries {
		if licenseFileName.MatchString(entry) {
			names = append(names, entry)
		}
	}
	sort.Strings(names)
	return names
}

func render(modules []module) string {
	var out strings.Builder
	out.WriteString("# Third-party notices\n\n")
	out.WriteString("onboard-kit is built from the Go modules below. Each one is compiled into the binary for darwin/arm64, linux/amd64 or linux/arm64, and each is used under its own license. The license text follows each module.\n\n")
	out.WriteString("`just notices` generates this file, and `just check-notices` fails when it is out of date. Do not edit it by hand.\n")
	for _, m := range modules {
		fmt.Fprintf(&out, "\n## %s %s\n", m.Path, m.Version)
		for _, file := range m.Files {
			fmt.Fprintf(&out, "\n### %s\n\n```text\n%s\n```\n", file.Name, strings.TrimSpace(file.Text))
		}
	}
	return out.String()
}
