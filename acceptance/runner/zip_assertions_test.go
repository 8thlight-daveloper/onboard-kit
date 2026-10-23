package runner_test

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"onboard-kit/acceptance/runner"
)

type archiveFixtureEntry struct {
	name              string
	payload           string
	method            uint16
	mode              fs.FileMode
	modified          time.Time
	raw               bool
	creator           uint16
	attributes        uint32
	attributeOverride bool
}

func ordinaryZIPEntry(name, payload string) archiveFixtureEntry {
	return archiveFixtureEntry{name: name, payload: payload, method: zip.Store, mode: 0o600, modified: time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)}
}

func archiveFixture(t *testing.T, entries []archiveFixtureEntry, prefixBytes int) string {
	t.Helper()
	var output bytes.Buffer
	if prefixBytes > 0 {
		output.Write(make([]byte, prefixBytes))
	}
	writer := zip.NewWriter(&output)
	if prefixBytes > 0 {
		writer.SetOffset(int64(prefixBytes))
	}
	for _, entry := range entries {
		header := zip.FileHeader{Name: entry.name, Method: entry.method, Modified: entry.modified}
		header.SetModTime(entry.modified)
		header.SetMode(entry.mode)
		if entry.attributeOverride {
			header.CreatorVersion = entry.creator<<8 | 20
			header.ExternalAttrs = entry.attributes
		}
		if entry.raw {
			if entry.method != zip.Store {
				t.Fatal("raw fixture helper supports only stored payloads")
			}
			header.CRC32 = crc32.ChecksumIEEE([]byte(entry.payload))
			header.CompressedSize64 = uint64(len(entry.payload))
			header.UncompressedSize64 = uint64(len(entry.payload))
			payload, err := writer.CreateRaw(&header)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := payload.Write([]byte(entry.payload)); err != nil {
				t.Fatal(err)
			}
			continue
		}
		payload, err := writer.CreateHeader(&header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := payload.Write([]byte(entry.payload)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return output.String()
}

func zipExpectationEntry(name, fixture, method, mode, modified string) string {
	return fmt.Sprintf("{name: %q, same-as: %q, method: %q, mode: %q, modified: %q}", name, fixture, method, mode, modified)
}

func zipScenario(entries ...string) string {
	return "name: zip-assertion\nsteps:\n  - run: [onboard-kit, copy, input.zip, actual.zip]\n    files:\n      input.zip: {from: archive.zip}\n    expect:\n      exit: 0\n      unchanged: [input.zip]\n      files:\n        actual.zip:\n          zip:\n            entries: [" + strings.Join(entries, ", ") + "]\n"
}

func oneZIPExpectation() string {
	return zipExpectationEntry("plan.obk.yaml", "golden/plan.txt", "store", "0600", "1980-01-01T00:00:00Z")
}

func TestZIPStoreAndDeflatePayloadMetadataMatch(t *testing.T) {
	t.Parallel()
	for _, method := range []struct {
		name string
		code uint16
	}{{"store", zip.Store}, {"deflate", zip.Deflate}} {
		t.Run(method.name, func(t *testing.T) {
			t.Parallel()
			entry := ordinaryZIPEntry("plan.obk.yaml", "literal plan\n")
			entry.method = method.code
			expect := zipExpectationEntry("plan.obk.yaml", "golden/plan.txt", method.name, "0600", "1980-01-01T00:00:00Z")
			failures := runScenarioWith(t, zipScenario(expect), runner.Options{TimeLimit: 20 * time.Second}, map[string]string{"archive.zip": archiveFixture(t, []archiveFixtureEntry{entry}, 0), "golden/plan.txt": "literal plan\n"})
			requirePass(t, failures)
		})
	}
}

func TestZIPEntryListIsCompleteAndOrdered(t *testing.T) {
	t.Parallel()
	first := ordinaryZIPEntry("first", "one")
	second := ordinaryZIPEntry("second", "two")
	for _, c := range []struct {
		name    string
		entries []archiveFixtureEntry
	}{{"correct", []archiveFixtureEntry{first, second}}, {"reordered", []archiveFixtureEntry{second, first}}, {"missing", []archiveFixtureEntry{first}}, {"extra", []archiveFixtureEntry{first, second, ordinaryZIPEntry("third", "three")}}} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			text := zipScenario(zipExpectationEntry("first", "golden/first", "store", "0600", "1980-01-01T00:00:00Z"), zipExpectationEntry("second", "golden/second", "store", "0600", "1980-01-01T00:00:00Z"))
			failures := runScenarioWith(t, text, runner.Options{TimeLimit: 20 * time.Second}, map[string]string{"archive.zip": archiveFixture(t, c.entries, 0), "golden/first": "one", "golden/second": "two"})
			if c.name == "correct" {
				requirePass(t, failures)
			} else {
				requireFailure(t, failures, 1, "file actual.zip zip")
			}
		})
	}
}

