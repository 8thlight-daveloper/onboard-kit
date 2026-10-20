package model

type Variable struct {
	Name        string
	Description string
	Example     string
	Type        VariableType
	Options     []string
	Pattern     string
	PatternHint string
	Required    bool
	Multiple    bool
	Default     Value
	Audience    Audience
	AskIf       *Condition
}
