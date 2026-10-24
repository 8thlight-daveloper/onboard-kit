package creator_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type creatorFollowupOutput struct {
	storage bytes.Buffer
	failed  bool
	marker  string
}

func (writer *creatorFollowupOutput) Write(body []byte) (int, error) {
	if writer.failed {
		return 0, errors.New("owned followup prompt fault")
	}
	if len(body) > 2097152-writer.storage.Len() {
		return 0, errors.New("owned followup output cap")
	}
	if writer.marker != "" && bytes.Contains(append(append([]byte(nil), writer.storage.Bytes()...), body...), []byte(writer.marker)) {
		writer.failed = true
		return 0, errors.New("owned followup prompt fault")
	}
	return writer.storage.Write(body)
}

func creatorFollowupAbsent(t *testing.T, root, target string) {
	t.Helper()
	if _, err := os.Lstat(target); !os.IsNotExist(err) {
		t.Error("unconsented target exists")
	}
	preserved(t, root)
}

func creatorFollowupRun(t *testing.T, root, target, actions string) string {
	t.Helper()
	input := &countedInput{source: []byte(actions)}
	out, stderr := &creatorFollowupOutput{}, &creatorFollowupOutput{}
	if newCommand().Run([]string{"--id=plan-one", "--out=" + target}, input, out, stderr) != 0 {
		t.Fatalf("supported followup interaction refused: %q", stderr.storage.String())
	}
	assertReport(t, stderr.storage.Bytes(), "written", "create.written")
	if input.position != len(actions)-len("TAIL") {
		t.Error("successful save consumed queued tail or skipped required fields")
	}
	preserved(t, root)
	return out.storage.String()
}

func creatorFollowupLocal(t *testing.T, target, id string) map[string]any {
	t.Helper()
	answers, ok := savedPlan(t, target)["answers"].(map[string]any)
	if !ok {
		t.Fatal("explicit scoped answers absent")
	}
	sections, ok := answers["sections"].(map[string]any)
	if !ok {
		t.Fatal("explicit section answer scopes absent")
	}
	local, ok := sections[id].(map[string]any)
	if !ok {
		t.Fatal("explicit selected answer scope absent")
	}
	return local
}

func TestCreatorFollowupCLIExactGrammarAndDecodedText(t *testing.T) {
	for _, kind := range []string{"alias", "separate-option-value", "invalid-id-utf8", "nul-id", "invalid-path-utf8", "nul-path"} {
		t.Run(kind, func(t *testing.T) {
			root, target := sandbox(t)
			args := []string{"--id=plan-one", "--out=" + target}
			switch kind {
			case "alias":
				args[0] = "--plan=plan-one"
			case "separate-option-value":
				args = []string{"--id", "--opaque", "--out=" + target}
			case "invalid-id-utf8":
				args[0] = "--id=" + string([]byte{255})
			case "nul-id":
				args[0] = "--id=opaque\x00suffix"
			case "invalid-path-utf8":
				args[1] = "--out=" + target + string([]byte{255})
			case "nul-path":
				args[1] = "--out=" + target + "\x00suffix"
			}
			input := &countedInput{source: []byte("cancel\nTAIL")}
			out, stderr := &creatorFollowupOutput{}, &creatorFollowupOutput{}
			if newCommand().Run(args, input, out, stderr) != 2 || input.position != 0 || out.storage.Len() != 0 {
				t.Error("invalid CLI crossed usage boundary")
			}
			assertReport(t, stderr.storage.Bytes(), "refused", "create.usage")
			creatorFollowupAbsent(t, root, target)
		})
	}
	for _, id := range []string{strings.Repeat("x", 4096), "--opaque"} {
		root, target := sandbox(t)
		input := &countedInput{source: []byte("cancel\nTAIL")}
		out, stderr := &creatorFollowupOutput{}, &creatorFollowupOutput{}
		if newCommand().Run([]string{"--id=" + id, "--out=" + target}, input, out, stderr) != 0 || input.position != len("cancel\n") {
			t.Error("inclusive ID or literal first-equals value was refused")
		}
		assertReport(t, stderr.storage.Bytes(), "cancelled", "create.cancelled")
		creatorFollowupAbsent(t, root, target)
	}
}

func TestCreatorFollowupActionsRequireExactConsent(t *testing.T) {
	for _, action := range []string{"", " save ", "SAVE"} {
		t.Run("review-"+action, func(t *testing.T) {
			root, target := sandbox(t)
			input := &countedInput{source: []byte("review\n" + action + "\ncancel\nTAIL")}
			out, stderr := &creatorFollowupOutput{}, &creatorFollowupOutput{}
			if newCommand().Run([]string{"--id=plan-one", "--out=" + target}, input, out, stderr) != 0 || input.position != len(input.source)-len("TAIL") {
				t.Error("nonexact or empty review action became consent")
			}
			assertReport(t, stderr.storage.Bytes(), "cancelled", "create.cancelled")
			if !strings.HasSuffix(out.storage.String(), "Choose one of the displayed actions.\nReview complete. Choose save, back, or cancel:\n") {
				t.Error("unknown review action did not repeat only the review prompt")
			}
			creatorFollowupAbsent(t, root, target)
		})
	}
}

