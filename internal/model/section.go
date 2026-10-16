package model

type Section struct {
	ID            string
	Title         string
	Summary       string
	Owner         string
	Timing        Timing
	Verification  string
	Body          string
	IncludeIf     *Condition
	After         []string
	Variables     []Variable
	Steps         []Step
	Source        *Source
	UnknownFields map[string]any
}

type Source struct {
	Template string
	Hash     string
}
