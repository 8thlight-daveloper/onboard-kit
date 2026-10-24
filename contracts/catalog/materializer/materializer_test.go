package materializer_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	api "onboard-kit/contracts/catalog/materializer"
	"onboard-kit/internal/model"
)

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	b, e := os.ReadFile("testdata/" + name)
	if e != nil {
		t.Fatal(e)
	}
	return b
}

func snapshot(t *testing.T, v any) []byte {
	t.Helper()
	b, e := json.Marshal(v)
	if e != nil {
		t.Fatal(e)
	}
	return b
}

func base() api.Request {
	return api.Request{PlanID: "organization-plan", Provenance: model.Provenance{TemplateSet: "reviewed-set", GeneratedOn: "2026-10-24"}}
}

func noErrors(t *testing.T, ds []api.Diagnostic) {
	t.Helper()
	for _, d := range ds {
		if d.Problem.Severity == model.SeverityError {
			t.Fatalf("unexpected %s error %s at %s", d.Stage, d.Problem.Code, d.Path)
		}
	}
}

func own(t *testing.T, got api.Result, ds []api.Diagnostic, stage, code, path, message string) {
	t.Helper()
	want := []api.Diagnostic{{Stage: stage, Path: path, Problem: model.Problem{Code: code, Severity: model.SeverityError, Message: message, Detail: map[string]string{"path": path}}}}
	if !reflect.DeepEqual(got, api.Result{}) || !reflect.DeepEqual(ds, want) {
		t.Errorf("own refusal differs: %#v", ds)
	}
}

func delegated(t *testing.T, got api.Result, ds []api.Diagnostic, stage, code string) {
	t.Helper()
	if !reflect.DeepEqual(got, api.Result{}) {
		t.Error("delegated refusal returned partial result")
	}
	found := false
	for _, d := range ds {
		if d.Stage == stage && d.Problem.Code == code && d.Problem.Severity == model.SeverityError && d.Problem.Message != "" {
			found = true
		}
	}
	if !found {
		t.Errorf("missing actual %s/%s error: %#v", stage, code, ds)
	}
}

func expectedPlan(r api.Request) model.Plan {
	return model.Plan{SchemaVersion: 1, ID: r.PlanID, Provenance: r.Provenance, Directory: model.Directory{Contacts: []model.Contact{}, Roles: []model.Role{}, Teams: []model.Team{}}, Variables: []model.Variable{}, Answers: model.Answers{Plan: map[string]model.Value{}, Sections: map[string]map[string]model.Value{}}, Sections: []model.Section{}}
}

func expectedRepository(t *testing.T, id string) model.Section {
	t.Helper()
	name := "repository-a.canonical"
	if id == "repo-b" {
		name = "repository-b.canonical"
	}
	canonical := fixture(t, name)
	if len(canonical) == 0 || canonical[len(canonical)-1] != '\n' {
		t.Fatal("canonical fixture requires one final framing LF")
	}
	canonical = canonical[:len(canonical)-1]
	sum := sha256.Sum256(canonical)
	return model.Section{ID: id, Title: "Repository access", Summary: "Use the required capability for that specific repository", Timing: model.TimingAnytime, Body: string(fixture(t, "repository.body.txt")), Variables: []model.Variable{
		{Name: "repository_scope", Description: "Which repository and exact capability does this activity cover?", Audience: model.AudienceOnboarder, Required: true},
		{Name: "repository_grant_route", Description: "Who grants this capability and what safe verification procedure applies?", Audience: model.AudienceOnboarder, Required: true},
	}, Steps: []model.Step{{ID: "verify-repository-capability", Name: "Verify the specific approved repository capability", Checks: []model.Check{{Ask: "Have you used the approved capability in the intended repository without exceeding its permission scope?"}}}}, Source: &model.Source{Template: "repository-access", Hash: hex.EncodeToString(sum[:])}}
}

type guideOracle struct {
	ID      string   `json:"id"`
	Reasons []string `json:"reasons"`
}

func guides(t *testing.T, got []api.Guidance, want []guideOracle) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("guidance count: got %d want %d", len(got), len(want))
	}
	for i, g := range got {
		if g.ID != want[i].ID || !reflect.DeepEqual(g.InstanceIDs, want[i].Reasons) {
			t.Errorf("guide/reason ordering differs at %d", i)
		}
		b, e := os.ReadFile("../reader/content/preparation/" + want[i].ID + ".md")
		if e != nil {
			t.Fatal(e)
		}
		if !bytes.Equal(g.Markdown, b) {
			t.Errorf("guide %s content changed", want[i].ID)
		}
	}
}

func TestEmptySelectionIsAnExplicitNewPlan(t *testing.T) {
	r := base()
	r.PlanID = " Reviewed Unicode plan Ω "
	got, ds := newMaterializer().Create(r)
	noErrors(t, ds)
	want := api.Result{Plan: expectedPlan(r), Guidance: []api.Guidance{}}
	if !reflect.DeepEqual(got, want) {
		t.Error("empty selection added defaults, omitted identity, or changed exact new-container shape")
	}
}

