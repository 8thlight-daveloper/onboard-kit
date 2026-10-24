package reviewer

import "onboard-kit/internal/model"

type Reviewer interface {
	Inspect(archive []byte) (Incoming, []Diagnostic)
	Review(incoming Incoming, local Local, choice PlanChoice) (ReviewResult, []Diagnostic)
	Begin(proposal Proposal, event model.Timestamp) (Publication, []Diagnostic)
}

type Diagnostic struct {
	Stage   string
	Display string
	Problem model.Problem
}

type PlanChoice string

const (
	PlanChoiceUndecided   PlanChoice = ""
	PlanChoiceUseIncoming PlanChoice = "use-incoming"
)

type Role string

const (
	RolePlan        Role = "plan"
	RolePreparation Role = "preparation"
	RoleRecord      Role = "record"
)

type Action string

const (
	ActionRetain    Action = "retain"
	ActionInstall   Action = "install"
	ActionReplace   Action = "replace"
	ActionCreate    Action = "create"
	ActionReconcile Action = "reconcile"
	ActionBind      Action = "bind"
)

type Paths struct {
	Record      string
	Preparation string
}

type Incoming struct {
	Archive   []byte
	Files     model.InvitationFiles
	PlanID    string
	Reference model.PlanReference
	Paths     Paths
	Warnings  []Diagnostic
}

type Body struct {
	Present bool
	Bytes   []byte
}

type Local struct {
	Plan               Body
	Preparation        Body
	Record             Body
	KnownRecordMissing bool
}

type Evidence struct {
	Incoming Incoming
	Local    Local
	Choice   PlanChoice
}

type ReviewResult struct {
	PlanChoiceNeeded bool
	Proposal         *Proposal
}

type Proposal struct {
	Evidence Evidence
	Review   Review
}

type Artifact struct {
	Role   Role
	Action Action
	Bytes  []byte
}

type Comparison struct {
	Available bool
	Reason    string
	Problems  []Diagnostic
}

type Review struct {
	Account               string
	PlanID                string
	Paths                 Paths
	Artifacts             []Artifact
	Comparison            Comparison
	Reconciliation        model.Reconciliation
	NeedsFirstOpenedAt    bool
	BindsUnboundRecord    bool
	PreparationProblems   []Diagnostic
	IncomingWarnings      []Diagnostic
	EffectivePlanWarnings []Diagnostic
	SessionProblems       []Diagnostic
	Session               Session
	Support               Support
	InvitationFirstSteps  []byte
}

type Session struct {
	Activities []Activity
	Outcomes   []Outcome
	Questions  []model.Question
	Blockers   []model.Blocker
}

type Activity struct {
	Section     model.Section
	CurrentHash string
	WaitingOn   []string
	Escalation  []string
}

type Outcome struct {
	Outcome     model.Outcome
	CurrentHash string
	Comparison  string
}

type Support struct {
	Available    bool
	Recipient    model.PreparationRecipient
	Arrangements []model.Arrangement
}

type Publication struct {
	OriginalReview Review
	Artifacts      []Artifact
	FirstOpenedAt  model.Timestamp
}