func TestZIPWrongPayloadOrMetadataFails(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"name", "payload", "mode", "special-mode", "modified", "method", "symlink", "directory", "unknown-unix-type", "unknown-macos-type"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			entry := ordinaryZIPEntry("plan.obk.yaml", "literal plan\n")
			switch name {
			case "name":
				entry.name = "other.obk.yaml"
			case "payload":
				entry.payload = "different plan\n"
			case "mode":
				entry.mode = 0o644
			case "special-mode":
				entry.mode = 0o600 | fs.ModeSetuid
			case "modified":
				entry.modified = entry.modified.Add(2 * time.Second)
			case "method":
				entry.method = zip.Deflate
			case "symlink":
				entry.mode = 0o600 | fs.ModeSymlink
			case "directory":
				entry.mode = 0o600 | fs.ModeDir
			case "unknown-unix-type":
				entry.attributeOverride = true
				entry.creator = 3
				entry.attributes = (0x3000 | 0o600) << 16
			case "unknown-macos-type":
				entry.attributeOverride = true
				entry.creator = 19
				entry.attributes = (0xe000 | 0o600) << 16
			}
			failures := runScenarioWith(t, zipScenario(oneZIPExpectation()), runner.Options{TimeLimit: 20 * time.Second}, map[string]string{"archive.zip": archiveFixture(t, []archiveFixtureEntry{entry}, 0), "golden/plan.txt": "literal plan\n"})
			requireFailure(t, failures, 1, "file actual.zip zip")
		})
	}
}

func centralCRC(t *testing.T, archive string, value uint32) string {
	t.Helper()
	data := []byte(archive)
	index := bytes.Index(data, []byte{'P', 'K', 1, 2})
	if index < 0 || index+20 > len(data) {
		t.Fatal("CRC fixture lacks central header")
	}
	binary.LittleEndian.PutUint32(data[index+16:index+20], value)
	return string(data)
}

func TestZIPReadsFullPayloadAndChecksEvenZeroCentralCRC(t *testing.T) {
	t.Parallel()
	entry := ordinaryZIPEntry("plan.obk.yaml", "literal plan\n")
	entry.raw = true
	actualCRC := crc32.ChecksumIEEE([]byte(entry.payload))
	if actualCRC == 0 {
		t.Fatal("CRC-zero negative fixture would not reach mismatch")
	}
	archive := archiveFixture(t, []archiveFixtureEntry{entry}, 0)
	for _, c := range []struct {
		name  string
		data  string
		valid bool
	}{{"raw-valid", archive, true}, {"wrong-crc", centralCRC(t, archive, actualCRC^1), false}, {"zero-crc", centralCRC(t, archive, 0), false}, {"truncated", archive[:len(archive)-8], false}, {"not-zip", "not a zip", false}} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			failures := runScenarioWith(t, zipScenario(oneZIPExpectation()), runner.Options{TimeLimit: 20 * time.Second}, map[string]string{"archive.zip": c.data, "golden/plan.txt": entry.payload})
			if c.valid {
				requirePass(t, failures)
			} else {
				requireFailure(t, failures, 1, "file actual.zip zip")
			}
		})
	}
}

func TestZIPEmptyPayloadAndEquivalentModifiedInstant(t *testing.T) {
	t.Parallel()
	entry := ordinaryZIPEntry("empty", "")
	entry.mode = 0
	entry.raw = true
	text := zipScenario(zipExpectationEntry("empty", "golden/empty", "store", "0000", "1980-01-01T01:00:00+01:00"))
	failures := runScenarioWith(t, text, runner.Options{TimeLimit: 20 * time.Second}, map[string]string{"archive.zip": archiveFixture(t, []archiveFixtureEntry{entry}, 0), "golden/empty": ""})
	requirePass(t, failures)
}