func TestFullRepeatedRepositoryBaselineAndScopedFacts(t *testing.T) {
	r := base()
	r.PlanVariables = []model.Variable{{Name: "shared_fact", Default: model.Value{Kind: model.ValueKindText, Text: "suggested only"}}}
	r.Instances = []api.Instance{
		{SourceID: "repository-access", PermanentSectionID: "repo-b"},
		{SourceID: "repository-access", PermanentSectionID: "repo-a", Bindings: []api.Binding{{Name: "repository_scope", Value: model.Value{Kind: model.ValueKindText, Text: "  approved independent capability  "}}}},
	}
	before := snapshot(t, r)
	got, ds := newMaterializer().Create(r)
	noErrors(t, ds)
	want := expectedPlan(r)
	want.Variables = r.PlanVariables
	want.Sections = []model.Section{expectedRepository(t, "repo-a"), expectedRepository(t, "repo-b")}
	want.Answers.Sections["repo-a"] = map[string]model.Value{"repository_scope": r.Instances[1].Bindings[0].Value}
	if !reflect.DeepEqual(got.Plan, want) {
		t.Error("full unresolved plan, Source baseline, preserved step IDs or explicit scoped values differ")
	}
	guides(t, got.Guidance, []guideOracle{{"invitation-prepare", []string{"repo-a", "repo-b"}}, {"resource-access-arrange", []string{"repo-a", "repo-b"}}})
	if !bytes.Equal(snapshot(t, r), before) {
		t.Error("request mutated")
	}
	if len(got.Plan.Sections) == 2 && got.Plan.Sections[0].Source.Hash == got.Plan.Sections[1].Source.Hash {
		t.Error("final permanent identity was absent from the baseline")
	}
}

func TestSelectionReorderAdditionAndOmissionDoNotRenameHeldInstances(t *testing.T) {
	r := base()
	r.Instances = []api.Instance{{SourceID: "repository-access", PermanentSectionID: "repo-b"}, {SourceID: "repository-access", PermanentSectionID: "repo-a"}}
	m := newMaterializer()
	first, ds := m.Create(r)
	noErrors(t, ds)
	r.Instances[0], r.Instances[1] = r.Instances[1], r.Instances[0]
	second, ds := m.Create(r)
	noErrors(t, ds)
	if !reflect.DeepEqual(first, second) || len(first.Plan.Sections) != 2 {
		t.Fatal("request ordering changed the selected proposal")
	}
	r.Instances = append(r.Instances, api.Instance{SourceID: "document-reading", PermanentSectionID: "zz-reading"})
	third, ds := m.Create(r)
	noErrors(t, ds)
	if len(third.Plan.Sections) != 3 || !reflect.DeepEqual(third.Plan.Sections[:2], first.Plan.Sections) {
		t.Error("adding independent instance changed repository identities or baselines")
	}
	r.Instances = []api.Instance{{SourceID: "repository-access", PermanentSectionID: "repo-b"}}
	fourth, ds := m.Create(r)
	noErrors(t, ds)
	if len(fourth.Plan.Sections) != 1 || !reflect.DeepEqual(fourth.Plan.Sections[0], first.Plan.Sections[1]) {
		t.Error("new smaller proposal renamed remaining instance")
	}
}

func TestEveryCurrentSectionCarrierCanRepeat(t *testing.T) {
	rows := strings.Split(strings.TrimSuffix(string(fixture(t, "relations.tsv")), "\n"), "\n")
	if len(rows) != 60 {
		t.Fatal("independent complete59-row oracle changed")
	}
	for _, line := range rows[1:] {
		fields := strings.Split(line, "\t")
		if len(fields) != 5 {
			t.Fatal("invalid literal relation row")
		}
		t.Run(fields[0], func(t *testing.T) {
			r := base()
			r.Instances = []api.Instance{{SourceID: fields[0], PermanentSectionID: fields[2]}, {SourceID: fields[0], PermanentSectionID: fields[1]}}
			got, ds := newMaterializer().Create(r)
			noErrors(t, ds)
			if len(got.Plan.Sections) != 2 {
				t.Fatal("did not materialize both selected instances")
			}
			for i, id := range []string{fields[1], fields[2]} {
				s := got.Plan.Sections[i]
				if s.ID != id || s.Source == nil || s.Source.Template != fields[0] || len(s.Source.Hash) != 64 || s.Title == "" || s.Summary == "" || s.Body == "" || s.Enabled != nil || s.Parent != "" || len(s.Requires) != 0 || len(s.After) != 0 {
					t.Errorf("section %d identity, content or graph-free projection differs", i)
				}
			}
			want := []guideOracle{}
			for _, id := range strings.Split(fields[3], ",") {
				want = append(want, guideOracle{id, strings.Split(fields[4], ",")})
			}
			guides(t, got.Guidance, want)
		})
	}
}

func TestLiteralMixedGuidanceUnions(t *testing.T) {
	var cases []struct {
		Name      string `json:"name"`
		Instances []struct {
			Source string `json:"source"`
			ID     string `json:"id"`
		} `json:"instances"`
		Guidance []guideOracle `json:"guidance"`
	}
	if e := json.Unmarshal(fixture(t, "unions.json"), &cases); e != nil {
		t.Fatal(e)
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			r := base()
			for _, i := range c.Instances {
				r.Instances = append(r.Instances, api.Instance{SourceID: i.Source, PermanentSectionID: i.ID})
			}
			got, ds := newMaterializer().Create(r)
			noErrors(t, ds)
			if got.Plan.SchemaVersion != 1 || len(got.Plan.Sections) != len(c.Instances) {
				t.Fatal("selection proposal missing")
			}
			guides(t, got.Guidance, c.Guidance)
		})
	}
}

