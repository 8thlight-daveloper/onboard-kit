package start

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"onboard-kit/internal/command/planfile"
	"onboard-kit/internal/model"
)

func writeOrganizationPlan(t *testing.T, path string) {
	t.Helper()
	plan := model.Plan{
		SchemaVersion: 1,
		ID:            "0d6c1b9e-5c0e-4b57-9a60-1d3a3f6f0a11",
		Answers:       model.Answers{Plan: map[string]model.Value{"organization": {Kind: model.ValueKindText, Text: "Globex"}}},
	}
	if !planfile.Save(path, plan, &bytes.Buffer{}) {
		t.Fatal("could not write the organization's plan")
	}
}

func TestAnUnreachableOrganizationPlanFallsBackToThePersonalCopy(t *testing.T) {
	folder := t.TempDir()
	organization := filepath.Join(folder, "org.obk.yaml")
	home := filepath.Join(folder, "home")
	writeOrganizationPlan(t, organization)
	var stdout, stderr bytes.Buffer
	args := []string{"--org-plan", organization, "--home", home, "--no-walk", "--refreshed-on", "2026-10-16"}
	if code := Run(args, strings.NewReader(""), &stdout, &stderr); code != exitSuccess {
		t.Fatalf("first run exited %d: %s", code, stderr.String())
	}
	if err := os.Remove(organization); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	if code := Run(args, strings.NewReader(""), &stdout, &stderr); code != exitSuccess {
		t.Fatalf("second run exited %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "cannot be reached. Using your copy as of 2026-10-16") {
		t.Fatalf("the copy's date is not said: %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "globex.obk.yaml") {
		t.Fatalf("the personal plan is not named: %q", stdout.String())
	}
}

func TestAnUnreachableOrganizationPlanWithNoCopyIsAProblem(t *testing.T) {
	folder := t.TempDir()
	var stdout, stderr bytes.Buffer
	args := []string{"--org-plan", filepath.Join(folder, "missing.obk.yaml"), "--home", filepath.Join(folder, "home"), "--no-walk"}
	if code := Run(args, strings.NewReader(""), &stdout, &stderr); code != exitProblem {
		t.Fatalf("exited %d, want a problem", code)
	}
	if !strings.Contains(stderr.String(), "no earlier copy") {
		t.Fatalf("the reason is not said: %q", stderr.String())
	}
}