func TestZIPAggregateActualUncompressedLimit(t *testing.T) {
	for _, size := range []int{1 << 20, (1 << 20) + 1} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			payload := strings.Repeat("x", size)
			first := ordinaryZIPEntry("first", payload[:size/2])
			second := ordinaryZIPEntry("second", payload[size/2:])
			first.method = zip.Deflate
			second.method = zip.Deflate
			text := zipScenario(zipExpectationEntry("first", "golden/first", "deflate", "0600", "1980-01-01T00:00:00Z"), zipExpectationEntry("second", "golden/second", "deflate", "0600", "1980-01-01T00:00:00Z"))
			failures := runScenarioWith(t, text, runner.Options{TimeLimit: 20 * time.Second}, map[string]string{"archive.zip": archiveFixture(t, []archiveFixtureEntry{first, second}, 0), "golden/first": first.payload, "golden/second": second.payload})
			if size == 1<<20 {
				requirePass(t, failures)
			} else {
				requireFailure(t, failures, 1, "file actual.zip zip")
			}
		})
	}
}

func TestZIPRawArchiveLimitInclusive(t *testing.T) {
	entry := ordinaryZIPEntry("plan.obk.yaml", "literal plan\n")
	bare := archiveFixture(t, []archiveFixtureEntry{entry}, 0)
	for _, size := range []int{2 << 20, (2 << 20) + 1} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			archive := archiveFixture(t, []archiveFixtureEntry{entry}, size-len(bare))
			if len(archive) != size {
				t.Fatal("raw-size boundary fixture is not the literal requested size")
			}
			failures := runScenarioWith(t, zipScenario(oneZIPExpectation()), runner.Options{TimeLimit: 20 * time.Second}, map[string]string{"archive.zip": archive, "golden/plan.txt": entry.payload})
			if size == 2<<20 {
				requirePass(t, failures)
			} else {
				requireFailure(t, failures, 1, "file actual.zip zip")
			}
		})
	}
}

func TestZIPLiteralNamesDuplicatesAndNoExtraction(t *testing.T) {
	t.Parallel()
	temporaryRoot := t.TempDir()
	entries := []archiveFixtureEntry{ordinaryZIPEntry("../outside.txt", "one"), ordinaryZIPEntry("../outside.txt", "two")}
	text := zipScenario(zipExpectationEntry("../outside.txt", "golden/one", "store", "0600", "1980-01-01T00:00:00Z"), zipExpectationEntry("../outside.txt", "golden/two", "store", "0600", "1980-01-01T00:00:00Z"))
	failures := runScenarioWith(t, text, runner.Options{TimeLimit: 20 * time.Second, TemporaryRoot: temporaryRoot}, map[string]string{"archive.zip": archiveFixture(t, entries, 0), "golden/one": "one", "golden/two": "two"})
	requirePass(t, failures)
	if _, err := os.Stat(filepath.Join(temporaryRoot, "outside.txt")); !os.IsNotExist(err) {
		t.Fatalf("ZIP assertion extracted an entry outside the run folder: %v", err)
	}
}

func TestZIPMissingArchiveAndExpectedFixtureAreFailures(t *testing.T) {
	t.Parallel()
	text := zipScenario(oneZIPExpectation())
	entry := ordinaryZIPEntry("plan.obk.yaml", "literal plan\n")
	failures := runScenarioWith(t, text, runner.Options{TimeLimit: 20 * time.Second}, map[string]string{"archive.zip": archiveFixture(t, []archiveFixtureEntry{entry}, 0)})
	requireFailure(t, failures, 1, "file actual.zip zip")
	missing := "name: zip-missing\nsteps:\n  - run: [onboard-kit, say, '', '', '0']\n    expect:\n      files:\n        actual.zip: {zip: {entries: [" + oneZIPExpectation() + "]}}\n"
	failures = runScenarioWith(t, missing, runner.Options{TimeLimit: 20 * time.Second}, map[string]string{"golden/plan.txt": entry.payload})
	requireFailure(t, failures, 1, "file actual.zip")
}

func TestZIPExpectationParsesThroughExistingPublicAPI(t *testing.T) {
	t.Parallel()
	if _, err := runner.Parse([]byte(zipScenario(oneZIPExpectation()))); err != nil {
		t.Fatalf("valid additive ZIP expectation must parse: %v", err)
	}
}

