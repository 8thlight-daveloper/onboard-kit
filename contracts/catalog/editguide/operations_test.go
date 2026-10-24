package editguide_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

func TestAddedInstanceUsesIndependentFinalIDHashAndCompleteNewGuidance(t *testing.T) {
	root, target := setup(t, 0755)
	input := "add\nrepository-access\nrepository-new\n1\nrepository_scope\ntext\nIndependent new repository\nretire\nresource-old\nreview\nsave\n"
	var out, stderr bytes.Buffer
	if newCommand().Run([]string{"--plan=" + target}, strings.NewReader(input), &out, &stderr) != 0 {
		t.Fatalf("addition/retirement failed: %q", stderr.String())
	}
	var want map[string]any
	if err := yaml.Unmarshal(fixture(t, "original.yaml"), &want); err != nil {
		t.Fatal(err)
	}
	sections := want["sections"].([]any)
	sections[0].(map[string]any)["enabled"] = false
	canonical := bytes.TrimSuffix(fixture(t, "repository-new.canonical"), []byte("\n"))
	sum := sha256.Sum256(canonical)
	var added map[string]any
	if err := json.Unmarshal(canonical, &added); err != nil {
		t.Fatal(err)
	}
	added["body"] = added["body"].(string) + "\n"
	for _, raw := range added["variables"].([]any) {
		delete(raw.(map[string]any), "type")
	}
	added["source"] = map[string]any{"template": "repository-access", "hash": hex.EncodeToString(sum[:])}
	want["sections"] = append(sections, added)
	want["answers"].(map[string]any)["sections"].(map[string]any)["repository-new"] = map[string]any{"repository_scope": "Independent new repository"}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	var actual map[string]any
	if err := yaml.Unmarshal(data, &actual); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(actual, want) {
		t.Errorf("whole saved plan changed old metadata/facts/baseline or new identity: %#v", actual)
	}
	previous := -1
	for _, guide := range []string{"invitation-prepare", "resource-access-arrange"} {
		block := "Guide: \"" + guide + "\"\nInstance reasons: 1\n  - \"repository-new\"\n" + document("Advisory instructions", fixture(t, guide+".md"))
		index := strings.Index(out.String(), block)
		if index < 0 || index <= previous {
			t.Errorf("literal advice/reason/order missing: %s", guide)
		}
		previous = index
	}
	if !strings.Contains(out.String(), "Preparation guidance for newly added instances only: 2\n") {
		t.Error("new-only guidance boundary missing")
	}
	personalPreserved(t, root)
}

func TestReactivationAndRepairCanBeStagedWithoutInvalidPrefixRejection(t *testing.T) {
	root, target := setup(t, 0755)
	if err := os.WriteFile(target, fixture(t, "reactivation-original.yaml"), 0600); err != nil {
		t.Fatal(err)
	}
	input := "reactivate\nresource-old\nsubmit\nresource-old\nresource_name\ntext\nExplicit same resource\nreview\nsave\n"
	var out, stderr bytes.Buffer
	if newCommand().Run([]string{"--plan=" + target}, strings.NewReader(input), &out, &stderr) != 0 {
		t.Fatalf("ordered whole batch refused: %q", stderr.String())
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	var actual, want map[string]any
	if yaml.Unmarshal(data, &actual) != nil || yaml.Unmarshal(fixture(t, "reactivation-original.yaml"), &want) != nil {
		t.Fatal("literal fixture YAML")
	}
	want["answers"].(map[string]any)["sections"].(map[string]any)["resource-old"].(map[string]any)["resource_name"] = "Explicit same resource"
	delete(want["sections"].([]any)[0].(map[string]any), "enabled")
	if !reflect.DeepEqual(actual, want) {
		t.Error("sequential same-identity edit changed unrelated data")
	}
	personalPreserved(t, root)
}

func TestNoReadAheadInputFramingFailuresPreserveOriginalAndPersonalFiles(t *testing.T) {
	for _, data := range [][]byte{[]byte("review\nsave\xff\n"), []byte(strings.Repeat("x", 65537)), []byte(strings.Repeat(strings.Repeat("x", 65535)+"\n", 16) + "review\nsave\n")} {
		root, target := setup(t, 0755)
		input := &countedInput{source: data}
		var out, stderr bytes.Buffer
		if newCommand().Run([]string{"--plan=" + target}, input, &out, &stderr) != 1 {
			t.Error("invalid/bounded input accepted")
		}
		report(t, stderr.Bytes(), "refused", "edit.input", "Use complete valid input within the displayed limits.")
		if input.position > 1048576 {
			t.Error("read beyond total input budget")
		}
		unchanged(t, target)
		personalPreserved(t, root)
	}
}

type failReport struct {
	calls    int
	accepted bytes.Buffer
}

func (w *failReport) Write(p []byte) (int, error) {
	w.calls++
	if w.calls == 1 {
		return 0, io.ErrClosedPipe
	}
	return w.accepted.Write(p)
}

func TestReportFailureAfterReplacementKeepsSavedPlanAndPersonalFiles(t *testing.T) {
	root, target := setup(t, 0755)
	var out bytes.Buffer
	var stderr failReport
	if newCommand().Run([]string{"--plan=" + target}, bytes.NewReader(fixture(t, "retire.txt")), &out, &stderr) != 1 {
		t.Error("failed report did not return failure")
	}
	report(t, stderr.accepted.Bytes(), "written", "edit.report-written", "The plan edits were saved, but reporting failed.")
	data, err := os.ReadFile(target)
	if err != nil || !bytes.Equal(data, fixture(t, "retired.yaml")) {
		t.Error("successful replacement was rolled back after reporting failure")
	}
	personalPreserved(t, root)
}

func TestDecoderRefusalDoesNotExposePrivateUnknownKeyOrReadConsent(t *testing.T) {
	root, target := setup(t, 0755)
	bad := []byte("schemaVersion: 1\nid: held-plan\nprivate-rejected-key: first\nprivate-rejected-key: second\n")
	if err := os.WriteFile(target, bad, 0600); err != nil {
		t.Fatal(err)
	}
	input := &countedInput{source: []byte("review\nsave\n")}
	var out, stderr bytes.Buffer
	if newCommand().Run([]string{"--plan=" + target}, input, &out, &stderr) != 1 || input.position != 0 {
		t.Error("decoder failure crossed input admission")
	}
	want := "Problems: 1\n  - Stage: \"decode\" Path: \"plan\" Code: \"format.duplicate-key\" Severity: \"error\" Repair: \"Repair the selected plan's format; duplicate keys and invalid spellings must be fixed explicitly.\"\n"
	if out.String() != want || strings.Contains(stderr.String(), "private-rejected-key") {
		t.Errorf("private raw decoder path was echoed or safe repair suppressed: %q %q", out.String(), stderr.String())
	}
	report(t, stderr.Bytes(), "refused", "edit.data", "Repair the displayed plan data before editing.")
	data, err := os.ReadFile(target)
	if err != nil || !bytes.Equal(data, bad) {
		t.Error("invalid input was silently repaired")
	}
	personalPreserved(t, root)
}
