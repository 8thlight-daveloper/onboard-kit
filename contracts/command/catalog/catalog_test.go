package catalog_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	contract "onboard-kit/contracts/command/catalog"
)

type summary struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	Goal        string `json:"goal"`
	Result      string `json:"result"`
	Carrier     string `json:"carrier"`
	ContentPath string `json:"contentPath"`
}

type inventoryFixture struct {
	Source  string    `json:"source"`
	Entries []summary `json:"entries"`
}

type digestRow struct {
	ID     string `json:"id"`
	Size   int    `json:"size"`
	SHA256 string `json:"sha256"`
}

type digestFixture struct {
	Source  string      `json:"source"`
	Entries []digestRow `json:"entries"`
}

type argumentCase struct {
	Name   string   `json:"name"`
	Source string   `json:"source"`
	Args   []string `json:"args"`
	Exit   int      `json:"exit"`
	Code   string   `json:"code"`
}

func fixtureBytes(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func readFixture(t *testing.T, name string, target any) {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(fixtureBytes(t, name)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		t.Fatal(err)
	}
}

func inventory(t *testing.T) []summary {
	t.Helper()
	var expected inventoryFixture
	readFixture(t, "inventory.json", &expected)
	if expected.Source == "" || len(expected.Entries) != 80 {
		t.Fatal("unfrozen independent eighty-entry inventory")
	}
	return expected.Entries
}

func digests(t *testing.T) map[string]digestRow {
	t.Helper()
	var expected digestFixture
	readFixture(t, "content-digests.json", &expected)
	if expected.Source == "" || len(expected.Entries) != 80 {
		t.Fatal("unfrozen independent content digests")
	}
	result := make(map[string]digestRow, 80)
	for _, row := range expected.Entries {
		if row.Size <= 0 || len(row.SHA256) != 64 {
			t.Fatal("invalid content digest oracle")
		}
		result[row.ID] = row
	}
	if len(result) != 80 {
		t.Fatal("duplicate digest identities")
	}
	return result
}

type result struct {
	exit     int
	out, err string
}

func run(command contract.Command, args []string) result {
	var stdout, stderr bytes.Buffer
	exit := command.Run(args, &stdout, &stderr)
	return result{exit: exit, out: stdout.String(), err: stderr.String()}
}

func TestLiteralFullJSONInventory(t *testing.T) {
	t.Parallel()
	wantBytes, err := json.Marshal(inventory(t))
	if err != nil {
		t.Fatal(err)
	}
	var want any
	if err := json.Unmarshal(wantBytes, &want); err != nil {
		t.Fatal(err)
	}
	shared := newCommand()
	for range 2 {
		got := run(shared, []string{"list"})
		if got.exit != 0 || got.err != "" || !strings.HasPrefix(got.out, "[") || !strings.HasSuffix(got.out, "]\n") {
			t.Fatalf("list output boundary: %#v", got)
		}
		var actual any
		if err := json.Unmarshal([]byte(got.out), &actual); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(actual, want) {
			t.Fatal("list differs from exact independent inventory/key set")
		}
	}
}

func assertDigest(t *testing.T, id string, data []byte, oracle digestRow) {
	t.Helper()
	hash := sha256.Sum256(data)
	if len(data) != oracle.Size || hex.EncodeToString(hash[:]) != oracle.SHA256 {
		t.Fatalf("%s content differs from independently frozen size/digest", id)
	}
}

func TestEveryExportMatchesIndependentContent(t *testing.T) {
	t.Parallel()
	oracles := digests(t)
	shared := newCommand()
	for _, row := range inventory(t) {
		t.Run(row.ID, func(t *testing.T) {
			for _, args := range [][]string{{"export", "--id", row.ID}, {"export", "--id=" + row.ID}} {
				got := run(shared, args)
				if got.exit != 0 || got.err != "" {
					t.Fatalf("export refused: %#v", got)
				}
				assertDigest(t, row.ID, []byte(got.out), oracles[row.ID])
			}
		})
	}
}

func showPrefix(row summary) string {
	return "# " + row.Title + "\n\nID: `" + row.ID + "`\nCarrier: `" + row.Carrier + "`\nCategory: `" + row.Category + "`\nGoal: " + row.Goal + "\nResult: " + row.Result + "\nContent: `" + row.ContentPath + "`\n\n## Content\n\n"
}

func TestEveryShowHasCompleteMetadataAndContent(t *testing.T) {
	t.Parallel()
	oracles := digests(t)
	shared := newCommand()
	for _, row := range inventory(t) {
		got := run(shared, []string{"show", "--id", row.ID})
		prefix := showPrefix(row)
		if got.exit != 0 || got.err != "" || !strings.HasPrefix(got.out, prefix) {
			t.Fatalf("show %s metadata or output boundary: %#v", row.ID, got)
		}
		content := strings.TrimPrefix(got.out, prefix)
		if row.Carrier == "section-template" {
			opening, rest, ok := strings.Cut(content, "\n")
			fence := strings.TrimSuffix(opening, "yaml")
			if !ok || !strings.HasSuffix(opening, "yaml") || len(fence) < 3 || strings.Trim(fence, "`") != "" || !strings.HasSuffix(rest, fence+"\n") {
				t.Fatal("missing safe YAML show fence")
			}
			content = strings.TrimSuffix(rest, fence+"\n")
			longest, current := 0, 0
			for _, character := range content {
				if character == '`' {
					current++
					longest = max(longest, current)
				} else {
					current = 0
				}
			}
			if len(fence) != max(3, longest+1) {
				t.Fatal("wrong dynamic fence length")
			}
		}
		assertDigest(t, row.ID, []byte(content), oracles[row.ID])
	}
}

func TestThreeLiteralShowGoldens(t *testing.T) {
	t.Parallel()
	for _, id := range []string{"company-email", "pair-assign", "invitation-start"} {
		got := run(newCommand(), []string{"show", "--id=" + id})
		if got.exit != 0 || got.err != "" || got.out != string(fixtureBytes(t, "show-"+id+".md")) {
			t.Fatalf("literal show %s differs: %#v", id, got)
		}
	}
}

var reportMessages = map[string]string{
	"catalog.usage":         "Use catalog list, show --id ID, or export --id ID.",
	"catalog.unknown-id":    "Choose an exact ID from catalog list.",
	"catalog.result-output": "The catalog result could not be fully displayed.",
	"catalog.help-output":   "The catalog help could not be fully displayed.",
}

func assertReport(t *testing.T, data, code string) {
	t.Helper()
	var got map[string]any
	if err := json.Unmarshal([]byte(data), &got); err != nil {
		t.Fatalf("missing JSON report: %q %v", data, err)
	}
	want := map[string]any{"code": code, "message": reportMessages[code]}
	if !strings.HasSuffix(data, "\n") || strings.HasSuffix(data, "\n\n") || !reflect.DeepEqual(got, want) {
		t.Fatalf("private report differs: got%#v want%#v", got, want)
	}
}

func TestLiteralMalformedArgumentsAndUnknownIDs(t *testing.T) {
	t.Parallel()
	var cases []argumentCase
	readFixture(t, "arguments.json", &cases)
	if len(cases) < 30 {
		t.Fatal("incomplete independent argv matrix")
	}
	shared := newCommand()
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if c.Source == "" {
				t.Fatal("argument fixture missing provenance")
			}
			args := make([]string, len(c.Args))
			copy(args, c.Args)
			got := run(shared, c.Args)
			if got.exit != c.Exit || got.out != "" || !reflect.DeepEqual(args, c.Args) {
				t.Fatalf("argv refusal boundary: %#v", got)
			}
			assertReport(t, got.err, c.Code)
		})
	}
}

