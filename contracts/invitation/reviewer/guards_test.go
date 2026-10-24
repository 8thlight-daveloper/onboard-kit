package reviewer_test

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	contract "onboard-kit/contracts/invitation/reviewer"
	"onboard-kit/internal/model"
)

func TestStrictSeedProfile(t *testing.T) {
	seed := string(fixture(t, "minimal.seed.yaml"))
	cases := []struct{ name, before, after, suffix, path string }{
		{"account", "account: jules", "account: other", "", "record.account"},
		{"plan", "id: plan-a", "id: other", "", "record.plan.id"},
		{"hash", planHash, "OLD", "", "record.plan.hash"},
		{"url", "", "", "  url: https://example.invalid\n", "record.plan.url"},
		{"start", "", "", "startDate: Monday\n", "record.startDate"},
		{"opened", "", "", "firstOpenedAt: already\n", "record.firstOpenedAt"},
		{"false-answer", "", "", "answers:\n  plan:\n    hidden: false\n", "record.answers.plan"},
		{"deferred-hidden-answer", "", "", "answers:\n  sections:\n    hidden:\n      old: deferred\n", "record.answers.sections"},
		{"outcome", "", "", "outcomes:\n  - section: setup\n    outcome: verified\n    on: yesterday\n", "record.outcomes"},
		{"unknown", "", "", "private-note: held\n", "record.unknownFields"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			raw := seed
			if c.before != "" {
				raw = strings.Replace(raw, c.before, c.after, 1)
			}
			raw += c.suffix
			got, problems := newReviewer().Inspect(archive(t, map[string][]byte{"records/jules/record.obk.yaml": []byte(raw)}))
			ownProblem(t, problems, "inspect.seed", "invitation.seed", c.path, "Use an identity-only seed matching the received invitation.")
			if !reflect.DeepEqual(got, contract.Incoming{}) {
				t.Error("fatal seed inspection leaked a partial result")
			}
		})
	}
	for _, suffix := range []string{"answers: {}\noutcomes: []\n", "answers:\n  sections:\n    hidden: {}\n"} {
		_, problems := newReviewer().Inspect(archive(t, map[string][]byte{"records/jules/record.obk.yaml": []byte(seed + suffix)}))
		if len(problems) != 0 {
			t.Error("empty collections do not carry answer entries")
		}
	}
}

func TestFirstStepsAreOpaqueNonemptyUTF8(t *testing.T) {
	for _, b := range [][]byte{nil, {0xff}} {
		_, problems := newReviewer().Inspect(archive(t, map[string][]byte{"first-steps.md": b}))
		ownProblem(t, problems, "inspect.first-steps", "invitation.first-steps", "firstSteps", "Supply nonempty UTF-8 invitation guidance.")
	}
	want := []byte(" \r\n\x00\x1b\u202e")
	got, problems := newReviewer().Inspect(archive(t, map[string][]byte{"first-steps.md": want}))
	if len(problems) != 0 || !bytes.Equal(got.Files.FirstSteps, want) {
		t.Error("accepted guidance must remain exact document data")
	}
}

func TestRecordIdentityCollectsBothDefects(t *testing.T) {
	raw := strings.Replace(string(fixture(t, "minimal.seed.yaml")), "account: jules", "account: other", 1)
	raw = strings.Replace(raw, "id: plan-a", "id: other", 1)
	r := newReviewer()
	result, problems := r.Review(incoming(t, r), contract.Local{Record: contract.Body{Present: true, Bytes: []byte(raw)}}, contract.PlanChoiceUseIncoming)
	if !reflect.DeepEqual(result, contract.ReviewResult{}) || len(problems) != 2 {
		t.Fatalf("expected zero result and two identity diagnostics, got %d", len(problems))
	}
	ownProblem(t, problems[:1], "review.record-identity", "invitation.record-account", "record.account", "Use the record for this invitation account.")
	ownProblem(t, problems[1:], "review.record-identity", "invitation.record-plan", "record.plan.id", "Use this plan's record or explicitly bind an unbound record.")
}

func TestHistoricalUTCAndNewEventProfilesDiffer(t *testing.T) {
	valid := []string{
		"0001-01-01T00:00:00Z", "9999-12-31T23:59:59+00:00",
		"2024-02-29T12:34:56.12345678901234567890Z",
	}
	invalid := []string{
		"0000-01-01T00:00:00Z", "2026-02-29T00:00:00Z", "2026-01-01t00:00:00Z",
		"2026-01-01T24:00:00Z", "2026-01-01T00:60:00Z", "2026-01-01T00:00:60Z",
		"2026-01-01T00:00:00-00:00", "2026-01-01T00:00:00+01:00",
		"2026-01-01T00:00:00.Z", "2026-01-01T00:00:00,1Z", "2026-01-01 00:00:00Z",
		"2026-01-01T00:00:00Z ", "2026-01-01T00:00:00z",
	}
	r := newReviewer()
	for _, stamp := range append(append([]string{}, valid...), invalid...) {
		t.Run(stamp, func(t *testing.T) {
			raw := string(fixture(t, "minimal.seed.yaml")) + "firstOpenedAt: \"" + stamp + "\"\n"
			got, problems := r.Review(incoming(t, r), contract.Local{Record: contract.Body{Present: true, Bytes: []byte(raw)}}, contract.PlanChoiceUseIncoming)
			isValid := false
			for _, value := range valid {
				isValid = isValid || stamp == value
			}
			if !isValid {
				ownProblem(t, problems, "review.record-history", "invitation.history-time", "record.firstOpenedAt", "Preserve a real UTC time for interpreted history.")
				return
			}
			if len(problems) != 0 || got.Proposal == nil {
				t.Fatal("supported historical UTC text refused")
			}
			pub, problems := r.Begin(*got.Proposal, "")
			if len(problems) != 0 || string(pub.FirstOpenedAt) != stamp || !bytes.Equal(pub.Artifacts[2].Bytes, []byte(raw)) {
				t.Error("historical UTC text must not be normalized")
			}
		})
	}
	p := proposal(t, r, contract.Local{})
	for _, stamp := range []model.Timestamp{"", "2026-01-01T00:00:00+00:00", "2026-01-01T00:00:00.1234567890Z", "2026-02-29T00:00:00Z"} {
		pub, problems := r.Begin(p, stamp)
		ownProblem(t, problems, "begin.event", "invitation.event", "event", "Supply one supported begin event only for a new record.")
		if !reflect.DeepEqual(pub, contract.Publication{}) {
			t.Error("invalid event returned publication")
		}
	}
}

func TestSourceBoundsBeforeLaterGuards(t *testing.T) {
	r := newReviewer()
	for _, role := range []string{"plan", "preparation", "record"} {
		for _, n := range []int{1048576, 1048577} {
			local := contract.Local{}
			body := contract.Body{Present: true, Bytes: bytes.Repeat([]byte{' '}, n)}
			switch role {
			case "plan":
				local.Plan = body
			case "preparation":
				local.Preparation = body
			case "record":
				local.Record = body
			}
			_, problems := r.Review(incoming(t, r), local, contract.PlanChoice("other"))
			if n == 1048576 {
				ownProblem(t, problems, "review.choice", "invitation.plan-choice", "choice", "Choose undecided or use-incoming explicitly.")
			} else {
				ownProblem(t, problems, "review.local", "invitation.source-size", "local."+role, "Keep each local source within the supported byte limit.")
			}
		}
	}
}