func TestCreatorFollowupLiteralTextIsNotTrimmedOrCasefolded(t *testing.T) {
	root, target := sandbox(t)
	actions := "add\nrepository-access\nrepo-a\nbind\nrepo-a\nrepository_scope\ntext\n  ReadMe shared scope  \nreview\nsave\nTAIL"
	creatorFollowupRun(t, root, target, actions)
	if creatorFollowupLocal(t, target, "repo-a")["repository_scope"] != "  ReadMe shared scope  " {
		t.Error("literal mixed-case surrounding-space fact changed")
	}
}

func TestCreatorFollowupLogicalLineEdgesAndPositivePartialEOF(t *testing.T) {
	for _, kind := range []string{"exact-crlf", "excess-partial", "positive-partial-save"} {
		t.Run(kind, func(t *testing.T) {
			root, target := sandbox(t)
			out, stderr := &creatorFollowupOutput{}, &creatorFollowupOutput{}
			switch kind {
			case "exact-crlf":
				input := &countedInput{source: []byte(strings.Repeat("x", 65536) + "\r\ncancel\nTAIL")}
				if newCommand().Run([]string{"--id=plan-one", "--out=" + target}, input, out, stderr) != 0 || input.position != 65536+2+len("cancel\n") {
					t.Error("exact-limit CRLF payload was not accepted as one unknown action")
				}
				assertReport(t, stderr.storage.Bytes(), "cancelled", "create.cancelled")
			case "excess-partial":
				input := &countedInput{source: []byte(strings.Repeat("x", 65537) + "UNREAD")}
				if newCommand().Run([]string{"--id=plan-one", "--out=" + target}, input, out, stderr) != 1 || input.position != 65537 {
					t.Error("oversize partial line did not stop at the first excess non-CR payload byte")
				}
				assertReport(t, stderr.storage.Bytes(), "refused", "create.input")
			case "positive-partial-save":
				input := &finalFaultReader{source: []byte("review\nsave"), final: io.EOF}
				if newCommand().Run([]string{"--id=plan-one", "--out=" + target}, input, out, stderr) != 0 || input.position != len(input.source) {
					t.Error("positive EOF completed an unterminated save")
				}
				assertReport(t, stderr.storage.Bytes(), "cancelled", "create.cancelled")
			}
			creatorFollowupAbsent(t, root, target)
		})
	}
}

func TestCreatorFollowupRequiredSourcePromptFaultLeavesFieldUnread(t *testing.T) {
	root, target := sandbox(t)
	input := &countedInput{source: []byte("add\nrepository-access\nrepo-a\nreview\nsave\n")}
	out := &creatorFollowupOutput{marker: "Source ID:\n"}
	stderr := &creatorFollowupOutput{}
	if newCommand().Run([]string{"--id=plan-one", "--out=" + target}, input, out, stderr) != 1 || !out.failed || input.position != len("add\n") {
		t.Error("failed source prompt consumed a field")
	}
	assertReport(t, stderr.storage.Bytes(), "refused", "create.output")
	creatorFollowupAbsent(t, root, target)
}

func TestCreatorFollowupPromptFaultFixtureCannotBypassWriteString(t *testing.T) {
	for _, stringPath := range []bool{false, true} {
		writer := &creatorFollowupOutput{marker: "Source ID:\n"}
		write := func(text string) (int, error) {
			if stringPath {
				return io.WriteString(writer, text)
			}
			return writer.Write([]byte(text))
		}
		if n, err := write("Source "); n != len("Source ") || err != nil {
			t.Fatal("owned safe prompt prefix refused")
		}
		if n, err := write("ID:\n"); n != 0 || err == nil || !writer.failed || writer.storage.String() != "Source " {
			t.Fatal("owned prompt fault not reached")
		}
	}
}

