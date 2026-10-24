package runner_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"onboard-kit/acceptance/runner"
)

func yamlScalarScenario(expectation string) string {
	return "name: yaml-scalar\nsteps:\n  - run: [onboard-kit, say, '', '', '0']\n    files:\n      actual.yaml: {from: payload.yaml}\n    expect:\n      exit: 0\n      unchanged: [actual.yaml]\n      files:\n        actual.yaml: " + expectation + "\n"
}

func yamlScalarExpectation() string {
	return `{yaml-scalars: [{path: [firstOpenedAt], utc-timestamp: true}]}`
}

func runYAMLScalar(t *testing.T, text, expectation string) []runner.Failure {
	t.Helper()
	return runScenarioWith(t, yamlScalarScenario(expectation), runner.Options{TimeLimit: 20 * time.Second}, map[string]string{"payload.yaml": text})
}

func requireYAMLScalarFailure(t *testing.T, failures []runner.Failure) {
	t.Helper()
	requireFailure(t, failures, 1, "file actual.yaml yaml-scalars")
}

func TestYAMLScalarPublicParse(t *testing.T) {
	if _, err := runner.Parse([]byte(yamlScalarScenario(yamlScalarExpectation()))); err != nil {
		t.Fatalf("valid additive YAML scalar expectation must parse: %v", err)
	}
	requirePass(t, runYAMLScalar(t, "firstOpenedAt: \"2026-10-24T12:34:56.123Z\"\n", yamlScalarExpectation()))
}

func TestYAMLScalarSchema(t *testing.T) {
	cases := []struct{ name, value string }{
		{"null", `{yaml-scalars: null}`},
		{"scalar", `{yaml-scalars: text}`},
		{"mapping", `{yaml-scalars: {path: [firstOpenedAt], utc-timestamp: true}}`},
		{"empty", `{yaml-scalars: []}`},
		{"null-entry", `{yaml-scalars: [null]}`},
		{"scalar-entry", `{yaml-scalars: [text]}`},
		{"empty-entry", `{yaml-scalars: [{}]}`},
		{"missing-path", `{yaml-scalars: [{utc-timestamp: true}]}`},
		{"missing-predicate", `{yaml-scalars: [{path: [firstOpenedAt]}]}`},
		{"unknown-entry-key", `{yaml-scalars: [{path: [firstOpenedAt], utc-timestamp: true, extra: true}]}`},
		{"path-scalar", `{yaml-scalars: [{path: firstOpenedAt, utc-timestamp: true}]}`},
		{"path-null", `{yaml-scalars: [{path: null, utc-timestamp: true}]}`},
		{"path-empty", `{yaml-scalars: [{path: [], utc-timestamp: true}]}`},
		{"path-empty-key", `{yaml-scalars: [{path: [""], utc-timestamp: true}]}`},
		{"path-numeric", `{yaml-scalars: [{path: [1], utc-timestamp: true}]}`},
		{"path-boolean", `{yaml-scalars: [{path: [true], utc-timestamp: true}]}`},
		{"path-null-item", `{yaml-scalars: [{path: [null], utc-timestamp: true}]}`},
		{"path-timestamp", `{yaml-scalars: [{path: [2026-10-24], utc-timestamp: true}]}`},
		{"path-list-item", `{yaml-scalars: [{path: [[firstOpenedAt]], utc-timestamp: true}]}`},
		{"predicate-false", `{yaml-scalars: [{path: [firstOpenedAt], utc-timestamp: false}]}`},
		{"predicate-text", `{yaml-scalars: [{path: [firstOpenedAt], utc-timestamp: "true"}]}`},
		{"predicate-number", `{yaml-scalars: [{path: [firstOpenedAt], utc-timestamp: 1}]}`},
		{"predicate-null", `{yaml-scalars: [{path: [firstOpenedAt], utc-timestamp: null}]}`},
		{"predicate-map", `{yaml-scalars: [{path: [firstOpenedAt], utc-timestamp: {yes: true}}]}`},
		{"outer-repeat", `{yaml-scalars: [{path: [firstOpenedAt], utc-timestamp: true}], yaml-scalars: [{path: [firstOpenedAt], utc-timestamp: true}]}`},
		{"path-repeat", `{yaml-scalars: [{path: [firstOpenedAt], path: [firstOpenedAt], utc-timestamp: true}]}`},
		{"predicate-repeat", `{yaml-scalars: [{path: [firstOpenedAt], utc-timestamp: true, utc-timestamp: true}]}`},
		{"outer-anchor", `&held {yaml-scalars: [{path: [firstOpenedAt], utc-timestamp: true}]}`},
		{"list-anchor", `{yaml-scalars: &held [{path: [firstOpenedAt], utc-timestamp: true}]}`},
		{"entry-anchor", `{yaml-scalars: [&held {path: [firstOpenedAt], utc-timestamp: true}]}`},
		{"path-anchor", `{yaml-scalars: [{path: [&held firstOpenedAt], utc-timestamp: true}]}`},
		{"entry-alias", `{yaml-scalars: [&held {path: [firstOpenedAt], utc-timestamp: true}, *held]}`},
	}
	for _, other := range []string{`same-as: payload.yaml`, `json: {}`, `absent: true`, `absent: false`, `zip: {entries: [{name: a, same-as: a, method: store, mode: "0600", modified: "1980-01-01T00:00:00Z"}]}`} {
		body := strings.TrimSuffix(strings.TrimPrefix(yamlScalarExpectation(), "{"), "}")
		cases = append(cases, struct{ name, value string }{"other-first-" + other, "{" + other + ", " + body + "}"}, struct{ name, value string }{"other-last-" + other, "{" + body + ", " + other + "}"})
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := runner.Parse([]byte(yamlScalarScenario(c.value))); err == nil {
				t.Fatal("invalid new YAML scalar assertion must fail Parse")
			}
		})
	}
}

