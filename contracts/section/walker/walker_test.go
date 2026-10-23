package walker_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"

	"onboard-kit/contracts/harness/jsonform"
	contract "onboard-kit/contracts/section/walker"
	"onboard-kit/internal/model"
)

type historyExpectation struct {
	Section    string
	Comparison contract.Comparison
}

type walkCase struct {
	Name, Source string
	Plan, Record json.RawMessage
	Sequence     []string
	Waiting      map[string][]string
	History      []historyExpectation
}

func fixtureCases(t *testing.T) []walkCase {
	t.Helper()
	data, err := os.ReadFile("testdata/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []walkCase
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 15 {
		t.Fatal("expected fifteen independent initial semantic cases")
	}
	return cases
}

func inputs(t *testing.T, c walkCase) (model.Plan, model.Record) {
	t.Helper()
	plan, err := jsonform.Plan(c.Plan)
	if err != nil {
		t.Fatal(err)
	}
	record, err := jsonform.Record(c.Record)
	if err != nil {
		t.Fatal(err)
	}
	return plan, record
}

func TestLiteralActionsWaitingAndHistory(t *testing.T) {
	t.Parallel()
	for _, c := range fixtureCases(t) {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			if c.Name == "" || c.Source == "" || c.Waiting == nil {
				t.Fatal("fixture lacks source or explicit oracle")
			}
			plan, record := inputs(t, c)
			view, problems := newWalker().Walk(plan, record)
			if len(problems) != 0 {
				t.Fatalf("unexpected problems: %#v", problems)
			}
			var ids []string
			for _, row := range view.Sections {
				ids = append(ids, row.Section.ID)
				want, ok := c.Waiting[row.Section.ID]
				if !ok || !slices.Equal(row.WaitingOn, want) {
					t.Fatalf("waiting %q: got %q want %q", row.Section.ID, row.WaitingOn, want)
				}
				if !validDigest(row.CurrentHash) {
					t.Fatalf("missing actual definition hash: %#v", row)
				}
			}
			if !slices.Equal(ids, c.Sequence) {
				t.Fatalf("action order: got %q want %q", ids, c.Sequence)
			}
			if len(view.Outcomes) != len(c.History) {
				t.Fatalf("history: got %#v want %#v", view.Outcomes, c.History)
			}
			for i, want := range c.History {
				got := view.Outcomes[i]
				if got.Outcome.Section != want.Section || got.Comparison != want.Comparison {
					t.Fatalf("history row: got %#v want %#v", got, want)
				}
				var held *model.Outcome
				for j := range record.Outcomes {
					if record.Outcomes[j].Section == want.Section {
						held = &record.Outcomes[j]
					}
				}
				if held == nil || !reflect.DeepEqual(got.Outcome, *held) {
					t.Fatal("lost or rewrote held outcome fields")
				}
				present := false
				for _, section := range plan.Sections {
					if section.ID == want.Section {
						present = true
					}
				}
				if present && !validDigest(got.CurrentHash) || !present && got.CurrentHash != "" {
					t.Fatal("wrong current definition availability")
				}
			}
		})
	}
}

func validDigest(text string) bool {
	if len(text) != 64 {
		return false
	}
	for _, c := range text {
		if !strings.ContainsRune("0123456789abcdef", c) {
			return false
		}
	}
	return true
}

