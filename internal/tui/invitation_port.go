package tui

import "onboard-kit/internal/model"

type InvitationPort interface {
	People() ([]InvitationPerson, InvitationOutcome)
	Create(request InvitationRequest) InvitationResult
}

type InvitationPerson struct {
	Account           string
	Refusal           InvitationOutcome
	Plan              model.Plan
	Preparation       model.Preparation
	PreparationErrors []model.Problem
	Blockers          []model.Blocker
	OwnerNamed        bool
	PreaccessContacts []string
	Output            string
}

type InvitationHelp struct {
	Kind string
	ID   string
}

type InvitationRequest struct {
	Account   string
	Output    string
	Preaccess string
	Help      []InvitationHelp
}

type InvitationResult struct {
	Outcome    InvitationOutcome
	Path       string
	FreeOutput string
	Problems   []model.Problem
	Blockers   []model.Blocker
}

type InvitationOutcome int

const (
	InvitationDone InvitationOutcome = iota
	InvitationWrittenLeftovers
	InvitationOutputExists
	InvitationOutputPath
	InvitationAccountRefused
	InvitationUnreadable
	InvitationPlanWithoutID
	InvitationPlanInvalid
	InvitationPreparationInvalid
	InvitationBlocked
	InvitationOwnerMissing
	InvitationPreaccessRefused
	InvitationGuideRefused
	InvitationTooLarge
	InvitationNotWritten
	InvitationNotWrittenLeftovers
)

func (outcome InvitationOutcome) written() bool {
	return outcome == InvitationDone || outcome == InvitationWrittenLeftovers
}

func (outcome InvitationOutcome) sentence() string {
	switch outcome {
	case InvitationWrittenLeftovers:
		return "The file was written, but a temporary file could not be removed from its folder. Keep the file, and check the folder."
	case InvitationOutputExists:
		return "A file with that name already exists, and it is never replaced."
	case InvitationOutputPath:
		return "The file cannot be made there. Choose a folder that exists, with no link on the way."
	case InvitationAccountRefused:
		return "The account cannot name an invitation. Use letters, digits, dots, underscores, or hyphens."
	case InvitationUnreadable:
		return "The plan or this preparation cannot be read. Check that both files exist, and that their paths have no .. part and no link on the way."
	case InvitationPlanWithoutID:
		return "The plan has no id, so no invitation can belong to it."
	case InvitationPlanInvalid:
		return "The plan has errors. The Review lists them."
	case InvitationPreparationInvalid:
		return "This preparation has problems. Fix them on the Preparation tab."
	case InvitationBlocked:
		return "The plan has required questions with no answer. Answer or defer them first."
	case InvitationOwnerMissing:
		return "The preparation names no onboarding owner with a name. Choose one on the Preparation tab."
	case InvitationPreaccessRefused:
		return "The pre-access help must be a contact with a name and an email or Slack handle."
	case InvitationGuideRefused:
		return "The first steps guide cannot be built from this help. The guide problems say why."
	case InvitationTooLarge:
		return "The invitation would be larger than 1 MiB, so it cannot be written."
	case InvitationNotWritten:
		return "The invitation could not be written. Nothing was replaced."
	case InvitationNotWrittenLeftovers:
		return "The invitation was not written, and a temporary file could not be removed from its folder."
	}
	return ""
}