func TestYAMLScalarTimestampProfile(t *testing.T) {
	valid := []string{"0001-01-01T00:00:00Z", "9999-12-31T23:59:59Z", "2000-02-29T12:34:56Z", "2024-02-29T00:00:00.0Z", "2026-10-24T12:34:56.123456789Z"}
	invalid := []string{"", "0000-01-01T00:00:00Z", "10000-01-01T00:00:00Z", "1900-02-29T00:00:00Z", "2026-02-29T00:00:00Z", "2026-04-31T00:00:00Z", "2026-00-01T00:00:00Z", "2026-13-01T00:00:00Z", "2026-01-00T00:00:00Z", "2026-01-32T00:00:00Z", "2026-01-01T24:00:00Z", "2026-01-01T00:60:00Z", "2026-01-01T00:00:60Z", "2026-1-01T00:00:00Z", "2026-01-1T00:00:00Z", "2026-01-01T1:00:00Z", "2026-01-01T00:0:00Z", "2026-01-01T00:00:0Z", "2026-01-01t00:00:00Z", "2026-01-01T00:00:00z", "2026-01-01T00:00:00+00:00", "2026-01-01T00:00:00-00:00", "2026-01-01T00:00:00+01:00", "2026-01-01T00:00:00.Z", "2026-01-01T00:00:00,1Z", "2026-01-01T00:00:00.1234567890Z", " 2026-01-01T00:00:00Z", "2026-01-01T00:00:00Z ", "２０２６-01-01T00:00:00Z", "2026-01-01T00:00:00Zsuffix"}
	for i, value := range valid {
		t.Run(fmt.Sprintf("valid-%d", i), func(t *testing.T) {
			requirePass(t, runYAMLScalar(t, fmt.Sprintf("firstOpenedAt: %q\n", value), yamlScalarExpectation()))
		})
	}
	for i, value := range invalid {
		t.Run(fmt.Sprintf("invalid-%d", i), func(t *testing.T) {
			requireYAMLScalarFailure(t, runYAMLScalar(t, fmt.Sprintf("firstOpenedAt: %q\n", value), yamlScalarExpectation()))
		})
	}
}

