package reviewer_test

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	contract "onboard-kit/contracts/session/reviewer"
)

func padded(raw string, size int) []byte {
	return []byte("#" + strings.Repeat("x", size-len(raw)-2) + "\n" + raw)
}

func resourceError(t *testing.T, diagnostics []contract.Diagnostic, stage, path string) {
	t.Helper()
	requireOwnError(t, diagnostics, stage, "session.resource", "Use a smaller onboarding session or fewer repeated values.", path)
}

func TestRawBodyBoundsBeforeAnyDecoderAndInclusiveAggregate(t *testing.T) {
	input := source(accessPlan)
	input.Plan, input.Preparation, input.Record = padded(accessPlan, 1048576), padded(preparation, 1048576), padded(recordBefore, 1048576)
	session, diagnostics := newReviewer().Inspect(input)
	success(t, diagnostics)
	if session.Account != "newcomer" || len(session.Activities) != 1 || session.Activities[0].CurrentHash != accessHash {
		t.Fatal("three inclusive1MiB sources were not actually reviewed")
	}
	for _, role := range []string{"plan", "preparation", "record"} {
		input := source(accessPlan)
		oversized := []byte(strings.Repeat("x", 1048577))
		switch role {
		case "plan":
			input.Plan = oversized
		case "preparation":
			input.Preparation = oversized
		case "record":
			input.Record = oversized
		}
		got, diagnostics := newReviewer().Inspect(input)
		if !reflect.DeepEqual(got, contract.Session{}) {
			t.Fatal("oversized source returned partial session")
		}
		resourceError(t, diagnostics, "source.admission", role)
	}
}

func TestIdentityByteBoundsBeforeContextAndDecode(t *testing.T) {
	for _, role := range []string{"account", "planId"} {
		input := source(accessPlan)
		value := strings.Repeat("x", 4096)
		if role == "account" {
			input.Account = value
			input.Record = []byte(strings.Replace(recordBefore, "account: newcomer", "account: "+value, 1))
			input.Preparation = []byte(strings.Replace(preparation, "account: newcomer", "account: "+value, 1))
		} else {
			input.PlanID = value
			input.Plan = []byte(strings.Replace(accessPlan, "id: plan-one", "id: "+value, 1))
			input.Record = []byte(strings.Replace(recordBefore, "id: plan-one", "id: "+value, 1))
			input.Preparation = []byte(strings.Replace(preparation, "planId: plan-one", "planId: "+value, 1))
		}
		session, diagnostics := newReviewer().Inspect(input)
		success(t, diagnostics)
		if role == "account" && session.Account != value || role == "planId" && session.PlanID != value {
			t.Fatal("inclusive identity was not used exactly")
		}
		if role == "account" {
			input.Account += "x"
		} else {
			input.PlanID += "x"
		}
		input.Plan = []byte("invalid source not reached")
		got, diagnostics := newReviewer().Inspect(input)
		if !reflect.DeepEqual(got, contract.Session{}) {
			t.Fatal("oversized identity returned session")
		}
		resourceError(t, diagnostics, "source.admission", role)
	}
}

func TestIntentAndEventAdmissionPrecedeSourceDecode(t *testing.T) {
	for _, mutate := range []func(*contract.Intent){
		func(intent *contract.Intent) { intent.Operation = contract.Operation(string([]byte{0xff})) },
		func(intent *contract.Intent) { intent.Question = string([]byte{0xff}) },
		func(intent *contract.Intent) { intent.Section = strings.Repeat("s", 4097) },
		func(intent *contract.Intent) { intent.Kind = "verified\xff" },
		func(intent *contract.Intent) { intent.Note = strings.Repeat("n", 65537) },
		func(intent *contract.Intent) { intent.Differently = string([]byte{0xff}) },
		func(intent *contract.Intent) { intent.Value.Text = strings.Repeat("x", 65537) },
		func(intent *contract.Intent) { intent.Value.List = make([]string, 4097) },
		func(intent *contract.Intent) { intent.Value.List = []string{string([]byte{0xff})} },
	} {
		input := source(accessPlan)
		input.Plan = []byte("not reached")
		intent := contract.Intent{Operation: contract.OperationOutcome, Section: "access", Kind: "failed"}
		mutate(&intent)
		got, diagnostics := newReviewer().Propose(input, intent)
		if !reflect.DeepEqual(got, contract.Proposal{}) {
			t.Fatal("intent admission returned proposal")
		}
		resourceError(t, diagnostics, "intent.admission", "intent")
	}
	proposal := contract.Proposal{Source: source(accessPlan), Intent: contract.Intent{Operation: contract.OperationOutcome, Section: "access", Kind: "failed"}}
	proposal.Source.Plan = []byte("not reached")
	got, diagnostics := newReviewer().Finalize(proposal, "\xff")
	if !reflect.DeepEqual(got, contract.Publication{}) {
		t.Fatal("invalid event resource returned publication")
	}
	resourceError(t, diagnostics, "event.admission", "event")
}