func TestAllCurrentSourcesHaveTheLiteralNineteenGuideUnion(t *testing.T) {
	r := base()
	rows := strings.Split(strings.TrimSuffix(string(fixture(t, "relations.tsv")), "\n"), "\n")
	if len(rows) != 60 {
		t.Fatal("complete source selection fixture differs")
	}
	for _, row := range rows[1:] {
		fields := strings.Split(row, "\t")
		if len(fields) != 5 {
			t.Fatal("invalid source row")
		}
		r.Instances = append(r.Instances, api.Instance{SourceID: fields[0], PermanentSectionID: fields[1]})
	}
	want := []guideOracle{}
	guideRows := strings.Split(strings.TrimSuffix(string(fixture(t, "all-guidance.tsv")), "\n"), "\n")
	if len(guideRows) != 20 {
		t.Fatal("literal nineteen-guide union fixture differs")
	}
	for _, row := range guideRows[1:] {
		fields := strings.Split(row, "\t")
		if len(fields) != 2 {
			t.Fatal("invalid guide-union row")
		}
		want = append(want, guideOracle{ID: fields[0], Reasons: strings.Split(fields[1], ",")})
	}
	got, ds := newMaterializer().Create(r)
	noErrors(t, ds)
	if len(got.Plan.Sections) != 59 {
		t.Fatal("complete literal selection omitted activities")
	}
	guides(t, got.Guidance, want)
}

func TestExactIdentityAndCarrierRefusals(t *testing.T) {
	for _, id := range []string{"", "../repository-access", "Repository-access", "repository-access ", "repository-access/", "\x00", string([]byte{255})} {
		r := base()
		r.Instances = []api.Instance{{SourceID: id, PermanentSectionID: "activity"}}
		got, ds := newMaterializer().Create(r)
		if id == string([]byte{255}) {
			own(t, got, ds, "request", "materialize.text", "request", "Use valid UTF8 creation data.")
			continue
		}
		want := []api.Diagnostic{{Stage: "source", Path: "instances[0].sourceId", Problem: model.Problem{Code: "catalog.unknown-id", Severity: model.SeverityError, Message: "Choose an exact ID from catalog list.", Detail: map[string]string{"field": "id"}}}}
		if !reflect.DeepEqual(got, api.Result{}) || !reflect.DeepEqual(ds, want) {
			t.Errorf("exact selector refusal differs: %#v", ds)
		}
	}
	for _, source := range []string{"invitation-prepare", "reading-assign", "invitation-start", "onboarding-resume"} {
		r := base()
		r.Instances = []api.Instance{{SourceID: source, PermanentSectionID: "activity"}}
		got, ds := newMaterializer().Create(r)
		own(t, got, ds, "source", "materialize.carrier", "instances[0].sourceId", "Select a section-template source for an activity.")
	}
	for _, id := range []string{"", "plan", "yes", "no", "y", "n", "true", "false", "on", "off", "null", "With-Caps", "white space", "a.b", "-start", " padded"} {
		r := base()
		r.Instances = []api.Instance{{SourceID: "repository-access", PermanentSectionID: id}}
		got, ds := newMaterializer().Create(r)
		own(t, got, ds, "request", "materialize.instance-id", "instances[0].permanentSectionId", "Supply a nonreserved lowercase permanent section ID.")
	}
	r := base()
	r.PlanID = ""
	got, ds := newMaterializer().Create(r)
	own(t, got, ds, "request", "materialize.plan-id", "planId", "Supply an explicit nonempty plan identity.")
}

func TestDuplicateRequestsNeverChooseAWinner(t *testing.T) {
	for _, sources := range [][2]string{{"repository-access", "repository-access"}, {"repository-access", "document-reading"}, {"document-reading", "repository-access"}} {
		r := base()
		r.Instances = []api.Instance{{SourceID: sources[0], PermanentSectionID: "same"}, {SourceID: sources[1], PermanentSectionID: "same"}}
		got, ds := newMaterializer().Create(r)
		own(t, got, ds, "request", "materialize.duplicate-instance", "instances[1].permanentSectionId", "Give each independent activity its own permanent section ID.")
	}
	for _, scoped := range []bool{false, true} {
		for _, equal := range []bool{false, true} {
			r := base()
			values := []api.Binding{{Name: "shared_fact", Value: model.Value{Kind: model.ValueKindText, Text: "first"}}, {Name: "shared_fact", Value: model.Value{Kind: model.ValueKindText, Text: "second"}}}
			if equal {
				values[1] = values[0]
			}
			path := "planBindings[1].name"
			if scoped {
				values[0].Name = "repository_scope"
				values[1].Name = "repository_scope"
				r.Instances = []api.Instance{{SourceID: "repository-access", PermanentSectionID: "repo-a", Bindings: values}}
				path = "instances[0].bindings[1].name"
			} else {
				r.PlanVariables = []model.Variable{{Name: "shared_fact"}}
				r.PlanBindings = values
			}
			got, ds := newMaterializer().Create(r)
			own(t, got, ds, "request", "materialize.duplicate-binding", path, "Supply each scoped binding name once.")
		}
	}
}