func TestLiteralHelpAndPrecedence(t *testing.T) {
	t.Parallel()
	want := string(fixtureBytes(t, "help.txt"))
	for _, args := range [][]string{{"--help"}, {"--help", "private-data"}, {"list", "--help", "--out", "private-path"}, {"show", "--help", "--id="}, {"export", "--help", "unknown", "--private"}} {
		got := run(newCommand(), args)
		if got.exit != 0 || got.err != "" || got.out != want {
			t.Fatalf("help differs or precedence lost: %#v", got)
		}
	}
}

func TestSuccessfulArgumentsRemainCallerOwned(t *testing.T) {
	t.Parallel()
	shared := newCommand()
	for _, args := range [][]string{
		{"list"},
		{"show", "--id", "company-email"},
		{"show", "--id=pair-assign"},
		{"export", "--id", "invitation-start"},
		{"export", "--id=company-email"},
		{"--help"},
		{"list", "--help", "ignored-private-token"},
		{"show", "--help", "--id="},
		{"export", "--help"},
	} {
		before := make([]string, len(args))
		copy(before, args)
		got := run(shared, args)
		if got.exit != 0 || got.err != "" || got.out == "" {
			t.Fatalf("missing successful ownership control: %#v", got)
		}
		if !reflect.DeepEqual(args, before) {
			t.Fatal("successful list/show/export/help mutated caller arguments")
		}
	}
}

type faultWriter struct{ short bool }

func (writer faultWriter) Write(data []byte) (int, error) {
	if writer.short {
		return max(0, len(data)-1), nil
	}
	return 0, errors.New("PRIVATE WRITER FAILURE")
}

func TestResultAndHelpWriterFailures(t *testing.T) {
	t.Parallel()
	for _, short := range []bool{false, true} {
		for _, args := range [][]string{{"list"}, {"show", "--id=company-email"}, {"export", "--id=pair-assign"}, {"--help"}} {
			var stderr bytes.Buffer
			exit := newCommand().Run(args, faultWriter{short: short}, &stderr)
			if exit != 1 {
				t.Fatal("stdout error/short write must refuse reported success")
			}
			code := "catalog.result-output"
			if args[0] == "--help" {
				code = "catalog.help-output"
			}
			assertReport(t, stderr.String(), code)
		}
	}
}

func TestRefusalWriterFailureAndUnusedStderr(t *testing.T) {
	t.Parallel()
	for _, short := range []bool{false, true} {
		for _, args := range [][]string{{}, {"show", "--id=unknown-private-selector"}} {
			var stdout bytes.Buffer
			if newCommand().Run(args, &stdout, faultWriter{short: short}) != 1 || stdout.Len() != 0 {
				t.Fatal("refusal report writer failure must return1 without stdout")
			}
		}
		if newCommand().Run([]string{"list"}, io.Discard, faultWriter{short: short}) != 0 {
			t.Fatal("unused stderr cannot fail successful operation")
		}
	}
}

func TestConcurrentCommandReuse(t *testing.T) {
	t.Parallel()
	shared := newCommand()
	var group sync.WaitGroup
	for range 12 {
		group.Add(1)
		go func() {
			defer group.Done()
			got := run(shared, []string{"show", "--id=company-email"})
			if got.exit != 0 || got.err != "" || got.out != string(fixtureBytes(t, "show-company-email.md")) {
				t.Error("shared command show state changed")
			}
			refused := run(shared, []string{"export", "--id=unknown-private-selector"})
			if refused.exit != 1 || refused.out != "" {
				t.Error("shared command unknown boundary changed")
			}
			assertReport(t, refused.err, "catalog.unknown-id")
		}()
	}
	group.Wait()
}
