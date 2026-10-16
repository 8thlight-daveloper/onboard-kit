package model

// Variable is a named, typed value that a person fills in once. It belongs to
// the plan or to one section.
type Variable struct {
	Name        string
	Description string
	Type        VariableType
	Options     []string
	Pattern     string
	PatternHint string
	Required    bool
	Default     Value
	Phase       Phase
	AskIf       *Condition
}