func TestBindingsAreExplicitAuthorizedAndScopeLocal(t *testing.T) {
	for _, audience := range []model.Audience{"", model.AudienceOnboarder, model.AudienceOnboardee} {
		for _, kind := range []model.ValueKind{model.ValueKindNone, model.ValueKindDeferred, model.ValueKindText} {
			r := base()
			v := model.Value{Kind: kind, Text: "held", Flag: true, List: []string{"inactive", "inactive"}}
			r.PlanVariables = []model.Variable{{Name: "shared_fact", Audience: audience}}
			r.PlanBindings = []api.Binding{{Name: "shared_fact", Value: v}}
			got, ds := newMaterializer().Create(r)
			if audience == model.AudienceOnboardee {
				delegated(t, got, ds, "authoring", "authoring.unused-onboardee")
				continue
			}
			noErrors(t, ds)
			if got.Plan.SchemaVersion != 1 || !reflect.DeepEqual(got.Plan.Answers.Plan, map[string]model.Value{"shared_fact": v}) {
				t.Error("binding value or inactive fields were normalized/dropped")
			}
		}
	}
	for _, name := range []string{"undeclared", "repository-access.repository_scope", "repo-a.repository_scope", "repository_scope ", "shared_fact"} {
		r := base()
		r.PlanVariables = []model.Variable{{Name: "shared_fact"}}
		r.Instances = []api.Instance{{SourceID: "repository-access", PermanentSectionID: "repo-a", Bindings: []api.Binding{{Name: name, Value: model.Value{Kind: model.ValueKindText, Text: "value"}}}}}
		got, ds := newMaterializer().Create(r)
		own(t, got, ds, "binding", "materialize.binding-name", "instances[0].bindings[0].name", "Bind an exact variable declared in this scope.")
	}
	r := base()
	r.PlanBindings = []api.Binding{{Name: "repository_scope", Value: model.Value{Kind: model.ValueKindText, Text: "value"}}}
	r.Instances = []api.Instance{{SourceID: "repository-access", PermanentSectionID: "repo-a"}}
	got, ds := newMaterializer().Create(r)
	own(t, got, ds, "binding", "materialize.binding-name", "planBindings[0].name", "Bind an exact variable declared in this scope.")
}

func TestAllExplicitValueShapesStayOwnedAndDefaultsUnaccepted(t *testing.T) {
	r := base()
	r.PlanVariables = []model.Variable{
		{Name: "text_fact", Default: model.Value{Kind: model.ValueKindText, Text: "suggestion"}},
		{Name: "flag_fact", Type: model.VariableTypeBoolean, Default: model.Value{Kind: model.ValueKindFlag, Flag: true}},
		{Name: "command_fact", Type: model.VariableTypeCommand},
		{Name: "unanswered_fact", Default: model.Value{Kind: model.ValueKindText, Text: "never accepted"}},
	}
	r.PlanBindings = []api.Binding{{Name: "text_fact", Value: model.Value{Kind: model.ValueKindText, Text: ""}}, {Name: "flag_fact", Value: model.Value{Kind: model.ValueKindFlag, Flag: false}}, {Name: "command_fact", Value: model.Value{Kind: model.ValueKindList, List: []string{"one", "one", ""}}}}
	got, ds := newMaterializer().Create(r)
	noErrors(t, ds)
	want := map[string]model.Value{"text_fact": r.PlanBindings[0].Value, "flag_fact": r.PlanBindings[1].Value, "command_fact": r.PlanBindings[2].Value}
	if got.Plan.SchemaVersion != 1 || !reflect.DeepEqual(got.Plan.Answers.Plan, want) || !reflect.DeepEqual(got.Plan.Variables, r.PlanVariables) {
		t.Error("explicit false/empty/repeated values or unanswered defaults changed")
	}
}

func TestActualDeclarationAndAnswerValidation(t *testing.T) {
	for _, variant := range []string{"type-before-answer", "pattern-before-answer", "duplicate-declaration", "duplicate-directory", "unknown-directory-submission", "wrong-kind", "pattern-answer", "choice-answer"} {
		t.Run(variant, func(t *testing.T) {
			r := base()
			stage, code := "authoring", "authoring.variable-type"
			switch variant {
			case "type-before-answer":
				r.PlanVariables = []model.Variable{{Name: "shared_fact", Type: "invented"}}
				r.PlanBindings = []api.Binding{{Name: "missing", Value: model.Value{Kind: model.ValueKindFlag}}}
			case "pattern-before-answer":
				r.PlanVariables = []model.Variable{{Name: "shared_fact", Pattern: "["}}
				r.PlanBindings = []api.Binding{{Name: "shared_fact", Value: model.Value{Kind: model.ValueKindFlag}}}
				code = "authoring.pattern"
			case "duplicate-declaration":
				r.PlanVariables = []model.Variable{{Name: "shared_fact"}, {Name: "shared_fact", Type: model.VariableTypeBoolean}}
				r.PlanBindings = []api.Binding{{Name: "shared_fact", Value: model.Value{Kind: model.ValueKindText, Text: "value"}}}
				stage = "structure"
				code = "plan.duplicate-variable-name"
			case "duplicate-directory":
				r.Directory.Contacts = []model.Contact{{ID: "helper", Name: "One", Email: "one@example.invalid"}, {ID: "helper", Name: "Two", Email: "two@example.invalid"}}
				stage = "structure"
				code = "directory.duplicate-id"
			case "unknown-directory-submission":
				r.PlanVariables = []model.Variable{{Name: "shared_fact", Type: model.VariableTypeContact}}
				r.PlanBindings = []api.Binding{{Name: "shared_fact", Value: model.Value{Kind: model.ValueKindText, Text: "missing-helper"}}}
				stage = "answer"
				code = "directory.unknown-id"
			case "wrong-kind":
				r.Instances = []api.Instance{{SourceID: "repository-access", PermanentSectionID: "repo-a", Bindings: []api.Binding{{Name: "repository_scope", Value: model.Value{Kind: model.ValueKindFlag, Flag: true}}}}}
				stage = "answer"
				code = "answer.wrong-kind"
			case "pattern-answer":
				r.PlanVariables = []model.Variable{{Name: "shared_fact", Pattern: "^accepted$"}}
				r.PlanBindings = []api.Binding{{Name: "shared_fact", Value: model.Value{Kind: model.ValueKindText, Text: "different"}}}
				stage = "answer"
				code = "answer.pattern-mismatch"
			case "choice-answer":
				r.PlanVariables = []model.Variable{{Name: "shared_fact", Type: model.VariableTypeChoice, Options: []string{"first", "second"}}}
				r.PlanBindings = []api.Binding{{Name: "shared_fact", Value: model.Value{Kind: model.ValueKindText, Text: "Third"}}}
				stage = "answer"
				code = "answer.not-an-option"
			}
			got, ds := newMaterializer().Create(r)
			delegated(t, got, ds, stage, code)
		})
	}
}