func TestZIPDuplicateMappingKeysAreAmbiguous(t *testing.T) {
	t.Parallel()
	base := "name: duplicate-zip-key\nsteps:\n  - run: [onboard-kit, say, '', '', '0']\n    expect:\n      files:\n        actual.zip: "
	first := oneZIPExpectation()
	second := zipExpectationEntry("other.txt", "golden/other.txt", "store", "0600", "1980-01-01T00:00:00Z")
	firstZIP := "{entries: [" + first + "]}"
	secondZIP := "{entries: [" + second + "]}"
	identicalName := strings.Replace(first, `name: "plan.obk.yaml"`, `name: "plan.obk.yaml", name: "plan.obk.yaml"`, 1)
	conflictingName := strings.Replace(first, `name: "plan.obk.yaml"`, `name: "plan.obk.yaml", name: "other.txt"`, 1)
	if identicalName == first || conflictingName == first {
		t.Fatal("duplicate metadata fixture did not repeat its intended key")
	}
	cases := []struct {
		name, value string
	}{
		{"outer-zip-identical", "{zip: " + firstZIP + ", zip: " + firstZIP + "}"},
		{"outer-zip-conflicting", "{zip: " + firstZIP + ", zip: " + secondZIP + "}"},
		{"entries-identical", "{zip: {entries: [" + first + "], entries: [" + first + "]}}"},
		{"entries-conflicting", "{zip: {entries: [" + first + "], entries: [" + second + "]}}"},
		{"entry-name-identical", "{zip: {entries: [" + identicalName + "]}}"},
		{"entry-name-conflicting", "{zip: {entries: [" + conflictingName + "]}}"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if _, err := runner.Parse([]byte(base + c.value + "\n")); err == nil {
				t.Fatal("repeated ZIP mapping key must fail Parse, even when values agree")
			}
		})
	}
}

func TestZIPExpectationStrictSchema(t *testing.T) {
	t.Parallel()
	valid := oneZIPExpectation()
	base := "name: malformed-zip\nsteps:\n  - run: [onboard-kit, say, '', '', '0']\n    expect:\n      files:\n        actual.zip: "
	cases := []struct{ name, value string }{
		{"zip-scalar", `{zip: text}`},
		{"zip-null", `{zip: null}`},
		{"missing-entries", `{zip: {}}`},
		{"empty-entries", `{zip: {entries: []}}`},
		{"null-entries", `{zip: {entries: null}}`},
		{"mapping-entries", `{zip: {entries: {name: plan}}}`},
		{"scalar-entry", `{zip: {entries: [text]}}`},
		{"null-entry", `{zip: {entries: [null]}}`},
		{"empty-entry", `{zip: {entries: [{}]}}`},
		{"unknown-zip-key", "{zip: {entries: [" + valid + "], extra: true}}"},
		{"unknown-entry-key", "{zip: {entries: [" + strings.TrimSuffix(valid, "}") + ", extra: true}]}}"},
		{"unsupported-method", "{zip: {entries: [" + strings.Replace(valid, `method: "store"`, `method: "bzip2"`, 1) + "]}}"},
		{"method-case", "{zip: {entries: [" + strings.Replace(valid, `method: "store"`, `method: "Store"`, 1) + "]}}"},
		{"mode-not-octal", "{zip: {entries: [" + strings.Replace(valid, `mode: "0600"`, `mode: "0888"`, 1) + "]}}"},
		{"mode-too-short", "{zip: {entries: [" + strings.Replace(valid, `mode: "0600"`, `mode: "600"`, 1) + "]}}"},
		{"mode-special-bits", "{zip: {entries: [" + strings.Replace(valid, `mode: "0600"`, `mode: "04600"`, 1) + "]}}"},
		{"mode-numeric", "{zip: {entries: [" + strings.Replace(valid, `mode: "0600"`, `mode: 0600`, 1) + "]}}"},
		{"modified-invalid", "{zip: {entries: [" + strings.Replace(valid, `modified: "1980-01-01T00:00:00Z"`, `modified: "not a timestamp"`, 1) + "]}}"},
		{"modified-invalid-date", "{zip: {entries: [" + strings.Replace(valid, `modified: "1980-01-01T00:00:00Z"`, `modified: "2026-02-30T00:00:00Z"`, 1) + "]}}"},
		{"fixture-parent-escape", "{zip: {entries: [" + strings.Replace(valid, `same-as: "golden/plan.txt"`, `same-as: "../outside.txt"`, 1) + "]}}"},
		{"fixture-absolute", "{zip: {entries: [" + strings.Replace(valid, `same-as: "golden/plan.txt"`, `same-as: "/outside.txt"`, 1) + "]}}"},
		{"zip-with-same-as", "{same-as: golden.zip, zip: {entries: [" + valid + "]}}"},
		{"zip-with-json", "{json: {}, zip: {entries: [" + valid + "]}}"},
		{"zip-with-absent-true", "{absent: true, zip: {entries: [" + valid + "]}}"},
		{"zip-with-absent-false", "{absent: false, zip: {entries: [" + valid + "]}}"},
	}
	fields := []struct{ name, fragment, value string }{
		{"name", `name: "plan.obk.yaml", `, `"plan.obk.yaml"`},
		{"same-as", `same-as: "golden/plan.txt", `, `"golden/plan.txt"`},
		{"method", `method: "store", `, `"store"`},
		{"mode", `mode: "0600", `, `"0600"`},
		{"modified", `, modified: "1980-01-01T00:00:00Z"`, `"1980-01-01T00:00:00Z"`},
	}
	for _, field := range fields {
		missing := strings.Replace(valid, field.fragment, "", 1)
		if missing == valid {
			t.Fatal("missing-field fixture did not remove its intended key")
		}
		cases = append(cases, struct{ name, value string }{"missing-" + field.name, "{zip: {entries: [" + missing + "]}}"})
		empty := strings.Replace(valid, field.name+": "+field.value, field.name+`: ""`, 1)
		if empty == valid {
			t.Fatal("empty-field fixture did not empty its intended value")
		}
		cases = append(cases, struct{ name, value string }{"empty-" + field.name, "{zip: {entries: [" + empty + "]}}"})
		null := strings.Replace(valid, field.name+": "+field.value, field.name+": null", 1)
		cases = append(cases, struct{ name, value string }{"null-" + field.name, "{zip: {entries: [" + null + "]}}"})
		wrong := strings.Replace(valid, field.name+": "+field.value, field.name+": true", 1)
		cases = append(cases, struct{ name, value string }{"nontext-" + field.name, "{zip: {entries: [" + wrong + "]}}"})
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if _, err := runner.Parse([]byte(base + c.value + "\n")); err == nil {
				t.Fatal("malformed ZIP expectation must fail Parse")
			}
		})
	}
}