func TestActualHeldDefinitionHashNotSource(t *testing.T) {
	t.Parallel()
	const canonical = `{"id":"access","title":"Access","summary":"Request access","timing":"anytime"}`
	const digest = "6c0301d995d0a3b0dab95a641a087120c5266746e59c50d934c575e135ca2fdc"
	sum := sha256.Sum256([]byte(canonical))
	if hex.EncodeToString(sum[:]) != digest {
		t.Fatal("independent canonical oracle is inconsistent")
	}
	section := model.Section{ID: "access", Title: "Access", Summary: "Request access", Source: &model.Source{Hash: "template-baseline"}}
	plan := model.Plan{Sections: []model.Section{section}}
	record := model.Record{Outcomes: []model.Outcome{{Section: "access", Kind: model.OutcomeVerified, SectionHash: digest}}}
	view, problems := newWalker().Walk(plan, record)
	if len(problems) != 0 || len(view.Sections) != 1 || len(view.Outcomes) != 1 {
		t.Fatalf("missing view: %#v %#v", view, problems)
	}
	if view.Sections[0].CurrentHash != digest || view.Outcomes[0].CurrentHash != digest || view.Outcomes[0].Comparison != contract.ComparisonSame {
		t.Fatal("actual hash or exact equality wrong")
	}
	plan.Sections[0].Source = nil
	view, problems = newWalker().Walk(plan, record)
	if len(problems) != 0 || len(view.Sections) != 1 || view.Sections[0].CurrentHash != digest || len(view.Outcomes) != 1 || view.Outcomes[0].Comparison != contract.ComparisonSame {
		t.Fatal("custom section was skipped")
	}
	plan.Sections[0].Summary = "Request access through help desk"
	const editedCanonical = `{"id":"access","title":"Access","summary":"Request access through help desk","timing":"anytime"}`
	const editedDigest = "423c3fce4113082df2b87850806c7a035c56ab76890942e350cf91e27f750dd7"
	sum = sha256.Sum256([]byte(editedCanonical))
	if hex.EncodeToString(sum[:]) != editedDigest {
		t.Fatal("independent edited oracle is inconsistent")
	}
	view, problems = newWalker().Walk(plan, record)
	if len(problems) != 0 || len(view.Sections) != 1 || view.Sections[0].CurrentHash != editedDigest || len(view.Outcomes) != 1 || view.Outcomes[0].Comparison != contract.ComparisonDifferent {
		t.Fatal("manual definition edit escaped comparison")
	}
}

func TestDuplicateOutcomesRefusedIncludingRemovedAndInactive(t *testing.T) {
	t.Parallel()
	plan := model.Plan{Sections: []model.Section{{ID: "a"}}}
	record := model.Record{Outcomes: []model.Outcome{{Section: "removed", Kind: model.OutcomeVerified}, {Section: "a", Kind: model.OutcomeVerified}, {Section: "removed", Kind: model.OutcomeFailed}, {Section: "a", Kind: model.OutcomeVerified}}}
	want := []model.Problem{
		{Code: "walk.duplicate-outcome", Severity: model.SeverityError, Section: "a", Message: "Keep one recorded outcome for this section.", Detail: map[string]string{"field": "outcomes"}},
		{Code: "walk.duplicate-outcome", Severity: model.SeverityError, Section: "removed", Message: "Keep one recorded outcome for this section.", Detail: map[string]string{"field": "outcomes"}},
	}
	for range 2 {
		view, problems := newWalker().Walk(plan, record)
		if !reflect.DeepEqual(problems, want) || len(view.Sections)+len(view.Outcomes)+len(view.Questions)+len(view.Blockers) != 0 {
			t.Fatalf("ambiguous record not refused: %#v %#v", view, problems)
		}
		slices.Reverse(record.Outcomes)
	}
	off := false
	plan.Sections[0].Enabled = &off
	view, problems := newWalker().Walk(plan, record)
	if !reflect.DeepEqual(problems, want) || len(view.Sections)+len(view.Outcomes)+len(view.Questions)+len(view.Blockers) != 0 {
		t.Fatal("inactive outcome ambiguity was hidden")
	}
}