func TestKnownDirectoryBindingsAndFinalWarnings(t *testing.T) {
	r := base()
	r.Directory = model.Directory{Contacts: []model.Contact{{ID: "helper", Name: " Helper ", Email: "helper@example.invalid"}, {ID: "unreachable", Name: "No route"}}, Roles: []model.Role{{ID: "support", Name: "Support", Holders: []string{"helper", "helper"}}}, Teams: []model.Team{{ID: "team-a", Name: " Team A ", Contact: "helper", Holders: map[string][]string{"support": {"helper", "helper"}}}}}
	r.PlanVariables = []model.Variable{{Name: "team_fact", Type: model.VariableTypeTeam}, {Name: "contacts_fact", Type: model.VariableTypeContact, Multiple: true}}
	r.PlanBindings = []api.Binding{{Name: "team_fact", Value: model.Value{Kind: model.ValueKindText, Text: "team-a"}}, {Name: "contacts_fact", Value: model.Value{Kind: model.ValueKindList, List: []string{"helper", "helper"}}}}
	got, ds := newMaterializer().Create(r)
	noErrors(t, ds)
	if got.Plan.SchemaVersion != 1 || !reflect.DeepEqual(got.Plan.Directory, r.Directory) || !reflect.DeepEqual(got.Plan.Variables, r.PlanVariables) {
		t.Error("directory/declaration order, repetitions or spaces changed")
	}
	if len(ds) != 1 || ds[0].Stage != "plan" || ds[0].Path != "" || ds[0].Problem.Code != "directory.contact-unreachable" || ds[0].Problem.Severity != model.SeverityWarning || ds[0].Problem.Message == "" {
		t.Error("actual final-plan warning lost, duplicated or made fatal")
	}
}

func TestProposedCountAndTextAdmissionPrecedesLaterSemantics(t *testing.T) {
	for _, count := range []int{256, 257} {
		r := base()
		for i := 0; i < count; i++ {
			r.Instances = append(r.Instances, api.Instance{SourceID: "application-local", PermanentSectionID: fmt.Sprintf("application-%03d", i)})
		}
		got, ds := newMaterializer().Create(r)
		if count == 257 {
			own(t, got, ds, "request", "materialize.limit", "request", "Reduce the selection or held creation data to the supported limits.")
		} else {
			noErrors(t, ds)
			if len(got.Plan.Sections) != 256 {
				t.Error("constructible inclusive repeat count rejected/truncated")
			}
		}
	}
	for _, count := range []int{8192, 8193} {
		r := base()
		for i := 0; i < count; i++ {
			r.PlanBindings = append(r.PlanBindings, api.Binding{Name: "aaa"})
		}
		got, ds := newMaterializer().Create(r)
		if count == 8193 {
			own(t, got, ds, "request", "materialize.limit", "request", "Reduce the selection or held creation data to the supported limits.")
		} else {
			own(t, got, ds, "request", "materialize.duplicate-binding", "planBindings[1].name", "Supply each scoped binding name once.")
		}
	}
	for _, extra := range []int{0, 1} {
		r := api.Request{PlanID: "p"}
		for i := 0; i < 8; i++ {
			text := strings.Repeat("x", 1048576)
			if i == 0 {
				text = text[:len(text)-25+extra]
			}
			r.PlanVariables = append(r.PlanVariables, model.Variable{Name: fmt.Sprintf("aa%c", 'a'+i), Default: model.Value{Kind: model.ValueKindNone, Text: text}})
		}
		got, ds := newMaterializer().Create(r)
		if extra == 1 {
			own(t, got, ds, "request", "materialize.limit", "request", "Reduce the selection or held creation data to the supported limits.")
		} else {
			noErrors(t, ds)
			if got.Plan.SchemaVersion != 1 || len(got.Plan.Variables) != 8 {
				t.Error("inclusive held-text cap refused/truncated inactive payload")
			}
		}
	}
	r := base()
	r.PlanVariables = []model.Variable{{Name: "aaa", Default: model.Value{Kind: model.ValueKindNone, Text: string([]byte{255})}}}
	got, ds := newMaterializer().Create(r)
	own(t, got, ds, "request", "materialize.text", "request", "Use valid UTF8 creation data.")
	r = base()
	r.PlanVariables = []model.Variable{{Name: "aaa", Default: model.Value{Kind: model.ValueKind(99)}}}
	got, ds = newMaterializer().Create(r)
	own(t, got, ds, "request", "materialize.value-kind", "planVariables[0].default", "Choose a supported value kind.")
}

