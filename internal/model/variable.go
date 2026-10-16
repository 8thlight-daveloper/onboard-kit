package model

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