func TestLiteralQuestionsResolutionAndSupport(t *testing.T) {
	t.Parallel()
	const raw = `{
		"variables":[{"name":"team","type":"string"},{"name":"squad","type":"team"}],
		"answers":{"plan":{"team":"not-a-team","squad":"platform"},"sections":{"a":{"grantor":"support"}}},
		"directory":{"contacts":[{"id":"desk","name":"Support desk","slack":"#support"}],"roles":[{"id":"support","name":"Support","holders":["desk"]}],"teams":[{"id":"platform","name":"Platform","holders":{"support":[]}}]},
		"sections":[{"id":"a","owner":"$grantor","body":"Sign in as $account.","escalation":["$grantor","$backup","$grantor"],"variables":[{"name":"grantor","type":"role"},{"name":"backup","type":"contact","required":true},{"name":"account","audience":"onboardee","required":true,"default":"suggested"}]}]
	}`
	plan, err := jsonform.Plan([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	view, problems := newWalker().Walk(plan, model.Record{})
	if len(view.Sections) != 1 {
		t.Fatal("missing active row")
	}
	row := view.Sections[0]
	if row.Section.Owner != "Support (vacant)" || row.Section.Body != "Sign in as <account>." || !slices.Equal(row.Escalation, []string{"Support (vacant)", "_unanswered_", "Support (vacant)"}) {
		t.Fatalf("wrong real team/scope/default/escalation resolution: %#v", row)
	}
	if !slices.Equal(row.Section.Escalation, []string{"$grantor", "$backup", "$grantor"}) {
		t.Fatal("rewrote definition escalation")
	}
	var questionIDs, blockerIDs []string
	for _, q := range view.Questions {
		questionIDs = append(questionIDs, q.ID)
		if q.ID == "a.account" && (q.Answer.Kind != model.ValueKindNone || q.DeclaredDefault.Kind != model.ValueKindText || q.DeclaredDefault.Text != "suggested" || q.Audience != model.AudienceOnboardee) {
			t.Fatal("default became accepted answer or lost ownership")
		}
	}
	for _, blocker := range view.Blockers {
		blockerIDs = append(blockerIDs, blocker.Question)
	}
	if !slices.Equal(questionIDs, []string{"plan.team", "plan.squad", "a.grantor", "a.backup", "a.account"}) || !slices.Equal(blockerIDs, []string{"a.backup", "a.account"}) {
		t.Fatalf("both-audience full finder lists wrong: %q %q", questionIDs, blockerIDs)
	}
	if len(problems) != 2 {
		t.Fatalf("expected unique main/escalation gaps: %#v", problems)
	}
	for i, want := range []struct{ question, audience string }{{"a.backup", "onboarder"}, {"a.account", "onboardee"}} {
		p := problems[i]
		if p.Code != "gap.unanswered" || p.Severity != model.SeverityWarning || p.Section != "a" || p.Question != want.question || p.Detail["audience"] != want.audience || strings.TrimSpace(p.Message) == "" {
			t.Fatalf("wrong useful gap: %#v", p)
		}
	}
}

func TestStructuralRefusalPrecedesOutcomeAmbiguity(t *testing.T) {
	t.Parallel()
	plan := model.Plan{Sections: []model.Section{{ID: "a", Requires: []string{"private-target"}}}}
	record := model.Record{Outcomes: []model.Outcome{{Section: "a", Kind: model.OutcomeVerified}, {Section: "a", Kind: model.OutcomeFailed, Note: "private-note"}}}
	want := []model.Problem{{Code: "plan.requires-unknown", Severity: model.SeverityError, Message: "Choose an existing prerequisite, or remove this requires item.", Section: "a", Detail: map[string]string{"field": "requires", "item": "0"}}}
	view, problems := newWalker().Walk(plan, record)
	if !reflect.DeepEqual(problems, want) || len(view.Sections)+len(view.Outcomes)+len(view.Questions)+len(view.Blockers) != 0 {
		t.Fatalf("wrong structural refusal precedence: %#v %#v", view, problems)
	}
}

func TestPurityStateAndPermutation(t *testing.T) {
	t.Parallel()
	shared := newWalker()
	for _, c := range fixtureCases(t) {
		plan, record := inputs(t, c)
		beforePlan, beforeRecord := inputs(t, c)
		first, firstProblems := shared.Walk(plan, record)
		if !reflect.DeepEqual(plan, beforePlan) || !reflect.DeepEqual(record, beforeRecord) {
			t.Fatal("walk changed arguments")
		}
		_, _ = shared.Walk(model.Plan{}, model.Record{})
		slices.Reverse(plan.Sections)
		slices.Reverse(record.Outcomes)
		again, againProblems := shared.Walk(plan, record)
		if !reflect.DeepEqual(first, again) || !reflect.DeepEqual(firstProblems, againProblems) {
			t.Fatalf("state or declaration order affected %s", c.Name)
		}
	}
}