func TestResultsDoNotAliasRequestsRepeatedInstancesOrCalls(t *testing.T) {
	r := base()
	r.Directory = model.Directory{Contacts: []model.Contact{{ID: "helper", Name: "Helper", Email: "helper@example.invalid"}}, Roles: []model.Role{{ID: "role", Name: "Role", Holders: []string{"helper"}}}, Teams: []model.Team{{ID: "team", Name: "Team", Holders: map[string][]string{"role": {"helper"}}}}}
	r.PlanVariables = []model.Variable{{Name: "choices_fact", Type: model.VariableTypeChoice, Options: []string{"one", "two"}}}
	r.PlanBindings = []api.Binding{{Name: "choices_fact", Value: model.Value{Kind: model.ValueKindText, Text: "one", List: []string{"held"}}}}
	r.Instances = []api.Instance{{SourceID: "repository-access", PermanentSectionID: "repo-a"}, {SourceID: "repository-access", PermanentSectionID: "repo-b"}}
	m := newMaterializer()
	first, ds := m.Create(r)
	noErrors(t, ds)
	if len(first.Plan.Sections) != 2 || len(first.Guidance) != 2 {
		t.Fatal("nonempty ownership control missing")
	}
	before := snapshot(t, first)
	requestBefore := snapshot(t, r)
	first.Plan.Directory.Roles[0].Holders[0] = "changed"
	first.Plan.Directory.Teams[0].Holders["role"][0] = "changed"
	first.Plan.Variables[0].Options[0] = "changed"
	first.Plan.Answers.Plan["choices_fact"].List[0] = "changed"
	first.Plan.Sections[0].Variables[0].Description = "changed"
	first.Plan.Sections[0].Steps[0].Checks[0].Ask = "changed"
	first.Plan.Sections[0].Source.Template = "changed"
	first.Guidance[0].Markdown[0] = '!'
	first.Guidance[0].InstanceIDs[0] = "changed"
	if !bytes.Equal(snapshot(t, r), requestBefore) || first.Plan.Sections[1].Variables[0].Description == "changed" || first.Plan.Sections[1].Steps[0].Checks[0].Ask == "changed" {
		t.Error("returned data alias request or another repeat")
	}
	second, ds := m.Create(r)
	noErrors(t, ds)
	if !bytes.Equal(snapshot(t, second), before) {
		t.Error("earlier result mutation poisoned compiled source or later call")
	}
	r.Directory.Teams[0].Holders["role"][0] = "request-edit"
	r.PlanVariables[0].Options[0] = "request-edit"
	r.PlanBindings[0].Value.List[0] = "request-edit"
	if !bytes.Equal(snapshot(t, second), before) {
		t.Error("request mutation reached returned plan")
	}
}

func TestConcurrentDeterministicReuseAndDiagnosticOwnership(t *testing.T) {
	m := newMaterializer()
	r := base()
	r.Instances = []api.Instance{{SourceID: "document-reading", PermanentSectionID: "reading"}}
	want, ds := m.Create(r)
	noErrors(t, ds)
	if len(want.Plan.Sections) != 1 {
		t.Fatal("nonempty control missing")
	}
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			got, problems := m.Create(r)
			if !reflect.DeepEqual(got, want) || len(problems) != len(ds) {
				t.Error("shared materializer call changed output")
			}
		}()
	}
	group.Wait()
	bad := base()
	bad.PlanID = ""
	_, first := m.Create(bad)
	if len(first) != 1 || first[0].Problem.Detail == nil {
		t.Fatal("own diagnostic control missing")
	}
	first[0].Problem.Detail["path"] = "changed"
	got, second := m.Create(bad)
	own(t, got, second, "request", "materialize.plan-id", "planId", "Supply an explicit nonempty plan identity.")
}

func TestBootstrapAndEscalationRemainUnresolved(t *testing.T) {
	r := base()
	r.Instances = []api.Instance{{SourceID: "onboarding-help", PermanentSectionID: "help"}, {SourceID: "onboarding-welcome", PermanentSectionID: "welcome"}}
	got, ds := newMaterializer().Create(r)
	noErrors(t, ds)
	if len(got.Plan.Sections) != 2 {
		t.Fatal("selected bootstrap sections missing")
	}
	help, welcome := got.Plan.Sections[0], got.Plan.Sections[1]
	if !help.Bootstrap || !welcome.Bootstrap || !reflect.DeepEqual(help.Escalation, []string{"$general_contacts"}) || help.Timing != model.TimingAnytime || welcome.Timing != model.TimingWeekOne {
		t.Error("bootstrap, escalation or held timing changed")
	}
	if len(help.Variables) != 3 || help.Variables[0].Type != model.VariableTypeContact || !help.Variables[0].Multiple || help.Variables[0].Name != "general_contacts" || len(got.Plan.Answers.Sections) != 0 || len(got.Plan.Answers.Plan) != 0 {
		t.Error("support declaration changed or acquired an inferred binding")
	}
	if !strings.Contains(help.Body, "$preaccess_route") || !strings.Contains(help.Body, "$fallback_route") || !strings.Contains(welcome.Body, "$initial_arrangements") {
		t.Error("unresolved instructions were resolved or dropped")
	}
}

