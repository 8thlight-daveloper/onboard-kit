package materializer

import "onboard-kit/internal/model"

type Materializer interface {
	Create(request Request) (Result, []Diagnostic)
}

type Request struct {
	PlanID        string
	Provenance    model.Provenance
	Directory     model.Directory
	PlanVariables []model.Variable
	PlanBindings  []Binding
	Instances     []Instance
}

type Instance struct {
	SourceID           string
	PermanentSectionID string
	Bindings           []Binding
}

type Binding struct {
	Name  string
	Value model.Value
}

type Result struct {
	Plan     model.Plan
	Guidance []Guidance
}

type Guidance struct {
	ID          string
	InstanceIDs []string
	Markdown    []byte
}

type Diagnostic struct {
	Stage   string
	Path    string
	Problem model.Problem
}
