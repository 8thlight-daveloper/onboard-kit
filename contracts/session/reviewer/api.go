package reviewer

import "onboard-kit/internal/model"

type Reviewer interface {
	Inspect(Source) (Session, []Diagnostic)
	Propose(Source, Intent) (Proposal, []Diagnostic)
	Finalize(Proposal, model.Timestamp) (Publication, []Diagnostic)
}

type Source struct {
	Plan        []byte
	Preparation []byte
	Record      []byte
	Account     string
	PlanID      string
}

type Operation string

const (
	OperationAnswer     Operation = "answer"
	OperationFollowPlan Operation = "follow-plan"
	OperationOutcome    Operation = "outcome"
)

type Intent struct {
	Operation   Operation
	Question    string
	Value       model.Value
	Section     string
	Kind        model.OutcomeKind
	Note        string
	Differently string
}

type Diagnostic struct {
	Stage   string
	Problem model.Problem
}

type Activity struct {
	Section     model.Section
	CurrentHash string
	WaitingOn   []string
	Escalation  []string
}

type HeldOutcome struct {
	Outcome     model.Outcome
	CurrentHash string
	Comparison  string
}

type Support struct {
	Available    bool
	Recipient    model.PreparationRecipient
	Arrangements []model.Arrangement
	Problems     []Diagnostic
}

type Session struct {
	Account              string
	PlanID               string
	Reference            model.PlanReference
	PlanReferenceMatches bool
	Activities           []Activity
	Outcomes             []HeldOutcome
	Questions            []model.Question
	Blockers             []model.Blocker
	Support              Support
}

type Proposal struct {
	Source       Source
	Intent       Intent
	Before       Session
	After        Session
	BeforeRecord model.Record
	AfterRecord  model.Record
	NeedsEvent   bool
}

type Publication struct {
	Record  []byte
	Changed bool
	Event   model.Timestamp
}
