package model

type Section struct {
	ID            string
	Title         string
	Summary       string
	Owner         string
	Timing        Timing
	Enabled       *bool
	Verification  string
	Body          string
	IncludeIf     *Condition
	After         []string
	Escalation    []string
	Variables     []Variable
	Steps         []Step
	Source        *Source
	UnknownFields map[string]any
}

type Source struct {
	Template string
	Hash     string
}