func TestCreatorFollowupRejectedSelectorsStayLiteralAndAtomic(t *testing.T) {
	for _, kind := range []string{"source-spaces", "mixed-id"} {
		t.Run(kind, func(t *testing.T) {
			root, target := sandbox(t)
			source, id, code := " repository-access ", "repo-a", "catalog.unknown-id"
			if kind == "mixed-id" {
				source, id, code = "repository-access", "Repo-A", "materialize.instance-id"
			}
			out := creatorFollowupRun(t, root, target, "add\n"+source+"\n"+id+"\nreview\nsave\nTAIL")
			body, err := os.ReadFile(target)
			if err != nil || !bytes.Equal(body, fixture(t, "empty-plan.yaml")) {
				t.Error("rejected selector silently repaired or installed")
			}
			if !strings.Contains(out, "Change not applied.\n") || !strings.Contains(out, "Code: \""+code+"\"") {
				t.Error("exact selector refusal absent")
			}
		})
	}
}

func TestCreatorFollowupUnknownInstanceDoesNotConsumeVariable(t *testing.T) {
	for _, action := range []string{"bind", "unbind"} {
		root, target := sandbox(t)
		input := &countedInput{source: []byte(action + "\nmissing\ncancel\nTAIL")}
		out, stderr := &creatorFollowupOutput{}, &creatorFollowupOutput{}
		if newCommand().Run([]string{"--id=plan-one", "--out=" + target}, input, out, stderr) != 0 || input.position != len(action+"\nmissing\ncancel\n") {
			t.Error("unknown instance consumed next draft action as a variable")
		}
		assertReport(t, stderr.storage.Bytes(), "cancelled", "create.cancelled")
		if strings.Contains(out.storage.String(), "Variable name:") || !strings.Contains(out.storage.String(), "Select an existing draft instance.\nDraft — not saved.") {
			t.Error("unknown instance did not return directly to Draft")
		}
		creatorFollowupAbsent(t, root, target)
	}
}

func TestCreatorFollowupRemoveOnePreservesLaterBoundInstances(t *testing.T) {
	root, target := sandbox(t)
	actions := "add\nrepository-access\nrepo-a\nadd\nrepository-access\nrepo-b\nbind\nrepo-b\nrepository_scope\ntext\nRetained later scope\nremove\nrepo-a\nreview\nsave\nTAIL"
	creatorFollowupRun(t, root, target, actions)
	plan := savedPlan(t, target)
	sections, ok := plan["sections"].([]any)
	if !ok || len(sections) != 1 || sections[0].(map[string]any)["id"] != "repo-b" {
		t.Error("removal discarded later selection")
	}
	if creatorFollowupLocal(t, target, "repo-b")["repository_scope"] != "Retained later scope" {
		t.Error("removal discarded later independent binding")
	}
}

func TestCreatorFollowupDeclarationsUseSelectedFreshResultAndEffectiveType(t *testing.T) {
	root, target := sandbox(t)
	actions := "add\nrepository-access\nrepo-a\nadd\ndocument-reading\nread-a\nbind\nrepo-a\nrepository_scope\ntext\nshared scope\nreview\nsave\nTAIL"
	out := creatorFollowupRun(t, root, target, actions)
	want := "Variables: 2\n  - Name: \"repository_scope\" Type: \"string\" Audience: \"onboarder\" Required: true\n  - Name: \"repository_grant_route\" Type: \"string\" Audience: \"onboarder\" Required: true\nVariable name:\n"
	if !strings.Contains(out, want) || strings.Count(out, "Variables:") != 1 || strings.Contains(out, "Name: \"reading_material\"") {
		t.Error("bind declarations were stale, incomplete or from another scope")
	}
}

func TestCreatorFollowupUnbindNamesAreExact(t *testing.T) {
	root, target := sandbox(t)
	actions := "add\nrepository-access\nrepo-a\nbind\nrepo-a\nrepository_scope\ntext\nretained exact binding\nunbind\nrepo-a\nREPOSITORY_SCOPE\nreview\nsave\nTAIL"
	out := creatorFollowupRun(t, root, target, actions)
	if !strings.Contains(out, "Select an existing draft binding.\n") || creatorFollowupLocal(t, target, "repo-a")["repository_scope"] != "retained exact binding" {
		t.Error("case-different unbind removed an exact held name")
	}
}

func TestCreatorFollowupFlagAndListRetriesHaveExactGrammar(t *testing.T) {
	for _, kind := range []string{"flag", "list"} {
		t.Run(kind, func(t *testing.T) {
			root, target := sandbox(t)
			suffix := "flag\nTRUE\nfalse\n"
			want := "Flag value: true or false:\nUse true or false.\nFlag value: true or false:\n"
			if kind == "list" {
				suffix = "list\n+0\n00\n 0\n0\n"
				want = "List length: 0-4096:\nUse a canonical count from 0 through 4096.\nList length: 0-4096:\nUse a canonical count from 0 through 4096.\nList length: 0-4096:\nUse a canonical count from 0 through 4096.\nList length: 0-4096:\n"
			}
			out := creatorFollowupRun(t, root, target, "add\nrepository-access\nrepo-a\nbind\nrepo-a\nrepository_scope\n"+suffix+"review\nsave\nTAIL")
			if !strings.Contains(out, want) {
				t.Error("invalid value spelling did not repeat exact current prompt")
			}
		})
	}
}

