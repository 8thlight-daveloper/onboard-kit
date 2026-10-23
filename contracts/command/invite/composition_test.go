package invite_test

import (
	"bytes"
	"strings"
	"testing"
)

func TestBootstrapWarningsArePreservedWithoutSeedingAnswers(t *testing.T) {
	s := newSandbox(t)
	section := "sections:\n  - id: access\n    title: Public access\n    bootstrap: true\n    summary: Use $account.\n    variables:\n      - name: account\n        audience: onboardee\n"
	writeLiteral(t, s.plan, append(literal(t, "plan.yaml"), []byte(section)...))
	before := tree(t, s.root)
	readReport(t, run(newCommand(), s.args()), "invite.written", 0, []projectedProblem{{Code: "gap.unanswered", Severity: "warning"}}, nil)
	preserved(t, s, before, true)
	payloads := exportedPayloads(t, s.out)
	fragment := "## Do these first\n\n### Gaps to fill\n\n| Gap | Who fills it | By when |\n| --- | --- | --- |\n| account |  | Anytime |\n\n### Public access\n\nUse `<account>`.\n\n## Connect to the shared repository"
	if !strings.Contains(string(payloads["first-steps.md"]), fragment) {
		t.Fatal("exact real Bootstrap fragment/warning behavior lost")
	}
	checkSeed(t, payloads)
}

func TestPreparationTeamDoesNotBecomeBootstrapAnswer(t *testing.T) {
	s := newSandbox(t)
	addition := "  teams:\n    - id: blue\n      name: Blue\nvariables:\n  - name: squad\n    type: team\n    audience: onboardee\nsections:\n  - id: access\n    title: PRIVATE prepared-only access\n    bootstrap: true\n    includeIf:\n      all:\n        - question: plan.squad\n          answer:\n            is: blue\n"
	writeLiteral(t, s.plan, append(literal(t, "plan.yaml"), []byte(addition)...))
	writeLiteral(t, s.preparation, append(literal(t, "preparation.yaml"), []byte("  team: blue\n")...))
	before := tree(t, s.root)
	readReport(t, run(newCommand(), s.args()), "invite.written", 0, nil, nil)
	preserved(t, s, before, true)
	payloads := exportedPayloads(t, s.out)
	if !bytes.Equal(payloads["first-steps.md"], literal(t, "first-steps.md")) || bytes.Contains(payloads["plan.obk.yaml"], []byte("\nanswers:")) {
		t.Fatal("preparation was implicitly bound into original plan/Bootstrap")
	}
	checkSeed(t, payloads)
}

func TestOrganizationOnboardeePreseedStaysOnlyInPlan(t *testing.T) {
	s := newSandbox(t)
	addition := "variables:\n  - name: personalanswer\n    audience: onboardee\nanswers:\n  plan:\n    personalanswer: orgseed\n"
	writeLiteral(t, s.plan, append(literal(t, "plan.yaml"), []byte(addition)...))
	before := tree(t, s.root)
	readReport(t, run(newCommand(), s.args()), "invite.written", 0, nil, nil)
	preserved(t, s, before, true)
	payloads := exportedPayloads(t, s.out)
	if !bytes.Contains(payloads["plan.obk.yaml"], []byte("personalanswer: orgseed")) {
		t.Fatal("original preseed was lost")
	}
	checkSeed(t, payloads)
}

func TestCustomDestinationsDoNotChangeArchiveNames(t *testing.T) {
	s := newSandbox(t)
	writeLiteral(t, s.plan, append(literal(t, "plan.yaml"), []byte("resultsPath: people/person-{account}/progress.yaml\n")...))
	before := tree(t, s.root)
	readReport(t, run(newCommand(), s.args()), "invite.written", 0, nil, nil)
	preserved(t, s, before, true)
	payloads := exportedPayloads(t, s.out)
	guide := string(payloads["first-steps.md"])
	for _, expected := range []string{"Relative record destination: `people/person-newcomer/progress.yaml`.", "Relative preparation destination: `people/person-newcomer/preparation.obk.yaml`.", "- `records/newcomer/record.obk.yaml`:", "- `records/newcomer/preparation.obk.yaml`:"} {
		if !strings.Contains(guide, expected) {
			t.Fatalf("literal destination/archive distinction lost: %q", expected)
		}
	}
	checkSeed(t, payloads)
}