func TestYAMLScalarValueAndShapeFailures(t *testing.T) {
	cases := []struct{ name, data string }{
		{"missing", "other: \"2026-10-24T12:34:56Z\"\n"},
		{"null", "firstOpenedAt: null\n"},
		{"boolean", "firstOpenedAt: true\n"},
		{"number", "firstOpenedAt: 2026\n"},
		{"implicit-timestamp", "firstOpenedAt: 2026-10-24T12:34:56Z\n"},
		{"explicit-timestamp", "firstOpenedAt: !!timestamp \"2026-10-24T12:34:56Z\"\n"},
		{"sequence", "firstOpenedAt: [\"2026-10-24T12:34:56Z\"]\n"},
		{"mapping", "firstOpenedAt: {value: \"2026-10-24T12:34:56Z\"}\n"},
		{"root-sequence", "- firstOpenedAt: \"2026-10-24T12:34:56Z\"\n"},
		{"root-scalar", "\"2026-10-24T12:34:56Z\"\n"},
		{"bad-date", "firstOpenedAt: \"2026-02-30T12:34:56Z\"\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			requireYAMLScalarFailure(t, runYAMLScalar(t, c.data, yamlScalarExpectation()))
		})
	}
}

func TestYAMLScalarGlobalDocumentValidation(t *testing.T) {
	good := "firstOpenedAt: \"2026-10-24T12:34:56Z\"\n"
	cases := []struct{ name, data string }{
		{"empty", ""},
		{"only-comment", "# nothing here\n"},
		{"malformed", good + "private: [unterminated\n"},
		{"invalid-utf8", good + "private: \"" + string([]byte{0xff}) + "\"\n"},
		{"extra-document", good + "---\nother: 1\n"},
		{"extra-empty-document", good + "---\n"},
		{"duplicate-selected", good + good},
		{"duplicate-unselected", good + "nested: {private: one, private: two}\n"},
		{"duplicate-same-value", good + "nested: {private: one, private: one}\n"},
		{"duplicate-string-spelling", good + "nested: {private: one, \"private\": two}\n"},
		{"duplicate-escaped-spelling", good + "nested: {\"pri\\u0076ate\": one, private: two}\n"},
		{"unused-anchor", good + "nested: &unused {value: held}\n"},
		{"selected-anchor", "firstOpenedAt: &held \"2026-10-24T12:34:56Z\"\n"},
		{"alias", "held: &held \"2026-10-24T12:34:56Z\"\nfirstOpenedAt: *held\n"},
		{"unselected-alias", good + "held: &held text\nother: *held\n"},
		{"complex-key", good + "? [one, two]\n: value\n"},
		{"numeric-key", good + "nested: {1: held}\n"},
		{"boolean-key", good + "nested: {true: held}\n"},
		{"null-key", good + "nested: {null: held}\n"},
		{"timestamp-key", good + "nested: {2026-10-24: held}\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			requireYAMLScalarFailure(t, runYAMLScalar(t, c.data, yamlScalarExpectation()))
		})
	}
	for _, c := range []struct{ name, data string }{{"explicit-one-document", "---\n" + good + "...\n"}, {"unselected-nontext-values", good + "\"1\": false\nnested: [false, null, 12]\n"}, {"explicit-string", "firstOpenedAt: !!str 2026-10-24T12:34:56Z\n"}} {
		t.Run(c.name, func(t *testing.T) {
			requirePass(t, runYAMLScalar(t, c.data, yamlScalarExpectation()))
		})
	}
}

