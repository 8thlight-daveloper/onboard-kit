package walk

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	firststepscontract "onboard-kit/contracts/firststeps/assembler"
	contract "onboard-kit/contracts/preparation/store"
	"onboard-kit/internal/command/invite"
	"onboard-kit/internal/preparation/store"
	"onboard-kit/internal/tui"
)

const (
	invitationPrefix    = "obk-invitation-"
	invitationExtension = ".zip"
	freeNameAttempts    = 1000
)

type invitationFiles struct {
	root         string
	plan         string
	store        contract.Store
	destinations map[string]string
}

func newInvitationFiles(root string, plan string) *invitationFiles {
	return &invitationFiles{root: root, plan: plan, store: store.New(), destinations: map[string]string{}}
}

func (files *invitationFiles) People() ([]tui.InvitationPerson, tui.InvitationOutcome) {
	listed := files.store.List(contract.ListRequest{Root: files.root, Plan: files.plan})
	if listed.Code != contract.CodeListed {
		if listed.Code == contract.CodePlan {
			return nil, tui.InvitationPlanWithoutID
		}
		return nil, tui.InvitationUnreadable
	}
	people := make([]tui.InvitationPerson, 0, len(listed.People))
	for _, person := range listed.People {
		files.destinations[person.Account] = person.Destination
		if person.Problem != "" {
			people = append(people, tui.InvitationPerson{Account: person.Account, Refusal: tui.InvitationUnreadable})
			continue
		}
		people = append(people, files.checked(person.Account))
	}
	return people, tui.InvitationDone
}

func (files *invitationFiles) request(account string) invite.Request {
	return invite.Request{Plan: files.plan, Preparation: filepath.Join(files.root, filepath.FromSlash(files.destinations[account])), Account: account}
}

func (files *invitationFiles) checked(account string) tui.InvitationPerson {
	readiness := invite.Check(files.request(account))
	return tui.InvitationPerson{
		Account:           account,
		Refusal:           invitationOutcomeOf(readiness.Code),
		Plan:              readiness.Plan,
		Preparation:       readiness.Preparation,
		PreparationErrors: readiness.PreparationErrors,
		Blockers:          readiness.Blockers,
		OwnerNamed:        readiness.OwnerNamed,
		PreaccessContacts: readiness.PreaccessContacts,
		Output:            freeName(filepath.Join(filepath.Dir(files.plan), invitationPrefix+account+invitationExtension)),
	}
}

func (files *invitationFiles) Create(request tui.InvitationRequest) tui.InvitationResult {
	exported := files.request(request.Account)
	exported.Output, exported.Preaccess = request.Output, request.Preaccess
	for _, help := range request.Help {
		exported.Help = append(exported.Help, firststepscontract.HelpReference{Kind: help.Kind, ID: help.ID})
	}
	result := invite.Export(exported)
	created := tui.InvitationResult{Outcome: invitationOutcomeOf(result.Code), Path: result.Path, Problems: result.Problems, Blockers: result.Blockers}
	if created.Outcome == tui.InvitationOutputExists {
		created.FreeOutput = freeName(request.Output)
	}
	return created
}

func freeName(path string) string {
	if !taken(path) {
		return path
	}
	stem := strings.TrimSuffix(path, invitationExtension)
	for number := 2; number <= freeNameAttempts; number++ {
		candidate := stem + "-" + strconv.Itoa(number) + invitationExtension
		if !taken(candidate) {
			return candidate
		}
	}
	return ""
}

func taken(path string) bool {
	_, err := os.Lstat(path)
	return !os.IsNotExist(err)
}

func invitationOutcomeOf(code string) tui.InvitationOutcome {
	switch code {
	case "", "written":
		return tui.InvitationDone
	case "cleanup-written":
		return tui.InvitationWrittenLeftovers
	case "output-exists":
		return tui.InvitationOutputExists
	case "output-path":
		return tui.InvitationOutputPath
	case "account":
		return tui.InvitationAccountRefused
	case "plan-read", "preparation-read", "plan-decode", "preparation-decode":
		return tui.InvitationUnreadable
	case "plan-identity":
		return tui.InvitationPlanWithoutID
	case "plan-invalid":
		return tui.InvitationPlanInvalid
	case "preparation-invalid":
		return tui.InvitationPreparationInvalid
	case "onboarder-blocked":
		return tui.InvitationBlocked
	case "owner":
		return tui.InvitationOwnerMissing
	case "preaccess-help":
		return tui.InvitationPreaccessRefused
	case "firststeps":
		return tui.InvitationGuideRefused
	case "archive":
		return tui.InvitationTooLarge
	case "cleanup-refused":
		return tui.InvitationNotWrittenLeftovers
	}
	return tui.InvitationNotWritten
}
