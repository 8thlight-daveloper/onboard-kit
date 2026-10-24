package reviewer_test

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
	"testing"

	contract "onboard-kit/contracts/invitation/reviewer"
	"onboard-kit/internal/model"
)

const resourceMessage = "Use a smaller onboarding plan or fewer repeated values."

func resourceProblem(t *testing.T, got []contract.Diagnostic, stage, path string) {
	t.Helper()
	ownProblem(t, got, stage, "invitation.resource", path, resourceMessage)
}

func storedArchive(t *testing.T, plan, preparation, seed, guidance []byte) []byte {
	t.Helper()
	roles := []struct {
		name string
		body []byte
	}{
		{"plan.obk.yaml", plan},
		{"records/jules/preparation.obk.yaml", preparation},
		{"records/jules/record.obk.yaml", seed},
		{"first-steps.md", guidance},
	}
	var out bytes.Buffer
	w := zip.NewWriter(&out)
	for _, role := range roles {
		entry, err := w.CreateHeader(&zip.FileHeader{Name: role.name, Method: zip.Store})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write(role.body); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	payload := len(plan) + len(preparation) + len(seed) + len(guidance)
	if out.Len() != payload+570 {
		t.Fatalf("Store transport framing differs: raw=%d payload=%d, expected overhead570", out.Len(), payload)
	}
	return out.Bytes()
}

func paddedAdmissionPlan(t *testing.T, dollars, padding int) []byte {
	t.Helper()
	if dollars%2 != 0 {
		t.Fatal("body must contain complete escaped-dollar pairs")
	}
	plan := string(fixture(t, "admission.base.plan.yaml"))
	plan = strings.Replace(plan, `body: ""`, `body: "`+strings.Repeat("$", dollars)+`"`, 1)
	plan = strings.Replace(plan, `pad: ""`, `pad: "`+strings.Repeat("a", padding)+`"`, 1)
	if len(plan) != 101+dollars+padding {
		t.Fatal("independent fixed plan byte count differs")
	}
	return []byte(plan)
}

func admissionArchive(t *testing.T, plan []byte) []byte {
	t.Helper()
	hash := sha256.Sum256(plan)
	seed := bytes.Replace(fixture(t, "admission.seed.template.yaml"), []byte(strings.Repeat("0", 64)), []byte(hex.EncodeToString(hash[:])), 1)
	return storedArchive(t, plan, fixture(t, "admission.preparation.yaml"), seed, fixture(t, "admission.first-steps.md"))
}

func TestAdmissionExactPublicBoundaryAndPayloadRefusal(t *testing.T) {
	plan := paddedAdmissionPlan(t, 440000, 10629)
	raw := admissionArchive(t, plan)
	if len(raw) != 451498 {
		t.Fatal("exact-boundary source count changed")
	}
	r := newReviewer()
	in, problems := r.Inspect(raw)
	if len(problems) != 0 || in.PlanID != "p" {
		t.Fatalf("exact B must admit: %d diagnostics", len(problems))
	}
	result, problems := r.Review(in, contract.Local{}, contract.PlanChoiceUndecided)
	if len(problems) != 0 || result.Proposal == nil {
		t.Fatal("exact B must also admit effective source context")
	}
	p := *result.Proposal
	if len(p.Review.Session.Activities) != 1 || p.Review.Session.Activities[0].Section.Body != strings.Repeat("$", 220000) || len(p.Review.Session.Questions) != 0 || len(p.Review.Session.Blockers) != 0 {
		t.Error("conservative escaped-dollar admission changed actual literal rendering")
	}
	pub, problems := r.Begin(p, beginAt)
	if len(problems) != 0 || pub.FirstOpenedAt != beginAt {
		t.Error("reserved short event must not invalidate admitted boundary review")
	}
	got, problems := r.Inspect(append(copied(raw), 'x'))
	resourceProblem(t, problems, "inspect.payload-admission", "invitation")
	if !reflect.DeepEqual(got, contract.Incoming{}) {
		t.Error("resource refusal returned partial incoming")
	}
}

func TestAdmissionEarlyPlanPrecedesSemanticValidation(t *testing.T) {
	plan := paddedAdmissionPlan(t, 442000, 10629)
	for _, disabled := range []bool{false, true} {
		candidate := plan
		if disabled {
			candidate = bytes.Replace(plan, []byte("    title: Task\n"), []byte("    title: Task\n    enabled: false\n"), 1)
		}
		got, problems := newReviewer().Inspect(admissionArchive(t, candidate))
		resourceProblem(t, problems, "inspect.plan-admission", "plan")
		if !reflect.DeepEqual(got, contract.Incoming{}) {
			t.Error("early resource refusal returned partial incoming")
		}
	}
	bad := bytes.Replace(plan, []byte("    summary: Task\n"), nil, 1)
	_, problems := newReviewer().Inspect(admissionArchive(t, bad))
	resourceProblem(t, problems, "inspect.plan-admission", "plan")
}

func TestAdmissionCountsIdenticalPresentLocalSource(t *testing.T) {
	r := newReviewer()
	plan := paddedAdmissionPlan(t, 440000, 10629)
	in, problems := r.Inspect(admissionArchive(t, plan))
	if len(problems) != 0 {
		t.Fatal("exact boundary incoming refused")
	}
	got, problems := r.Review(in, contract.Local{Plan: contract.Body{Present: true, Bytes: copied(plan)}}, contract.PlanChoiceUndecided)
	resourceProblem(t, problems, "review.effective-admission", "effectivePlan")
	if !reflect.DeepEqual(got, contract.ReviewResult{}) {
		t.Error("effective resource refusal returned a proposal/choice")
	}
}

func TestAdmissionPreviousResourceWarningAndChoiceOrder(t *testing.T) {
	r := newReviewer()
	local := contract.Local{Plan: contract.Body{Present: true, Bytes: paddedAdmissionPlan(t, 442000, 10629)}}
	in := incoming(t, r)
	got, problems := r.Review(in, local, contract.PlanChoiceUndecided)
	if len(problems) != 0 || !got.PlanChoiceNeeded || got.Proposal != nil {
		t.Fatal("explicit incoming choice must precede previous admission")
	}
	got, problems = r.Review(in, local, contract.PlanChoiceUseIncoming)
	if len(problems) != 0 || got.Proposal == nil {
		t.Fatal("inadmissible previous comparison must not block accepted replacement")
	}
	v := got.Proposal.Review
	want := []contract.Diagnostic{{Stage: "previous.plan-admission", Display: "Previous plan comparison is unavailable within the resource limit.", Problem: model.Problem{
		Code: "invitation.previous-resource", Severity: model.SeverityWarning,
		Message: "Previous plan comparison is unavailable within the resource limit.", Detail: map[string]string{"path": "previousPlan"},
	}}}
	if v.Comparison.Available || v.Comparison.Reason != "resource" || !reflect.DeepEqual(v.Comparison.Problems, want) || !bytes.Equal(got.Proposal.Evidence.Local.Plan.Bytes, local.Plan.Bytes) || v.Artifacts[0].Action != contract.ActionReplace {
		t.Error("previous resource result lost warning meaning or guarded preimage")
	}
}

func nestedUnknown(levels int) string {
	return "private: " + strings.Repeat("[", levels) + "null" + strings.Repeat("]", levels) + "\n"
}

func TestAdmissionLocalDepthGuardsBeforeSemanticErrors(t *testing.T) {
	r := newReviewer()
	in := incoming(t, r)
	for _, role := range []string{"preparation", "record"} {
		limit := 124
		name := "minimal.preparation.yaml"
		stage := "review.preparation-admission"
		if role == "record" {
			limit, name, stage = 125, "minimal.seed.yaml", "review.record-admission"
		}
		for _, extra := range []int{0, 1} {
			raw := string(fixture(t, name)) + nestedUnknown(limit+extra)
			if extra == 1 {
				if role == "record" {
					raw = strings.Replace(raw, "account: jules", "account: other", 1)
				} else {
					raw += "recipient:\n  workLocation: invalid\n"
				}
			}
			local := contract.Local{}
			if role == "record" {
				local.Record = contract.Body{Present: true, Bytes: []byte(raw)}
			} else {
				local.Preparation = contract.Body{Present: true, Bytes: []byte(raw)}
			}
			got, problems := r.Review(in, local, contract.PlanChoiceUseIncoming)
			if extra == 0 {
				if len(problems) != 0 || got.Proposal == nil {
					t.Errorf("%s depth128 control refused", role)
				}
				continue
			}
			resourceProblem(t, problems, stage, "local."+role)
			if got.Proposal != nil {
				t.Error("depth129 returned proposal")
			}
		}
	}
}

func nullForest(items int) string {
	return "private: [" + strings.Repeat("null,", items-1) + "null]\n"
}

func TestAdmissionSharedUnknownCounterAcrossRoots(t *testing.T) {
	r := newReviewer()
	in := incoming(t, r)
	prep := string(fixture(t, "minimal.preparation.yaml")) + nullForest(32764)
	for _, extra := range []int{0, 1} {
		record := string(fixture(t, "minimal.seed.yaml")) + nullForest(32764+extra)
		got, problems := r.Review(in, contract.Local{
			Preparation: contract.Body{Present: true, Bytes: []byte(prep)},
			Record:      contract.Body{Present: true, Bytes: []byte(record)},
		}, contract.PlanChoiceUseIncoming)
		if extra == 0 {
			if len(problems) != 0 || got.Proposal == nil {
				t.Fatal("exact65536 shared unknown occurrences must admit")
			}
			continue
		}
		resourceProblem(t, problems, "review.record-admission", "local.record")
		if got.Proposal != nil {
			t.Error("one extra unknown node must refuse the combined forest")
		}
	}
}

func contactFanoutPlan(contacts int) []byte {
	var p strings.Builder
	p.WriteString("schemaVersion: 1\nid: plan-a\nsections:\n  - id: setup\n    title: Set up\n    summary: \"Use")
	for i := range 64 {
		fmt.Fprintf(&p, " $contact_%03d", i)
	}
	p.WriteString(".\"\n    variables:\n")
	for i := range 64 {
		fmt.Fprintf(&p, "      - name: contact_%03d\n        description: Select a contact\n        type: contact\n        audience: onboardee\n", i)
	}
	p.WriteString("directory:\n  contacts:\n")
	for i := range contacts {
		fmt.Fprintf(&p, "    - id: c%04d\n      name: %s\n      email: c%04d@company.example\n", i, strings.Repeat("a", 256), i)
	}
	return []byte(p.String())
}

func TestAdmissionDirectoryChoiceFanoutWithoutLargeRendering(t *testing.T) {
	r := newReviewer()
	plan := contactFanoutPlan(512)
	hash := sha256.Sum256(plan)
	seed := bytes.Replace(fixture(t, "minimal.seed.yaml"), []byte(planHash), []byte(hex.EncodeToString(hash[:])), 1)
	raw := archive(t, map[string][]byte{"plan.obk.yaml": plan, "records/jules/record.obk.yaml": seed})
	_, problems := r.Inspect(raw)
	resourceProblem(t, problems, "inspect.plan-admission", "plan")
	in := inspectPlan(t, r, contactFanoutPlan(200))
	got, problems := r.Review(in, contract.Local{}, contract.PlanChoiceUndecided)
	if len(problems) != 0 || got.Proposal == nil || len(got.Proposal.Review.Session.Questions) != 64 {
		t.Fatal("smaller concrete directory choice control must remain usable")
	}
	for _, q := range got.Proposal.Review.Session.Questions {
		if len(q.Choices) != 200 || q.Choices[0].Value != "c0000" || q.Choices[199].Value != "c0199" {
			t.Error("admission must not truncate real directory choices")
		}
	}
}