func TestCreatorFollowupRepeatedEmptyContactItemsStayExplicit(t *testing.T) {
	root, target := sandbox(t)
	out := creatorFollowupRun(t, root, target, "add\nonboarding-help\nhelp-a\nbind\nhelp-a\ngeneral_contacts\nlist\n3\n\n\n\nreview\nsave\nTAIL")
	if !reflect.DeepEqual(creatorFollowupLocal(t, target, "help-a")["general_contacts"], []any{"", "", ""}) {
		t.Error("literal empty repeated list items were dropped, deduplicated or converted")
	}
	if !strings.Contains(out, "List item 1:\nList item 2:\nList item 3:\nChange staged.\n") {
		t.Error("not every explicit list item was read before adoption")
	}
}

func TestCreatorFollowupRejectedListRetainsOrderedSecretRepairs(t *testing.T) {
	root, target := sandbox(t)
	out := creatorFollowupRun(t, root, target, "add\nonboarding-help\nhelp-a\nbind\nhelp-a\ngeneral_contacts\nlist\n2\n-----BEGIN PRIVATE KEY-----\nhttps://fixture:fictional-password@invalid.example/\nreview\nsave\nTAIL")
	first, second := strings.Index(out, "Code: \"secret.private-key\""), strings.Index(out, "Code: \"secret.url-credential\"")
	if first < 0 || second <= first {
		t.Error("entered list order was not preserved through actual complete diagnostic batch")
	}
	if strings.Count(out, "Repair: \"Remove credential-like content; supply a nonsecret organization fact.\"") != 2 {
		t.Error("actual secret code-prefix projection was lost")
	}
	if strings.Contains(out, "-----BEGIN PRIVATE KEY-----") || strings.Contains(out, "fictional-password") {
		t.Error("rejected literal private fixture text leaked")
	}
	if _, ok := savedPlan(t, target)["answers"]; ok {
		t.Error("rejected list was installed")
	}
}

func TestCreatorFollowupUnsafeUnicodeUsesLiteralIndependentSurrogateOrder(t *testing.T) {
	root, target := sandbox(t)
	id := "label\u202e\u2028\u2029\U000e0001"
	out, stderr := &creatorFollowupOutput{}, &creatorFollowupOutput{}
	if newCommand().Run([]string{"--id=" + id, "--out=" + target}, strings.NewReader("review\nsave\n"), out, stderr) != 0 {
		t.Fatal("valid opaque Unicode ID refused")
	}
	assertReport(t, stderr.storage.Bytes(), "written", "create.written")
	if !strings.Contains(out.storage.String(), `Plan: "label\u202e\u2028\u2029\udb40\udc01"`+"\n") {
		t.Error("unsafe Unicode display or supplementary surrogate order differs")
	}
	for _, r := range []rune{'\u202e', '\u2028', '\u2029', '\U000e0001'} {
		if strings.ContainsRune(out.storage.String(), r) {
			t.Error("unsafe terminal rune emitted literally")
		}
	}
	if savedPlan(t, target)["id"] != id {
		t.Error("display quotation altered stored opaque identity")
	}
	preserved(t, root)
}

func TestCreatorFollowupTrailingAndIntermediateSymlinkPathsRemainRefusals(t *testing.T) {
	for _, kind := range []string{"trailing", "intermediate-symlink"} {
		t.Run(kind, func(t *testing.T) {
			root, target := sandbox(t)
			if kind == "trailing" {
				target = root + string(os.PathSeparator)
			} else {
				real := filepath.Join(root, "real")
				child := filepath.Join(real, "child")
				if err := os.Mkdir(real, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(child, 0700); err != nil {
					t.Fatal(err)
				}
				alias := filepath.Join(root, "alias")
				if err := os.Symlink(real, alias); err != nil {
					t.Fatal(err)
				}
				info, err := os.Lstat(alias)
				if err != nil || info.Mode()&os.ModeSymlink == 0 {
					t.Fatal("actual owned symlink control missing")
				}
				target = filepath.Join(alias, "child", "new.yaml")
			}
			input := &countedInput{source: []byte("cancel\nTAIL")}
			out, stderr := &creatorFollowupOutput{}, &creatorFollowupOutput{}
			if newCommand().Run([]string{"--id=plan-one", "--out=" + target}, input, out, stderr) != 1 || input.position != 0 || out.storage.Len() != 0 {
				t.Error("unsafe native path crossed initial admission")
			}
			assertReport(t, stderr.storage.Bytes(), "refused", "create.path")
			preserved(t, root)
		})
	}
}
