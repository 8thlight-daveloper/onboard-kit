package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseModulesReadsOneModulePerLine(t *testing.T) {
	got, err := parseModules("example.test/a v1.0.0 /cache/a\nexample.test/b v2.1.0 /cache/b\n")
	if err != nil {
		t.Fatal(err)
	}
	want := []moduleRef{{"example.test/a", "v1.0.0", "/cache/a"}, {"example.test/b", "v2.1.0", "/cache/b"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestParseModulesRefusesAShortLine(t *testing.T) {
	if _, err := parseModules("example.test/a v1.0.0\n"); err == nil {
		t.Fatal("expected an error")
	}
}

func TestUnionSortsAndKeepsEachModuleOnce(t *testing.T) {
	first := []moduleRef{{"example.test/b", "v1", "/b"}, {"example.test/a", "v1", "/a"}}
	second := []moduleRef{{"example.test/a", "v1", "/a"}, {"example.test/c", "v1", "/c"}}
	got := union(first, second)
	paths := []string{}
	for _, ref := range got {
		paths = append(paths, ref.Path)
	}
	if want := []string{"example.test/a", "example.test/b", "example.test/c"}; !reflect.DeepEqual(paths, want) {
		t.Fatalf("got %v, want %v", paths, want)
	}
}

func TestUnionKeepsTwoVersionsOfOneModule(t *testing.T) {
	got := union([]moduleRef{{"example.test/a", "v2", "/a2"}}, []moduleRef{{"example.test/a", "v1", "/a1"}})
	if len(got) != 2 || got[0].Version != "v1" || got[1].Version != "v2" {
		t.Fatalf("got %v", got)
	}
}

func TestLicenseNamesKeepsLicenseAndNoticeFilesOnly(t *testing.T) {
	got := licenseNames([]string{"README.md", "main.go", "NOTICE", "LICENSE.txt", "license", "COPYING", "licensing_test.go"})
	want := []string{"COPYING", "LICENSE.txt", "NOTICE", "license"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestRenderListsEachModuleWithItsFiles(t *testing.T) {
	got := render([]module{{Path: "example.test/a", Version: "v1.0.0", Files: []licenseFile{{Name: "LICENSE", Text: "MIT text\n"}, {Name: "NOTICE", Text: "notice text"}}}})
	for _, want := range []string{"## example.test/a v1.0.0\n", "### LICENSE\n\n```text\nMIT text\n```\n", "### NOTICE\n\n```text\nnotice text\n```\n"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in\n%s", want, got)
		}
	}
}

func TestRenderIsDeterministic(t *testing.T) {
	modules := []module{{Path: "example.test/a", Version: "v1", Files: []licenseFile{{Name: "LICENSE", Text: "x"}}}}
	if render(modules) != render(modules) {
		t.Fatal("render changed between calls")
	}
}

func TestRunRefusesUnknownCommand(t *testing.T) {
	var out strings.Builder
	for _, args := range [][]string{nil, {"other"}, {"generate", "extra"}} {
		if err := run(args, &out); err == nil {
			t.Fatalf("expected an error for %v", args)
		}
	}
}
