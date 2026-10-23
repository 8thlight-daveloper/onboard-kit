package invite_test

import (
	"strings"
	"testing"
)

func TestHelpIsLiteralAndIgnoresRemainingTokens(t *testing.T) {
	for _, args := range [][]string{{"--help", "--PRIVATE=secret"}, {"export", "--help", "--plan", "/PRIVATE/missing"}} {
		got := run(newCommand(), args)
		if got.exit != 0 || got.stdout != usage || got.stderr != "" {
			t.Fatalf("help precedence or literal text differs: %#v", got)
		}
	}
}

func TestUsageRefusalsDoNotTouchFiles(t *testing.T) {
	for _, name := range []string{"missing-operation", "unknown-operation", "missing-out", "missing-assertion", "repeat-plan", "unknown-flag", "positional", "bare-terminator", "empty-account", "missing-value"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			s := newSandbox(t)
			args := s.args()
			switch name {
			case "missing-operation":
				args = nil
			case "unknown-operation":
				args[0] = "PRIVATE_OPERATION"
			case "missing-out":
				args = append(args[:7], args[9:]...)
			case "missing-assertion":
				args = args[:9]
			case "repeat-plan":
				args = append(args, "--plan", s.plan)
			case "unknown-flag":
				args = append(args, "--PRIVATE", "PRIVATE_VALUE")
			case "positional":
				args = append(args, "PRIVATE_POSITIONAL")
			case "bare-terminator":
				args = append(args, "--")
			case "empty-account":
				args[6] = ""
			case "missing-value":
				args = append(args, "--help-contact", "--help-role", "owner")
			}
			before := tree(t, s.root)
			readReport(t, run(newCommand(), args), "invite.usage", 2, nil, nil)
			preserved(t, s, before, false)
		})
	}
}

func TestEqualsSpellingAndDotComponentsAreAccepted(t *testing.T) {
	s := newSandbox(t)
	args := []string{"export", "--plan=" + s.root + "/./plan.yaml", "--preparation=" + s.preparation, "--account=newcomer", "--out=" + s.root + "//./out.zip", "--preaccess-help-contact=owner"}
	before := tree(t, s.root)
	readReport(t, run(newCommand(), args), "invite.written", 0, nil, nil)
	preserved(t, s, before, true)
	exportedPayloads(t, s.out)
}

func TestInvalidAccountRefusesBeforeFileReads(t *testing.T) {
	s := newSandbox(t)
	args := s.args()
	args[2] = "/PRIVATE/nonexistent-plan"
	args[6] = "PRIVATE/bad-account"
	before := tree(t, s.root)
	readReport(t, run(newCommand(), args), "invite.account", 1, nil, nil)
	preserved(t, s, before, false)
}

func TestLiteralDoubleDashValueUsesEqualsSpelling(t *testing.T) {
	s := newSandbox(t)
	args := append([]string{"export", "--plan=--PRIVATE-literal"}, s.args()[3:]...)
	before := tree(t, s.root)
	readReport(t, run(newCommand(), args), "invite.plan-read", 1, nil, nil)
	preserved(t, s, before, false)
}

func TestTypedHelpSelectorsKeepAppearanceOrder(t *testing.T) {
	s := newSandbox(t)
	addition := "    - id: helper\n      name: Example helper\n      slack: '@helper'\n  roles:\n    - id: support\n      name: Support\n      holders:\n        - owner\n"
	writeLiteral(t, s.plan, append(literal(t, "plan.yaml"), []byte(addition)...))
	args := append(s.args()[:9], "--help-contact", "helper", "--preaccess-help-contact", "owner", "--help-role", "support", "--help-contact", "helper")
	before := tree(t, s.root)
	readReport(t, run(newCommand(), args), "invite.written", 0, nil, nil)
	preserved(t, s, before, true)
	payloads := exportedPayloads(t, s.out)
	want := "## Who to ask\n\nOnboarding owner: Example owner, owner@example\\.invalid\n\n- Example helper, @helper\n- Example owner, owner@example\\.invalid\n- Example owner, owner@example\\.invalid\n\n## About this plan"
	if !strings.Contains(string(payloads["first-steps.md"]), want) {
		t.Fatal("explicit mixed contact/role/route selectors lost held order or deduped by rendered text")
	}
}
