package model

type Section struct {
	ID            string
	Title         string
	Summary       string
	Parent        string
	Bootstrap     bool
	Owner         string
	Timing        Timing
	Enabled       *bool
	Verification  string
	Body          string
	IncludeIf     *Condition
	After         []string
	Requires      []string
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
