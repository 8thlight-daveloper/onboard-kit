package runner_test

import (
	"strings"
	"testing"

	"onboard-kit/acceptance/runner"
)

func TestBenignLegacyTextKeepsExactOriginalParseDiagnostics(t *testing.T) {
	for _, item := range []struct{ name, text, want string }{
		{"name-word", "name: capture\n", "steps is missing"},
		{"description-word", "description: capture\nsteps:\n  - run: [onboard-kit]\n", "name is missing"},
		{"description-opener", "name: legacy\ndescription: '${'\nsteps:\n  - expect: {exit: 0}\n", "step 1: run is missing"},
		{"inline-unresolved-text", "name: legacy\nsteps:\n  - files: {literal.txt: '${PRIVATE_UNUSED}'}\n", "step 1: run is missing"},
		{"inline-word", "name: legacy\nsteps:\n  - files: {literal.txt: capture}\n", "step 1: run is missing"},
		{"stdin-words", "name: legacy\nsteps:\n  - stdin: capture yaml-scalars\n", "step 1: run is missing"},
		{"description-old-word", "name: legacy\ndescription: yaml-scalars\nsteps:\n  - expect: {exit: 0}\n", "step 1: run is missing"},
		{"path-error-keeps-literal", "name: yaml-scalars\nsteps:\n  - run: [onboard-kit]\n    files: {'../outside.txt': capture}\n", "step 1: file \"../outside.txt\" must stay inside the folder"},
		{"existing-timestamp-is-not-new-predicate", "name: legacy\nsteps:\n  - expect: {files: {actual.yaml: {yaml-scalars: [{path: [firstOpenedAt], utc-timestamp: true}]}}}\n", "step 1: run is missing"},
	} {
		t.Run(item.name, func(t *testing.T) {
			_, err := runner.Parse([]byte(item.text))
			if err == nil || err.Error() != item.want {
				t.Fatalf("literal pre-extension diagnostic: want %q, got %v", item.want, err)
			}
		})
	}
}

func TestValidLegacyLiteralMarkersRemainOrdinaryScenarioData(t *testing.T) {
	for _, text := range []string{
		"name: capture\ndescription: '${'\nsteps:\n  - run: [onboard-kit]\n",
		"name: yaml-scalars\nsteps:\n  - run: [onboard-kit]\n    stdin: capture yaml-scalars\n",
		"name: legacy\nsteps:\n  - run: [onboard-kit]\n    files: {literal.txt: '${UNRESOLVED} capture yaml-scalars'}\n",
		"name: legacy\nsteps:\n  - run: [onboard-kit]\n    expect: {stdout: {equals: '${UNRESOLVED} capture yaml-scalars'}}\n",
		"name: legacy\nsteps:\n  - run: [onboard-kit]\n    expect: {stdout: {json: {label: '${UNRESOLVED} capture yaml-scalars'}}}\n",
	} {
		scenario, err := runner.Parse([]byte(text))
		if err != nil {
			t.Fatalf("legacy literal data was mistaken for extension syntax: %v", err)
		}
		if scenario.Name == "" || len(scenario.Steps) != 1 || len(scenario.Steps[0].Run) != 1 || scenario.Steps[0].Run[0] != "onboard-kit" {
			t.Fatal("valid legacy Parse returned no actual scenario")
		}
	}
}

func TestActualExtensionParseErrorsArePrivateButUnknownReferencesRemainRuntime(t *testing.T) {
	for _, item := range []struct{ name, text string }{
		{"capture", "name: private\nPRIVATE_SCHEMA_CAPTURE: '${PRIVATE_DYNAMIC}'\nsteps:\n  - run: [onboard-kit]\n    capture: {HELD: {unique-directory-entry: entries}}\n"},
		{"equals", "name: private\nPRIVATE_SCHEMA_EQUALS: '${PRIVATE_DYNAMIC}'\nsteps:\n  - run: [onboard-kit]\n    expect: {files: {actual.yaml: {yaml-scalars: [{path: [value], equals: '${PRIVATE_DYNAMIC}'}]}}}\n"},
		{"absent", "name: private\nPRIVATE_SCHEMA_ABSENT: '${PRIVATE_DYNAMIC}'\nsteps:\n  - run: [onboard-kit]\n    expect: {files: {actual.yaml: {yaml-scalars: [{path: [value], absent: true}]}}}\n"},
	} {
		t.Run(item.name, func(t *testing.T) {
			_, err := runner.Parse([]byte(item.text))
			if err == nil || err.Error() == "" {
				t.Fatal("actual extension still requires a nonempty schema refusal")
			}
			for _, private := range []string{"PRIVATE_SCHEMA", "PRIVATE_DYNAMIC"} {
				if strings.Contains(err.Error(), private) {
					t.Fatal("extension schema diagnostic exposed unknown private field/value")
				}
			}
		})
	}
	for _, text := range []string{
		"name: runtime-only\nsteps:\n  - run: [onboard-kit, '${PRIVATE_UNKNOWN}']\n    capture: {HELD: {unique-directory-entry: entries}}\n",
		"name: runtime-only\nsteps:\n  - run: [onboard-kit]\n    expect: {files: {actual.yaml: {yaml-scalars: [{path: [value], equals: '${PRIVATE_UNKNOWN}'}]}}}\n",
	} {
		if _, err := runner.Parse([]byte(text)); err != nil {
			t.Fatal("valid unresolved interpolation belongs to Run, not schema refusal")
		}
	}
}
