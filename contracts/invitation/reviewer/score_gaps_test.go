package reviewer_test

import (
	"reflect"
	"strings"
	"testing"

	contract "onboard-kit/contracts/invitation/reviewer"
)

func TestScoreGapRenderedFieldBoundaries(t *testing.T) {
	for _, item := range []struct {
		name    string
		prefix  string
		suffix  string
		dollars int
		padding int
		inert   int
	}{
		{"run", "    steps:\n      - run: [\"", "\"]\n", 100000, 39782, 1},
		{"check-ask", "    steps:\n      - checks:\n          - ask: \"", "\"\n", 440000, 9581, 1},
		{"step-environment", "    steps:\n      - env:\n          X: \"", "\"\n", 440000, 9712, 0},
	} {
		t.Run(item.name, func(t *testing.T) {
			plan := string(paddedAdmissionPlan(t, 0, item.padding))
			plan = strings.Replace(plan, "pad:", item.prefix+strings.Repeat("$", item.dollars)+item.suffix+"pad:", 1)
			raw := admissionArchive(t, []byte(plan))
			raw = append(raw, strings.Repeat("x", item.inert)...)
			r := newReviewer()
			in, problems := r.Inspect(raw)
			if len(problems) != 0 || in.PlanID != "p" {
				t.Fatalf("independently derived exact B must admit: %v", problems)
			}
			result, problems := r.Review(in, contract.Local{}, contract.PlanChoiceUndecided)
			if len(problems) != 0 || result.Proposal == nil || len(result.Proposal.Review.Session.Activities) != 1 {
				t.Fatalf("literal carrier control must remain usable: %v", problems)
			}
			got, problems := r.Inspect(append(copied(raw), 'x'))
			resourceProblem(t, problems, "inspect.payload-admission", "invitation")
			if !reflect.DeepEqual(got, contract.Incoming{}) {
				t.Fatal("one inert source byte returned a partial incoming")
			}
		})
	}
}

func TestScoreGapCandidateAndLocalMarkerBoundaries(t *testing.T) {
	for _, item := range []struct {
		name    string
		local   bool
		dollars int
		padding int
		extra   string
		answers string
		inert   int
	}{
		{"plan", false, 440000, 9684, "", "", 0},
		{"local", true, 380000, 8858, "", "", 0},
		{"irrelevant-options", false, 440000, 9578, "    options: [a]\n", "", 0},
		{"command-default-list", false, 440000, 9561, "    type: command\n    default: [a, b]\n", "", 1},
		{"hidden-given-pool", false, 440000, 9328, "", "answers:\n  sections:\n    hidden:\n      xxx: a\n", 1},
		{"local-hidden-list", true, 348000, 12486, "", "answers:\n  sections:\n    task:\n      hid: [aaaaaaaaaa, bbbbbbbbbb]\n", 1},
		{"contact-directory", true, 320000, 28552, "    type: contact\n", "directory:\n  contacts:\n    - id: c\n      name: aaaaaaaaaaaaaaaaaaaa\n      slack: s\n", 0},
		{"vacant-role-directory", true, 260000, 28633, "    type: role\n", "directory:\n  roles:\n    - id: r\n      name: aaaaaaaaaaaaaaaaaaaa\n", 0},
	} {
		t.Run(item.name, func(t *testing.T) {
			plan := string(paddedAdmissionPlan(t, item.dollars, item.padding))
			if item.local {
				plan = strings.Replace(plan, "    body:", "    variables:\n      - name: xxx\n"+strings.ReplaceAll(item.extra, "    ", "        ")+"    body:", 1)
				plan = strings.Replace(plan, "sections:\n", item.answers+"sections:\n", 1)
			} else {
				plan = strings.Replace(plan, "sections:\n", "variables:\n  - name: xxx\n"+item.extra+item.answers+"sections:\n", 1)
			}
			raw := admissionArchive(t, []byte(plan))
			raw = append(raw, strings.Repeat("x", item.inert)...)
			r := newReviewer()
			in, problems := r.Inspect(raw)
			if len(problems) != 0 || in.PlanID != "p" {
				t.Fatalf("independently derived exact B must admit: %v", problems)
			}
			result, problems := r.Review(in, contract.Local{}, contract.PlanChoiceUndecided)
			if len(problems) != 0 || result.Proposal == nil {
				t.Fatalf("exact B effective context must admit: %v", problems)
			}
			if len(result.Proposal.Review.Session.Activities) != 1 || len(result.Proposal.Review.Session.Questions) != 1 {
				t.Fatal("full session must retain its one active onboarder question")
			}
			question := result.Proposal.Review.Session.Questions[0]
			scope := "plan"
			if item.local {
				scope = "task"
			}
			if question.ID != scope+".xxx" || question.Name != "xxx" || question.Scope != scope || question.Audience != "onboarder" {
				t.Fatal("full session changed the declaration's scope or audience")
			}
			got, problems := r.Inspect(append(copied(raw), 'x'))
			resourceProblem(t, problems, "inspect.payload-admission", "invitation")
			if !reflect.DeepEqual(got, contract.Incoming{}) {
				t.Fatal("one inert source byte returned a partial incoming")
			}
		})
	}
}

