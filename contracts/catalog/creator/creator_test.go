package creator_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

type countedInput struct {
	source   []byte
	position int
	hook     func()
}

func (r *countedInput) Read(p []byte) (int, error) {
	if r.hook != nil {
		hook := r.hook
		r.hook = nil
		hook()
	}
	if r.position == len(r.source) {
		return 0, io.EOF
	}
	if len(p) == 0 {
		return 0, nil
	}
	p[0] = r.source[r.position]
	r.position++
	return 1, nil
}

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func sandbox(t *testing.T) (string, string) {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chown(root, -1, os.Getegid()); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"preparation.obk.yaml", "record.obk.yaml", "organization.obk.yaml"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("PRESERVE ORIGINAL\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	return root, filepath.Join(root, "created.obk.yaml")
}

func preserved(t *testing.T, root string) {
	t.Helper()
	for _, name := range []string{"preparation.obk.yaml", "record.obk.yaml", "organization.obk.yaml"} {
		body, err := os.ReadFile(filepath.Join(root, name))
		if err != nil || string(body) != "PRESERVE ORIGINAL\n" {
			t.Errorf("unrelated artifact changed: %s", name)
		}
	}
}

func assertReport(t *testing.T, data []byte, status, code string) {
	t.Helper()
	var actual map[string]string
	if err := json.Unmarshal(data, &actual); err != nil {
		t.Fatalf("structured report required: %q", data)
	}
	messages := map[string]string{
		"create.cancelled": "Cancelled. No files changed.", "create.written": "The reviewed organization plan was created.", "create.usage": "Use catalog create --id ID --out FILE.", "create.path": "Choose an absent plan file in an existing safe directory.", "create.input": "Use complete valid input within the displayed limits.", "create.output": "Required output failed before save; no plan was published.", "create.stale": "The destination context changed; reload and review again.",
	}
	want := map[string]string{"status": status, "code": code, "message": messages[code]}
	if !reflect.DeepEqual(actual, want) || !bytes.HasSuffix(data, []byte("\n")) {
		t.Errorf("report differs: %q", data)
	}
}

func TestLiteralCases(t *testing.T) {
	var cases []struct {
		Name             string   `json:"name"`
		Input            string   `json:"input"`
		InputBase64      string   `json:"inputBase64"`
		Exit             int      `json:"exit"`
		Status           string   `json:"status"`
		Code             string   `json:"code"`
		Saved            bool     `json:"saved"`
		ExactEmptyReview bool     `json:"exactEmptyReview"`
		Contains         []string `json:"contains"`
		Consumed         int      `json:"consumed"`
	}
	if err := json.Unmarshal(fixture(t, "cases.json"), &cases); err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root, target := sandbox(t)
			data := []byte(c.Input)
			if c.InputBase64 != "" {
				var err error
				data, err = base64.StdEncoding.DecodeString(c.InputBase64)
				if err != nil {
					t.Fatal(err)
				}
			}
			input := &countedInput{source: data}
			var out, stderr bytes.Buffer
			exit := newCommand().Run([]string{"--id=plan-one", "--out=" + target}, input, &out, &stderr)
			if exit != c.Exit {
				t.Errorf("exit=%d want=%d", exit, c.Exit)
			}
			assertReport(t, stderr.Bytes(), c.Status, c.Code)
			if c.Consumed >= 0 && input.position != c.Consumed {
				t.Errorf("consumed=%d want=%d", input.position, c.Consumed)
			}
			startup := strings.NewReplacer("@ID@", `"plan-one"`, "@OUT@", strconv.Quote(target)).Replace(string(fixture(t, "startup.txt")))
			if !strings.HasPrefix(out.String(), startup) {
				t.Error("exact startup missing")
			}
			if c.ExactEmptyReview {
				review := strings.ReplaceAll(string(fixture(t, "empty-review.txt")), "@OUT@", strconv.Quote(target))
				if out.String() != startup+review {
					t.Errorf("exact empty review differs: %q", out.String())
				}
			}
			for _, text := range c.Contains {
				if !strings.Contains(out.String(), text) {
					t.Errorf("missing literal %q", text)
				}
			}
			body, err := os.ReadFile(target)
			if c.Saved {
				if err != nil || !bytes.Equal(body, fixture(t, "empty-plan.yaml")) {
					t.Errorf("exact saved plan differs: %q err=%v", body, err)
				}
			} else if !os.IsNotExist(err) {
				t.Error("unsaved target exists")
			}
			preserved(t, root)
		})
	}
}

