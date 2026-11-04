package tui

import (
	"strconv"
	"strings"
	"testing"

	"onboard-kit/internal/model"
	"onboard-kit/internal/question/finder"
)

type fakeInvitations struct {
	people   []InvitationPerson
	written  map[string]bool
	requests []InvitationRequest
	outcomes []InvitationOutcome
	listed   int
}

func outputOf(account string) string { return "plans/obk-invitation-" + account + ".zip" }

func (port *fakeInvitations) freeName(path string) string {
	if !port.written[path] {
		return path
	}
	stem := strings.TrimSuffix(path, ".zip")
	for number := 2; ; number++ {
		if candidate := stem + "-" + strconv.Itoa(number) + ".zip"; !port.written[candidate] {
			return candidate
		}
	}
}

func (port *fakeInvitations) People() ([]InvitationPerson, InvitationOutcome) {
	port.listed++
	people := make([]InvitationPerson, 0, len(port.people))
	for _, person := range port.people {
		person.Output = port.freeName(outputOf(person.Account))
		people = append(people, person)
	}
	return people, InvitationDone
}

func (port *fakeInvitations) Create(request InvitationRequest) InvitationResult {
	port.requests = append(port.requests, request)
	if len(port.outcomes) > 0 {
		outcome := port.outcomes[0]
		port.outcomes = port.outcomes[1:]
		return InvitationResult{Outcome: outcome}
	}
	if port.written[request.Output] {
		return InvitationResult{Outcome: InvitationOutputExists, FreeOutput: port.freeName(request.Output)}
	}
	port.written[request.Output] = true
	return InvitationResult{Outcome: InvitationDone, Path: request.Output}
}

func readyInvitee(plan model.Plan) InvitationPerson {
	preparation := examplePreparation(plan)
	preparation.Recipient.OnboardingOwner = "alex-rivera"
	return InvitationPerson{
		Account: preparation.Account, Plan: plan, Preparation: preparation, OwnerNamed: true,
		PreaccessContacts: []string{"alex-rivera", "it-help"},
	}
}

func blockedInvitee(t *testing.T) InvitationPerson {
	t.Helper()
	plan := blankPlan(t)
	blockers := finder.New().Blockers(plan, model.AudienceOnboarder)
	if len(blockers) < 2 {
		t.Fatalf("the blank plan has blockers to show, got %d", len(blockers))
	}
	return InvitationPerson{
		Account: "pat-kim", Plan: plan, Blockers: blockers[:2], OwnerNamed: true,
		Preparation:       model.Preparation{SchemaVersion: 1, Kind: "preparation", Account: "pat-kim", PlanID: plan.ID, Recipient: model.PreparationRecipient{Name: "Pat Kim"}},
		PreparationErrors: []model.Problem{{Severity: model.SeverityError, Message: "The preparation belongs to another account."}},
	}
}

func newFakeInvitations(t *testing.T) *fakeInvitations {
	t.Helper()
	plan := acmePlan(t)
	return &fakeInvitations{people: []InvitationPerson{blockedInvitee(t), readyInvitee(plan)}, written: map[string]bool{}}
}

func startInviting(t *testing.T, port *fakeInvitations) App {
	t.Helper()
	app, _ := startWith(t, acmePlan(t), model.AudienceOnboarder)
	return resize(app.WithPreparation(withSamLee(t)).WithInvitation(port), 120, 30)
}

func onInvitation(t *testing.T, port *fakeInvitations) App {
	t.Helper()
	return press(startInviting(t, port), typed("i"))
}