func TestScoreGapEligibilityChargeBoundary(t *testing.T) {
	plan := string(paddedAdmissionPlan(t, 440000, 8342))
	plan = strings.Replace(plan, "sections:\n", "variables:\n  - name: xxx\nsections:\n", 1)
	plan = strings.Replace(plan, "    body:", "    includeIf:\n      all:\n        - section: task\n          enabled: true\n    body:", 1)
	raw := append(admissionArchive(t, []byte(plan)), 'x')
	r := newReviewer()
	in, problems := r.Inspect(raw)
	if len(problems) != 0 || in.PlanID != "p" {
		t.Fatalf("independently derived exact B must admit: %v", problems)
	}
	result, problems := r.Review(in, contract.Local{}, contract.PlanChoiceUndecided)
	if len(problems) != 0 || result.Proposal == nil || len(result.Proposal.Review.Session.Activities) != 1 {
		t.Fatalf("section switch condition must remain usable: %v", problems)
	}
	got, problems := r.Inspect(append(copied(raw), 'x'))
	resourceProblem(t, problems, "inspect.payload-admission", "invitation")
	if !reflect.DeepEqual(got, contract.Incoming{}) {
		t.Fatal("one inert source byte returned a partial incoming")
	}
}

func TestScoreGapEscalationBoundary(t *testing.T) {
	plan := string(paddedAdmissionPlan(t, 320000, 28427))
	plan = strings.Replace(plan, "sections:\n", "directory:\n  contacts:\n    - id: c\n      name: aaaaaaaaaaaaaaaaaaaa\n      slack: s\nsections:\n", 1)
	plan = strings.Replace(plan, "    body:", "    variables:\n      - name: xxx\n        type: contact\n    escalation: [$xxx]\n    body:", 1)
	raw := admissionArchive(t, []byte(plan))
	r := newReviewer()
	in, problems := r.Inspect(raw)
	if len(problems) != 0 || in.PlanID != "p" {
		t.Fatalf("independently derived exact B escalation must admit: %v", problems)
	}
	result, problems := r.Review(in, contract.Local{}, contract.PlanChoiceUndecided)
	if len(problems) != 0 || result.Proposal == nil || len(result.Proposal.Review.Session.Activities) != 1 {
		t.Fatalf("escalation control must remain usable: %v", problems)
	}
	_, problems = r.Inspect(append(copied(raw), 'x'))
	resourceProblem(t, problems, "inspect.payload-admission", "invitation")
}

func TestScoreGapWaitingBoundary(t *testing.T) {
	plan := string(paddedAdmissionPlan(t, 440000, 9167))
	plan = strings.Replace(plan, "    body:", "    requires: [other]\n    body:", 1)
	plan = strings.Replace(plan, "pad:", "  - id: other\n    title: A\n    summary: A\npad:", 1)
	raw := append(admissionArchive(t, []byte(plan)), 'x')
	r := newReviewer()
	in, problems := r.Inspect(raw)
	if len(problems) != 0 || in.PlanID != "p" {
		t.Fatalf("independently derived exact B waiting must admit: %v", problems)
	}
	result, problems := r.Review(in, contract.Local{}, contract.PlanChoiceUndecided)
	if len(problems) != 0 || result.Proposal == nil || len(result.Proposal.Review.Session.Activities) != 2 {
		t.Fatalf("prerequisite control must remain usable: %v", problems)
	}
	_, problems = r.Inspect(append(copied(raw), 'x'))
	resourceProblem(t, problems, "inspect.payload-admission", "invitation")
}

func TestScoreGapOutcomeMetadataBoundary(t *testing.T) {
	r := newReviewer()
	raw := append(admissionArchive(t, paddedAdmissionPlan(t, 440000, 9941)), 'x')
	in, problems := r.Inspect(raw)
	if len(problems) != 0 {
		t.Fatalf("smaller incoming must admit: %v", problems)
	}
	record := append(copied(in.Files.Record), []byte("outcomes:\n  - section: task\n    outcome: verified\n    on: \"2026-10-23T12:34:56Z\"\n")...)
	if len(record) != 215 {
		t.Fatal("independent record source count changed")
	}
	result, problems := r.Review(in, contract.Local{Record: contract.Body{Present: true, Bytes: record}}, contract.PlanChoiceUndecided)
	if len(problems) != 0 || result.Proposal == nil || len(result.Proposal.Review.Session.Outcomes) != 1 {
		t.Fatalf("independently derived exact B history must admit: %v", problems)
	}
	got, problems := r.Review(in, contract.Local{Record: contract.Body{Present: true, Bytes: append(copied(record), '\n')}}, contract.PlanChoiceUndecided)
	resourceProblem(t, problems, "review.effective-admission", "effectivePlan")
	if !reflect.DeepEqual(got, contract.ReviewResult{}) {
		t.Fatal("one inert local source byte returned a partial review")
	}
}