func TestInclusiveIntentTotalReachesLaterIntentRefusal(t *testing.T) {
	intent := contract.Intent{
		Operation: contract.OperationAnswer, Question: "personal.choice",
		Note: strings.Repeat("n", 65536), Differently: strings.Repeat("d", 65536),
		Value: text(strings.Repeat("x", 65536)),
	}
	heldBytes := len(intent.Operation) + len(intent.Question) + len(intent.Note) + len(intent.Differently) + len(intent.Value.Text)
	intent.Value.List = []string{strings.Repeat("l", 262144-heldBytes)}
	got, diagnostics := newReviewer().Propose(source(answerPlan), intent)
	if !reflect.DeepEqual(got, contract.Proposal{}) {
		t.Fatal("irrelevant fields returned proposal")
	}
	requireOwnError(t, diagnostics, "intent", "session.intent", "Choose exactly one session change.", "")
	intent.Value.List[0] += "x"
	got, diagnostics = newReviewer().Propose(source(answerPlan), intent)
	if !reflect.DeepEqual(got, contract.Proposal{}) {
		t.Fatal("oversized aggregate returned proposal")
	}
	resourceError(t, diagnostics, "intent.admission", "intent")
}

func sectionPlan(count int) string {
	var body strings.Builder
	body.WriteString("schemaVersion: 1\nid: plan-one\nsections:\n")
	for i := 0; i < count; i++ {
		body.WriteString("  - id: part-" + strconv.Itoa(i) + "\n    title: Part\n    summary: Read this part\n")
	}
	return body.String()
}

func TestSectionCapAndUnknownDepthBeforeExpandingCores(t *testing.T) {
	session, diagnostics := newReviewer().Inspect(source(sectionPlan(256)))
	success(t, diagnostics)
	if session.Account != "newcomer" || len(session.Activities) != 256 {
		t.Fatal("inclusive section count failed actual review")
	}
	got, diagnostics := newReviewer().Inspect(source(sectionPlan(257)))
	if !reflect.DeepEqual(got, contract.Session{}) {
		t.Fatal("over-cap sections returned session")
	}
	resourceError(t, diagnostics, "plan.admission", "plan")
	for _, lists := range []int{125, 126} {
		plan := accessPlan + "future: " + strings.Repeat("[", lists) + "leaf" + strings.Repeat("]", lists) + "\n"
		got, diagnostics := newReviewer().Inspect(source(plan))
		if lists == 125 {
			success(t, diagnostics)
			if got.Account != "newcomer" || len(got.Activities) != 1 {
				t.Fatal("depth128 did not reach actual review")
			}
		} else {
			if !reflect.DeepEqual(got, contract.Session{}) {
				t.Fatal("depth129 returned session")
			}
			resourceError(t, diagnostics, "plan.admission", "plan")
		}
	}
}

func TestChargedFanoutIncludesDisabledDeclarations(t *testing.T) {
	plan := strings.Replace(answerPlan, "choice: org-a", "choice: "+strings.Repeat("x", 4096), 1)
	plan = strings.Replace(plan, "    summary: Use $choice.\n", "    summary: Use $choice.\n    enabled: false\n    body: "+strings.Repeat("$choice", 4096)+"\n", 1)
	got, diagnostics := newReviewer().Inspect(source(plan))
	if !reflect.DeepEqual(got, contract.Session{}) {
		t.Fatal("charged unsafe fanout returned partial session")
	}
	resourceError(t, diagnostics, "plan.admission", "plan")
	positive := strings.Replace(answerPlan, "choice: org-a", "choice: small", 1)
	positive = strings.Replace(positive, "    summary: Use $choice.\n", "    summary: Use $choice.\n    body: $choice$choice\n", 1)
	session, diagnostics := newReviewer().Inspect(source(positive))
	success(t, diagnostics)
	if len(session.Activities) != 1 || session.Activities[0].Section.Body != "smallsmall" {
		t.Fatalf("small actual render positive control %#v", session.Activities)
	}
}

func TestExactChargedBoundaryCountsAllThreeSourcesAndRoots(t *testing.T) {
	const minimalRecord = "schemaVersion: 1\nkind: record\naccount: newcomer\nplan:\n  id: plan-one\n"
	const minimalPreparation = "schemaVersion: 1\nkind: preparation\naccount: newcomer\nplanId: plan-one\n"
	plan := "schemaVersion: 1\nid: plan-one\nanswers:\n  sections:\n    access:\n      choice: " + strings.Repeat("x", 4096) +
		"\nsections:\n  - id: access\n    title: Access\n    summary: Request access\n    body: " + strings.Repeat("$choice", 1500) +
		"\n    variables:\n      - name: choice\n        type: string\n        audience: onboardee\n"
	input := contract.Source{
		Plan: padded(plan, 1048576), Preparation: padded(minimalPreparation, 1048576), Record: padded(minimalRecord, 103578),
		Account: "newcomer", PlanID: "plan-one",
	}
	session, diagnostics := newReviewer().Inspect(input)
	success(t, diagnostics)
	if session.Account != "newcomer" || len(session.Activities) != 1 || len(session.Activities[0].Section.Body) != 6144000 || strings.Trim(session.Activities[0].Section.Body, "x") != "" {
		t.Fatal("exact B did not reach the independently bounded actual render")
	}
	input.Record = padded(minimalRecord, 103579)
	got, diagnostics := newReviewer().Inspect(input)
	if !reflect.DeepEqual(got, contract.Session{}) {
		t.Fatal("Score B+8 returned usable session")
	}
	resourceError(t, diagnostics, "payload.admission", "session")
}