func TestHelpAndCLIAdmission(t *testing.T) {
	for _, args := range [][]string{{"--help", "--invalid"}, {"-h", "--out=/missing"}} {
		input := &countedInput{source: []byte("MUST NOT READ")}
		var out, stderr bytes.Buffer
		if newCommand().Run(args, input, &out, &stderr) != 0 || !bytes.Equal(out.Bytes(), fixture(t, "help.txt")) || stderr.Len() != 0 || input.position != 0 {
			t.Error("help must precede all facts and input")
		}
	}
	for _, args := range [][]string{nil, {"--id=plan-one"}, {"--out=somewhere"}, {"--id=one", "--id=one", "--out=where"}, {"--id=one", "--out=where", "--help"}, {"--id", "--out=where"}, {"--id=one", "--out="}, {"--id=one", "--out=where", "extra"}, {"--id=" + strings.Repeat("x", 4097), "--out=where"}} {
		input := &countedInput{source: []byte("MUST NOT READ")}
		var out, stderr bytes.Buffer
		if newCommand().Run(args, input, &out, &stderr) != 2 || input.position != 0 || out.Len() != 0 {
			t.Error("invalid CLI crossed admission")
		}
		assertReport(t, stderr.Bytes(), "refused", "create.usage")
	}
}

func canonicalSection(t *testing.T, source, id string) map[string]any {
	t.Helper()
	seedID := "repository-a"
	if source == "document-reading" {
		seedID = "reading-a"
	}
	canonical := bytes.TrimSuffix(fixture(t, source+".canonical"), []byte("\n"))
	canonical = bytes.Replace(canonical, []byte(`"id":"`+seedID+`"`), []byte(`"id":"`+id+`"`), 1)
	sum := sha256.Sum256(canonical)
	var section map[string]any
	if err := json.Unmarshal(canonical, &section); err != nil {
		t.Fatal(err)
	}
	section["body"] = section["body"].(string) + "\n"
	for _, raw := range section["variables"].([]any) {
		delete(raw.(map[string]any), "type")
	}
	section["source"] = map[string]any{"template": source, "hash": hex.EncodeToString(sum[:])}
	return section
}

func savedPlan(t *testing.T, path string) map[string]any {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var plan map[string]any
	if err := yaml.Unmarshal(body, &plan); err != nil {
		t.Fatal(err)
	}
	return plan
}

func TestRepeatedRepositoryReadingAndExactScopedFacts(t *testing.T) {
	root, target := sandbox(t)
	input := &countedInput{source: fixture(t, "repeated.txt")}
	var out, stderr bytes.Buffer
	if newCommand().Run([]string{"--out", target, "--id", "plan-one"}, input, &out, &stderr) != 0 {
		t.Fatalf("repeated creation failed: %q", stderr.String())
	}
	assertReport(t, stderr.Bytes(), "written", "create.written")
	want := map[string]any{"schemaVersion": 1, "id": "plan-one", "answers": map[string]any{"sections": map[string]any{
		"repository-a": map[string]any{"repository_scope": "first repository"}, "repository-b": map[string]any{"repository_scope": "deferred"}, "reading-a": map[string]any{"reading_material": "current guide"},
	}}, "sections": []any{canonicalSection(t, "document-reading", "reading-a"), canonicalSection(t, "document-reading", "reading-b"), canonicalSection(t, "repository-access", "repository-a"), canonicalSection(t, "repository-access", "repository-b")}}
	if actual := savedPlan(t, target); !reflect.DeepEqual(actual, want) {
		t.Errorf("whole independently authored saved plan differs: %#v", actual)
	}
	body, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(body, []byte("    repository-b:\n      repository_scope: deferred\n")) {
		t.Error("explicit Deferred was not encoded as its marker kind")
	}
	for _, literal := range []string{
		"Preparation guidance: 3\n", "Guide: \"invitation-prepare\"\nInstance reasons: 4\n  - \"reading-a\"\n  - \"reading-b\"\n  - \"repository-a\"\n  - \"repository-b\"\n", "Guide: \"reading-assign\"\nInstance reasons: 2\n  - \"reading-a\"\n  - \"reading-b\"\n", "Guide: \"resource-access-arrange\"\nInstance reasons: 2\n  - \"repository-a\"\n  - \"repository-b\"\n",
	} {
		if !strings.Contains(out.String(), literal) {
			t.Errorf("advisory reasons differ: %q", literal)
		}
	}
	if input.position != len(input.source) {
		t.Error("complete save consumption differs")
	}
	preserved(t, root)
}