func TestProposedCollectionCapsHaveIndependentLaterControls(t *testing.T) {
	type limitCase struct {
		name                  string
		build                 func(int) api.Request
		bound                 int
		laterStage, laterCode string
	}
	cases := []limitCase{
		{"plan-variables", func(n int) api.Request {
			r := base()
			for i := 0; i < n; i++ {
				r.PlanVariables = append(r.PlanVariables, model.Variable{Name: fmt.Sprintf("variable_%04d", i)})
			}
			return r
		}, 1024, "", ""},
		{"options-per-list", func(n int) api.Request {
			r := base()
			r.PlanVariables = []model.Variable{{Name: "aaa", Options: make([]string, n)}}
			return r
		}, 1024, "", ""},
		{"options-total", func(n int) api.Request {
			r := base()
			for i := 0; n > 0; i++ {
				count := n
				if count > 1024 {
					count = 1024
				}
				r.PlanVariables = append(r.PlanVariables, model.Variable{Name: fmt.Sprintf("variable_%04d", i), Options: make([]string, count)})
				n -= count
			}
			return r
		}, 16384, "", ""},
		{"value-list-per-list", func(n int) api.Request {
			r := base()
			r.PlanVariables = []model.Variable{{Name: "aaa", Default: model.Value{Kind: model.ValueKindNone, List: make([]string, n)}}}
			return r
		}, 4096, "", ""},
		{"value-list-total", func(n int) api.Request {
			r := base()
			for i := 0; n > 0; i++ {
				count := n
				if count > 4096 {
					count = 4096
				}
				r.PlanVariables = append(r.PlanVariables, model.Variable{Name: fmt.Sprintf("variable_%04d", i), Default: model.Value{Kind: model.ValueKindNone, List: make([]string, count)}})
				n -= count
			}
			return r
		}, 32768, "", ""},
		{"clauses-per-condition", func(n int) api.Request {
			r := base()
			r.PlanVariables = []model.Variable{{Name: "aaa", AskIf: &model.Condition{All: make([]model.Clause, n)}}}
			return r
		}, 256, "authoring", "authoring.plan-ask-if"},
		{"clauses-total", func(n int) api.Request {
			r := base()
			for i := 0; n > 0; i++ {
				count := n
				if count > 256 {
					count = 256
				}
				r.PlanVariables = append(r.PlanVariables, model.Variable{Name: fmt.Sprintf("variable_%04d", i), AskIf: &model.Condition{All: make([]model.Clause, count)}})
				n -= count
			}
			return r
		}, 4096, "authoring", "authoring.plan-ask-if"},
		{"comparison-per-list", func(n int) api.Request {
			r := base()
			r.PlanVariables = []model.Variable{{Name: "aaa", AskIf: &model.Condition{All: []model.Clause{{Answer: model.AnswerTest{IsOneOf: make([]model.Value, n)}}}}}}
			return r
		}, 256, "authoring", "authoring.plan-ask-if"},
		{"comparison-total", func(n int) api.Request {
			r := base()
			condition := &model.Condition{}
			for n > 0 {
				count := n
				if count > 256 {
					count = 256
				}
				condition.All = append(condition.All, model.Clause{Answer: model.AnswerTest{IsOneOf: make([]model.Value, count)}})
				n -= count
			}
			r.PlanVariables = []model.Variable{{Name: "aaa", AskIf: condition}}
			return r
		}, 16384, "authoring", "authoring.plan-ask-if"},
		{"holders-per-list", func(n int) api.Request {
			r := base()
			r.Directory.Contacts = []model.Contact{{ID: "helper", Name: "Helper", Email: "helper@example.invalid"}}
			holders := make([]string, n)
			for i := range holders {
				holders[i] = "helper"
			}
			r.Directory.Roles = []model.Role{{ID: "support", Name: "Support", Holders: holders}}
			return r
		}, 1024, "", ""},
		{"holders-total", func(n int) api.Request {
			r := base()
			r.Directory.Contacts = []model.Contact{{ID: "helper", Name: "Helper", Email: "helper@example.invalid"}}
			for i := 0; n > 0; i++ {
				count := n
				if count > 1024 {
					count = 1024
				}
				holders := make([]string, count)
				for j := range holders {
					holders[j] = "helper"
				}
				r.Directory.Roles = append(r.Directory.Roles, model.Role{ID: fmt.Sprintf("role-%02d", i), Name: "Support", Holders: holders})
				n -= count
			}
			return r
		}, 8192, "", ""},
		{"directory-per-list", func(n int) api.Request {
			r := base()
			for i := 0; i < n; i++ {
				r.Directory.Contacts = append(r.Directory.Contacts, model.Contact{ID: fmt.Sprintf("helper-%04d", i), Name: "Helper", Email: "helper@example.invalid"})
			}
			return r
		}, 2048, "", ""},
		{"directory-total", func(n int) api.Request {
			r := base()
			for i := 0; i < 2048; i++ {
				r.Directory.Contacts = append(r.Directory.Contacts, model.Contact{ID: fmt.Sprintf("helper-%04d", i), Name: "Helper", Email: "helper@example.invalid"})
			}
			for i := 0; i < n-2048; i++ {
				if i < 2048 {
					r.Directory.Roles = append(r.Directory.Roles, model.Role{ID: fmt.Sprintf("role-%04d", i), Name: "Support", Holders: []string{"helper-0000"}})
				} else {
					r.Directory.Teams = append(r.Directory.Teams, model.Team{ID: "team", Name: "Team"})
				}
			}
			return r
		}, 4096, "", ""},
		{"team-holder-map-per-map", func(n int) api.Request {
			r := base()
			r.Directory.Contacts = []model.Contact{{ID: "helper", Name: "Helper", Email: "helper@example.invalid"}}
			holders := map[string][]string{}
			for i := 0; i < 2048; i++ {
				id := fmt.Sprintf("role-%04d", i)
				r.Directory.Roles = append(r.Directory.Roles, model.Role{ID: id, Name: "Support", Holders: []string{"helper"}})
			}
			for i := 0; i < n; i++ {
				holders[fmt.Sprintf("role-%04d", i)] = []string{}
			}
			r.Directory.Teams = []model.Team{{ID: "team", Name: "Team", Holders: holders}}
			return r
		}, 2048, "", ""},
		{"team-holder-map-total", func(n int) api.Request {
			r := base()
			r.Directory.Contacts = []model.Contact{{ID: "helper", Name: "Helper", Email: "helper@example.invalid"}}
			for i := 0; i < 2048; i++ {
				r.Directory.Roles = append(r.Directory.Roles, model.Role{ID: fmt.Sprintf("role-%04d", i), Name: "Support", Holders: []string{"helper"}})
			}
			for i := 0; n > 0; i++ {
				count := n
				if count > 2048 {
					count = 2048
				}
				holders := map[string][]string{}
				for j := 0; j < count; j++ {
					holders[fmt.Sprintf("role-%04d", j)] = []string{}
				}
				r.Directory.Teams = append(r.Directory.Teams, model.Team{ID: fmt.Sprintf("team-%02d", i), Name: "Team", Holders: holders})
				n -= count
			}
			return r
		}, 8192, "", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			for _, extra := range []int{0, 1} {
				r := c.build(c.bound + extra)
				got, ds := newMaterializer().Create(r)
				if extra == 1 {
					own(t, got, ds, "request", "materialize.limit", "request", "Reduce the selection or held creation data to the supported limits.")
				} else if c.laterCode != "" {
					delegated(t, got, ds, c.laterStage, c.laterCode)
				} else {
					noErrors(t, ds)
					if got.Plan.SchemaVersion != 1 {
						t.Error("constructible inclusive collection control did not produce a plan")
					}
				}
			}
		})
	}
}

