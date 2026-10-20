package start

import (
	"path/filepath"
	"testing"

	"onboard-kit/internal/model"
)

func TestSlugsAreLowerCaseWordsJoinedByHyphens(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"ACME Corp":        "acme-corp",
		"  Globex,  Inc. ": "globex-inc",
		"Über 2 Go":        "über-2-go",
		"???":              "",
		"a/b":              "a-b",
	}
	for text, want := range cases {
		if got := slug(text); got != want {
			t.Errorf("slug(%q) = %q, want %q", text, got, want)
		}
	}
}

func TestTheOrganizationNameComesFromASettledTextAnswer(t *testing.T) {
	t.Parallel()
	named := func(value model.Value) model.Plan {
		return model.Plan{Answers: model.Answers{Plan: map[string]model.Value{"organization": value}}}
	}
	cases := []struct {
		name string
		plan model.Plan
		want string
	}{
		{"text", named(model.Value{Kind: model.ValueKindText, Text: "ACME Corp"}), "acme-corp"},
		{"deferred", named(model.Value{Kind: model.ValueKindDeferred}), "org"},
		{"punctuation only", named(model.Value{Kind: model.ValueKindText, Text: "!!"}), "org"},
		{"absent", model.Plan{}, "org"},
	}
	for _, each := range cases {
		if got := organizationName(each.plan); got != each.want {
			t.Errorf("%s: %q, want %q", each.name, got, each.want)
		}
	}
}

func TestTheGivenFolderWinsOverTheEnvironment(t *testing.T) {
	t.Setenv("OBK_HOME", "/from/environment")
	folder, err := stateHome("/given")
	if err != nil || folder != "/given" {
		t.Fatalf("got %q, %v", folder, err)
	}
	folder, err = stateHome("")
	if err != nil || folder != "/from/environment" {
		t.Fatalf("got %q, %v", folder, err)
	}
}

func TestTheDefaultFolderEndsInObk(t *testing.T) {
	t.Setenv("OBK_HOME", "")
	folder, err := stateHome("")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(folder) != "obk" {
		t.Fatalf("got %q", folder)
	}
}