func TestScopedReplacementUnbindAndRejectedChanges(t *testing.T) {
	for _, c := range []struct {
		name, input string
		answers     map[string]any
		problems    []string
	}{
		{"replace", "bind\nrepository-a\nrepository_scope\ntext\nfirst\nbind\nrepository-a\nrepository_scope\ntext\nsecond\n", map[string]any{"repository_scope": "second"}, nil},
		{"literal-cancel", "bind\nrepository-a\nrepository_scope\ntext\ncancel\n", map[string]any{"repository_scope": "cancel"}, nil},
		{"empty-text", "bind\nrepository-a\nrepository_scope\ntext\n\n", map[string]any{"repository_scope": ""}, nil},
		{"unbind", "bind\nrepository-a\nrepository_scope\ntext\nfirst\nunbind\nrepository-a\nrepository_scope\n", nil, nil},
		{"wrong-kind", "bind\nrepository-a\nrepository_scope\nflag\nfalse\n", nil, []string{"Problems: 1\n  - Stage: \"answer\" Path: \"instances[0].bindings[0].value\" Code: \"answer.wrong-kind\" Severity: \"error\" Repair: \"Use a value kind that matches the declared variable type.\"\n"}},
		{"unknown-name", "bind\nrepository-a\nnot_declared\ntext\nvalue\n", nil, []string{"Problems: 1\n  - Stage: \"binding\" Path: \"instances[0].bindings[0].name\" Code: \"materialize.binding-name\" Severity: \"error\" Repair: \"Bind an exact variable declared in this scope.\"\n"}},
		{"duplicate-id", "add\ndocument-reading\nrepository-a\n", nil, []string{"Problems: 1\n  - Stage: \"request\" Path: \"instances[1].permanentSectionId\" Code: \"materialize.duplicate-instance\" Severity: \"error\" Repair: \"Give each independent activity its own permanent section ID.\"\n"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			root, target := sandbox(t)
			var out, stderr bytes.Buffer
			text := "add\nrepository-access\nrepository-a\n" + c.input + "review\nsave\n"
			if newCommand().Run([]string{"--id=plan-one", "--out=" + target}, strings.NewReader(text), &out, &stderr) != 0 {
				t.Fatalf("creation failed: %q", stderr.String())
			}
			actual := savedPlan(t, target)
			want := map[string]any{"schemaVersion": 1, "id": "plan-one", "sections": []any{canonicalSection(t, "repository-access", "repository-a")}}
			if c.answers != nil {
				want["answers"] = map[string]any{"sections": map[string]any{"repository-a": c.answers}}
			}
			if !reflect.DeepEqual(actual, want) {
				t.Errorf("scoped result differs: %#v", actual)
			}
			for _, problem := range c.problems {
				if !strings.Contains(out.String(), problem) {
					t.Errorf("missing actual dependency problem %s", problem)
				}
			}
			preserved(t, root)
		})
	}
}

func TestFailedCurrentReviewDoesNotReadQueuedSave(t *testing.T) {
	for mode := 0; mode < 3; mode++ {
		root, target := sandbox(t)
		input := &countedInput{source: []byte("review\nsave\nTAIL")}
		out := &reviewFailure{mode: mode}
		var stderr bytes.Buffer
		if newCommand().Run([]string{"--id=plan-one", "--out=" + target}, input, out, &stderr) != 1 || input.position != len("review\n") {
			t.Error("failed review consumed queued consent")
		}
		assertReport(t, stderr.Bytes(), "refused", "create.output")
		if _, err := os.Lstat(target); !os.IsNotExist(err) {
			t.Error("failed review published a plan")
		}
		preserved(t, root)
	}
}

func TestInputBoundsAndNoHiddenPartialSave(t *testing.T) {
	for _, input := range [][]byte{[]byte(strings.Repeat("x", 65537)), []byte("review\nsave\xff\n"), []byte(strings.Repeat(strings.Repeat("x", 65535)+"\n", 16) + "review\nsave\n")} {
		root, target := sandbox(t)
		reader := &countedInput{source: input}
		var out, stderr bytes.Buffer
		if newCommand().Run([]string{"--id=plan-one", "--out=" + target}, reader, &out, &stderr) != 1 {
			t.Error("invalid/bounded input was accepted")
		}
		assertReport(t, stderr.Bytes(), "refused", "create.input")
		if reader.position > 1048576 {
			t.Error("read beyond total input bound")
		}
		if _, err := os.Lstat(target); !os.IsNotExist(err) {
			t.Error("invalid input published a plan")
		}
		preserved(t, root)
	}
}