func TestProposedCombinedCollectionEntryBudget(t *testing.T) {
	for _, extra := range []int{0, 1} {
		r := base()
		for i := 0; i < 8; i++ {
			r.PlanVariables = append(r.PlanVariables, model.Variable{Name: fmt.Sprintf("list_%02d", i), Default: model.Value{Kind: model.ValueKindNone, List: make([]string, 4096)}})
		}
		for i := 0; i < 16; i++ {
			r.PlanVariables = append(r.PlanVariables, model.Variable{Name: fmt.Sprintf("options_%02d", i), Options: make([]string, 1024)})
		}
		condition := &model.Condition{}
		for i := 0; i < 64; i++ {
			count := 256
			if i == 63 {
				count = 167 + extra
			}
			condition.All = append(condition.All, model.Clause{Answer: model.AnswerTest{IsOneOf: make([]model.Value, count)}})
		}
		r.PlanVariables = append(r.PlanVariables, model.Variable{Name: "conditional", AskIf: condition})
		got, ds := newMaterializer().Create(r)
		if extra == 1 {
			own(t, got, ds, "request", "materialize.limit", "request", "Reduce the selection or held creation data to the supported limits.")
		} else {
			delegated(t, got, ds, "authoring", "authoring.plan-ask-if")
		}
	}
}

func TestProposedTextBoundsAndGuardPrecedence(t *testing.T) {
	for _, size := range []int{1048576, 1048577} {
		r := base()
		r.PlanVariables = []model.Variable{{Name: "aaa", Default: model.Value{Kind: model.ValueKindNone, Text: strings.Repeat("x", size)}}}
		got, ds := newMaterializer().Create(r)
		if size == 1048577 {
			own(t, got, ds, "request", "materialize.limit", "request", "Reduce the selection or held creation data to the supported limits.")
		} else {
			noErrors(t, ds)
			if got.Plan.SchemaVersion != 1 {
				t.Error("inclusive scalar size rejected")
			}
		}
	}
	for _, size := range []int{4096, 4097} {
		r := base()
		r.PlanID = strings.Repeat("x", size)
		got, ds := newMaterializer().Create(r)
		if size == 4097 {
			own(t, got, ds, "request", "materialize.limit", "request", "Reduce the selection or held creation data to the supported limits.")
		} else {
			noErrors(t, ds)
			if got.Plan.ID != r.PlanID {
				t.Error("opaque plan identity truncated")
			}
		}
	}
	r := base()
	r.PlanID = string([]byte{255})
	r.Instances = make([]api.Instance, 257)
	got, ds := newMaterializer().Create(r)
	own(t, got, ds, "request", "materialize.limit", "request", "Reduce the selection or held creation data to the supported limits.")
	r = base()
	r.PlanVariables = []model.Variable{{Name: "aaa", Options: []string{string([]byte{255})}}}
	got, ds = newMaterializer().Create(r)
	own(t, got, ds, "request", "materialize.text", "request", "Use valid UTF8 creation data.")
}