func TestYAMLScalarLiteralPathsAndConjunction(t *testing.T) {
	data := "root:\n  \"a.b/c~[0]\":\n    \" with spaces \": \"2026-10-24T12:34:56Z\"\nother: \"2000-02-29T00:00:00.123Z\"\n"
	first := `{path: [root, "a.b/c~[0]", " with spaces "], utc-timestamp: true}`
	second := `{path: [other], utc-timestamp: true}`
	expect := "{yaml-scalars: [" + first + ", " + second + ", " + first + "]}"
	requirePass(t, runYAMLScalar(t, data, expect))
	requireYAMLScalarFailure(t, runYAMLScalar(t, strings.Replace(data, "2000-02-29T00:00:00.123Z", "missing", 1), expect))
	requireYAMLScalarFailure(t, runYAMLScalar(t, data, `{yaml-scalars: [{path: [root, a, b], utc-timestamp: true}]}`))
	for _, intermediate := range []string{"null", "[]", "[text]", "text"} {
		requireYAMLScalarFailure(t, runYAMLScalar(t, "root: "+intermediate+"\n", `{yaml-scalars: [{path: [root, firstOpenedAt], utc-timestamp: true}]}`))
	}
}

func TestYAMLScalarInclusiveByteLimit(t *testing.T) {
	prefix := "firstOpenedAt: \"2026-10-24T12:34:56Z\"\n#"
	for _, size := range []int{1 << 20, (1 << 20) + 1} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			data := prefix + strings.Repeat("x", size-len(prefix))
			if len(data) != size {
				t.Fatal("independent byte fixture has wrong size")
			}
			failures := runYAMLScalar(t, data, yamlScalarExpectation())
			if size == 1<<20 {
				requirePass(t, failures)
			} else {
				requireYAMLScalarFailure(t, failures)
			}
		})
	}
}

func TestYAMLScalarCheckingPreservesBytesAndOtherAssertions(t *testing.T) {
	data := "# Preserve this exact spelling.\nfirstOpenedAt : !!str 2026-10-24T12:34:56Z\nnotes: untouched\n"
	text := strings.Replace(yamlScalarScenario(yamlScalarExpectation()), "        actual.yaml:", "        old.txt: {same-as: golden.txt}\n        object.json: {json: {accepted: true}}\n        absent.txt: {absent: true}\n        actual.yaml:", 1)
	text = strings.Replace(text, "      actual.yaml: {from: payload.yaml}", "      actual.yaml: {from: payload.yaml}\n      old.txt: {from: golden.txt}\n      object.json: '{\"accepted\":true,\"extra\":1}'", 1)
	text += "  - run: [onboard-kit, say, '', '', '0']\n    expect:\n      exit: 0\n      files:\n        actual.yaml: {same-as: payload.yaml}\n"
	requirePass(t, runScenarioWith(t, text, runner.Options{TimeLimit: 20 * time.Second}, map[string]string{"payload.yaml": data, "golden.txt": "held existing text\n"}))
}

func TestYAMLScalarFailuresDoNotExposeHeldText(t *testing.T) {
	secret := "FABRICATED_PRIVATE_SENTINEL"
	for _, data := range []string{"firstOpenedAt: \"" + secret + "\"\n", "firstOpenedAt: \"2026-10-24T12:34:56Z\"\nprivate: [" + secret + "\n", "firstOpenedAt: \"2026-10-24T12:34:56Z\"\n" + secret + ": one\n" + secret + ": two\n"} {
		failures := runYAMLScalar(t, data, yamlScalarExpectation())
		requireYAMLScalarFailure(t, failures)
		for _, failure := range failures {
			if strings.Contains(failure.Message, secret) || strings.Contains(failure.String(), secret) {
				t.Fatal("YAML scalar failure exposed held fabricated private text")
			}
		}
	}
}

func TestYAMLScalarMissingOrDirectoryIsFailure(t *testing.T) {
	missing := "name: yaml-missing\nsteps:\n  - run: [onboard-kit, say, '', '', '0']\n    expect:\n      files:\n        actual.yaml: " + yamlScalarExpectation() + "\n"
	failures := runScenarioWith(t, missing, runner.Options{TimeLimit: 20 * time.Second}, nil)
	requireFailure(t, failures, 1, "file actual.yaml")
	directory := strings.Replace(missing, "    expect:", "    files:\n      actual.yaml/child: held\n    expect:", 1)
	failures = runScenarioWith(t, directory, runner.Options{TimeLimit: 20 * time.Second}, nil)
	requireFailure(t, failures, 1, "file actual.yaml")
}
