package editor

import "onboard-kit/internal/model"

type Editor interface {
	Edit(plan model.Plan, operations []Operation) (model.Plan, []Diagnostic)
}

type Operation struct {
	Kind      Kind
	SectionID string
	SourceID  string
	Bindings  []Binding
	Name      string
	Value     model.Value
}

type Kind string

const (
	Add        Kind = "add"
	Submit     Kind = "submit"
	Unset      Kind = "unset"
	Retire     Kind = "retire"
	Reactivate Kind = "reactivate"
)

type Binding struct {
	Name  string
	Value model.Value
}

type Diagnostic struct {
	Stage    string
	Path     string
	Code     string
	Severity model.Severity
	Message  string
}
