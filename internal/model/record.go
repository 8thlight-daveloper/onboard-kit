package model

type Timestamp string

type OutcomeKind string

const (
	OutcomeVerified OutcomeKind = "verified"
	OutcomeFailed   OutcomeKind = "failed"
	OutcomeDeferred OutcomeKind = "deferred"
)

type Record struct {
	SchemaVersion int
	Account       string
	StartDate     Date
	Plan          PlanReference
	Answers       Answers
	Outcomes      []Outcome
	UnknownFields map[string]any
}

type PlanReference struct {
	ID   string
	Hash string
	URL  string
}

type Outcome struct {
	Section     string
	Kind        OutcomeKind
	On          Timestamp
	SectionHash string
	Note        string
	Differently string
}

type Reconciliation struct {
	NewQuestions     []string
	RemovedQuestions []string
	AnswersToReview  []string
	NewSections      []string
	RemovedSections  []string
	ChangedOutcomes  []string
}