func TestZIPExpectedFixtureSymlinkConfinement(t *testing.T) {
	t.Parallel()
	for _, escape := range []bool{false, true} {
		t.Run(fmt.Sprint(escape), func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			targetRoot := root
			if escape {
				targetRoot = t.TempDir()
			}
			target := filepath.Join(targetRoot, "payload.txt")
			if err := os.WriteFile(target, []byte("literal plan\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(root, "golden"), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(target, filepath.Join(root, "golden", "plan.txt")); err != nil {
				t.Fatal(err)
			}
			entry := ordinaryZIPEntry("plan.obk.yaml", "literal plan\n")
			if err := os.WriteFile(filepath.Join(root, "archive.zip"), []byte(archiveFixture(t, []archiveFixtureEntry{entry}, 0)), 0o600); err != nil {
				t.Fatal(err)
			}
			scenario, err := runner.Parse([]byte(zipScenario(oneZIPExpectation())))
			if err != nil {
				t.Fatalf("valid ZIP scenario did not parse: %v", err)
			}
			failures := runner.Run(scenario, runner.Options{BinaryPath: fakeBinaryPath, RepositoryRoot: root, TimeLimit: 20 * time.Second})
			if escape {
				requireFailure(t, failures, 1, "file actual.zip zip")
			} else {
				requirePass(t, failures)
			}
		})
	}
}

func TestZIPAssertionsPreserveOtherFileAssertions(t *testing.T) {
	t.Parallel()
	text := strings.Replace(zipScenario(oneZIPExpectation()), "        actual.zip:", "        old.txt: {same-as: golden/old.txt}\n        copied.json: {json: {accepted: true}}\n        missing.txt: {absent: true}\n        actual.zip:", 1)
	text = strings.Replace(text, "      input.zip: {from: archive.zip}", "      input.zip: {from: archive.zip}\n      old.txt: {from: golden/old.txt}\n      copied.json: '{\"accepted\":true,\"extra\":1}'", 1)
	entry := ordinaryZIPEntry("plan.obk.yaml", "literal plan\n")
	failures := runScenarioWith(t, text, runner.Options{TimeLimit: 20 * time.Second}, map[string]string{"archive.zip": archiveFixture(t, []archiveFixtureEntry{entry}, 0), "golden/plan.txt": entry.payload, "golden/old.txt": "held old file"})
	requirePass(t, failures)
}

func TestOldFileAssertionsWithoutZIPRemainUnchanged(t *testing.T) {
	t.Parallel()
	text := `name: old-assertions
steps:
  - run: [onboard-kit, copy, source.txt, result.txt]
    files:
      source.txt: {from: golden/text.txt}
      object.json: '{"accepted":true,"extra":1}'
    expect:
      exit: 0
      unchanged: [source.txt]
      files:
        result.txt: {same-as: golden/text.txt}
        object.json: {json: {accepted: true}}
        absent.txt: {absent: true}
`
	requirePass(t, runScenarioWith(t, text, runner.Options{TimeLimit: 20 * time.Second}, map[string]string{"golden/text.txt": "held exact bytes\n"}))
}
